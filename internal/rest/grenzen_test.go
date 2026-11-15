package rest_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Die Grenzen dieser Fläche** — Werkstück O5 von Phase 5.2 (#1429).

func TestDieGrenzenSindMaschinenlesbar(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, neuerDienst(t), endpunkt{methode: http.MethodGet, weg: "/grenzen"}, gueltig)
	if antwort.Code != http.StatusOK {
		t.Fatalf("%d: %s", antwort.Code, antwort.Body.String())
	}

	var grenzen rest.Grenzen
	if err := json.Unmarshal(antwort.Body.Bytes(), &grenzen); err != nil {
		t.Fatalf("keine Grenzen: %v", err)
	}

	vorgabe := domain.Vorgabegrenzen()
	if grenzen.Parallel["rendern"] != vorgabe.Parallel[domain.RessourceRendern] ||
		grenzen.Kosten["lesen"] != 1 || grenzen.Kosten["masse"] != domain.KostenMasse.Kosten() ||
		grenzen.Principal.Burst != vorgabe.Principal.Burst || grenzen.Principal.FensterSekunden != 60 ||
		grenzen.Global.Burst != vorgabe.Global.Burst || grenzen.Reserve != vorgabe.Reserve ||
		grenzen.AntwortBytes != vorgabe.Antwort || grenzen.Kontingente != domain.KontingenteAn {
		t.Errorf("Grenzen: %+v", grenzen)
	}
}

// **Auftrags- und Verbindungsgrenze und die Kostentafel** — #2003: dieselbe
// Auskunft wie an den anderen Flächen.
func TestDieGrenzenNennenAuftraegeVerbindungenUndKostentafel(t *testing.T) {
	t.Parallel()

	grenzen := rest.GrenzenVon(domain.Vorgabegrenzen())

	if grenzen.Verbindungen != 16 ||
		grenzen.Auftraege != (rest.Auftragsgrenzen{Laufend: 4, Behalten: 16, AufbewahrungSekunden: 3600}) ||
		len(grenzen.Kostentafel) != len(domain.Kostentafel()) {
		t.Errorf("Grenzen: %+v", grenzen)
	}
}

// **Abgeschaltete Kontingente sagt die Auskunft** — QA-08.AK-03 (#1739): wer
// fragt, soll nicht aus dem Ausbleiben eines 429 schließen müssen.
func TestDieGrenzenNennenAbgeschalteteKontingente(t *testing.T) {
	t.Parallel()

	ohne := domain.Vorgabegrenzen()
	ohne.OhneKontingente = true

	if grenzen := rest.GrenzenVon(ohne); grenzen.Kontingente != domain.KontingenteAus ||
		grenzen.Principal.Burst != ohne.Principal.Burst || grenzen.AntwortBytes != ohne.Antwort {
		t.Errorf("Grenzen: %+v", grenzen)
	}

	roh, fehler := json.Marshal(rest.GrenzenVon(ohne))
	if fehler != nil || !strings.Contains(string(roh), `"kontingente":"aus"`) {
		t.Errorf("JSON: %s, %v", roh, fehler)
	}
}

func TestDieSperreTrifftNurDenRatendenNachweis(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)
	ziel := endpunkt{methode: http.MethodGet, weg: "/sets"}
	raten := func(nachweis string) gestalt {
		return func(anfrage *http.Request) {
			anfrage.Host = "127.0.0.1:7777"
			anfrage.Header.Set("Authorization", "Bearer "+nachweis)
		}
	}

	// Drei Versuche frei, der vierte sperrt — der fünfte trifft die Sperre.
	var letzter int
	for range 5 {
		letzter = ruf(t, dienst, ziel, raten("falsch")).Code
	}

	if letzter != http.StatusTooManyRequests {
		t.Errorf("der Ratende: %d", letzter)
	}

	// **Der Berechtigte bleibt frei, und ein anderer Rater beginnt bei null.**
	if antwort := ruf(t, dienst, ziel, gueltig); antwort.Code != http.StatusOK {
		t.Errorf("der Berechtigte: %d", antwort.Code)
	}

	if antwort := ruf(t, dienst, ziel, raten("anders")); antwort.Code != http.StatusUnauthorized {
		t.Errorf("ein anderer Nachweis: %d", antwort.Code)
	}
}

