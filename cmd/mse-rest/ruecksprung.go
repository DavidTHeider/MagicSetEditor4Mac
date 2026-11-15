package main

// **Der Rücksprung am Dienst** — Werkstück O7 von Phase 5.2
// (#1431).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// ruecksprungauskunft reicht die vier Wege an den Dienst.
type ruecksprungauskunft struct {
	db  *sql.DB
	ort string
}

func (r ruecksprungauskunft) Ruecksprungarten(ctx context.Context) (any, error) {
	antwort := app.RuecksprungUeber(r.db, r.ort).Ruecksprungarten(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (r ruecksprungauskunft) Ruecksprunglage(ctx context.Context, pfad, kennwort string) (any, error) {
	antwort := app.RuecksprungUeber(r.db, r.ort).Ruecksprunglage(ctx, pfad, kennwort)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (r ruecksprungauskunft) Rueckkehren(ctx context.Context) (any, error) {
	antwort := app.RuecksprungUeber(r.db, r.ort).Rueckkehren(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (r ruecksprungauskunft) RueckkehrQuittieren(ctx context.Context, kennung int64) error {
	antwort := app.RuecksprungUeber(r.db, r.ort).RueckkehrQuittieren(ctx, kennung)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}
