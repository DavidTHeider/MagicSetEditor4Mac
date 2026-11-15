// **Die Wege nach draußen am MCP-Server** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// **Dieselbe Naht wie an den anderen Flächen** (§16): welche Adresse zu
// welcher Karte gehört, entscheidet die Domäne, und nicht drei Übersetzungen
// nebeneinander.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// marktlinke erfüllt `mcp.Linkauskunft` über der Naht von `app`.
type marktlinke struct{ db *sql.DB }

func (m marktlinke) Marktlinks(
	ctx context.Context, printing, name string,
) ([]mcp.Marktlink, error) {
	antwort := app.PeerflaecheUeber(m.db).Marktlinks(ctx, printing, name)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	links := make([]mcp.Marktlink, 0, len(antwort.Links))

	for _, link := range antwort.Links {
		links = append(links, mcp.Marktlink{
			Anbieter: link.Anbieter, Art: link.Art,
			Herkunft: link.Herkunft, Adresse: link.Adresse,
		})
	}

	return links, nil
}
