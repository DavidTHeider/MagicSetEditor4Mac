package mcp_test

// **Ein gescheiterter Ruf trägt seinen Code auch maschinenlesbar** — F03
// (#2342).
//
// Bis hierher stand die Klasse eines Fehlers nur im Satz; ein Gegenüber, das
// nach ihr verzweigen wollte, musste ihn lesen. Jetzt steht sie zusätzlich in
// `structuredContent.fehler` — in der Gestalt, die die Wurzel einsetzt
// (`internal/app`, dieselbe wie im Fenster und im Dienst). **Der Satz bleibt.**

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// rufergebnis ruft das scheiternde Werkzeug und gibt das Ergebnis.
func rufergebnis(t *testing.T, server *mcp.Server) map[string]any {
	t.Helper()

	var aus strings.Builder

	zeile := `{"jsonrpc":"2.0","id":1,"method":"tools/call",` +
		`"params":{"name":"lesen.probe","arguments":{}}}`

	if fehler := server.Bediene(t.Context(), strings.NewReader(zeile+"\n"), &aus); fehler != nil {
		t.Fatalf("bedienen: %v", fehler)
	}

	var antwort struct {
		Result map[string]any `json:"result"`
	}

	if fehler := json.Unmarshal([]byte(aus.String()), &antwort); fehler != nil {
		t.Fatalf("Antwort lesen: %v\n%s", fehler, aus.String())
	}

	return antwort.Result
}

func TestEinFehlerTraegtSeineFormStrukturiert(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer([]mcp.Werkzeug{scheiterndesWerkzeug()}).
		MitFehlerform(func(_ context.Context, fehler error) any {
			return map[string]string{"code": "unbekannt", "grund": fehler.Error()}
		})

	ergebnis := rufergebnis(t, server)

	strukturiert, _ := ergebnis["structuredContent"].(map[string]any) //nolint:errcheck // geprüft darunter
	form, _ := strukturiert["fehler"].(map[string]any)                //nolint:errcheck // geprüft darunter

	if form["code"] != "unbekannt" || form["grund"] != "das ging schief" {
		t.Errorf("structuredContent %v trägt die Form nicht", ergebnis["structuredContent"])
	}

	if ergebnis["isError"] != true {
		t.Errorf("isError %v", ergebnis["isError"])
	}
}

func TestOhneFehlerformBleibtDasErgebnisWieEsWar(t *testing.T) {
	t.Parallel()

	ergebnis := rufergebnis(t, mcp.NeuerServer([]mcp.Werkzeug{scheiterndesWerkzeug()}))

	if _, da := ergebnis["structuredContent"]; da {
		t.Errorf("ohne Fehlerform steht structuredContent da: %v", ergebnis)
	}
}
