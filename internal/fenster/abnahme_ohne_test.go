//go:build !abnahme || !darwin

package fenster

// **Außerhalb des Abnahmebaus gibt es kein Rufbuch** — B01
// (#2220):
// die Hülle des Asset-Servers bleibt, wie sie ist.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestOhneAbnahmeBleibtDieHuelleDieInnere(t *testing.T) {
	t.Parallel()

	gerufen := 0
	innen := application.Middleware(func(weiter http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			gerufen++
			weiter.ServeHTTP(rw, req)
		})
	})

	huelle := rufbuchHuelle(innen)(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusTeapot)
	}))

	antwort := httptest.NewRecorder()
	huelle.ServeHTTP(antwort, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody))

	if gerufen != 1 || antwort.Code != http.StatusTeapot {
		t.Errorf("gerufen %d mal, Antwort %d", gerufen, antwort.Code)
	}
}
