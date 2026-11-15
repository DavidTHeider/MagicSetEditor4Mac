// **Auszählen nach einer erklärten Dimension** — Werkstück C von Phase 4.11
// (#1338).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// auszaehlung erfüllt `mcp.Auszaehlung`.
type auszaehlung struct{ db *sql.DB }

func (a auszaehlung) Angebot(ctx context.Context, setID int64) (any, error) {
	naht, fehler := app.AuszaehlungUeber(a.db)
	if fehler != nil {
		return nil, fehler
	}

	return naht.Angebot(ctx, setID)
}

func (a auszaehlung) Auszaehlen(ctx context.Context, setID int64, name string) (any, error) {
	naht, fehler := app.AuszaehlungUeber(a.db)
	if fehler != nil {
		return nil, fehler
	}

	return naht.Auszaehlen(ctx, setID, name)
}
