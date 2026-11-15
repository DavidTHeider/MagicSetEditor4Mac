// **Die Autoersetzungen verwalten** — Werkstück E von Phase 4.11
// (#1341).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// autoersetzungen führt die fünf Werkzeuge an die Naht dieser Schicht.
type autoersetzungen struct{ db *sql.DB }

func (a autoersetzungen) Autoersetzungen(ctx context.Context, setID int64) (any, error) {
	weg, fehler := app.AutoersetzungspflegeUeber(a.db)
	if fehler != nil {
		return nil, fehler
	}

	regeln, an, lesefehler := weg.Liste(ctx, setID)
	if lesefehler != nil {
		return nil, lesefehler
	}

	return map[string]any{"regeln": regeln, "an": an}, nil
}

func (a autoersetzungen) AutoersetzungSetzen(
	ctx context.Context, setID int64, muster, ersatz string, ganzesWort, aktiv bool,
) error {
	weg, fehler := app.AutoersetzungspflegeUeber(a.db)
	if fehler != nil {
		return fehler
	}

	return weg.Setzen(ctx, setID, app.AutoersetzungsregelDTO{
		Muster: muster, Ersatz: ersatz, GanzesWort: ganzesWort, Aktiv: aktiv,
	})
}

func (a autoersetzungen) AutoersetzungLoeschen(ctx context.Context, setID int64, muster string) error {
	weg, fehler := app.AutoersetzungspflegeUeber(a.db)
	if fehler != nil {
		return fehler
	}

	return weg.Loeschen(ctx, setID, muster)
}

func (a autoersetzungen) AutoersetzungenZuruecksetzen(ctx context.Context, setID int64) error {
	weg, fehler := app.AutoersetzungspflegeUeber(a.db)
	if fehler != nil {
		return fehler
	}

	return weg.Zuruecksetzen(ctx, setID)
}

func (a autoersetzungen) AutoersetzungSchalten(ctx context.Context, setID int64, an bool) error {
	weg, fehler := app.AutoersetzungspflegeUeber(a.db)
	if fehler != nil {
		return fehler
	}

	return weg.SchalterSetzen(ctx, setID, an)
}
