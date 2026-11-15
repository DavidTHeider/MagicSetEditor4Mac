package mcp_test

// **Geschrieben ist geschrieben — auch wo das Ergebnis eine Liste ist** —
// §7, Regel 3 (#1970).
//
// # Wogegen hier geprüft wird
//
// Ein Werkzeug, dessen Ergebnis eine Liste ist, hat kein Feld für einen
// Nachladefehler; die Liste in ein Objekt zu packen wäre brechend (0033).
// **Der Server gibt dann zwei Teile**: die Liste, wie sie war, und einen
// Satz dahinter — als geglücktes Ergebnis und nicht als Fehler, denn ein
// Modell, das `isError` liest, versuchte es noch einmal.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestEinNachgeladenesErgebnisHatZweiTeileUndIstKeinFehler(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "schreiben.probe", Art: mcp.Schreibend, Erklaerung: "Glückt halb.",
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			return "", &mcp.Nachgeladen{
				Ergebnis: "[]",
				Grund:    mcp.Nachladefehler{Code: "validierung", Meldung: "klemmt"},
			}
		},
	}})

	var aus strings.Builder

	if fehler := server.Bediene(t.Context(), strings.NewReader(
		`{"jsonrpc":"2.0","id":1,"method":"tools/call",`+
			`"params":{"name":"schreiben.probe","arguments":{}}}`+"\n"), &aus); fehler != nil {
		t.Fatalf("bedienen: %v", fehler)
	}

	ergebnis := abbildung(t, entpacke(t, strings.TrimSpace(aus.String()))["result"])
	teile, _ := ergebnis["content"].([]any) //nolint:errcheck // geprüft darunter

	var satz string
	if len(teile) == 2 {
		satz, _ = abbildung(t, teile[1])["text"].(string) //nolint:errcheck // geprüft darunter
	}

	switch {
	case ergebnis["isError"] != false:
		t.Fatalf("als Fehler gemeldet: %v", ergebnis)
	case len(teile) != 2:
		t.Fatalf("es sind %d Teile: %v", len(teile), ergebnis)
	case abbildung(t, teile[0])["text"] != "[]":
		t.Errorf("der erste Teil ist %v", teile[0])
	case !strings.Contains(satz, "klemmt"):
		t.Errorf("der zweite Teil ist %v", teile[1])
	}
}

func TestDasPlatzwerkzeugGibtDenNachladefehlerWeiter(t *testing.T) {
	t.Parallel()

	probe := &platzprobe{nachladen: &mcp.Nachladefehler{Code: "validierung", Meldung: "klemmt"}}

	_, fehler := rufePlatz(t, "schreiben.platz.freigeben", probe, `{"posten": "vorsicherungen"}`)

	var nachgeladen *mcp.Nachgeladen
	if !errors.As(fehler, &nachgeladen) {
		t.Fatalf("kein Nachgeladen: %v", fehler)
	}

	if nachgeladen.Grund.Meldung != "klemmt" || !strings.HasPrefix(nachgeladen.Ergebnis, "[") ||
		!strings.Contains(nachgeladen.Error(), "klemmt") || probe.geraeumt != "vorsicherungen" {
		t.Errorf("nachgeladen: %+v (%s)", nachgeladen, nachgeladen.Error())
	}
}
