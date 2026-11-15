package main

// **Die abgestufte Rücksetzung auf der Kommandozeile** — Werkstück O4 von
// Phase 5.2 (#1428).
//
// # Drei Befehle, und der dritte fasst an
//
//   - `mse ruecksetzungsarten` nennt die sechs in der Reihe ihrer Schwere.
//   - `mse ruecksetzung-vorschau <art> [profilID] [nutzerID]` sieht nach und
//     **ändert nichts**. Der Rückgabewert ist 1, wenn etwas im Weg steht —
//     ein Skript soll den Fall erkennen, ohne den Text zu lesen.
//   - `mse zuruecksetzen <art> <name> [nachweis]` führt aus.
//
// **Der Name steht im Argument, und das ist die Bindung**: wer ein Profil
// oder einen Menschen löschen will, tippt seinen Namen ab. Ein Schalter
// `--ja` wäre keine.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func ruecksetzungsartenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("ruecksetzungsarten nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(_ context.Context, db *sql.DB) int {
		arten := app.RuecksetzungsAnwendungUeber(db, bestandsort).Ruecksetzungsarten()

		if an.alsJSON {
			return alsJSONSchreiben(an, arten)
		}

		for _, art := range arten {
			an.sag("%s\n", art)
		}

		return an.abschluss()
	})
}

func ruecksetzungsvorschaubefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 || len(rest) > 3 {
		an.klage("ruecksetzung-vorschau braucht die Art, dazu wahlweise "+
			"Profil- und Nutzerkennung.\n%s", hilfe)

		return CodeBedienung
	}

	profilID, nutzerID := kennungenAus(rest)

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RuecksetzungsAnwendungUeber(db, bestandsort).
			Ruecksetzungsvorschau(ctx, rest[0], profilID, nutzerID)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort); code != 0 {
				return code
			}

			return befundscode(len(antwort.Hindernisse))
		}

		vorschauSchreiben(an, antwort)

		if code := an.abschluss(); code != 0 {
			return code
		}

		return befundscode(len(antwort.Hindernisse))
	})
}

// vorschauSchreiben gibt die Vorschau in Zeilen.
func vorschauSchreiben(an *schreiber, antwort app.RuecksetzungsvorschauAntwort) {
	an.sag("art\t%s\n", antwort.Art)
	an.sag("bereit\t%t\n", antwort.Bereit)
	an.sag("nachweisnoetig\t%t\n", antwort.NachweisNoetig)
	an.sag("sicherungempfohlen\t%t\n", antwort.SicherungEmpfohlen)

	// **Und die Revision steht in den Zeilen** — Befund 9 des Nachlaufs zum
	// Zweitblick über Phase 5.2. Sie ist Pflicht, wo etwas gelöscht wird; sie
	// nur im JSON zu nennen hieße, den zeilenweisen Weg an derselben Fähigkeit
	// scheitern zu lassen (§16).
	an.sag("revision\t%s\n", antwort.Revision)

	for _, gruppe := range antwort.Betroffen {
		an.sag("betroffen\t%s\t%d\t%s\n",
			gruppe.Was, gruppe.Anzahl, strings.Join(gruppe.Namen, ", "))
	}

	for _, hindernis := range antwort.Hindernisse {
		an.sag("hindernis\t%s\t%s\n", hindernis.Art, hindernis.Stelle)
	}
}

func zuruecksetzenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	// **Die Revision steht hinter einem Schalter** — Befund 12 des
	// Zweitblicks über Phase 5.2. Sie ist die Marke aus der Vorschau und
	// bindet die Ausführung an das, was jemand gesehen hat.
	//
	// **Und wo etwas gelöscht wird, ist sie Pflicht** — Befund 9 des
	// Nachlaufs. Ohne sie liefe der Befehl an der Bindung vorbei, und eine
	// Zusage mit einer Tür daneben ist keine. Wahlfrei bleibt sie für die drei
	// Arten, die nur eine Wahl zurücknehmen.
	rest, revision := revisionAus(rest)

	if len(rest) < 2 || len(rest) > 5 {
		an.klage("zuruecksetzen braucht die Art und den Namen zur Bestätigung, "+
			"dazu --revision aus der Vorschau (Pflicht, wo etwas gelöscht "+
			"wird), Nachweis, Profil- und Nutzerkennung.\n%s", hilfe)

		return CodeBedienung
	}

	var nachweis string
	if len(rest) > 2 {
		nachweis = rest[2]
	}

	profilID, nutzerID := kennungenAus(rest[2:])

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RuecksetzungsAnwendungUeber(db, bestandsort).
			Zuruecksetzen(ctx, rest[0], profilID, nutzerID, rest[1], nachweis, revision)

		// **Die Bilanz wird auch mit einem Fehler geschrieben** (§7, Regel 3):
		// der Ort des beiseitegelegten Bestands ist die Auskunft, ohne die
		// niemand ihn wiederfände.
		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort); code != 0 {
				return code
			}

			return fachcode(antwort.Fehler != nil)
		}

		if antwort.Beiseite != "" {
			an.sag("beiseite\t%s\n", antwort.Beiseite)
		}

		an.sag("zurueckgesetzt\t%d\n", antwort.Zurueckgesetzt)

		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		return an.abschluss()
	})
}

// kennungenAus liest Profil- und Nutzerkennung aus den letzten Argumenten.
//
// **Stellenweise und nicht über Schalter**: `mse` kennt keine benannten
// Argumente, und ein leerer Platz heißt „meines" — dieselbe Zusage wie an
// jeder anderen Fläche.
func kennungenAus(rest []string) (profilID, nutzerID string) {
	if len(rest) > 1 {
		profilID = rest[1]
	}

	if len(rest) > 2 {
		nutzerID = rest[2]
	}

	return profilID, nutzerID
}

// fachcode gibt den Rückgabewert zu einem fachlichen Fehler.
func fachcode(gescheitert bool) int {
	if gescheitert {
		return CodeFachlich
	}

	return 0
}

// revisionAus zieht `--revision <marke>` aus den Argumenten.
//
// **Vor der Stellenzählung**, damit der Schalter an jeder Stelle stehen darf:
// wer ihn zwischen Art und Namen schriebe, verschöbe sonst alles dahinter.
func revisionAus(rest []string) (ohne []string, revision string) {
	ohne = make([]string, 0, len(rest))

	for stelle := 0; stelle < len(rest); stelle++ {
		if rest[stelle] != "--revision" {
			ohne = append(ohne, rest[stelle])

			continue
		}

		if stelle+1 < len(rest) {
			revision, stelle = rest[stelle+1], stelle+1
		}
	}

	return ohne, revision
}
