package app

// **Aus einem Set ein Deck** — FMT-04 (R52,
// #1693).
// Die Karten eines Sets werden Plätze eines neuen Decks; entschieden wird im
// Dienst (`service.Setdeck`), hier wird übersetzt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setdecknaht legt aus einem Set ein Deck an.
type Setdecknaht interface {
	Anlegen(ctx context.Context, setID int64, name string, menge int) (domain.Deck, error)
}

// MitSetdeck setzt die Naht — an **beiden** Composition Roots.
func (a *App) MitSetdeck(naht Setdecknaht) { a.setdeck = naht }

// DeckAusSet legt aus den Karten eines Sets ein Deck an — jede Karte ein
// Platz im Hauptdeck mit `menge`; ohne Namen heißt es wie das Set.
func (a *App) DeckAusSet(ctx context.Context, setID, name string, menge int) DeckAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return DeckAntwort{Fehler: befund}
	}

	deck, err := a.setdeck.Anlegen(ctx, kennung, name, menge)
	if err != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return einDeck(deck)
}

// DeckAusSet — die Fensterbindung.
func (b *Bindungen) DeckAusSet(ctx context.Context, setID, name string, menge int) DeckAntwort {
	return b.app(ctx).DeckAusSet(b.ruf(ctx), setID, name, menge)
}

// DeckAusSet — für die Flächen ohne Fenster.
func (d *Deckflaeche) DeckAusSet(ctx context.Context, setID, name string, menge int) DeckAntwort {
	return d.app.DeckAusSet(ctx, setID, name, menge)
}
