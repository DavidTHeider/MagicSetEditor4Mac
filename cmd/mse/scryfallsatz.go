// **Ein ganzer Satz von Scryfall** —
// #1033.
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// scryfallsatzbefehl liest einen ganzen Satz als Set ein.
//
// **Das Spiel gehört dazu**: Scryfall beschreibt Magic-Karten, nicht
// MSE-Kartenfelder — die Zuordnung dazwischen hängt am Spiel.
func scryfallsatzbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("scryfall-set braucht den Satzcode und das Spiel.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.ScryfallsatznahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		ergebnis, einlesefehler := naht.Einlesen(ctx, rest[0], rest[1])
		if einlesefehler != nil {
			an.klageUeber(einlesefehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, ergebnis)
		}

		if ergebnis.SchonOffen {
			an.sag("Der Satz %q liegt schon als Set %d im Bestand (%d Karten).\n",
				rest[0], ergebnis.SetID, ergebnis.Karten)

			return an.abschluss()
		}

		an.sag("Set %d angelegt: %d Karten aus %q.\n",
			ergebnis.SetID, ergebnis.Karten, rest[0])

		// **Die Befunde stehen dabei** (§8): ein Satz mit 300 lesbaren und
		// zwei auffälligen Karten wird eingelesen, und die zwei stehen hier.
		for _, befund := range ergebnis.Befunde {
			an.sag("  Befund: %s\n", befund)
		}

		return an.abschluss()
	})
}
