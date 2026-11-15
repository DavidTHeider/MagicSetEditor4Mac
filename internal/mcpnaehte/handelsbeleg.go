package mcpnaehte

// **Der Beleg eines Handels für MCP** — gelesen, ohne einen Schritt
// (Hüllenprobe `nat04handel`,
// #1672).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// Handelsbeleg gibt Stand und Beleg eines Vorgangs.
func (h handelsstand) Handelsbeleg(ctx context.Context, vorgang string) (any, error) {
	antwort := app.PeerflaecheUeber(h.db).Handelsbeleg(ctx, vorgang)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
