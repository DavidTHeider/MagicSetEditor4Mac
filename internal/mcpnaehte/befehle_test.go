package mcpnaehte

import (
	"encoding/json"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Dieselbe Liste wie der Server** — Werkstück U12 von Phase 5.1 (#1423).
func TestDieBefehleSindDieWerkzeugeDesServers(t *testing.T) {
	db := frischerBestand(t)

	befehle, err := Befehle(db, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	if len(befehle) < 150 {
		t.Fatalf("nur %d Befehle", len(befehle))
	}

	// **Jeder Befehl hat einen Verweis in der Fähigkeitentabelle** — die
	// Vollständigkeit des Katalogs gegen die Registrierung.
	verweise := app.Befehlsverweise()

	var gelesen bool

	for _, befehl := range befehle {
		if _, da := verweise[befehl.Name]; !da {
			t.Errorf("%s hat keinen Verweis in FAEHIGKEITEN.md", befehl.Name)
		}

		if befehl.Name == "lesen.sets" {
			gelesen = true

			if befehl.Schreibend || !befehl.Idempotent || befehl.Ruf == nil {
				t.Errorf("lesen.sets: %+v", befehl)
			}

			// Der Ruf antwortet wie über MCP: mit JSON.
			text, err := befehl.Ruf(probekontext(t), json.RawMessage(`{}`))
			if err != nil || !json.Valid([]byte(text)) {
				t.Errorf("lesen.sets ruft: %q %v", text, err)
			}
		}
	}

	if !gelesen {
		t.Error("lesen.sets fehlt")
	}
}
