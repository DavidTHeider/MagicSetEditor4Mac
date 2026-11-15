package mcpnaehte

// **Der Preisanhalt eines Handels für MCP** — TRADE-02b
// (#1951).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Handelsanhalt gibt, was die Seiten eines Vorgangs ungefähr wert sind.
func (h handelsstand) Handelsanhalt(ctx context.Context, vorgang string) (any, error) {
	antwort := app.PeerflaecheUeber(h.db).HandelPreisanhaltVon(ctx, vorgang)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort.Anhalt, nil
}
