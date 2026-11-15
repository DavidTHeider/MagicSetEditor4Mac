package mcpnaehte

// **Die Metriken des MCP-Servers** — Werkstück O9 von Phase 5.2
// (#1433),
// seit OPS-05 (#1722)
// mit den Zahlen der Verbindung.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// metrikauskunft reicht den Stand an das Werkzeug — gezählt im
// Arbeitsspeicher **dieses** Prozesses, und die Verbindung ist die, über die
// er arbeitet.
type metrikauskunft struct{ db *sql.DB }

func (m metrikauskunft) Metriken(ctx context.Context) (any, error) {
	return app.MetrikenUeber(ctx, m.db), nil
}
