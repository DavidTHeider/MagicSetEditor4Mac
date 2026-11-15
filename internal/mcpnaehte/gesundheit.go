package mcpnaehte

// **Der Betriebszustand des MCP-Servers** — Werkstück O9 von Phase 5.2
// (#1433).
//
// **Ohne Fremdschau**: der MCP-Server fragt keinen fremden Dienst von sich
// aus, und eine Auskunft, die dabei ins Netz ginge, machte aus „lebst du" eine
// Netzanfrage (§16, Regel 3). Die Liste bleibt leer — und leer heißt hier
// „nichts bekannt", nicht „alle aus".

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// gesundheitsauskunft reicht den Betriebszustand an das Werkzeug.
type gesundheitsauskunft struct{ db *sql.DB }

func (g gesundheitsauskunft) Gesundheit(ctx context.Context) (mcp.Gesundheit, error) {
	zustand := app.GesundheitsnahtUeber(g.db).Zustand(ctx)

	antwort := mcp.Gesundheit{
		Lebt:   zustand.Lebt,
		Bereit: zustand.Bereit(),
		Grund:  string(zustand.Grund),
		Fremde: make([]mcp.Fremdzustand, 0, len(zustand.Fremde)),
	}

	for _, fremd := range zustand.Fremde {
		antwort.Fremde = append(antwort.Fremde, mcp.Fremdzustand{
			Dienst:     string(fremd.Dienst),
			Erreichbar: fremd.Erreichbar,
			Grund:      fremd.Grund,
		})
	}

	return antwort, nil
}