func TestDieWartezeitLaeuftMitDerUhr(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	uhr := func() time.Time { return jetzt }
	dienst := mitHaushalt(t, neuerDienstUm(t, uhr), engeGrenzen(), uhr)
	ziel := endpunkt{methode: http.MethodGet, weg: "/sets"}

	for range 3 {
		ruf(t, dienst, ziel, gueltig)
	}

	if antwort := ruf(t, dienst, ziel, gueltig); antwort.Code != http.StatusTooManyRequests ||
		antwort.Header().Get("Retry-After") != "1" ||
		!strings.Contains(antwort.Body.String(), "in 1 s wieder") {
		t.Fatalf("leer: %d %q %s", antwort.Code, antwort.Header().Get("Retry-After"), antwort.Body.String())
	}

	// Eine Sekunde später ist eine Einheit nachgelaufen.
	jetzt = jetzt.Add(time.Second)

	if antwort := ruf(t, dienst, ziel, gueltig); antwort.Code != http.StatusOK {
		t.Errorf("nach einer Sekunde: %d", antwort.Code)
	}
}

func TestEineZuGrosseAntwortBrichtAb(t *testing.T) {
	t.Parallel()

	grenzen := domain.Vorgabegrenzen()
	grenzen.Antwort = 8
	dienst := neuerDienst(t).MitGrenzen(grenzen)

	defer func() {
		// **Abbruch, nicht Kürzung**: `net/http` schließt die Verbindung ohne
		// Stapelspur, und der Aufrufer bekommt kein halbes JSON.
		fehler, istFehler := recover().(error)
		if !istFehler || !errors.Is(fehler, http.ErrAbortHandler) {
			t.Errorf("kein Abbruch: %v", fehler)
		}
	}()

	ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/sets"}, gueltig)
}

func TestEinZuGrosserRumpfWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	// **Die Grenze gilt dem Rumpf wie der Antwort** — hier so gewählt, dass
	// die Abweisung selbst noch hinauspasst.
	grenzen := domain.Vorgabegrenzen()
	grenzen.Antwort = 512
	dienst := neuerDienst(t).MitGrenzen(grenzen)

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/sets",
		rumpf: `{"name":"` + strings.Repeat("x", 2000) + `","spiel":"magic"}`,
	}, gueltig)

	if antwort.Code < http.StatusBadRequest || antwort.Code == http.StatusTooManyRequests {
		t.Errorf("ein Rumpf über der Grenze kam durch: %d %s", antwort.Code, antwort.Body.String())
	}
}

func TestVieleVerschiedeneNachweiseTreffenDieDrosselFuerAlle(t *testing.T) {
	t.Parallel()

	dienst := neuerDienst(t)
	ziel := endpunkt{methode: http.MethodGet, weg: "/sets"}

	// **Je Nachweis bleibt es bei drei freien Versuchen — aber nicht für
	// beliebig viele Nachweise**: die Drossel für alle Ratenden zusammen steht
	// dahinter, sechzig in der Minute, damit ein Wörterbuch nicht durchläuft.
	var letzter int

	for i := range 61 {
		nachweis := "rate" + strconv.Itoa(i)
		letzter = ruf(t, dienst, ziel, func(anfrage *http.Request) {
			anfrage.Host = "127.0.0.1:7777"
			anfrage.Header.Set("Authorization", "Bearer "+nachweis)
		}).Code
	}

	if letzter != http.StatusTooManyRequests {
		t.Errorf("der 61. Nachweis: %d", letzter)
	}
}

func TestOhneHaushaltGibtEsKeinGuthabenUndKeineSperre(t *testing.T) {
	t.Parallel()

	// **Nur die Drossel für alle, wie vor O5** — der Wurzelknoten reicht den
	// Haushalt herein; ein Dienst ohne ihn sperrt niemanden und bucht nichts.
	dienst := rest.NeuerDienst(probemarke, festeUhr, rest.Naehte{Leser: williger{}, Grenzen: williger{}})
	ziel := endpunkt{methode: http.MethodGet, weg: "/sets"}

	for range 5 {
		if antwort := ruf(t, dienst, ziel, ohneMarke); antwort.Code != http.StatusUnauthorized {
			t.Fatalf("ohne Haushalt: %d", antwort.Code)
		}
	}

	for range 130 {
		if antwort := ruf(t, dienst, ziel, gueltig); antwort.Code != http.StatusOK {
			t.Fatalf("ohne Guthaben gedrosselt: %d", antwort.Code)
		}
	}

	// Auch ein vorgezeigter falscher Nachweis wird nicht gesperrt — nur gezählt.
	for range 5 {
		antwort := ruf(t, dienst, ziel, func(anfrage *http.Request) {
			anfrage.Host = "127.0.0.1:7777"
			anfrage.Header.Set("Authorization", "Bearer falsch")
		})
		if antwort.Code != http.StatusUnauthorized {
			t.Fatalf("ohne Sperre: %d", antwort.Code)
		}
	}
}
