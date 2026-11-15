// **Karten aus einem Skript** — Werkstück B von Phase 4.11
// (#1337).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kartenskripte erfüllt `mcp.Kartenskripte`.
type kartenskripte struct{ db *sql.DB }

func (k kartenskripte) Angebot(ctx context.Context, setID int64) (any, error) {
	naht, fehler := app.KartenskripteUeber(k.db)
	if fehler != nil {
		return nil, fehler
	}

	return naht.Angebot(ctx, setID)
}

func (k kartenskripte) Ausfuehren(ctx context.Context, setID int64, name string) (any, error) {
	naht, fehler := app.KartenskripteUeber(k.db)
	if fehler != nil {
		return nil, fehler
	}

	return naht.Ausfuehren(ctx, setID, name)
}
