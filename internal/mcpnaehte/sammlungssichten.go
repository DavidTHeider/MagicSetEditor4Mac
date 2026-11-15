// **Die Sammlungssichten des MCP-Servers** —
// #1292.
//
// **Ohne Wache** (0044): alle vier lesen nur.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// sammlungssichten erfüllt `mcp.Sammlungssichten`.
type sammlungssichten struct{ db *sql.DB }

func (s sammlungssichten) Besitz(ctx context.Context, printing string) (any, error) {
	return app.SichtenUeber(s.db).Besitz(ctx, printing)
}

func (s sammlungssichten) Lagerorte(ctx context.Context) (any, error) {
	return app.SichtenUeber(s.db).Lagerorte(ctx)
}

func (s sammlungssichten) Dubletten(ctx context.Context, behalten int) (any, error) {
	return app.SichtenUeber(s.db).Dubletten(ctx, behalten)
}

func (s sammlungssichten) Wuensche(ctx context.Context) (any, error) {
	return app.SichtenUeber(s.db).Wuensche(ctx)
}
