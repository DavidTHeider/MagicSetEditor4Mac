package main

// **Die Kartenliste für den Dienst** — Werkstück 3 von Phase 3.7
// (#1168).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// fremdliste reicht die Kartenliste an die Naht durch.
type fremdliste struct{ db *sql.DB }

// Listenaufbau gibt Spalten, Aufschriften, Zeilen und die Ordnung.
func (f fremdliste) Listenaufbau(ctx context.Context, setID int64) (
	domain.Kartenlistenaufbau, error,
) {
	naht, err := app.Listenflaechennaht(f.db)
	if err != nil {
		return domain.Kartenlistenaufbau{}, err
	}

	return naht.Aufbau(ctx, setID)
}

// KartenFiltern legt den Schnellfilter über die Karten eines Sets (U5, #1416).
func (f fremdliste) KartenFiltern(ctx context.Context, setID int64, anfrage string) ([]int64, error) {
	naht, err := app.Listenflaechennaht(f.db)
	if err != nil {
		return nil, err
	}

	return naht.Filtern(ctx, setID, anfrage)
}

// ListenspaltenSetzen schreibt die Wahl fort.
func (f fremdliste) ListenspaltenSetzen(
	ctx context.Context, setID int64, wahl []domain.Spaltenwahl,
) error {
	naht, err := app.Listenflaechennaht(f.db)
	if err != nil {
		return err
	}

	return naht.SpaltenSetzen(ctx, setID, wahl)
}

// ListensortierungSetzen schreibt die Ordnung fort.
func (f fremdliste) ListensortierungSetzen(
	ctx context.Context, setID int64, ordnung domain.Kartenordnung,
) error {
	naht, err := app.Listenflaechennaht(f.db)
	if err != nil {
		return err
	}

	return naht.SortierungSetzen(ctx, setID, ordnung)
}
