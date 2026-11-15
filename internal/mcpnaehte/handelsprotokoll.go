package mcpnaehte

// **Das Protokoll eines Handels für MCP** — TRADE-03.AK-03
// (#1717).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// Handelsprotokoll nennt die Zustandswechsel eines Vorgangs.
func (h handelsstand) Handelsprotokoll(
	ctx context.Context, vorgang string,
) ([]mcp.Handelsprotokollzeile, error) {
	antwort := app.PeerflaecheUeber(h.db).HandelProtokoll(ctx, vorgang)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]mcp.Handelsprotokollzeile, 0, len(antwort.Zeilen))
	for _, zeile := range antwort.Zeilen {
		zeilen = append(zeilen, mcp.Handelsprotokollzeile(zeile))
	}

	return zeilen, nil
}
