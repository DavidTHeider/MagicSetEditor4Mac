// Die Stilwahl und die Stiloptionen für den Dienst
// (#1107).
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// stile führt die sechs Endpunkte an die Naht dieser Schicht.
type stile struct{ db *sql.DB }

func (s stile) Stile(ctx context.Context, setID int64) ([]rest.Stil, error) {
	weg, err := app.StilnahtUeber(s.db)
	if err != nil {
		return nil, err
	}

	angebote, lesefehler := weg.Stile(ctx, setID)
	if lesefehler != nil {
		return nil, lesefehler
	}

	umgesetzt := make([]rest.Stil, 0, len(angebote))
	for _, angebot := range angebote {
		umgesetzt = append(umgesetzt, rest.Stil{
			Paket: angebot.Paket, Spiel: angebot.Spiel, Version: angebot.Version,
			EigenesSpiel: angebot.EigenesSpiel, Gewaehlt: angebot.Gewaehlt,
		})
	}

	return umgesetzt, nil
}

func (s stile) Folgen(ctx context.Context, setID int64, paket string) (
	rest.Wechselfolgen, error,
) {
	weg, err := app.StilnahtUeber(s.db)
	if err != nil {
		return rest.Wechselfolgen{}, err
	}

	folgen, lesefehler := weg.Folgen(ctx, setID, paket)
	if lesefehler != nil {
		return rest.Wechselfolgen{}, lesefehler
	}

	return rest.Wechselfolgen{
		UngedeuteteFelder: folgen.UngedeuteteFelder,
		VerloreneOptionen: folgen.VerloreneOptionen,
		Spielwechsel:      folgen.Spielwechsel,
	}, nil
}

func (s stile) Wechseln(ctx context.Context, setID int64, paket string) error {
	weg, err := app.StilnahtUeber(s.db)
	if err != nil {
		return err
	}

	return weg.Wechseln(ctx, setID, paket)
}

func (s stile) Optionen(ctx context.Context, setID int64) ([]rest.Stiloption, error) {
	weg, err := app.StilnahtUeber(s.db)
	if err != nil {
		return nil, err
	}

	gelesen, lesefehler := weg.Optionen(ctx, setID)
	if lesefehler != nil {
		return nil, lesefehler
	}

	optionen := make([]rest.Stiloption, 0, len(gelesen))
	for i := range gelesen {
		option := &gelesen[i]
		optionen = append(optionen, rest.Stiloption{
			Name: option.Name, Art: option.Art, Beschreibung: option.Beschreibung,
			Wert: option.Wert, Gesetzt: option.Gesetzt,
			HatVorgabe: option.HatVorgabe, Vorgabe: option.Vorgabe,
		})
	}

	return optionen, nil
}

func (s stile) OptionSetzen(ctx context.Context, setID int64, feld, wert string) error {
	weg, err := app.StilnahtUeber(s.db)
	if err != nil {
		return err
	}

	return weg.OptionSetzen(ctx, setID, feld, wert)
}

func (s stile) OptionZuruecksetzen(ctx context.Context, setID int64, feld string) error {
	weg, err := app.StilnahtUeber(s.db)
	if err != nil {
		return err
	}

	return weg.OptionZuruecksetzen(ctx, setID, feld)
}
