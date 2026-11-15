package main

// **Der lokale Suchindex auf der Kommandozeile** — Werkstück A15 von Phase
// 5.3 (#1514).
//
// # Drei Befehle
//
//   - `mse indexstand` sagt, wogegen gesucht wird — und ob semantisch.
//   - `mse indexsuche <frage> [grenze]` sucht.
//   - `mse indexaufbau <modell> <sammeldatei> <revision> [datum] [--delta]`
//     baut neu. **Mit `--delta`** baut er auf der aktiven Generation auf: die
//     Sammeldatei liefert dann nur, was sich geändert hat, und der Rest wird
//     **kopiert** statt neu eingebettet.
//
// **Der Aufbau steht hier und nicht über MCP**, und der Grund ist derselbe
// wie dort: er läuft Minuten und bindet einen Kern. Wer ihn hier anstößt, hat
// ihn angestoßen; ein Modell, das ihn anstieße, nähme jemandem die Maschine
// weg.
//
// **Abgebrochen wird mit `^C`** — der Befehl hängt seinen Kontext daran, und
// der Aufbau verwirft dann seine halbe Generation. Ein eigener
// Abbruchbefehl wäre einer für einen Aufbau in einem **anderen** Prozess, und
// den hält dieser Riegel nicht.

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func indexstandbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("indexstand nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RetrievalAnwendungUeber(db).Indexstand(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		if !antwort.Aktiv {
			an.sag("Kein Suchindex aktiv. %s\n", antwort.Grund)

			return an.abschluss()
		}

		an.sag("Modell:   %s (%s)\n", antwort.Manifest.Modell,
			antwort.Manifest.Laufzeit)
		an.sag("Verfahren: %s, %s, %d Stellen\n",
			antwort.Manifest.Suchkonfiguration, antwort.Manifest.Abstandsmass,
			antwort.Manifest.Dimension)

		for _, stand := range antwort.Manifest.Korpusstaende {
			an.sag("Stand:    %s %s %s\n", stand.Quelle, stand.Revision, stand.Datum)
		}

		an.sag("Semantisch suchbar: %v %s\n", antwort.Suchbar, antwort.Grund)

		return an.abschluss()
	})
}

func indexsuchbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 || len(rest) > 2 {
		an.klage("indexsuche braucht die Frage, dazu wahlweise die "+
			"Trefferzahl.\n%s", hilfe)

		return CodeBedienung
	}

	grenze, gedeutet := grenzeAus(rest)
	if !gedeutet {
		an.klage("die Trefferzahl ist keine Zahl.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.RetrievalAnwendungUeber(db).Indexsuche(
			ctx, rest[0], grenze, app.IndexsuchfilterDTO{})
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("Modus: %s %s\n", antwort.Modus, antwort.Grund)

		for _, treffer := range antwort.Treffer {
			an.sag("%2d  %s  [%s %s]\n", treffer.Rang, treffer.Titel,
				treffer.Quelle, treffer.Quellrevision)
		}

		return an.abschluss()
	})
}

func indexaufbaubefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	// **Der Schalter steht vorn oder hinten** — `--delta` ist eine Angabe
	// über den Lauf und keine Stelle in der Reihe.
	rest, fortschreiben := ohneDelta(rest)

	if len(rest) < 3 || len(rest) > 4 {
		an.klage("indexaufbau braucht Modell, Sammeldatei und Revision, dazu "+
			"wahlweise das Gültigkeitsdatum und --delta.\n%s", hilfe)

		return CodeBedienung
	}

	datum := ""
	if len(rest) == 4 {
		datum = rest[3]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		// **Mit dem Ort des Bestands** — daran hängt die Absprache über die
		// Prozesse ([app.Indexsperre]). Ohne ihn bauten zwei
		// `mse indexaufbau` gleichzeitig gegen dieselbe Datei; gefunden vom
		// Flächenlauf zum laufenden Rebuild (A14, #1453).
		antwort := app.RetrievalAnwendungUeber(db).
			MitBestandsort(bestandsort).IndexAufbauen(
			ctx, rest[0], rest[1], rest[2], datum, fortschreiben)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("Generation %d: %d Abschnitte in %d ms (fortgeschrieben: %v)\n",
			antwort.Generation, antwort.Geschrieben, antwort.Dauer,
			antwort.Fortgeschrieben)

		return an.abschluss()
	})
}

// ohneDelta zieht `--delta` aus den Argumenten.
//
// **Der Schalter heißt nicht `--fortschreiben`**, weil niemand ihn so tippt;
// was er tut, steht in der Hilfe.
func ohneDelta(rest []string) ([]string, bool) {
	uebrig := make([]string, 0, len(rest))
	gefunden := false

	for _, teil := range rest {
		if teil == "--delta" {
			gefunden = true

			continue
		}

		uebrig = append(uebrig, teil)
	}

	return uebrig, gefunden
}

// grenzeAus liest die wahlweise Trefferzahl — zehn, wo keine steht.
func grenzeAus(rest []string) (int, bool) {
	if len(rest) < 2 {
		return 10, true
	}

	grenze, fehler := strconv.Atoi(rest[1])

	return grenze, fehler == nil
}
