// **Die Schlüsselwörter eines Sets** —
// #1062.
//
// Zwei Befehle: einer zeigt, einer pflegt. Die Trennung ist dieselbe wie bei
// den Wortlisten — nichts wird still geändert.
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schluesselwoerterbefehl zeigt die des Sets und die des Spiels.
func schluesselwoerterbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 1 {
		an.klage("schluesselwoerter braucht die Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		pflege, err := app.SchluesselwortpflegeAls(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		liste, lesefehler := pflege.Liste(ctx, setID)
		if lesefehler != nil {
			an.klageUeber(lesefehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, liste)
		}

		zeigeSchluesselwoerter(an, liste.Woerter)

		return an.abschluss()
	})
}

// zeigeSchluesselwoerter schreibt die Liste in lesbarer Form.
//
// **Die des Spiels tragen keine Kennung** und stehen mit einem Strich dort, wo
// bei den eigenen die Zahl steht — so ist auf einen Blick zu sehen, welche sich
// ändern lassen.
func zeigeSchluesselwoerter(an *schreiber, woerter []app.SchluesselwortDTO) {
	if len(woerter) == 0 {
		an.sag("Dieses Set und sein Spiel führen keine Schlüsselwörter.\n")

		return
	}

	for _, wort := range woerter {
		kennung := wort.ID
		if kennung == "" {
			kennung = "—"
		}

		marke := " "
		if !wort.Gueltig {
			marke = "!"
		}

		an.sag("%-6s %s %-24s %s\n", kennung, marke, wort.Wort, wort.Muster)
	}

	an.sag("\n» — « heißt: aus dem Spiel, unveränderlich. " +
		"» ! « heißt: das Muster erkennt nichts.\n")
}

// schluesselwortbefehl pflegt eines des Sets.
func schluesselwortbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("schluesselwort braucht das Set und eine Handlung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		pflege, err := app.SchluesselwortpflegeAls(db, domain.UrheberKommandozeile)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return schluesselworthandlung(ctx, an, pflege, setID, rest[1:])
	})
}
