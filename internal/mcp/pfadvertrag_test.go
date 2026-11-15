package mcp_test

// **Ein Pfad an den Werkzeugserver geht durch den Pfadvertrag** —
// OPS-02.AK-04 (#1719).
// Dieselbe Regel wie beim Dienst: vor dem Werkzeug, an den Namen aus
// `domain.IstPfadfeld`. **Gesetzt wird er für den Transport über HTTP**; über
// stdio spricht nur, wer den Prozess selbst gestartet hat.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestEinPfadAusserhalbDesVertragsErreichtKeinWerkzeug(t *testing.T) {
	gerufen := 0
	gefragt := []string{}

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "schreiben.probe", Art: mcp.Schreibend, Erklaerung: "zählt",
		Schema: map[string]any{"type": "object"},
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			gerufen++

			return "gut", nil
		},
	}}).MitPfadvertrag(func(_ context.Context, pfad string) error {
		gefragt = append(gefragt, pfad)
		if strings.HasPrefix(pfad, "/erlaubt/") {
			return nil
		}

		return errors.New("liegt außerhalb dessen, was der Dienst anfassen darf")
	})

	ruf := func(argumente string) string {
		antwort := antworten(t, server,
			`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.probe","arguments":`+argumente+`}}`)
		if len(antwort) != 1 {
			t.Fatalf("%v", antwort)
		}

		return antwort[0]
	}

	if abgewiesen := ruf(`{"ziel":"/Users/ich/.ssh/x","name":"egal"}`); !strings.Contains(abgewiesen, "außerhalb") ||
		!strings.Contains(abgewiesen, `"isError":true`) {
		t.Errorf("ein fremder Pfad: %s", abgewiesen)
	}

	if gerufen != 0 {
		t.Fatal("das Werkzeug lief mit einem fremden Pfad")
	}

	// **Dass jemand zuhört**: derselbe Ruf mit einem erlaubten Pfad kommt an.
	if erlaubt := ruf(`{"pfad":"/erlaubt/probe.mse-set"}`); !strings.Contains(erlaubt, "gut") || gerufen != 1 {
		t.Errorf("ein erlaubter Pfad: %s", erlaubt)
	}

	if strings.Join(gefragt, "|") != "/Users/ich/.ssh/x|/erlaubt/probe.mse-set" {
		t.Errorf("gefragt wurde nach %v — »name« ist kein Pfad", gefragt)
	}
}

// **Ein Name im Paket ist kein Ort** — Befund 4 des Zweitblicks über Phase
// 6.4: `lesen.paketdatei` bekommt in `datei` einen Namen **im** Paket, und
// der Vertrag wies ihn als unvollständigen Pfad ab.
func TestEinNameImPaketGehtNichtDurchDenVertrag(t *testing.T) {
	gerufen := 0

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "lesen.probe", Art: mcp.Lesend, Idempotent: true, Erklaerung: "zählt",
		Schema: map[string]any{"type": "object"}, Namensfelder: []string{"datei"},
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			gerufen++

			return "gut", nil
		},
	}}).MitPfadvertrag(func(context.Context, string) error {
		return errors.New("liegt außerhalb dessen, was der Dienst anfassen darf")
	})

	antwort := antworten(t, server,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.probe","arguments":{"datei":"style"}}}`)
	if len(antwort) != 1 || !strings.Contains(antwort[0], "gut") || gerufen != 1 {
		t.Errorf("ein Name im Paket: %v", antwort)
	}

	// **Und ein Pfad daneben geht weiter durch ihn** — der Name nimmt nur
	// sein eigenes Feld aus.
	antwort = antworten(t, server,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"lesen.probe","arguments":{"datei":"style","pfad":"/x"}}}`)
	if len(antwort) != 1 || !strings.Contains(antwort[0], "außerhalb") || gerufen != 1 {
		t.Errorf("ein Pfad daneben: %v", antwort)
	}
}
