//go:build !server

package fenster

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// **Die WebView läuft unter einer Schutzrichtlinie** — SEC-17
// (#2164).
// Bis dahin setzte weder `frontend/index.html` noch die Hülle eine CSP; ein
// eingeschleustes Skript hätte alles gedurft, was die Seite darf. Ob WebKit
// die Kopfzeile über Wails' eigenes Schema wirklich durchsetzt, belegt die
// Hüllenprobe `sec17`; hier steht, dass sie gesetzt wird und was sie sagt.

// weiter antwortet mit 418 — so ist zu sehen, ob eine Anfrage durchging.
var weiter = http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
	schreiben.WriteHeader(http.StatusTeapot)
})

func TestDieWeicheSetztDieSchutzrichtlinie(t *testing.T) {
	t.Parallel()

	weiche := schutzbau(assetweiche(vorschauweg.Neu(&vorschauweg.Fach{})), fstest.MapFS{}, slog.New(slog.DiscardHandler))(weiter)

	for _, weg := range []string{"/", "/assets/index.js", vorschauweg.Pfad + "1.png"} {
		antwort := httptest.NewRecorder()
		weiche.ServeHTTP(antwort, httptest.NewRequestWithContext(t.Context(), http.MethodGet, weg, http.NoBody))

		richtlinie := antwort.Header().Get("Content-Security-Policy")
		if richtlinie == "" {
			t.Errorf("%s: keine Schutzrichtlinie", weg)

			continue
		}

		if richtlinie != schutzrichtlinie(cspZusatz()) {
			t.Errorf("%s: %q", weg, richtlinie)
		}
	}
}

// **Was die Richtlinie sagt** — Skripte nur aus dem eigenen Bündel, kein
// `eval`, keine fremden Ziele; die Ausnahmen sind benannt und gemessen.
func TestDieSchutzrichtlinieLaesstNurEigeneSkripteZu(t *testing.T) {
	t.Parallel()

	richtlinie := schutzrichtlinie("")

	anweisungen := map[string]string{}
	for _, teil := range strings.Split(richtlinie, ";") {
		felder := strings.Fields(teil)
		if len(felder) > 0 {
			anweisungen[felder[0]] = strings.Join(felder[1:], " ")
		}
	}

	for name, erwartet := range map[string]string{
		"default-src": "'self'",
		"script-src":  "'self'",
		"connect-src": "'self'",
		// **Die Vorschaubilder der Kartensuche** kommen von den Diensten
		// selbst — gemessen am 2026-10-07, siehe [fremdeBildherkuenfte].
		"img-src":    "'self' data: https://cards.scryfall.io https://gatherer.wizards.com https://gatherer-static.wizards.com",
		"object-src": "'none'",
		"base-uri":   "'none'",
		"report-uri": vorschauweg.Berichtsweg,
	} {
		if anweisungen[name] != erwartet {
			t.Errorf("%s ist %q, erwartet %q", name, anweisungen[name], erwartet)
		}
	}

	if strings.Contains(anweisungen["script-src"], "unsafe") {
		t.Errorf("script-src erlaubt mehr als das Bündel: %q", anweisungen["script-src"])
	}
}

// **Der Abnahmebau lässt den Rückkanal der Probe zu — und nur ihn.**
func TestEinZusatzOeffnetNurConnectSrc(t *testing.T) {
	t.Parallel()

	ohne := schutzrichtlinie("")
	mit := schutzrichtlinie("http://localhost:9300")

	if !strings.Contains(mit, "connect-src 'self' http://localhost:9300;") {
		t.Errorf("der Zusatz fehlt: %q", mit)
	}

	if strings.Replace(mit, " http://localhost:9300", "", 1) != ohne {
		t.Errorf("der Zusatz änderte mehr als connect-src:\n%s\n%s", ohne, mit)
	}
}
