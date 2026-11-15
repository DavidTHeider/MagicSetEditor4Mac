package mcp_test

// **Falsche Marken haben ein eigenes Budget** — SEC-12
// (#2141).
//
// Der Werkzeugserver verglich die Marke in konstanter Zeit und ließ jeden
// beliebig oft raten. Jetzt gilt dieselbe Grenze wie beim Dienst — 60
// Fehlversuche je Minute —, und **nur** Fehlversuche zählen: wer die Marke
// hat, arbeitet in seinem eigenen Budget weiter, auch wenn ein anderer gerade
// rät.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestFalscheMarkenHabenEinEigenesBudget(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 7, 3, 0, 0, 0, time.UTC)
	server := mcp.NeuerServer(mcp.Werkzeuge(mcp.Naehte{Netzweg: &netzwegprobe{}})).
		MitUhr(func() time.Time { return jetzt })
	behandler := mcp.UeberHTTPMit(server, probemarke)

	frage := func(methode, marke string) *httptest.ResponseRecorder {
		anfrage := httptest.NewRequestWithContext(t.Context(), methode, mcp.Endpunkt,
			strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
		anfrage.Host = "127.0.0.1:7778"
		anfrage.Header.Set("Authorization", "Bearer "+marke)

		schreiben := httptest.NewRecorder()
		behandler.ServeHTTP(schreiben, anfrage)

		return schreiben
	}

	methoden := []string{http.MethodGet, http.MethodPost, http.MethodDelete}

	for nummer := range 60 {
		if antwort := frage(methoden[nummer%3], "falsch"); antwort.Code != http.StatusUnauthorized {
			t.Fatalf("Versuch %d: %d", nummer+1, antwort.Code)
		}
	}

	ueber := frage(http.MethodPost, "falsch")
	if ueber.Code != http.StatusTooManyRequests || ueber.Header().Get("Retry-After") == "" {
		t.Fatalf("der 61. Versuch: %d, Retry-After %q", ueber.Code, ueber.Header().Get("Retry-After"))
	}

	// **Die gültige Marke arbeitet weiter** — ihr Budget ist ein anderes.
	if gueltig := frage(http.MethodPost, probemarke); gueltig.Code == http.StatusUnauthorized ||
		gueltig.Code == http.StatusTooManyRequests {
		t.Errorf("die gültige Marke bekam %d", gueltig.Code)
	}

	// **Und nach dem Fenster wird wieder gezählt.**
	jetzt = jetzt.Add(time.Minute)
	if antwort := frage(http.MethodGet, "falsch"); antwort.Code != http.StatusUnauthorized {
		t.Errorf("nach dem Fenster: %d", antwort.Code)
	}
}
