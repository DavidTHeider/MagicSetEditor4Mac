package rest_test

// **Der Zugang über ein Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351).

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// profilschluessel ist der Träger, an dem dieser Lauf das Profil wiederfindet.
//
// **Ein eigener und nicht der aus `internal/storage`**: dieses Paket kennt
// `storage` nicht (§2), und was hier geprüft wird, ist, **dass** die Wache den
// Kontext austauscht — nicht, wie `storage` ihn beschriftet.
type profilschluessel struct{}

// tokenwache ist eine Wache mit **einem** bekannten Token.
type tokenwache struct {
	gilt      string
	profilID  int64
	nurLesen  bool
	gefragt   int
	schreibIt bool
}

func (w *tokenwache) Pruefe(ctx context.Context, klartext string, schreibend bool) (
	context.Context, string, error,
) {
	w.gefragt++
	w.schreibIt = schreibend

	if klartext != w.gilt {
		return ctx, "", fmt.Errorf("%w: dieses Token gilt nicht", domain.ErrValidation)
	}

	if schreibend && w.nurLesen {
		return ctx, "", fmt.Errorf(
			"%w: dieses Token darf lesen, verlangt ist schreiben",
			domain.ErrErlaubnis)
	}

	// **Die Maschinenidentität** — Profil und Principal, wie die echte Wache
	// sie bildet (Befund 5 des Zweitblicks über Phase 5.2).
	return context.WithValue(ctx, profilschluessel{}, w.profilID),
		fmt.Sprintf("%d:Probemaschine", w.profilID), nil
}

// anfrageMit baut eine Anfrage, die die Herkunftsprüfung besteht.
//
// **`Host` gehört dazu**: `httptest.NewRequest` setzt `example.com`, und die
// erste Wache weist das ab — zu Recht, denn genau so sähe ein
// DNS-Rebinding aus (`herkunft.go`).
func anfrageMit(t *testing.T, methode, kopfzeile string) *http.Request {
	t.Helper()

	anfrage := httptest.NewRequestWithContext(
		t.Context(), methode, "/sets", http.NoBody)
	anfrage.Host = "127.0.0.1:7777"
	anfrage.Header.Set("Authorization", kopfzeile)

	return anfrage
}

// mitWache baut einen Dienst, dessen Weg bis zum mitzählenden Doppel geht.
//
// **Eine `nil`-Wache ist ein Fall und kein Versehen**: so läuft der Dienst
// ohne technische Profile, und der Rückfall auf die Prozessmarke wird geprüft.
func mitWache(t *testing.T, wache rest.Tokenwache) (http.Handler, *zaehlender) {
	t.Helper()

	zeuge := &zaehlender{}
	dienst := dienstMit(t, zeuge)

	if wache != nil {
		dienst = dienst.MitTokenwache(wache)
	}

	return dienst.Behandler(), zeuge
}

// profilAus zieht das Profil aus dem Kontext, den das Doppel gesehen hat.
func profilAus(zeuge *zaehlender) any {
	if zeuge.kontext == nil {
		return nil
	}

	return zeuge.kontext.Value(profilschluessel{})
}

func TestEinGueltigesTokenKommtDurchUndBringtSeinProfilMit(t *testing.T) {
	t.Parallel()

	wache := &tokenwache{gilt: "mse_3_ABCDEFGH", profilID: 42}
	weg, zeuge := mitWache(t, wache)

	antwort := httptest.NewRecorder()
	anfrage := anfrageMit(t, http.MethodGet, "Bearer mse_3_ABCDEFGH")

	weg.ServeHTTP(antwort, anfrage)

	if zeuge.rufe != 1 {
		t.Fatalf("der Weg dahinter wurde %dmal gerufen, Status %d",
			zeuge.rufe, antwort.Code)
	}

	if profilAus(zeuge) != int64(42) {
		t.Errorf("das Profil dahinter ist %v, erwartet 42", profilAus(zeuge))
	}

	if wache.schreibIt {
		t.Error("ein GET hat eine Schreiberlaubnis verlangt")
	}
}

