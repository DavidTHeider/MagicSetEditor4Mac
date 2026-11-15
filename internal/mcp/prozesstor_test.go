package mcp_test

// **Der eigene Schalter vor jeder Anfrage** — Z7 des Zweitblicks über Phase
// 6.12 (#1905).
//
// Gefragt wurde er nur vor einem Werkzeug: `ping`, `initialize` und
// `tools/list` gingen weiter durch, und der Prozess lief.

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestEinAbgeschalteterProzessBeantwortetNichts(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer(mcp.Werkzeuge(mcp.Naehte{Netzweg: &netzwegprobe{}})).
		MitSchaltern(func(context.Context, string) error {
			return &domain.Abschaltung{Schalter: "werkzeugserver"}
		})
	behandler := mcp.UeberHTTPMit(server, probemarke)

	for _, methode := range []string{"ping", "initialize", "tools/list"} {
		antwort := frage(t, behandler, http.MethodPost,
			`{"jsonrpc":"2.0","id":1,"method":"`+methode+`"}`, probemarke)
		if antwort.Code != http.StatusForbidden || !strings.Contains(antwort.Body.String(), "abgeschaltet") {
			t.Errorf("%s: %d %s", methode, antwort.Code, antwort.Body.String())
		}
	}

	// **Erst die Marke, dann der Schalter** — wer keine hat, erfährt nichts
	// über ihn.
	if antwort := frage(t, behandler, http.MethodPost,
		`{"jsonrpc":"2.0","id":1,"method":"ping"}`, ""); antwort.Code != http.StatusUnauthorized {
		t.Errorf("ohne Marke: %d", antwort.Code)
	}
}

// **Lässt sich der Kontext nicht ausstatten, geht die Anfrage weiter** — sie
// sagt den Fehler selbst, und der Schalter wird nicht ohne Profil gefragt.
func TestOhneAusstattungFragtDerProzessKeinenSchalter(t *testing.T) {
	t.Parallel()

	gefragt := false
	server := mcp.NeuerServer(mcp.Werkzeuge(mcp.Naehte{Netzweg: &netzwegprobe{}})).
		MitAusstattung(func(context.Context) (context.Context, error) {
			return nil, errors.New("kein Profil")
		}).
		MitSchaltern(func(context.Context, string) error {
			gefragt = true

			return &domain.Abschaltung{Schalter: "werkzeugserver"}
		})

	antwort := frage(t, mcp.UeberHTTPMit(server, probemarke), http.MethodPost,
		`{"jsonrpc":"2.0","id":1,"method":"ping"}`, probemarke)
	if antwort.Code != http.StatusOK || gefragt {
		t.Errorf("%d, gefragt %v", antwort.Code, gefragt)
	}
}
