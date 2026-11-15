// **Die Decklisteneinfuhr des Dienstes** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// **Ohne Wache auf die Platte** (0044): beide Wege fassen nur den Bestand an,
// keine Datei. Was der Dienst **nicht** tut, ist entscheiden, welcher von
// mehreren Drucken gemeint ist — dieselbe Zusage wie bei der Listeneinfuhr
// daneben.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// deckeinfuhr führt die Decklisteneinfuhr an die Naht von `internal/rest`.
type deckeinfuhr struct{ db *sql.DB }

// DecklisteVorschau zeigt, was aus einer Deckliste würde.
func (d deckeinfuhr) DecklisteVorschau(
	ctx context.Context, deckID int64, text string,
) (any, error) {
	naht, err := app.DeckeinfuhrUeber(d.db)
	if err != nil {
		return nil, err
	}

	return naht.Vorschau(ctx, deckID, text)
}

// DecklisteUebernehmen legt die bestätigten Zeilen ins Deck.
func (d deckeinfuhr) DecklisteUebernehmen(
	ctx context.Context, deckID int64, stand string,
	zeilen []rest.Deckeinfuhrzeile,
) (any, error) {
	naht, err := app.DeckeinfuhrUeber(d.db)
	if err != nil {
		return nil, err
	}

	bestaetigt := make([]app.Deckeinfuhrzeile, 0, len(zeilen))
	for _, zeile := range zeilen {
		bestaetigt = append(bestaetigt, app.Deckeinfuhrzeile{
			Printing: zeile.Printing, Menge: zeile.Menge,
			Bereich: zeile.Bereich,
		})
	}

	return naht.Uebernehmen(ctx, deckID, stand, bestaetigt)
}
