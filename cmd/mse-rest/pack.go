// Das Zufallspack für den Dienst
// (#889).
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// pack führt die zwei Endpunkte an die Naht dieser Schicht.
type pack struct{ db *sql.DB }

func (p pack) Packarten(ctx context.Context, setID int64) (rest.Packangebot, error) {
	weg, err := app.PackUeber(p.db)
	if err != nil {
		return rest.Packangebot{}, err
	}

	antwort, lesefehler := weg.Arten(ctx, setID)
	if lesefehler != nil {
		return rest.Packangebot{}, lesefehler
	}

	arten := make([]rest.Packart, 0, len(antwort.Arten))
	for _, art := range antwort.Arten {
		arten = append(arten,
			rest.Packart{Name: art.Name, Waehlbar: art.Waehlbar, Karten: art.Karten})
	}

	return rest.Packangebot{Arten: arten, Befunde: antwort.Befunde}, nil
}

func (p pack) PackZiehen(ctx context.Context, setID int64, art, same string) (
	rest.Packziehung, error,
) {
	weg, err := app.PackUeber(p.db)
	if err != nil {
		return rest.Packziehung{}, err
	}

	antwort, ziehfehler := weg.Ziehen(ctx, setID, art, same)
	if ziehfehler != nil {
		return rest.Packziehung{}, ziehfehler
	}

	karten := make([]rest.Packkarte, 0, len(antwort.Karten))
	for _, karte := range antwort.Karten {
		karten = append(karten, rest.Packkarte{ID: karte.ID, Name: packname(karte)})
	}

	return rest.Packziehung{Karten: karten, Befunde: antwort.Befunde}, nil
}

// packname gibt den Namen einer gezogenen Karte.
//
// **Leer ist kein Fehler**: eine Karte ohne `name` gibt es, und ein
// Platzhalter verschleierte, dass sie so im Bestand steht.
func packname(karte app.KarteDTO) string {
	for _, wert := range karte.Werte {
		if wert.Feld == "name" {
			return wert.Wert
		}
	}

	return ""
}
