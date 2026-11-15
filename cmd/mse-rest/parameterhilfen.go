// **Die Parameterhilfen der Schlüsselwörter** — Werkstück D von Phase 4.11
// (#1342).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// parameterhilfen führt das Werkzeug an die Naht dieser Schicht.
type parameterhilfen struct{ db *sql.DB }

func (p parameterhilfen) Parameterhilfen(
	ctx context.Context, setID, wortID int64,
) (any, error) {
	pflege, err := app.SchluesselwortpflegeAls(p.db, domain.UrheberServer)
	if err != nil {
		return nil, err
	}

	return pflege.Parameterhilfen(ctx, setID, wortID)
}
