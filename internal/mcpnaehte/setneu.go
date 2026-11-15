// **Ein neues, leeres Set über MCP** —
// #891.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// setanleger legt ein leeres Set an (#891).
//
// **Über `app.Einfuhrnaht`**, wie das Einfahren daneben: was zusammengesetzt
// wird, gehört an eine Stelle (§2).
type setanleger struct{ db *sql.DB }

func (s setanleger) SetAnlegen(
	ctx context.Context, name, spiel, stil string,
) (int64, error) {
	return app.Einfuhrnaht(s.db).Neu(ctx, name, spiel, stil)
}
