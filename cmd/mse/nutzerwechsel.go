// **Wer arbeitet, und wie man ein anderer wird** — Werkstück 17 von Phase 4.12
// (#1400),
// nach [0074](../../docs/entscheidungen/0074-wer-arbeitet-steht-fest.md).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// nutzerwechselbefehl zeigt, wer arbeitet — oder stellt es um.
//
// # Ohne Argument zeigen, mit Argument wechseln
//
// **Dieselbe Form wie `mse profil`** daneben, und derselbe Grund: das Zeigen
// ist der häufigere und der harmlosere Ruf.
func nutzerwechselbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) > 2 {
		an.klage("nutzer-wechseln nimmt höchstens zwei Angaben: "+
			"die Kennung und den Nachweis.\n%s", hilfe)

		return CodeBedienung
	}

	// **Der Nachweis ist der zweite Faktor des Ziels** — nicht der eigene.
	// **Im Argument und nicht auf der Standardeingabe**, aus demselben Grund
	// wie beim Profilwechsel: ein Einmalkennwort gilt neunzig Sekunden.
	nachweis := ""
	if len(rest) == 2 {
		nachweis = rest[1]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		flaeche := app.NutzerflaecheUeber(db)

		if len(rest) == 0 {
			antwort := flaeche.WerArbeitet(ctx)
			if antwort.Fehler != nil {
				an.klageUeberFehler(antwort.Fehler)

				return CodeFachlich
			}

			if an.alsJSON {
				return alsJSONSchreiben(an, antwort)
			}

			an.sag("Es arbeitet der Nutzer %s.\n", antwort.NutzerID)

			return CodeGut
		}

		antwort := flaeche.NutzerWechseln(ctx, rest[0], nachweis)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("Es arbeitet jetzt der Nutzer %s, im Profil %s.\n",
			antwort.Profil.NutzerID, antwort.Profil.Name)

		return CodeGut
	})
}
