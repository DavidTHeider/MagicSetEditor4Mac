package mcpnaehte

// **Die Schalter über MCP** — OPS-10 (#1908), nur lesend.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// schalter liest die Schalter des gezeigten Profils.
type schalter struct{ db *sql.DB }

func (s schalter) Schalter(ctx context.Context) ([]mcp.Schalter, error) {
	antwort := app.SchalterwacheUeber(s.db).Stand(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	stand := make([]mcp.Schalter, 0, len(antwort.Schalter))
	for _, einer := range antwort.Schalter {
		stand = append(stand, mcp.Schalter(einer))
	}

	return stand, nil
}
