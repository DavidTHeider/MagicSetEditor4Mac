package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// zusatz reicht die Zusatzfelder einer Karte an MCP (#2298) — über
// [app.Zusatzweg], denselben Dienst wie im Fenster.
type zusatz struct{ db *sql.DB }

func (z zusatz) Zusatzfelder(ctx context.Context, setID, karteID int64) ([]mcp.Zusatzfeld, error) {
	weg, err := app.ZusatznahtUeber(z.db, domain.UrheberServer)
	if err != nil {
		return nil, err
	}

	felder, err := weg.Felder(ctx, setID, karteID)
	if err != nil {
		return nil, err
	}

	umgesetzt := make([]mcp.Zusatzfeld, 0, len(felder))
	for i := range felder {
		umgesetzt = append(umgesetzt, mcp.Zusatzfeld{
			Name: felder[i].Form.Name, Art: felder[i].Form.Art,
			Beschreibung: felder[i].Form.Beschreibung, Wahlen: felder[i].Form.Wahlen,
			Wert: felder[i].Wert, Gespeichert: felder[i].Gespeichert,
		})
	}

	return umgesetzt, nil
}

func (z zusatz) ZusatzwertSetzen(
	ctx context.Context, setID, karteID int64, feld, wert string,
) (mcp.Zusatzvorwert, error) {
	weg, err := app.ZusatznahtUeber(z.db, domain.UrheberServer)
	if err != nil {
		return mcp.Zusatzvorwert{}, err
	}

	vorher, err := weg.Setzen(ctx, setID, karteID, feld, wert)

	return mcp.Zusatzvorwert(vorher), err
}
