// **Die Naht zur Deckübernahme** — Werkstück 12 von Phase 4.12
// (#1351).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// deckuebernahme bedient den Weg über einem Bestand.
type deckuebernahme struct{ db *sql.DB }

func (d deckuebernahme) DeckUebernehmen(
	ctx context.Context, quellProfil, quellDeck string,
) (rest.Uebernommenes, error) {
	antwort := app.NutzerflaecheKopflos(d.db).DeckUebernehmen(
		ctx, quellProfil, quellDeck)
	if antwort.Fehler != nil {
		return rest.Uebernommenes{}, app.AlsFehler(antwort.Fehler)
	}

	return rest.Uebernommenes{
		DeckID:  antwort.Deck.ID,
		Name:    antwort.Deck.Name,
		Spiel:   antwort.Deck.Spiel,
		Plaetze: len(antwort.Deck.Plaetze),
		Befunde: antwort.Befunde,
	}, nil
}

// DeckeZurUebernahme nennt die Decks eines fremden Profils — als Liste und
// nicht als nacktes `null`, wenn es keine gibt.
func (d deckuebernahme) DeckeZurUebernahme(
	ctx context.Context, quellProfil string,
) (any, error) {
	antwort := app.NutzerflaecheKopflos(d.db).DeckeZurUebernahme(ctx, quellProfil)
	if antwort.Fehler != nil {
		return nil, app.AlsFehler(antwort.Fehler)
	}

	return map[string]any{"decks": antwort.Decks}, nil
}
