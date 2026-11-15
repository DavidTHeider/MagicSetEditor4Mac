package main

// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// # Drei Befehle, und der dritte schreibt
//
//   - `ruecksprung-arten` sagt, was zurückgehen kann und was davon diese
//     Anwendung selbst tut.
//   - `ruecksprung-lage <datei>` stellt eine Sicherung neben den Bestand.
//   - `rueckkehren` nennt jede verbuchte Rückkehr;
//     `rueckkehr-quittieren <id>` erklärt eine für nachgesehen.
//
// # Der Rückgabewert von `ruecksprung-lage`
//
// **1, wenn etwas unklar bliebe oder von Hand nachzusehen ist.** Ein Skript,
// das vor dem Zurückspielen fragt, soll den Fall erkennen können, ohne den
// Text zu lesen — und „unklar" ist der Fall, der Aufmerksamkeit braucht.

import (
	"context"
	"database/sql"
	"os"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func ruecksprungartenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("ruecksprung-arten nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RuecksprungUeber(db, bestandsort).Ruecksprungarten(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, art := range antwort.Arten {
			an.sag("%s\t%s\n", art.Art, art.Hindernis)
		}

		return an.abschluss()
	})
}

func ruecksprunglagebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("ruecksprung-lage braucht genau eine Datei.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RuecksprungUeber(db, bestandsort).Ruecksprunglage(ctx, rest[0], os.Getenv(app.Sicherungskennwortumgebung))
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort); code != 0 {
				return code
			}

			return lagecode(antwort)
		}

		an.sag("datenstand\t%s\n", antwort.Datenstand)

		for _, zeile := range antwort.Zeilen {
			an.sag("%s\t%d\t%d\t%d\n",
				zeile.Gegenstand, zeile.InSicherung, zeile.Jetzt, zeile.Verlust)
		}

		for _, unklar := range antwort.Unklar {
			an.sag("unklar\t%s\t%s\t%s\t%d\n",
				unklar.Schluessel, unklar.Zustand, unklar.Sprache, unklar.Menge)
		}

		if antwort.HandpruefungNoetig {
			an.sag("handpruefung\tja\n")
		}

		if code := an.abschluss(); code != 0 {
			return code
		}

		return lagecode(antwort)
	})
}

// lagecode macht aus der Lage einen Rückgabewert.
//
// **Unklar ist nicht dasselbe wie ein Fehler**, und trotzdem ist es nicht 0:
// wer ein Skript baut, das nach dem Zurückspielen weiterläuft, soll hier
// anhalten. Der Rückgabewert ist die Aufforderung, nachzusehen.
func lagecode(antwort app.RuecksprunglageAntwort) int {
	if len(antwort.Unklar) > 0 || antwort.HandpruefungNoetig {
		return CodeFachlich
	}

	return CodeGut
}

func rueckkehrenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("rueckkehren nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RuecksprungUeber(db, bestandsort).Rueckkehren(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, eine := range antwort.Rueckkehren {
			an.sag("%d\t%s\t%s\t%s\t%s\t%d\n", eine.ID, eine.Klaerung,
				eine.Zurueckgespielt, eine.Datenstand, eine.Quelle, len(eine.Mengen))
		}

		return an.abschluss()
	})
}

func rueckkehrquittierenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("rueckkehr-quittieren braucht genau eine Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	kennung, fehler := strconv.ParseInt(rest[0], 10, 64)
	if fehler != nil {
		an.klage("%q ist keine Kennung.\n", rest[0])

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RuecksprungUeber(db, bestandsort).RueckkehrQuittieren(ctx, kennung)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("quittiert\t%d\n", kennung)

		return an.abschluss()
	})
}
