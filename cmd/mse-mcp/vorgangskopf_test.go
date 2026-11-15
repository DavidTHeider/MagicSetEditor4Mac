package main

// **MCP über HTTP nimmt eine mitgebrachte Vorgangskennung** — Befund 7 des
// Zweitblicks über Phase 6.4: der Dienst tat es seit OPS-05, der
// Werkzeugserver über HTTP nicht.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func TestMCPUeberHTTPNimmtDieKennung(t *testing.T) {
	t.Parallel()

	var gesehen domain.Vorgangskennung

	behandler := mitVorgangskopf(http.HandlerFunc(func(_ http.ResponseWriter, anfrage *http.Request) {
		gesehen = app.VorgangAus(anfrage.Context())
	}))

	for kopf, erwartet := range map[string]domain.Vorgangskennung{
		"0123456789abcdef": "0123456789abcdef",
		"NICHT-HEX":        "",
	} {
		anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", http.NoBody)
		anfrage.Header.Set(rest.Vorgangskopf, kopf)
		behandler.ServeHTTP(httptest.NewRecorder(), anfrage)

		if gesehen != erwartet {
			t.Errorf("%q: %q", kopf, gesehen)
		}
	}
}
