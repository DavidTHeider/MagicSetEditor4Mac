package main

// **Einladen und annehmen an der Kommandozeile** — NET-01
// (#1705).
//
//	mse einladung-erstellen <name> <zweck>          eine Einladung ausstellen
//	mse einladung-pruefen <code>                    sehen, was sie sagt
//	mse einladung-annehmen <code> <name> [<vergleichscode> [freigeben]]
//
// **Der Vergleichscode ist ein Argument und kein Schalter** — wie der Abdruck
// bei `kontakt-bestaetigen`: wer bestätigt, tippt ab, was er verglichen hat.
// **Und `freigeben` ist ein Wort und kein Schalter**, damit es an der Stelle
// steht, an der es gilt: hinter dem Vergleich, ohne den es nichts freigibt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// einladungerstellenbefehl stellt eine Einladung aus.
func einladungerstellenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	adresse, rest := flaggeAus(rest, "--adresse")

	if len(rest) != 2 {
		an.klage("einladung-erstellen nimmt 2 Argumente [--adresse <adresse>].\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PeerflaecheUeber(db).KontakteinladungErstellenUnter(
			ctx, rest[0], rest[1], adresse)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("code\t%s\n", antwort.Einladung.Code)
		an.sag("ablauf\t%s\n", antwort.Einladung.Ablauf)
		an.sag("fingerabdruck\t%s\n", antwort.Einladung.Fingerabdruck)

		return an.abschluss()
	})
}

// einladungpruefenbefehl sagt, was eine Einladung ist — schreibt nichts.
func einladungpruefenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("einladung-pruefen nimmt 1 Argument.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PeerflaecheUeber(db).KontakteinladungPruefen(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		vorschau := antwort.Vorschau
		an.sag("name\t%s\nzweck\t%s\nlage\t%s\nvergleich\t%s\n",
			vorschau.Name, vorschau.Zweck, vorschau.Lage, vorschau.Vergleichscode)
		an.sag("kennung\t%s\nfingerabdruck\t%s\nablauf\t%s\n",
			vorschau.Kennung, vorschau.Fingerabdruck, vorschau.Ablauf)

		// **Nur, wo es etwas zu sagen gibt** — dieselbe Form wie bei den
		// Kontakten.
		if vorschau.Adresse != "" {
			an.sag("unter\t%s\n", vorschau.Adresse)
		}

		if vorschau.Vorgaenger != "" {
			an.sag("vorgaenger\t%s\n", vorschau.Vorgaenger)
		}

		if vorschau.Angenommen {
			an.sag("angenommen\tja\n")
		}

		return an.abschluss()
	})
}

// einladungannehmenbefehl nimmt eine Einladung an — ganz oder gar nicht.
func einladungannehmenbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	meinName, rest := flaggeAus(rest, "--mein-name")

	if len(rest) < 2 || len(rest) > 4 || len(rest) == 4 && rest[3] != "freigeben" {
		an.klage("einladung-annehmen nimmt <code> <name> [<vergleichscode> "+
			"[freigeben]] [--mein-name <name>].\n%s", hilfe)

		return CodeBedienung
	}

	verglichen := ""
	if len(rest) > 2 {
		verglichen = rest[2]
	}

	return kontaktlauf(ctx, bestandsort, an, rest, "einladung-annehmen", len(rest),
		func(ctx context.Context, anwendung *app.App, teile []string) app.PeerkontakteAntwort {
			antwort := anwendung.KontakteinladungAnnehmen(
				ctx, teile[0], teile[1], verglichen, len(teile) == 4, meinName)

			// **Die Meldung beim Einladenden steht vor der Liste** (NET-01b),
			// in derselben Form: ein Wort, dann der Satz. Wer JSON will,
			// findet sie dort unter `gemeldet`.
			if gemeldet := antwort.Gemeldet; gemeldet != nil && antwort.Fehler == nil && !an.alsJSON {
				an.sag("gemeldet\t%s\t%s\n", gemeldet.Stand, gemeldet.Meldung)
			}

			return antwort
		})
}

// flaggeAus nimmt `<flagge> <wert>` aus den Argumenten, wo immer es steht —
// `--mein-name` beim Annehmen, `--adresse` beim Einladen (NET-01b).
func flaggeAus(argumente []string, flagge string) (wert string, uebrig []string) {
	uebrig = make([]string, 0, len(argumente))

	for stelle := 0; stelle < len(argumente); stelle++ {
		if argumente[stelle] == flagge && stelle+1 < len(argumente) {
			wert = argumente[stelle+1]
			stelle++

			continue
		}

		uebrig = append(uebrig, argumente[stelle])
	}

	return wert, uebrig
}
