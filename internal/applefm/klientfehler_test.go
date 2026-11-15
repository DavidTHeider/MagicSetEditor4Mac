package applefm_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Die Ausgänge einer Antwort, je wie sie aussieht** — drei davon sind
// gemessen (Vermerk 0166), die übrigen sind die Gestalt, die die
// Schnittstelle zulässt: eine Verweigerung im Feld `refusal`, ein Abbruch
// am Fensterende, gar keine Wahl.
func TestDieAusgaengeEinerAntwort(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		datei   string
		ausgang applefm.Ausgang
	}{
		{"verweigert.json", applefm.AusgangAbgelehnt},
		{"abgeschnitten.json", applefm.AusgangUnvollstaendig},
	}

	for _, fall := range faelle {
		t.Run(fall.datei, func(t *testing.T) {
			t.Parallel()

			klient, _ := dienst(t, map[string]nachbildung{
				"POST /v1/chat/completions": {http.StatusOK, fall.datei, nil},
			})

			antwort, fehler := klient.Rede(t.Context(), applefm.Frage{Auftrag: "x"})
			if fehler != nil || antwort.Ausgang != fall.ausgang {
				t.Errorf("%+v %v", antwort, fehler)
			}
		})
	}
}

// **Ein verschwundenes Pflichtfeld bricht laut** (§16): eine Antwort ohne
// `choices` ist kein leerer Vorschlag, sondern eine andere Gestalt.
func TestEineAntwortOhneWahlIstEinFehler(t *testing.T) {
	t.Parallel()

	klient, _ := dienst(t, map[string]nachbildung{
		"POST /v1/chat/completions": {http.StatusOK, "ohnewahl.json", nil},
	})

	if _, fehler := klient.Rede(t.Context(), applefm.Frage{Auftrag: "x"}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("%v", fehler)
	}
}

// **Was sich nicht lesen lässt, ist ein Fehler mit Grund** — auf jedem Weg.
func TestUnlesbaresIstEinFehler(t *testing.T) {
	t.Parallel()

	klient, _ := dienst(t, map[string]nachbildung{
		"GET /health":               {http.StatusOK, "keinjson.txt", nil},
		"GET /v1/models":            {http.StatusInternalServerError, "keinjson.txt", nil},
		"POST /v1/chat/completions": {http.StatusOK, "keinjson.txt", nil},
	})

	if fehler := klient.Gesund(t.Context()); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("gesund: %v", fehler)
	}

	if _, fehler := klient.Modelle(t.Context()); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Modelle: %v", fehler)
	}

	if _, fehler := klient.Rede(t.Context(), applefm.Frage{Auftrag: "x"}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Rede: %v", fehler)
	}
}

// **Eine Frage, die sich nicht bauen lässt, geht nicht hinaus.**
func TestEineUnbaubareFrageGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	klient, _ := dienst(t, map[string]nachbildung{})

	schema := map[string]any{"kaputt": make(chan int)}
	if _, fehler := klient.Rede(t.Context(), applefm.Frage{Auftrag: "x", Schema: schema}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Schema: %v", fehler)
	}

	falsch := applefm.NeuerKlient(netz.KlientUeber(http.DefaultClient), "http://[::1")
	if fehler := falsch.Gesund(t.Context()); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Adresse: %v", fehler)
	}
}

// **Ein Dienst, der nicht mehr da ist, ist ein Fehler** — und kein Hängen.
func TestEinVerschwundenerDienstIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.NotFoundHandler())
	adresse := server.URL
	server.Close()

	klient := applefm.NeuerKlient(netz.KlientUeber(server.Client()), adresse)
	if _, fehler := klient.Modelle(t.Context()); fehler == nil {
		t.Error("kein Fehler")
	}
}
