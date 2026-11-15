package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zusatzfeldbefehl liest oder setzt die Zusatzfelder einer Karte — die Felder,
// die ihr Stil mitbringt und die sich von Hand setzen lassen (#2298).
//
// **Zwei Gestalten**: `zusatzfeld <set> <karte>` zeigt, `… setzen <feld>
// <wert>` schreibt — wie `stiloption`.
func zusatzfeldbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 2 && (len(rest) != 5 || rest[2] != "setzen") {
		an.klage("zusatzfeld nimmt <set> <karte> und wahlweise »setzen <feld> <wert>«.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gutesSet := kennung(an, "die Setkennung", rest[0])
	if !gutesSet {
		return CodeBedienung
	}

	karteID, guteKarte := kennung(an, "die Kartenkennung", rest[1])
	if !guteKarte {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.ZusatznahtUeber(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if len(rest) == 5 {
			_, fehler := weg.Setzen(ctx, setID, karteID, rest[3], rest[4])

			return getan(an, fehler, "Gesetzt.\n")
		}

		return zusatzfelderZeigen(ctx, an, weg, setID, karteID)
	})
}

// zusatzfelderAntwort ist die Gestalt unter `--json`.
type zusatzfelderAntwort struct {
	Felder []app.ZusatzfeldDTO `json:"felder"`
}

// zusatzfelderZeigen schreibt die Zusatzfelder samt Werten.
func zusatzfelderZeigen(ctx context.Context, an *schreiber, weg *app.Zusatzweg, setID, karteID int64) int {
	felder, fehler := weg.Felder(ctx, setID, karteID)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, zusatzfelderAntwort{Felder: felder})
	}

	if len(felder) == 0 {
		an.sag("Der Stil dieser Karte hat keine Zusatzfelder, die sich setzen lassen.\n")

		return an.abschluss()
	}

	for i := range felder {
		wert := felder[i].Wert
		if !felder[i].Gespeichert {
			wert = "(Vorgabe des Stils)"
		}

		an.sag("%-24s %-16s %s\n", felder[i].Form.Name, felder[i].Form.Art, wert)
	}

	return an.abschluss()
}
