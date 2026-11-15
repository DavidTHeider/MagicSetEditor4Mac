package main

// **Der Betriebszustand des Dienstes** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// gesundheitsauskunft reicht den Betriebszustand an den Dienst.
//
// **Sie gibt „bereit" getrennt heraus**, weil `GET /bereit` daraus seinen
// Statuscode macht — und eine Fläche soll die Regel nicht nachbauen (§16).
type gesundheitsauskunft struct{ db *sql.DB }

func (g gesundheitsauskunft) Gesundheit(ctx context.Context) (
	zustand any, bereit bool, fehler error,
) {
	anwendung := app.NeueApp("")
	anwendung.MitGesundheit(app.GesundheitsnahtUeber(g.db))

	antwort := anwendung.Gesundheit(ctx)

	return antwort, antwort.Bereit, nil
}
