package main

// **Aus einem Set ein Deck** — FMT-04
// (#1693).

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// deckAusSet legt aus den Karten eines Sets ein Deck an — `deck aus-set <set>
// <menge> [name]`. Eine Menge, die keine Zahl ist, ist ein Bedienfehler; eine
// unter eins ein fachliches Nein des Dienstes.
func deckAusSet(ctx context.Context, flaeche *app.Deckflaeche, an *schreiber, rest []string) int {
	if len(rest) < 2 || len(rest) > 3 {
		an.klage("deck aus-set braucht das Set und die Menge, der Name ist frei.\n%s", hilfe)

		return CodeBedienung
	}

	menge, err := strconv.Atoi(rest[1])
	if err != nil {
		an.klage("Die Menge %q ist keine Zahl.\n", rest[1])

		return CodeBedienung
	}

	name := ""
	if len(rest) == 3 {
		name = rest[2]
	}

	antwort := flaeche.DeckAusSet(ctx, rest[0], name, menge)
	if antwort.Fehler != nil {
		return klageUeberNaht(an, antwort.Fehler)
	}

	if an.alsJSON {
		return alsJSONSchreiben(an, antwort)
	}

	an.sag("%s\n", antwort.Deck.ID)

	return an.abschluss()
}
