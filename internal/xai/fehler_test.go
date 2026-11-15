package xai_test

// **Was ein Fehler von xAI heißt** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Der Fall, den dieses Paket anders löst als die anderen zwei
//
// **Die Gestalt des Fehlerrumpfs steht nirgends.** Nachgelesen am 2026-09-12
// unter <https://docs.x.ai/developers/debugging>: die Dokumentation nennt die
// Statuscodes und ihre Bedeutung — und **kein einziges Feld**.
//
// **Also kommt die Fehlerart aus dem Status**, denn der ist dokumentiert, und
// der Rumpf reist als Grund mit, so wie er ankommt. Wer hier eine Feldgestalt
// annähme, hätte sie erfunden (§16, Regel 1) — und das fiele erst auf, wenn
// eine Meldung im Protokoll leer bliebe.

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// **Jeder dokumentierte Status bekommt seine Art.**
func TestJederStatusBekommtSeineArt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		status   int
		erwartet error
	}{
		{
			"401 ist eine Erlaubnisfrage", http.StatusUnauthorized,
			domain.ErrErlaubnis,
		},
		{"403 auch", http.StatusForbidden, domain.ErrErlaubnis},
		{
			"404 heißt: das gibt es nicht", http.StatusNotFound,
			domain.ErrNotFound,
		},

		// **429 ist hier ein Konflikt und keine Erlaubnisfrage.** Anthropic
		// und OpenAI führen dahinter zwei Dinge — zu schnell gefragt **oder**
		// kein Guthaben; xAI nennt genau eines, und das geht vorüber.
		{
			"429 heißt: zu schnell", http.StatusTooManyRequests,
			domain.ErrConflict,
		},

		{
			"409 heißt: das passt nicht zusammen", http.StatusConflict,
			domain.ErrConflict,
		},
		{
			"400 ist ein Eingabefehler", http.StatusBadRequest,
			domain.ErrValidation,
		},

		// **Zwei, die die anderen nicht haben.**
		{"415 auch", http.StatusUnsupportedMediaType, domain.ErrValidation},
		{"422 auch", http.StatusUnprocessableEntity, domain.ErrValidation},

		{
			"und 500 ebenso", http.StatusInternalServerError,
			domain.ErrValidation,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := einFehlerBei(t, fall.status, `{"error":"so nicht"}`)
			if !errors.Is(fehler, fall.erwartet) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			// **Und der Status steht im Satz** — sonst stünde der Leser vor
			// „xAI hat abgelehnt" ohne zu wissen, wogegen.
			if !strings.Contains(fehler.Error(), http.StatusText(fall.status)) &&
				!strings.Contains(fehler.Error(), strconv.Itoa(fall.status)) {
				t.Errorf("der Satz nennt den Status nicht: %q", fehler)
			}
		})
	}
}

// **Was dasteht, reist mit — und wo nichts dasteht, der rohe Rumpf.**
//
// **Der rohe Rumpf ist hier die ehrlichere Auskunft**: eine leere Meldung
// verschwiege, dass der Dienst etwas gesagt hat, das dieses Programm nicht
// einordnen kann.
func TestDerGrundReistMit(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		rumpf    string
		erwartet string
	}{
		{
			"Meldung und Code", `{"error":"so nicht","code":"bad_thing"}`,
			"so nicht (bad_thing)",
		},
		{"nur die Meldung", `{"error":"so nicht"}`, "so nicht"},
		{
			"eine fremde Gestalt kommt roh durch",
			`{"detail":{"msg":"so nicht"}}`, `{"detail":{"msg":"so nicht"}}`,
		},
		{"kein JSON kommt roh durch", "Bad Gateway", "Bad Gateway"},
		{"und ein leerer Rumpf sagt das", "", "ohne Rumpf"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := einFehlerBei(t, http.StatusBadRequest, fall.rumpf)

			grund := domain.GrundVon(fehler)
			if grund == nil {
				t.Fatalf("der Fehler trägt keinen Grund: %v", fehler)
			}

			if !strings.Contains(grund.Error(), fall.erwartet) {
				t.Errorf("der Grund ist %q, erwartet war %q",
					grund, fall.erwartet)
			}

			// **Und der Grund nennt den Weg** — bei zwei Wegen ist die
			// Meldung allein keine Ortsangabe.
			if !strings.Contains(grund.Error(), xai.Modellweg) {
				t.Errorf("der Grund nennt den Weg nicht: %q", grund)
			}
		})
	}
}

// **Ein 429 wird gedeutet und nicht als Netzfehler durchgereicht.**
//
// **Sonst käme er als »die Antwort war 429« an**, und niemand wüsste, dass
// er vorübergeht.
func TestEin429WirdGedeutet(t *testing.T) {
	t.Parallel()

	fehler := einFehlerBei(t, http.StatusTooManyRequests,
		`{"error":"rate limit"}`)

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "429") {
		t.Errorf("der Satz nennt die 429 nicht: %q", fehler)
	}
}

// einFehlerBei fragt einen Dienst, der genau so ablehnt.
func einFehlerBei(t *testing.T, status int, rumpf string) error {
	t.Helper()

	klient := xai.NeuerKlient(einDienstDer(t, status, rumpf),
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	_, _, fehler := klient.Modelle(t.Context())
	if fehler == nil {
		t.Fatalf("der Status %d ist als Erfolg durchgegangen", status)
	}

	return fehler
}
