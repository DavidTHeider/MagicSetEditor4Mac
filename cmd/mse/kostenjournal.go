package main

// **Das Kostenjournal auf der Kommandozeile** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Drei Befehle
//
//   - `mse kosten <konto> [zeitraum]` — die vier Zahlen. Der Rückgabewert ist
//     **1**, wenn etwas ungeklärt ist: ein Skript soll den Fall erkennen, ohne
//     den Text zu lesen.
//   - `mse kostenposten <konto> [zeitraum]` — die letzten Aufträge.
//   - `mse kostenbudget <konto> <zeitraum> <betrag> <waehrung>` — die Grenze.
//
// **Der Betrag steht in der kleinsten Einheit**, und das steht in der Hilfe:
// „5000 EUR" wären fünfzig Euro. Wer es anders schriebe, müsste runden.

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kostenbefehle verteilt die drei Wege über Geld.
func kostenbefehle(
	ctx context.Context, befehl string, rest []string, bestandsort string, an *schreiber,
) (int, bool) {
	switch befehl {
	case "kosten":
		return kostenbefehl(ctx, bestandsort, an, rest), true
	case "kostenposten":
		return kostenpostenbefehl(ctx, bestandsort, an, rest), true
	case "kostenbudget":
		return kostenbudgetbefehl(ctx, bestandsort, an, rest), true
	}

	return 0, false
}

func kostenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 || len(rest) > 2 {
		an.klage("kosten braucht das Konto, dazu wahlweise den Zeitraum.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KostenAnwendungUeber(db).
			Kostenstand(ctx, rest[0], zweitesOderLeer(rest))
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			if code := alsJSONSchreiben(an, antwort); code != 0 {
				return code
			}

			return fachcode(antwort.Unklar.Wert > 0)
		}

		an.sag("konto\t%s\n", antwort.Konto)
		an.sag("zeitraum\t%s\n", antwort.Zeitraum)

		for name, betrag := range map[string]app.BetragDTO{
			"grenze": antwort.Grenze, "reserviert": antwort.Reserviert,
			"verbraucht": antwort.Verbraucht, "unklar": antwort.Unklar,
			"verfuegbar": antwort.Verfuegbar,
		} {
			an.sag("%s\t%d\t%s\n", name, betrag.Wert, betrag.Waehrung)
		}

		if code := an.abschluss(); code != 0 {
			return code
		}

		return fachcode(antwort.Unklar.Wert > 0)
	})
}

func kostenpostenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) == 0 || len(rest) > 2 {
		an.klage("kostenposten braucht das Konto, dazu wahlweise den Zeitraum.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KostenAnwendungUeber(db).
			Kostenposten(ctx, rest[0], zweitesOderLeer(rest))
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, posten := range antwort.Posten {
			an.sag("%s\t%s\t%s\t%d\t%d\t%s\t%s\n",
				posten.Kennung, posten.Stand, posten.Aufgabe,
				posten.Hoechstbetrag.Wert, posten.Betrag.Wert,
				posten.Hoechstbetrag.Waehrung, posten.Begonnen)
		}

		return an.abschluss()
	})
}

func kostenbudgetbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 4 {
		an.klage("kostenbudget braucht Konto, Zeitraum, Betrag und Währung.\n%s", hilfe)

		return CodeBedienung
	}

	wert, zahlfehler := strconv.ParseInt(rest[2], 10, 64)
	if zahlfehler != nil {
		an.klage("%q ist kein Betrag in der kleinsten Einheit.\n", rest[2])

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.KostenAnwendungUeber(db).
			KostenbudgetSetzen(ctx, rest[0], rest[1], wert, rest[3])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("gesetzt\t%s\t%s\t%d\t%s\n", rest[0], rest[1], wert, rest[3])

		return an.abschluss()
	})
}

// zweitesOderLeer gibt das zweite Argument — oder nichts.
func zweitesOderLeer(rest []string) string {
	if len(rest) > 1 {
		return rest[1]
	}

	return ""
}
