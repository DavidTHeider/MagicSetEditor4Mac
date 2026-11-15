package main

// **Die Verweise für den Dienst** — Werkstück 2 von Phase 3.6
// (#1167).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// fremdverweise reicht die Verweise an die Naht durch.
type fremdverweise struct{ db *sql.DB }

// VerweisSetzen bezieht eine Karte auf einen Druck.
func (f fremdverweise) VerweisSetzen(
	ctx context.Context, karteID int64, printing string,
) (rest.Verweissicht, error) {
	naht, err := app.Verweisflaechennaht(f.db)
	if err != nil {
		return rest.Verweissicht{}, err
	}

	verweis, setzfehler := naht.Setzen(ctx, karteID, printing)
	if setzfehler != nil {
		return rest.Verweissicht{}, setzfehler
	}

	return alsVerweissicht(verweis), nil
}

// Verweis nennt den Druck, auf den eine Karte verweist.
func (f fremdverweise) Verweis(ctx context.Context, karteID int64) (
	*rest.Verweissicht, error,
) {
	naht, err := app.Verweisflaechennaht(f.db)
	if err != nil {
		return nil, err
	}

	verweis, lesefehler := naht.Fuer(ctx, karteID)
	if lesefehler != nil || verweis == nil {
		return nil, lesefehler
	}

	sicht := alsVerweissicht(*verweis)

	return &sicht, nil
}

// VerweisLoesen nimmt den Verweis zurück.
func (f fremdverweise) VerweisLoesen(ctx context.Context, karteID int64) error {
	naht, err := app.Verweisflaechennaht(f.db)
	if err != nil {
		return err
	}

	return naht.Loesen(ctx, karteID)
}

// Verweise nennt die Verweise aller Karten eines Sets.
func (f fremdverweise) Verweise(ctx context.Context, setID int64) (
	[]rest.Verweissicht, error,
) {
	naht, err := app.Verweisflaechennaht(f.db)
	if err != nil {
		return nil, err
	}

	verweise, lesefehler := naht.JeSet(ctx, setID)
	if lesefehler != nil {
		return nil, lesefehler
	}

	sichten := make([]rest.Verweissicht, 0, len(verweise))
	for _, verweis := range verweise {
		sichten = append(sichten, alsVerweissicht(verweis))
	}

	return sichten, nil
}

// alsVerweissicht übersetzt Feld für Feld — nicht über JSON.
func alsVerweissicht(verweis app.KartenverweisDTO) rest.Verweissicht {
	return rest.Verweissicht{
		KarteID: verweis.KarteID, Printing: verweis.Printing,
		DruckID: verweis.DruckID, Vorlaeufig: verweis.Vorlaeufig,
		Name: verweis.Name, Setcode: verweis.Setcode, Nummer: verweis.Nummer,
		Adresse: verweis.Adresse, Gesetzt: verweis.Gesetzt,
	}
}
