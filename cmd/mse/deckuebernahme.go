// **Ein Deck aus einem fremden Profil übernehmen** — Werkstück 12 von
// Phase 4.12
// (#1351).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
//	mse deck-uebernehmen <quellProfil> <quellDeck>
//	mse decks-uebernehmbar <quellProfil>
//
// **Erweiternd** im Sinne von 0033: ein Befehl kommt hinzu, keiner ändert sich.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// deckuebernahmebefehl holt ein Deck aus einem fremden Profil.
func deckuebernahmebefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 2 {
		an.klage("deck-uebernehmen braucht die Kennung des fremden Profils "+
			"und die des Decks.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).DeckUebernehmen(ctx, rest[0], rest[1])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Deck.ID, antwort.Deck.Name)

		// **Was nicht mitkam, steht danach** — und zwar auf der Ausgabe und
		// nicht auf der Fehlerausgabe: es ist eine Auskunft, kein Fehler, und
		// der Rückgabewert bleibt 0 (§8).
		for _, befund := range antwort.Befunde {
			an.sag("  %s\n", befund)
		}

		return an.abschluss()
	})
}

// deckwahlbefehl nennt die Decks eines fremden Profils — zur Wahl für
// `deck-uebernehmen` (UX-08.AK-04, #1680). Dieselbe Erlaubnis wie die
// Übernahme, geprüft, bevor gelesen wird.
func deckwahlbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 1 {
		an.klage("decks-uebernehmbar braucht die Kennung des fremden "+
			"Profils.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).DeckeZurUebernahme(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		for _, deck := range antwort.Decks {
			an.sag("%s\t%s\t%s\n", deck.ID, deck.Name, deck.Spiel)
		}

		return an.abschluss()
	})
}
