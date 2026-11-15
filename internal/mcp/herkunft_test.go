package mcp_test

// **Herkunft vor Werkzeugwirkung** — SEC-11
// (#2140).
//
// Der Werkzeugserver über HTTP ließ jede Anfrage ohne `Origin` durch, auch
// mit fremdem `Host` und fremdem `Referer`: eine Seite, die einen Namen per
// DNS-Rebinding auf die Schleife zeigen lässt, schickt bei gleicher Herkunft
// **kein** `Origin` — wohl aber ihren Namen als `Host`. Mit der eigenen Marke
// gab das `DELETE` ein 204. Jetzt gilt dieselbe Regel wie beim Dienst
// (`rest.HerkunftTaugtUnter`): kein `Origin`, kein `Referer`, und `Host` ist
// einer der eigenen Namen.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestDieHerkunftEntscheidetVorDerWerkzeugwirkung(t *testing.T) {
	t.Parallel()

	behandler := mcp.UeberHTTPMit(
		mcp.NeuerServer(mcp.Werkzeuge(mcp.Naehte{Netzweg: &netzwegprobe{}})),
		probemarke, "192.168.1.5", "tisch.local")

	faelle := []struct {
		name                string
		host, herkunft, von string
		abgewiesen          bool
	}{
		{name: "Schleife", host: "127.0.0.1:7778"},
		{name: "localhost", host: "localhost:7778"},
		{name: "localhost ohne Port", host: "localhost"},
		{name: "eigene LAN-Adresse", host: "192.168.1.5:7778"},
		{name: "eigener Name aus dem Zeugnis", host: "tisch.local:7778"},
		{name: "fremder Host ohne Origin", host: "boese.example:7778", abgewiesen: true},
		{name: "fremder Referer", host: "127.0.0.1:7778", von: "https://boese.example/", abgewiesen: true},
		{name: "fremdes Origin", host: "127.0.0.1:7778", herkunft: "https://boese.example", abgewiesen: true},
		{name: "eigenes Origin aus einem Browser", host: "127.0.0.1:7778", herkunft: "http://127.0.0.1:7778", abgewiesen: true},
		{name: "IPv6-Schleife wird nicht gebunden", host: "[::1]:7778", abgewiesen: true},
	}

	for _, fall := range faelle {
		for _, methode := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
			t.Run(methode+" "+fall.name, func(t *testing.T) {
				t.Parallel()

				anfrage := httptest.NewRequestWithContext(t.Context(), methode, mcp.Endpunkt,
					strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
				anfrage.Host = fall.host
				anfrage.Header.Set("Authorization", "Bearer "+probemarke)

				if fall.herkunft != "" {
					anfrage.Header.Set("Origin", fall.herkunft)
				}

				if fall.von != "" {
					anfrage.Header.Set("Referer", fall.von)
				}

				schreiben := httptest.NewRecorder()
				behandler.ServeHTTP(schreiben, anfrage)

				if verboten := schreiben.Code == http.StatusForbidden; verboten != fall.abgewiesen {
					t.Errorf("Status %d, abgewiesen erwartet: %v — %s", schreiben.Code, fall.abgewiesen, schreiben.Body.String())
				}
			})
		}
	}
}
