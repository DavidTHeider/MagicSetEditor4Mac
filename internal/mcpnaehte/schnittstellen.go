package mcpnaehte

// **Die Schnittstellen über MCP** — OPS-09 (#1907), nur lesend.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// schnittstellen liest über eine Anwendung ohne Bestand — der Stand steht im
// Zuhause dieses Prozesses, nicht im Bestand.
type schnittstellen struct{}

func (schnittstellen) Schnittstellen(ctx context.Context) ([]mcp.Schnittstelle, error) {
	antwort := app.NeueApp("").Schnittstellen(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	stand := make([]mcp.Schnittstelle, 0, len(antwort.Schnittstellen))
	for _, eine := range antwort.Schnittstellen {
		stand = append(stand, mcp.Schnittstelle{
			Name: eine.Name, Stand: eine.Stand, Datei: eine.Datei, Art: eine.Art, Zeilen: eine.Zeilen,
		})
	}

	return stand, nil
}
