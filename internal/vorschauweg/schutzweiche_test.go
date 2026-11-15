package vorschauweg

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// **Die Schutzweiche** — SEC-17
// (#2164):
// wie die Richtlinie an jede Antwort kommt und was mit einem Bericht
// geschieht. Was die Richtlinie sagt, prüft `internal/fenster`.

// weiter antwortet mit 418 — so ist zu sehen, ob eine Anfrage durchging.
var weiter = http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
	schreiben.WriteHeader(http.StatusTeapot)
})

// **Beobachtend steht sie als Bericht da** — gemeldet, nicht durchgesetzt.
func TestBeobachtendWirdNurBerichtet(t *testing.T) {
	t.Parallel()

	weiche := Schutzweiche(weiter, probebuendel, "default-src 'self'", true, slog.New(slog.DiscardHandler))

	antwort := httptest.NewRecorder()
	weiche.ServeHTTP(antwort, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody))

	if antwort.Header().Get("Content-Security-Policy") != "" {
		t.Error("beobachtend wurde durchgesetzt")
	}

	if antwort.Header().Get("Content-Security-Policy-Report-Only") != "default-src 'self'" {
		t.Errorf("Kopfzeilen: %v", antwort.Header())
	}
}

// **Ein Bericht landet im Protokoll** — mit Anweisung und abgewiesener
// Quelle, und er geht nicht an Wails weiter.
func TestEinBerichtLandetImProtokoll(t *testing.T) {
	t.Parallel()

	var protokoll bytes.Buffer

	weiche := Schutzweiche(weiter, probebuendel, "default-src 'self'", false, slog.New(slog.NewTextHandler(&protokoll, nil)))

	for name, fall := range map[string]struct {
		rumpf    string
		erwartet []string
	}{
		"ein Bericht": {
			`{"csp-report":{"document-uri":"wails://localhost/","violated-directive":"script-src-elem","effective-directive":"script-src-elem","blocked-uri":"inline"}}`,
			[]string{"CSP", "script-src-elem", "inline"},
		},
		"einer ohne effective-directive": {
			`{"csp-report":{"violated-directive":"img-src 'self'","blocked-uri":"data"}}`,
			[]string{"CSP", "img-src", "data"},
		},
		"ein unlesbarer": {`kein JSON`, []string{"CSP", "unlesbar"}},
		"ein übergroßer": {`{"csp-report":{"blocked-uri":"` + strings.Repeat("x", 2*Berichtsgrenze) + `"}}`, []string{"CSP", "unlesbar"}},
	} {
		protokoll.Reset()

		antwort := httptest.NewRecorder()
		weiche.ServeHTTP(antwort, httptest.NewRequestWithContext(t.Context(), http.MethodPost, Berichtsweg, strings.NewReader(fall.rumpf)))

		if antwort.Code != http.StatusNoContent {
			t.Errorf("%s: Status %d, erwartet 204", name, antwort.Code)
		}

		for _, teil := range fall.erwartet {
			if !strings.Contains(protokoll.String(), teil) {
				t.Errorf("%s: das Protokoll nennt %q nicht: %s", name, teil, protokoll.String())
			}
		}
	}
}

// **Ins Protokoll kommt der Ursprung, nicht die Adresse** — was eine Seite
// laden wollte, kann tragen, was ein Mensch getippt hat.
func TestEinAbgewiesenerWegKommtOhnePfadInsProtokoll(t *testing.T) {
	t.Parallel()

	for roh, erwartet := range map[string]string{
		"inline": "inline",
		"eval":   "eval",
		"data":   "data",
		"https://sec17.invalid/fremd.js?karte=Ada": "https://sec17.invalid",
		"wails://localhost/vorschau/12.png":        "wails://localhost",
		"/vorschau/Geheimer%20Name.png":            "«Adresse»",
	} {
		if ist := ursprung(roh); ist != erwartet {
			t.Errorf("ursprung(%q) = %q, erwartet %q", roh, ist, erwartet)
		}
	}
}

// **Der Bau legt die Weiche vor die innere Leitung** — die Richtlinie steht an
// der Antwort, und die innere Leitung sieht die Anfrage.
func TestDerSchutzbauLegtDieWeicheVorDieInnereLeitung(t *testing.T) {
	t.Parallel()

	gesehen := false
	innen := func(weiter http.Handler) http.Handler {
		return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gesehen = true
			weiter.ServeHTTP(schreiben, anfrage)
		})
	}

	antwort := httptest.NewRecorder()
	Schutzbau(innen, probebuendel, "default-src 'self'", false, slog.New(slog.DiscardHandler))(weiter).ServeHTTP(
		antwort, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody))

	if !gesehen || antwort.Code != http.StatusTeapot {
		t.Errorf("die innere Leitung sah die Anfrage: %v, Status %d", gesehen, antwort.Code)
	}

	if antwort.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Errorf("Kopfzeilen: %v", antwort.Header())
	}
}
