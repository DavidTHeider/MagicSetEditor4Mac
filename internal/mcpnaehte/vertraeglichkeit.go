package mcpnaehte

// **Die Verträglichkeit über MCP** — Werkstück O6 von Phase 5.2
// (#1430).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// vertraeglichkeitsauskunft reicht das Urteil an das Werkzeug.
type vertraeglichkeitsauskunft struct{ db *sql.DB }

func (v vertraeglichkeitsauskunft) Vertraeglichkeit(
	ctx context.Context,
) (mcp.Vertraeglichkeit, error) {
	antwort := app.VertraeglichkeitUeber(v.db).Vertraeglichkeit(ctx)
	if antwort.Fehler != nil {
		return mcp.Vertraeglichkeit{}, antwort.Fehler.AlsFehler()
	}

	urteil := mcp.Vertraeglichkeit{
		Fassung: antwort.Fassung, Urteil: antwort.Urteil, Bereit: antwort.Bereit,
		Hindernisse: make([]mcp.Hindernis, 0, len(antwort.Hindernisse)),
	}

	for _, hindernis := range antwort.Hindernisse {
		urteil.Hindernisse = append(urteil.Hindernisse,
			mcp.Hindernis{Art: hindernis.Art, Stelle: hindernis.Stelle})
	}

	return urteil, nil
}
