package applefm_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Der Klient gegen die gemessene Gestalt** — AI-05 (#2042), Vermerk 0166.
// Die Antworten unter `testdaten/` sind selbst geschrieben, in der Form, die
// `fm serve` am 2026-10-03 gab.

// nachbildung antwortet je Weg mit einer Datei aus `testdaten/` und hält fest,
// was gefragt wurde.
type nachbildung struct {
	status  int
	datei   string
	gefragt map[string]any
}

func dienst(t *testing.T, je map[string]nachbildung) (klient *applefm.Klient, gefragt map[string]any) {
	t.Helper()

	gefragt = map[string]any{}

	server := httptest.NewServer(http.HandlerFunc(func(antwort http.ResponseWriter, anfrage *http.Request) {
		bild, da := je[anfrage.Method+" "+anfrage.URL.Path]
		if !da {
			http.NotFound(antwort, anfrage)

			return
		}

		if anfrage.Body != nil {
			roh, _ := io.ReadAll(anfrage.Body) //nolint:errcheck // Probe
			_ = json.Unmarshal(roh, &gefragt)  //nolint:errcheck // GET hat keinen Rumpf
		}

		inhalt, fehler := os.ReadFile(filepath.Join("testdaten", bild.datei))
		if fehler != nil {
			t.Errorf("Testdatei: %v", fehler)
		}

		antwort.WriteHeader(bild.status)
		_, _ = antwort.Write(inhalt) //nolint:errcheck // Probe
	}))
	t.Cleanup(server.Close)

	return applefm.NeuerKlient(netz.KlientUeber(server.Client()), server.URL), gefragt
}

func TestDerKlientFragtNachGesundheitUndModell(t *testing.T) {
	t.Parallel()

	klient, _ := dienst(t, map[string]nachbildung{
		"GET /health":    {http.StatusOK, "gesund.json", nil},
		"GET /v1/models": {http.StatusOK, "modelle.json", nil},
	})

	if fehler := klient.Gesund(t.Context()); fehler != nil {
		t.Errorf("gesund: %v", fehler)
	}

	modelle, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(modelle) != 1 || modelle[0] != applefm.Modell {
		t.Errorf("%v", modelle)
	}
}

// **Ein Modell, das nicht bereit ist, ist kein gesunder Dienst** — `fm
// available` kann zwischen Start und Frage umschlagen (Apple Intelligence aus).
func TestEinNichtBereitesModellIstNichtGesund(t *testing.T) {
	t.Parallel()

	klient, _ := dienst(t, map[string]nachbildung{"GET /health": {http.StatusOK, "nichtbereit.json", nil}})

	if fehler := klient.Gesund(t.Context()); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("%v", fehler)
	}
}

// **Eine Antwort nach Schema, ohne Strom** — `stream` fehlte, gäbe `fm serve`
// einen Strom (Vermerk 0166).
func TestDerKlientFragtOhneStromUndMitSchema(t *testing.T) {
	t.Parallel()

	klient, gefragt := dienst(t, map[string]nachbildung{
		"POST /v1/chat/completions": {http.StatusOK, "antwort.json", nil},
	})

	schema := map[string]any{"type": "object"}

	antwort, fehler := klient.Rede(t.Context(), applefm.Frage{
		System: "Du schreibst Flavortexte.", Auftrag: "Funkenschmied", Schema: schema,
	})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if antwort.Ausgang != applefm.AusgangFertig || antwort.Modell != applefm.Modell ||
		antwort.Inhalt != `{"name": "Funkenschmied", "flavor": "Aus Glut wird Werkzeug, aus Werkzeug wird Legende."}` {
		t.Errorf("%+v", antwort)
	}

	if gefragt["stream"] != false || gefragt["model"] != applefm.Modell {
		t.Errorf("gefragt: %v", gefragt)
	}

	format, istAbbildung := gefragt["response_format"].(map[string]any)
	if !istAbbildung || format["type"] != "json_schema" {
		t.Errorf("Format: %v", gefragt["response_format"])
	}
}

// **Zwei Arten der Ablehnung, ein Ausgang** — HTTP 500 mit der Meldung zu den
// Schutzregeln, und HTTP 200 mit dem Satz im Inhalt.
func TestEineAblehnungIstEinEigenerAusgang(t *testing.T) {
	t.Parallel()

	for name, bild := range map[string]nachbildung{
		"Schutzregeln": {http.StatusInternalServerError, "abgelehnt-schutz.json", nil},
		"Satz":         {http.StatusOK, "abgelehnt-text.json", nil},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			klient, _ := dienst(t, map[string]nachbildung{"POST /v1/chat/completions": bild})

			antwort, fehler := klient.Rede(t.Context(), applefm.Frage{Auftrag: "Kampf"})
			if fehler != nil {
				t.Fatalf("Rede: %v", fehler)
			}

			if antwort.Ausgang != applefm.AusgangAbgelehnt || antwort.Naeheres == "" {
				t.Errorf("%+v", antwort)
			}
		})
	}
}

// **Was kein Ablehnen ist, ist ein Fehler mit Grund** — zu lang, unbekanntes
// Modell, ein anderer Serverfehler.
func TestFehlerDesDienstesSindFehler(t *testing.T) {
	t.Parallel()

	for name, bild := range map[string]nachbildung{
		"zu lang":            {http.StatusInternalServerError, "zulang.json", nil},
		"unbekanntes Modell": {http.StatusBadRequest, "unbekannt.json", nil},
		"ohne Rumpf":         {http.StatusBadGateway, "gesund.json", nil},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			klient, _ := dienst(t, map[string]nachbildung{"POST /v1/chat/completions": bild})

			if _, fehler := klient.Rede(t.Context(), applefm.Frage{Auftrag: "x"}); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
