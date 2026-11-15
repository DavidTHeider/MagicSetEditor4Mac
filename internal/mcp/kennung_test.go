package mcp_test

// **Der Griff ins Protokoll am Fehlertext** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// scheiterndesWerkzeug gibt immer denselben fachlichen Fehler.
func scheiterndesWerkzeug() mcp.Werkzeug {
	return mcp.Werkzeug{
		Name: "lesen.probe", Art: mcp.Lesend, Idempotent: true,
		Erklaerung: "Scheitert.",
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			return "", errors.New("das ging schief")
		},
	}
}

// **Mit Naht steht die Kennung am Fehlertext.** Sie ist der Griff, mit dem
// jemand im Protokoll nachsieht.
func TestEinFehlerTraegtDieVorgangskennung(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer([]mcp.Werkzeug{scheiterndesWerkzeug()}).
		MitKennung(func(context.Context) string { return "0a1b2c3d" })

	if text := ruftext(t, server); !strings.Contains(text, "(Vorgang: 0a1b2c3d)") {
		t.Errorf("Antwort %q trägt keine Kennung", text)
	}
}

// **Ohne Naht steht keine da** — und das ist besser als eine erfundene: sie
// zeigte auf eine Protokollzeile, die es nicht gibt.
func TestOhneNahtStehtKeineKennungDa(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer([]mcp.Werkzeug{scheiterndesWerkzeug()})

	if text := ruftext(t, server); strings.Contains(text, "Vorgang:") {
		t.Errorf("Antwort %q trägt eine Kennung", text)
	}
}

// **Und eine leere Kennung hängt nichts an.** Ein Ruf, der keine Grenze
// durchlaufen hat, bekommt keinen Griff — kein „(Vorgang: )".
func TestEineLeereKennungHaengtNichtsAn(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer([]mcp.Werkzeug{scheiterndesWerkzeug()}).
		MitKennung(func(context.Context) string { return "" })

	if text := ruftext(t, server); strings.Contains(text, "Vorgang:") {
		t.Errorf("Antwort %q trägt eine leere Kennung", text)
	}
}

// ruftext ruft das scheiternde Werkzeug über den Strom und gibt die Antwort.
//
// **Über `Bediene` und nicht über `Werkzeug.Ruf`**: die Kennung hängt der
// **Server** an, nicht das Werkzeug — ein Test am Werkzeug ginge daran vorbei.
func ruftext(t *testing.T, server *mcp.Server) string {
	t.Helper()

	var aus strings.Builder

	zeile := `{"jsonrpc":"2.0","id":1,"method":"tools/call",` +
		`"params":{"name":"lesen.probe","arguments":{}}}`

	if fehler := server.Bediene(
		t.Context(), strings.NewReader(zeile+"\n"), &aus); fehler != nil {
		t.Fatalf("bedienen: %v", fehler)
	}

	return aus.String()
}
