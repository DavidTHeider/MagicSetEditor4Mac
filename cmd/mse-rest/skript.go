// Die Skriptkonsole für den Dienst
// (#1070).
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// skriptkonsole führt `POST /karten/{id}/skript` an die Naht dieser Schicht.
type skriptkonsole struct{ db *sql.DB }

// Auswerten rechnet einen Ausdruck im Kontext einer Karte aus.
func (s skriptkonsole) Auswerten(ctx context.Context, karteID int64, ausdruck string) (
	rest.Skriptergebnis, error,
) {
	naht, err := app.SkriptnahtUeber(s.db)
	if err != nil {
		return rest.Skriptergebnis{}, err
	}

	ergebnis, auswertfehler := naht.Auswerten(ctx, karteID, ausdruck)
	if auswertfehler != nil {
		return rest.Skriptergebnis{}, auswertfehler
	}

	return rest.Skriptergebnis{
		Quelltext: ergebnis.Quelltext,
		Text:      ergebnis.Text,
		Art:       ergebnis.Art,
		Befunde:   ergebnis.Befunde,
	}, nil
}

// Namen zählt auf, was ein Ausdruck an dieser Karte nennen darf
// (#1164).
func (s skriptkonsole) Namen(ctx context.Context, karteID int64) ([]string, error) {
	naht, err := app.SkriptnahtUeber(s.db)
	if err != nil {
		return nil, err
	}

	return naht.Namen(ctx, karteID)
}
