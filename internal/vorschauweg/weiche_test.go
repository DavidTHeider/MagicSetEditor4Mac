package vorschauweg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// **Die Weiche trennt Bild und Oberfläche** — ENG-02: die drei Vorsätze gehen
// an den Bildweg, alles andere an den Weg dahinter.
//
// **Am Bildweg selbst gemessen und nicht an einem Doppel**: ein unbekanntes
// Bild beantwortet er mit 404, und die Oberfläche dahinter sagt, dass sie
// gefragt wurde.
func TestDieWeicheTrenntBildUndOberflaeche(t *testing.T) {
	t.Parallel()

	weiter := http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
		schreiben.WriteHeader(http.StatusTeapot)
	})
	weiche := vorschauweg.Neu(&vorschauweg.Fach{}).Vor(weiter)

	for weg, erwartet := range map[string]int{
		vorschauweg.Pfad + "abc.png":      http.StatusNotFound,
		vorschauweg.Spielpfad + "abc.png": http.StatusNotFound,
		vorschauweg.Tischpfad + "7.png":   http.StatusNotFound,
		"/":                               http.StatusTeapot,
		"/index.html":                     http.StatusTeapot,
		"/assets/index.js":                http.StatusTeapot,
		"/vorschau":                       http.StatusTeapot,
		"/wails/runtime":                  http.StatusTeapot,
	} {
		antwort := httptest.NewRecorder()
		weiche.ServeHTTP(antwort, httptest.NewRequestWithContext(t.Context(), http.MethodGet, weg, http.NoBody))

		if antwort.Code != erwartet {
			t.Errorf("%s ergab %d, erwartet %d", weg, antwort.Code, erwartet)
		}
	}
}
