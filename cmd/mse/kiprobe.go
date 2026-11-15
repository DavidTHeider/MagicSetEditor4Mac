// **Ein Modell messen an der Kommandozeile** — Werkstück A9 von Phase 5.3
// (#1448).
//
// **Eine eigene Datei**, weil `ki.go` mit ihr über die Zeilengrenze aus §2
// ginge. Der Schnitt folgt der Frage: dort steht, was ein Modell **tut**,
// hier, was es **kann** — und wie man das misst.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kimodellpruefenbefehl misst ein Modell und merkt sich das Ergebnis.
//
// **Es kostet Geld**, wenn der Anbieter anderswo rechnet — und geht deshalb
// durch dieselbe Sperre wie ein Vorschlag: Schalter, »nur lokal«, Journal
// und Freigabe.
func kimodellpruefenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("ki-modell-pruefen braucht den Anbieter und das Modell.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KIAnwendungUeber(db).KIModellPruefen(ctx, rest[0], rest[1], "")
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort.Proben)
		}

		for _, probe := range antwort.Proben {
			an.sag("%s/%s\t%s\tgeprüft %s\t%s\t%s\n",
				probe.Anbieter, probe.Alias, probe.Faehigkeit, probe.Datum,
				probe.Modell, probe.Beleg)
		}

		return an.abschluss()
	})
}

// kimodellvergessenbefehl nimmt die Messungen eines Modells zurück.
func kimodellvergessenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("ki-modell-vergessen braucht den Anbieter und das Modell.\n%s",
			hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KIAnwendungUeber(db).KIModellVergessen(ctx, rest[0], rest[1])
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		// **Eine Zeile und nicht Schweigen.** Wer etwas wegnimmt, will
		// lesen, dass es geschehen ist — und die anderen drei Flächen sagen
		// es auch (§16).
		an.sag("vergessen\t%s/%s\n", rest[0], rest[1])

		return an.abschluss()
	})
}
