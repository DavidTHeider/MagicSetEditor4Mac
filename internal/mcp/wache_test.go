package mcp_test

// **Die Wache je Ruf** — Werkstück O2 von Phase 5.2
// (#1426):
// ein Token, das zwischen zwei Rufen zurückgezogen wird, trägt den nächsten
// nicht mehr.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die Wache sieht jeden Ruf mit der Art seines Werkzeugs — und ihr Nein
// kommt als Ergebnis zurück, bevor das Werkzeug läuft.**
func TestDieWacheStehtVorJedemRuf(t *testing.T) {
	var gesehen []mcp.Art

	gerufen := 0
	zurueckgezogen := false

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "schreiben.probe", Art: mcp.Schreibend, Erklaerung: "zählt",
		Schema: map[string]any{"type": "object"},
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			gerufen++

			return "gut", nil
		},
	}}).MitWache(func(ctx context.Context, art mcp.Art) (context.Context, error) {
		gesehen = append(gesehen, art)

		if zurueckgezogen {
			return ctx, errors.New("dieses Token gilt nicht")
		}

		return ctx, nil
	})

	ruf := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.probe","arguments":{}}}`

	erste := antworten(t, server, ruf)
	if gerufen != 1 || len(erste) != 1 || strings.Contains(erste[0], "gilt nicht") {
		t.Fatalf("der erste Ruf kam nicht durch: %v (gerufen %d)", erste, gerufen)
	}

	// **Zwischen zwei Rufen zurückgezogen** — der zweite scheitert, ohne
	// dass das Werkzeug noch einmal läuft.
	zurueckgezogen = true

	zweite := antworten(t, server, ruf)
	if gerufen != 1 || len(zweite) != 1 || !strings.Contains(zweite[0], "gilt nicht") ||
		!strings.Contains(zweite[0], `"isError":true`) {
		t.Errorf("der zweite Ruf lief trotz Widerruf: %v (gerufen %d)", zweite, gerufen)
	}

	if len(gesehen) != 2 || gesehen[0] != mcp.Schreibend {
		t.Errorf("die Wache sah %v", gesehen)
	}
}
