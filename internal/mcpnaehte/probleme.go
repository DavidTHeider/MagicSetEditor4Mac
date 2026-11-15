package mcpnaehte

// **Die Probleme eines ganzen Sets** — PERF02
// (#2352):
// ein dünner Adapter auf `app.Problemweg`, der Weg über dem Zeichenwerk des
// Bestands — **es bleibt je Bestand**, und mit ihm der letzte Stand.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type probleme struct{ db *sql.DB }

func (p probleme) Problemstand(ctx context.Context, setID int64) (mcp.Problemstand, error) {
	weg, err := app.ProblemwegUeber(p.db)
	if err != nil {
		return mcp.Problemstand{}, err
	}

	stand, err := weg.Stand(ctx, setID)

	return problemstandFuerMCP(stand), err
}

func (p probleme) ProblemePruefen(ctx context.Context, setID int64) (mcp.Problemstand, error) {
	weg, err := app.ProblemwegUeber(p.db)
	if err != nil {
		return mcp.Problemstand{}, err
	}

	stand, err := weg.Pruefen(ctx, setID, nil)

	return problemstandFuerMCP(stand), err
}

// problemstandFuerMCP übersetzt den Stand in die Gestalt des Servers.
func problemstandFuerMCP(stand app.ProblemstandDTO) mcp.Problemstand {
	karten := make([]mcp.Kartenproblem, 0, len(stand.Karten))
	for _, karte := range stand.Karten {
		karten = append(karten, mcp.Kartenproblem{KarteID: karte.KarteID, Befunde: karte.Befunde})
	}

	return mcp.Problemstand{
		Urteil: stand.Urteil, Zahl: stand.Zahl, Sicher: stand.Sicher, Geprueft: stand.Geprueft,
		Gesamt: stand.Gesamt, Beendet: stand.Beendet, Alter: stand.Alter, Karten: karten,
	}
}
