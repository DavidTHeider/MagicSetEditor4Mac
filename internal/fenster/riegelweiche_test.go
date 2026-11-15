package fenster

// **Eine Panik in der Bildweiche beendet die Anwendung nicht** — F02
// (#2341).
//
// Wails beantwortet jede Anfrage der Webansicht in einer eigenen Goroutine
// ohne `recover` (`assetserver_webview.go`, `ServeWebViewRequest`,
// v3.0.0-beta.27) — sein `PanicHandler` sieht sie nicht. Eine Panik in der
// Bildweiche, der Schutznaht oder dem Rufbuch endete im Absturz.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestEinePanikInDerWeicheWirdZuEiner500(t *testing.T) {
	gesammelt := protokolliert(t)

	wirft := application.Middleware(func(http.Handler) http.Handler {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("Probepanik in der Weiche") })
	})

	antwort := httptest.NewRecorder()
	riegelweiche(wirft)(http.NotFoundHandler()).ServeHTTP(antwort, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/bild/probe.png", http.NoBody))

	if antwort.Code != http.StatusInternalServerError {
		t.Errorf("Status %d, erwartet 500", antwort.Code)
	}

	zeile := gesammelt.String()
	if !strings.Contains(zeile, "Probepanik in der Weiche") || !strings.Contains(zeile, "kennung=") {
		t.Errorf("die Panik steht nicht mit Kennung im Protokoll:\n%s", zeile)
	}
}

func TestEineGuteAnfrageGehtDurchDieRiegelweiche(t *testing.T) {
	durch := application.Middleware(func(weiter http.Handler) http.Handler { return weiter })

	antwort := httptest.NewRecorder()
	riegelweiche(durch)(http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
		schreiben.WriteHeader(http.StatusTeapot)
	})).ServeHTTP(antwort, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody))

	if antwort.Code != http.StatusTeapot {
		t.Errorf("Status %d, erwartet der des Handlers", antwort.Code)
	}
}