func TestEinSchreibenderWegVerlangtSchreiben(t *testing.T) {
	t.Parallel()

	faelle := map[string]bool{
		http.MethodGet:    false,
		http.MethodHead:   false,
		http.MethodPost:   true,
		http.MethodPut:    true,
		http.MethodDelete: true,
		http.MethodPatch:  true,
	}

	for methode, schreibend := range faelle {
		t.Run(methode, func(t *testing.T) {
			t.Parallel()

			wache := &tokenwache{gilt: "mse_3_ABCDEFGH", profilID: 42}
			weg, _ := mitWache(t, wache)

			anfrage := anfrageMit(t, methode, "Bearer mse_3_ABCDEFGH")

			weg.ServeHTTP(httptest.NewRecorder(), anfrage)

			if wache.gefragt != 1 {
				t.Fatalf("die Wache wurde %dmal gefragt", wache.gefragt)
			}

			if wache.schreibIt != schreibend {
				t.Errorf("schreibend = %v, erwartet %v",
					wache.schreibIt, schreibend)
			}
		})
	}
}

func TestEinLesetokenDarfNichtSchreiben(t *testing.T) {
	t.Parallel()

	wache := &tokenwache{gilt: "mse_3_ABCDEFGH", profilID: 42, nurLesen: true}
	weg, zeuge := mitWache(t, wache)

	antwort := httptest.NewRecorder()
	anfrage := anfrageMit(t, http.MethodPost, "Bearer mse_3_ABCDEFGH")

	weg.ServeHTTP(antwort, anfrage)

	if zeuge.rufe != 0 {
		t.Fatal("der Weg dahinter wurde trotzdem gerufen")
	}

	if antwort.Code != http.StatusForbidden {
		t.Errorf("Status = %d, erwartet 403", antwort.Code)
	}

	t.Run("und die Meldung sagt, woran es lag", func(t *testing.T) {
		// **Die eine Ablehnung, die mehr sagt als „nein"** — der Grund steht
		// an [domain.ErrErlaubnis].
		if !strings.Contains(antwort.Body.String(), "lesen") {
			t.Errorf("Rumpf = %q", antwort.Body.String())
		}
	})

	t.Run("und lesen darf es weiterhin", func(t *testing.T) {
		// **Ohne diesen Fall stünde der Lauf oben auch dann grün da, wenn die
		// Wache jedes Token abwiese** (§10).
		zweite := anfrageMit(t, http.MethodGet, "Bearer mse_3_ABCDEFGH")

		antwort := httptest.NewRecorder()
		weg.ServeHTTP(antwort, zweite)

		if antwort.Code == http.StatusForbidden {
			t.Errorf("auch das Lesen wurde mit 403 abgewiesen")
		}
	})
}

func TestEinUngueltigesTokenKlingtWieEineFalscheMarke(t *testing.T) {
	t.Parallel()

	// **Damit aus der Antwort nicht hervorgeht, welcher Nachweis beinahe
	// gestimmt hätte.** Wer eine Kennung durchprobiert, soll nicht erfahren,
	// dass es sie gibt.
	wache := &tokenwache{gilt: "mse_3_ABCDEFGH", profilID: 42}

	// **`#nosec G101`**: gosec sieht `Bearer` und hält es für ein eingebautes
	// Geheimnis. Beide Werte sind ausgedachte **falsche** Nachweise — genau
	// das ist ihr Zweck.
	faelle := map[string]string{ // #nosec G101 -- ausgedachte falsche Nachweise
		"ein fremdes Token":  "Bearer mse_9_ZZZZZZZZ",
		"eine falsche Marke": "Bearer " + probemarke[:8],
	}

	rueckmeldungen := map[string]string{}

	for name, kopfzeile := range faelle {
		weg, zeuge := mitWache(t, wache)

		antwort := httptest.NewRecorder()
		anfrage := anfrageMit(t, http.MethodGet, kopfzeile)

		weg.ServeHTTP(antwort, anfrage)

		if zeuge.rufe != 0 {
			t.Fatalf("%s kam durch", name)
		}

		if antwort.Code != http.StatusUnauthorized {
			t.Errorf("%s: Status = %d, erwartet 401", name, antwort.Code)
		}

		rueckmeldungen[name] = antwort.Body.String()
	}

	if rueckmeldungen["ein fremdes Token"] != rueckmeldungen["eine falsche Marke"] {
		t.Errorf("die beiden Antworten unterscheiden sich:\n%q\n%q",
			rueckmeldungen["ein fremdes Token"],
			rueckmeldungen["eine falsche Marke"])
	}
}

