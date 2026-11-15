//go:build !server

package fenster

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// **Im Fensterbau ist die Weiche die Bildweiche** — QA-09 (#1740): ein Bild
// geht an den Bildweg, alles andere an Wails. Eine Wache steht hier nicht,
// weil kein Port lauscht (`assetweiche.go`); im Servermodus steht sie davor
// (`assetweiche_server_test.go`, mit `-tags server`).
func TestImFensterbauIstDieWeicheDieBildweiche(t *testing.T) {
	t.Parallel()

	weiche := assetweiche(vorschauweg.Neu(&vorschauweg.Fach{}))(
		http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusTeapot)
		}))

	for weg, erwartet := range map[string]int{
		vorschauweg.Pfad + "1.png": http.StatusNotFound,
		"/wails/runtime":           http.StatusTeapot,
	} {
		antwort := httptest.NewRecorder()
		weiche.ServeHTTP(antwort, httptest.NewRequestWithContext(
			t.Context(), http.MethodGet, weg, http.NoBody))

		if antwort.Code != erwartet {
			t.Errorf("%s: Status %d, erwartet %d", weg, antwort.Code, erwartet)
		}
	}
}
