// **Ein Bild als Symbol nachzeichnen** — Werkstück F von Phase 4.11
// (#1340).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// symbolnachzeichnung erfüllt `mcp.Symbolnachzeichnung`.
type symbolnachzeichnung struct{ db *sql.DB }

func (s symbolnachzeichnung) SymbolNachzeichnen(
	ctx context.Context, setID int64, bildpfad string,
) (any, error) {
	naht, fehler := app.SymbolnachzeichnungUeber(s.db)
	if fehler != nil {
		return nil, fehler
	}

	return naht.Nachzeichnen(ctx, setID, bildpfad)
}