func TestOhneWacheGiltNurDieMarke(t *testing.T) {
	t.Parallel()

	// **Der Rückfall, den es geben muss**: ein Bestand ohne technische Profile
	// setzt keine Wache, und dann ist ein `mse_…` nichts weiter als eine
	// falsche Marke.
	weg, zeuge := mitWache(t, nil)

	antwort := httptest.NewRecorder()
	anfrage := anfrageMit(t, http.MethodGet, "Bearer mse_3_ABCDEFGH")

	weg.ServeHTTP(antwort, anfrage)

	if antwort.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, erwartet 401", antwort.Code)
	}

	t.Run("und die Marke selbst weiterhin", func(t *testing.T) {
		mit := anfrageMit(t, http.MethodGet, "Bearer "+probemarke)

		weg.ServeHTTP(httptest.NewRecorder(), mit)

		if zeuge.rufe != 1 {
			t.Error("auch die richtige Marke kam nicht durch")
		}
	})
}

// **Eine Rotation gibt kein frisches Guthaben.**
//
// Befund 5 des Zweitblicks über Phase 5.2: der Dienst budgetierte nach dem
// **Hash des vorgezeigten Klartexts**. Wer sein Token rotierte, bekam damit
// einen neuen Principal und ein volles Guthaben — die Drossel ließ sich
// umgehen, indem man rotierte. Gezählt wird jetzt nach der Maschine.
func TestEineRotationGibtKeinFrischesGuthaben(t *testing.T) {
	t.Parallel()

	wache := &tokenwache{gilt: "mse_1_alt", profilID: 7}
	behandler, _ := mitWache(t, wache)

	erlaubt := rest.ErlaubteRufe(domain.Vorgabegrenzen(), http.MethodGet, "/sets")

	for versuch := 1; versuch <= erlaubt; versuch++ {
		if antwort := rufeBehandler(t, behandler, http.MethodGet, "/sets",
			"mse_1_alt"); antwort.Code >= http.StatusBadRequest {
			t.Fatalf("Aufruf %d endete mit %d", versuch, antwort.Code)
		}
	}

	// **Das Guthaben ist leer.**
	if antwort := rufeBehandler(t, behandler, http.MethodGet, "/sets",
		"mse_1_alt"); antwort.Code != http.StatusTooManyRequests {
		t.Fatalf("der Ruf danach endete mit %d", antwort.Code)
	}

	// **Jetzt rotiert das Token**: anderes Geheimnis, dieselbe Maschine.
	wache.gilt = "mse_1_neu"

	if antwort := rufeBehandler(t, behandler, http.MethodGet, "/sets",
		"mse_1_neu"); antwort.Code != http.StatusTooManyRequests {
		t.Errorf("nach der Rotation kam %d — das Guthaben war wieder voll",
			antwort.Code)
	}
}

// rufeBehandler fragt einen Behandler mit diesem Token.
func rufeBehandler(
	t *testing.T, behandler http.Handler, methode, weg, klartext string,
) *httptest.ResponseRecorder {
	t.Helper()

	anfrage := httptest.NewRequestWithContext(t.Context(), methode, weg, http.NoBody)
	anfrage.Host = "127.0.0.1:7777"
	anfrage.Header.Set("Authorization", "Bearer "+klartext)

	antwort := httptest.NewRecorder()
	behandler.ServeHTTP(antwort, anfrage)

	return antwort
}
