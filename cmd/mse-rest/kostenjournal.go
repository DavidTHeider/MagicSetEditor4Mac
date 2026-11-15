package main

// **Das Kostenjournal am Dienst** — Werkstück O12 von Phase 5.2
// (#1439).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kostenauskunft reicht die drei Wege an den Dienst.
type kostenauskunft struct{ db *sql.DB }

func (k kostenauskunft) Kostenstand(ctx context.Context, konto, zeitraum string) (
	any, error,
) {
	antwort := app.KostenAnwendungUeber(k.db).Kostenstand(ctx, konto, zeitraum)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (k kostenauskunft) Kostenposten(ctx context.Context, konto, zeitraum string) (
	any, error,
) {
	antwort := app.KostenAnwendungUeber(k.db).Kostenposten(ctx, konto, zeitraum)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}

func (k kostenauskunft) KostenbudgetSetzen(
	ctx context.Context, konto, zeitraum string, wert int64, waehrung string,
) (any, error) {
	antwort := app.KostenAnwendungUeber(k.db).
		KostenbudgetSetzen(ctx, konto, zeitraum, wert, waehrung)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
