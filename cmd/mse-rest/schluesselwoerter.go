// Die Schlüsselwörter für den Dienst
// (#1062).
package main

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// schluesselwoerter führt die vier Endpunkte an die Naht dieser Schicht.
type schluesselwoerter struct{ db *sql.DB }

func (s schluesselwoerter) Schluesselwoerter(ctx context.Context, setID int64) (
	rest.Schluesselwortliste, error,
) {
	pflege, err := app.SchluesselwortpflegeAls(s.db, domain.UrheberDienst)
	if err != nil {
		return rest.Schluesselwortliste{}, err
	}

	liste, lesefehler := pflege.Liste(ctx, setID)
	if lesefehler != nil {
		return rest.Schluesselwortliste{}, lesefehler
	}

	umgesetzt := make([]rest.Schluesselwort, 0, len(liste.Woerter))

	for _, wort := range liste.Woerter {
		umgesetzt = append(umgesetzt, rest.Schluesselwort{
			ID: kennungAus(wort.ID), Wort: wort.Wort, Muster: wort.Muster,
			Erinnerung: wort.Erinnerung, Regeln: wort.Regeln, Modus: wort.Modus,
			Fest: wort.Fest, Gueltig: wort.Gueltig,
		})
	}

	return rest.Schluesselwortliste{Woerter: umgesetzt, Modi: liste.Modi}, nil
}

func (s schluesselwoerter) SchluesselwortAnlegen(
	ctx context.Context, setID int64,
) (int64, error) {
	pflege, err := app.SchluesselwortpflegeAls(s.db, domain.UrheberDienst)
	if err != nil {
		return 0, err
	}

	return pflege.Anlegen(ctx, setID)
}

func (s schluesselwoerter) SchluesselwortSetzen(
	ctx context.Context, setID, wortID int64, feld, wert string,
) error {
	pflege, err := app.SchluesselwortpflegeAls(s.db, domain.UrheberDienst)
	if err != nil {
		return err
	}

	return pflege.Setzen(ctx, setID, wortID, feld, wert)
}

func (s schluesselwoerter) SchluesselwortLoeschen(
	ctx context.Context, setID, wortID int64,
) error {
	pflege, err := app.SchluesselwortpflegeAls(s.db, domain.UrheberDienst)
	if err != nil {
		return err
	}

	return pflege.Loeschen(ctx, setID, wortID)
}

// kennungAus macht aus der Textkennung eine Zahl.
//
// **Leer heißt null**, und das ist die Auskunft: ein Schlüsselwort des Spiels
// gehört keinem Set und trägt keine Kennung. Ein Fehler wäre hier die falsche
// Antwort — die leere Zeichenkette ist kein Versehen, sondern der Fall.
func kennungAus(text string) int64 {
	kennung, _ := strconv.ParseInt(text, 10, 64) //nolint:errcheck // leer heißt null

	return kennung
}
