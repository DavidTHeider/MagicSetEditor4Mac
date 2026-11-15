package main

// **Suchen und Ersetzen für den Dienst** — Werkstück 2 von Phase 3.7
// (#1168).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// fremdersetzung reicht das Ersetzen an die Naht durch.
type fremdersetzung struct{ db *sql.DB }

// ErsetzungVorschau sagt, was sich ändern würde.
func (f fremdersetzung) ErsetzungVorschau(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	naht, err := app.Ersetzungsflaechennaht(f.db, domain.UrheberDienst)
	if err != nil {
		return domain.Ersetzungsvorschau{}, err
	}

	return naht.Vorschau(ctx, setID, frage)
}

// ErsetzungAusfuehren ersetzt und gibt dieselbe Übersicht zurück.
func (f fremdersetzung) ErsetzungAusfuehren(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	naht, err := app.Ersetzungsflaechennaht(f.db, domain.UrheberDienst)
	if err != nil {
		return domain.Ersetzungsvorschau{}, err
	}

	return naht.Ausfuehren(ctx, setID, frage)
}
