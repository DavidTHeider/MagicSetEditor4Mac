// Der Befehl `zeichnen` — eine Karte als Bild
// (#886).

package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bildAntwort sagt, was geschrieben wurde.
type bildAntwort struct {
	Pfad   string `json:"pfad"`
	Breite int    `json:"breite"`
	Hoehe  int    `json:"hoehe"`

	// Befunde ist, was beim Zeichnen aufgefallen ist (#1137).
	//
	// **Nicht `omitempty`**: eine leere Liste ist eine Aussage — nichts ist
	// aufgefallen —, und ein fehlendes Feld wäre keine.
	Befunde []string `json:"befunde"`
}

// karteZeichnen schreibt eine Karte als Bilddatei.
func karteZeichnen(
	ctx context.Context, db *sql.DB, an *schreiber, karteID int64, ziel, guete string,
) int {
	gewaehlt, gueteFehler := app.GueteAusName(guete)
	if gueteFehler != nil {
		an.klageUeber(gueteFehler)

		return CodeBedienung
	}

	werk, err := app.Zeichennaht(db)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	geschrieben, fehler := werk.SchreibeBild(ctx, karteID, ziel, gewaehlt)
	if fehler != nil {
		an.klageUeber(fehler)

		// **Eine unbrauchbare Endung ist ein Bedienfehler**, kein fachliches
		// Nein: der Aufruf stimmt nicht, und ein Skript soll das
		// unterscheiden können.
		if errors.Is(fehler, domain.ErrValidation) {
			return CodeBedienung
		}

		return CodeFachlich
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, bildAntwort{
			Pfad: geschrieben.Pfad, Breite: geschrieben.Breite,
			Hoehe: geschrieben.Hoehe, Befunde: geschrieben.Befunde,
		})
	}

	an.sag("Geschrieben nach %s (%d × %d).\n", ziel, geschrieben.Breite, geschrieben.Hoehe)

	// **Und was dabei aufgefallen ist** (#1137). Je Zeile einer: wer sie in
	// ein Protokoll leitet, hat sie einzeln.
	for _, befund := range geschrieben.Befunde {
		an.sag("  %s\n", befund)
	}

	return an.abschluss()
}
