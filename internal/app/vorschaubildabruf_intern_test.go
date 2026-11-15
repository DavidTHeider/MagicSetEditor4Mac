package app

// Das Gegenstück zu `bildbytes` in `vorschauweg_test.go`, für die Tests **in**
// diesem Paket — die beiden können sich nicht teilen, weil `package app` und
// `package app_test` verschiedene Pakete sind.

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// bildausDemWeg holt ein Vorschaubild über den echten Weg.
func bildausDemWeg(t *testing.T, anwendung *App, adresse string) []byte {
	t.Helper()

	schreiben := httptest.NewRecorder()
	Bildweg(anwendung).ServeHTTP(schreiben,
		httptest.NewRequestWithContext(probekontext(t), http.MethodGet, adresse, http.NoBody))

	if schreiben.Code != http.StatusOK {
		t.Fatalf("Bild %q abholen: Code %d", adresse, schreiben.Code)
	}

	return schreiben.Body.Bytes()
}
