package main

// **Die Verträglichkeit auf der Kommandozeile** — Werkstück O6 von Phase 5.2
// (#1430).
//
// # Der Rückgabewert ist die Antwort
//
// **0, wenn die Fassung passt; 1, wenn nicht; 2, wenn es ungeklärt ist.** Ein
// Skript, das vor einem Wechsel fragt, soll die drei Ausgänge unterscheiden
// können — „ungeklärt" ist kein Nein, und es als eines zu melden hieße, alte
// Freigaben für unbrauchbar zu erklären.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// CodeUngeklaert heißt: es lässt sich nicht sagen.
const CodeUngeklaert = 2

func vertraeglichkeitsbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("vertraeglichkeit nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.VertraeglichkeitUeber(db).Vertraeglichkeit(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort); code != 0 {
				return code
			}

			return fassungscode(antwort.Urteil)
		}

		an.sag("fassung\t%s\n", antwort.Fassung)
		an.sag("urteil\t%s\n", antwort.Urteil)

		for _, hindernis := range antwort.Hindernisse {
			an.sag("hindernis\t%s\t%s\n", hindernis.Art, hindernis.Stelle)
		}

		if code := an.abschluss(); code != 0 {
			return code
		}

		return fassungscode(antwort.Urteil)
	})
}

// urteilscode macht aus dem Urteil einen Rückgabewert.
//
// **Ein leeres Urteil ist gut**: es heißt „es gibt keine neuere Fassung", und
// dafür gibt es nichts zu beanstanden.
func fassungscode(urteil string) int {
	switch urteil {
	case string(domain.UrteilUnvertraeglich):
		return CodeFachlich
	case string(domain.UrteilUngeklaert):
		return CodeUngeklaert
	}

	return CodeGut
}
