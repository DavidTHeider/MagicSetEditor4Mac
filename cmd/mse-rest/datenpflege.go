package main

// **Die Datenpflege am Dienst** — Werkstück O8 von Phase 5.2
// (#1432).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// pflegeauskunft reicht die zwei Wege an den Dienst.
type pflegeauskunft struct {
	db  *sql.DB
	ort string
}

func (p pflegeauskunft) Pflegevorschau(ctx context.Context) (any, error) {
	antwort := app.PflegeAnwendungUeber(p.db, p.ort).Pflegevorschau(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (p pflegeauskunft) Pflegen(ctx context.Context, erstellt string, wahl []rest.Pflegewahl) (
	any, error,
) {
	gewaehlt := make([]app.PflegewahlDTO, 0, len(wahl))
	for _, eine := range wahl {
		gewaehlt = append(gewaehlt, app.PflegewahlDTO{Art: eine.Art, Revision: eine.Revision})
	}

	antwort := app.PflegeAnwendungUeber(p.db, p.ort).Pflegen(ctx, erstellt, gewaehlt)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
