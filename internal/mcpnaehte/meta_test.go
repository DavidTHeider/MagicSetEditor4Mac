package mcpnaehte

// **Die Meta-Auskunft des MCP-Servers reicht den Bau durch** — ENG-06.AK-04
// (#1731):
// dieselben Angaben wie an Kommandozeile und Dienst, aus derselben Quelle.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestDieMetaauskunftReichtBauUndVertragDurch(t *testing.T) {
	t.Parallel()

	// **Jedes Feld mit einem eigenen Wert** — ein Testlauf hat keinen Stand,
	// und eine vergessene Zeile fiele bei leeren Feldern nicht auf.
	gebildet := metaAus(app.MetaAntwort{
		Fassung: "0.6.0", Plattform: "darwin/arm64", Go: "go1.27.1", Flaeche: "mcp",
		Bau:             app.Bau{Stand: "0123abcd", Gebaut: "2026-09-23T00:00:00Z", Veraendert: true},
		Bindungsvertrag: "a1b2c3d4e5f6",
	})

	if erwartet := (mcp.Meta{
		Fassung: "0.6.0", Plattform: "darwin/arm64", Go: "go1.27.1", Flaeche: "mcp",
		Stand: "0123abcd", Gebaut: "2026-09-23T00:00:00Z", Veraendert: true,
		Bindungsvertrag: "a1b2c3d4e5f6",
	}); gebildet != erwartet {
		t.Errorf("gebildet %+v, erwartet %+v", gebildet, erwartet)
	}

	if meta, fehler := (metaauskunft{}).Meta(t.Context()); fehler != nil ||
		meta.Bindungsvertrag != app.Vertragsabdruck() {
		t.Errorf("Meta = %+v, %v", meta, fehler)
	}
}
