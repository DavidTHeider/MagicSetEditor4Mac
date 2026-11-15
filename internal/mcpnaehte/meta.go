package mcpnaehte

// **Die Meta-Auskunft des MCP-Servers** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// metaauskunft reicht Fassung und Plattform an das Werkzeug.
type metaauskunft struct{}

func (metaauskunft) Meta(ctx context.Context) (mcp.Meta, error) {
	return metaAus(app.NeueApp("").Meta(ctx)), nil
}

// metaAus bildet die Antwort der Anwendung auf die des Werkzeugs ab.
func metaAus(angaben app.MetaAntwort) mcp.Meta {
	return mcp.Meta{
		Fassung: angaben.Fassung, Plattform: angaben.Plattform,
		Go: angaben.Go, Flaeche: angaben.Flaeche,
		Stand: angaben.Bau.Stand, Gebaut: angaben.Bau.Gebaut,
		Veraendert: angaben.Bau.Veraendert, Bindungsvertrag: angaben.Bindungsvertrag,
	}
}
