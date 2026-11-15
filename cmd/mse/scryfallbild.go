package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// scryfallbildbefehl holt ein Kartenbild von Scryfall und setzt es ein.
//
// # Warum er schreibt, wo `scryfall` erst zeigt
//
// Der Feldabgleich zeigt erst, weil ein Regeltext von Scryfall einen getippten
// überschreiben könnte — der Vorschlag ist die Grenze. **Ein Bild hat diesen
// Fall nicht**: es gibt nichts abzuwägen, entweder man will das Bild zu diesem
// Kartennamen oder nicht. Wer es nicht will, ruft den Befehl nicht.
func scryfallbildbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 4 {
		an.klage("scryfall-bild braucht Setkennung, Kartenkennung, Feldnamen und "+
			"den Kartennamen.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[1])
	if !gut {
		return CodeBedienung
	}

	art, gut := bildart(an, rest[4:])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		return scryfallBildEinsetzen(ctx, db, an, setID, karteID, rest[2], rest[3], art)
	})
}

// bildart liest `--art normal` aus den Restargumenten.
//
// **Ohne Angabe die ganze Karte.** Sie trägt Künstlername und Copyright im
// Bild selbst — die Bedingung von Scryfall ist damit ohne Zutun erfüllt.
func bildart(an *schreiber, rest []string) (string, bool) {
	if len(rest) == 0 {
		return "normal", true
	}

	if rest[0] != "--art" || len(rest) != 2 {
		an.klage("nach dem Kartennamen steht höchstens `--art klein|normal|ausschnitt`.\n%s",
			hilfe)

		return "", false
	}

	return rest[1], true
}

// scryfallBildEinsetzen führt die Naht und meldet, was danebensteht.
func scryfallBildEinsetzen(
	ctx context.Context, db *sql.DB, an *schreiber,
	setID, karteID int64, feld, name, art string,
) int {
	naht, err := app.Fremdbildnaht(db, domain.UrheberKommandozeile)
	if err != nil {
		an.klageUeber(err)

		return CodeFachlich
	}

	eingesetzt, bildfehler := naht.Einsetzen(ctx, setID, karteID, feld, name, art)
	if bildfehler != nil {
		an.klageUeber(bildfehler)

		return CodeFachlich
	}

	// **Die Nennung steht in der Ausgabe**, nicht in einer Fußnote: wer dieses
	// Bild weiterverwendet, muss sie mitnehmen, und die Kommandozeile ist die
	// Oberfläche, in der sie hier stehen kann.
	an.sag("Bild in %q eingesetzt.\n%s\nQuelle: %s", eingesetzt.Feld,
		eingesetzt.Nennung, eingesetzt.Adresse)

	if eingesetzt.IllustratorGesetzt {
		an.sag("»illustrator« steht jetzt auf %q.", eingesetzt.Kuenstler)
	}

	return CodeGut
}
