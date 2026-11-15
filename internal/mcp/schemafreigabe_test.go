package mcp_test

// **Jeder Ruf hält die Schemafreigabe** — Zweitblick über Phase 2 von 0.9.0,
// Befunde 1 und 2 (#2323): dieselbe Lage wie am Dienst. Die Fassungswache
// stand in der Schreibwache — hinter der Tokenwache, nicht vor den Werkzeugen,
// die den Zeiger bewegen, und sie fragte nur, statt die Antwort zu halten.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestJederRufHaeltDieSchemafreigabe(t *testing.T) {
	t.Parallel()

	var geholt, gehalten, waehrend atomic.Int32

	werkzeug := func(name string, art mcp.Art, zeiger bool) mcp.Werkzeug {
		return mcp.Werkzeug{
			Name: name, Art: art, Erklaerung: "misst", Zeigerwechsel: zeiger,
			Schema: map[string]any{"type": "object"},
			Ruf: func(context.Context, json.RawMessage) (string, error) {
				waehrend.Store(gehalten.Load())

				return "gut", nil
			},
		}
	}

	server := mcp.NeuerServer([]mcp.Werkzeug{
		werkzeug("lesen.probe", mcp.Lesend, false),
		werkzeug("schreiben.profil.probe", mcp.Schreibend, true),
	}).MitSchemafreigabe(func(context.Context) (func(), error) {
		geholt.Add(1)
		gehalten.Add(1)

		return func() { gehalten.Add(-1) }, nil
	})

	for _, name := range []string{"lesen.probe", "schreiben.profil.probe"} {
		waehrend.Store(0)

		antworten(t, server, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"`+name+`","arguments":{}}}`)

		if waehrend.Load() != 1 || gehalten.Load() != 0 {
			t.Errorf("%s: während %d, danach %d", name, waehrend.Load(), gehalten.Load())
		}
	}

	if geholt.Load() != 2 {
		t.Errorf("geholt %d, erwartet zweimal", geholt.Load())
	}
}

// **Ohne Freigabe läuft kein Werkzeug** — auch kein lesendes.
func TestOhneSchemafreigabeLaeuftKeinWerkzeug(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "lesen.probe", Art: mcp.Lesend, Erklaerung: "liest",
		Schema: map[string]any{"type": "object"},
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			t.Error("das Werkzeug lief ohne Freigabe")

			return "", nil
		},
	}}).MitSchemafreigabe(func(context.Context) (func(), error) {
		return nil, errors.New("der Bestand wird gerade gewandert")
	})

	antwort := antworten(t, server, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.probe","arguments":{}}}`)

	if len(antwort) != 1 || !strings.Contains(antwort[0], `"isError":true`) {
		t.Errorf("ohne Freigabe: %v", antwort)
	}
}

// httpRuf schickt eine Anfrage über HTTP an einen Server.
func httpRuf(t *testing.T, server *mcp.Server, methode, rumpf string) *httptest.ResponseRecorder {
	t.Helper()

	anfrage := httptest.NewRequestWithContext(t.Context(), methode, mcp.Endpunkt, strings.NewReader(rumpf))
	anfrage.Host = "127.0.0.1:7778"
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)
	anfrage.Header.Set("Accept", "application/json, text/event-stream")

	schreiben := httptest.NewRecorder()
	mcp.UeberHTTPMit(server, probemarke).ServeHTTP(schreiben, anfrage)

	return schreiben
}

// **Über HTTP vor allem, was den Bestand liest** — Nachlauf über Phase 2 von
// 0.9.0, Befund 3: `prozessTaugt` liest Profil und Schalter vor der Freigabe,
// und das betraf `initialize`, `ping`, `tools/list`, GET und DELETE. Und ein
// Werkzeug holt sie unter ihr nicht ein zweites Mal.
func TestUeberHTTPHaeltJedeAnfrageDieFreigabe(t *testing.T) {
	t.Parallel()

	nein := mcp.NeuerServer(nil).MitSchemafreigabe(func(context.Context) (func(), error) {
		return nil, errors.New("der Bestand wird gerade gewandert")
	})

	for _, methode := range []string{http.MethodPost, http.MethodGet, http.MethodDelete} {
		if antwort := httpRuf(t, nein, methode, `{"jsonrpc":"2.0","id":1,"method":"ping"}`); antwort.Code != http.StatusServiceUnavailable {
			t.Errorf("%s ohne Freigabe: %d %s", methode, antwort.Code, antwort.Body.String())
		}
	}

	var geholt atomic.Int32

	ja := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "lesen.probe", Art: mcp.Lesend, Erklaerung: "liest",
		Schema: map[string]any{"type": "object"},
		Ruf:    func(context.Context, json.RawMessage) (string, error) { return "gut", nil },
	}}).MitSchemafreigabe(func(context.Context) (func(), error) {
		geholt.Add(1)

		return func() {}, nil
	})

	httpRuf(t, ja, http.MethodPost, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.probe","arguments":{}}}`)

	if geholt.Load() != 1 {
		t.Errorf("geholt %d, erwartet einmal", geholt.Load())
	}
}
