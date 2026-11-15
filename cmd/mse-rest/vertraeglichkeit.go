package main

// **Die Verträglichkeit am Dienst** — Werkstück O6 von Phase 5.2
// (#1430).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// vertraeglichkeitsauskunft reicht das Urteil an den Dienst.
type vertraeglichkeitsauskunft struct{ db *sql.DB }

func (v vertraeglichkeitsauskunft) Vertraeglichkeit(ctx context.Context) (any, error) {
	antwort := app.VertraeglichkeitUeber(v.db).Vertraeglichkeit(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
