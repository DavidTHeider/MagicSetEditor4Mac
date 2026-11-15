package main

// **Die Zwischenablage für den Dienst** — Werkstück 1 von Phase 3.7
// (#1168).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// fremdablage reicht das Kopieren und Einfügen an die Naht durch.
type fremdablage struct{ db *sql.DB }

// AblageKopieren gibt die genannten Karten als Text.
func (f fremdablage) AblageKopieren(
	ctx context.Context, setID int64, karteIDs []int64,
) (string, error) {
	naht, err := app.Ablageflaechennaht(f.db, domain.UrheberDienst)
	if err != nil {
		return "", err
	}

	return naht.Kopieren(ctx, setID, karteIDs)
}

// AblageAusschneiden gibt sie als Text und entfernt sie aus dem Set.
func (f fremdablage) AblageAusschneiden(
	ctx context.Context, setID int64, karteIDs []int64,
) (string, error) {
	naht, err := app.Ablageflaechennaht(f.db, domain.UrheberDienst)
	if err != nil {
		return "", err
	}

	return naht.Ausschneiden(ctx, setID, karteIDs)
}

// AblageEinfuegen legt die Karten aus einem Text im Set an.
func (f fremdablage) AblageEinfuegen(
	ctx context.Context, setID int64, inhalt string,
) (karten, befunde []string, fehler error) {
	naht, err := app.Ablageflaechennaht(f.db, domain.UrheberDienst)
	if err != nil {
		return nil, nil, err
	}

	return naht.Einfuegen(ctx, setID, inhalt)
}
