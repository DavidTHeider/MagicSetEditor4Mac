package app

// **Ein Deck als Liste** — NET-04
// (#1708).
// Die Gestalt steht in [service.Deckliste]; wer ein Deck teilt, teilt diesen
// Text.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// DecklisteAntwort trägt ein Deck als Liste.
type DecklisteAntwort struct {
	Liste  string  `json:"liste"`
	Fehler *Fehler `json:"fehler"`
}

// DecklisteAusgeben gibt ein Deck in der Gestalt, die die Einfuhr versteht.
//
// **Sie schreibt nichts** — die Liste geht an den, der fragt.
func (a *App) DecklisteAusgeben(ctx context.Context, deckID string) DecklisteAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return DecklisteAntwort{Fehler: befund}
	}

	deck, err := a.decks.Deck(ctx, kennung)
	if err != nil {
		return DecklisteAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return DecklisteAntwort{Liste: service.Deckliste(deck)}
}

// Deckliste gibt ein Deck als Liste — für MCP und REST (NET-04, #1708).
func (d *Deckwerkstatt) Deckliste(ctx context.Context, deckID int64) (any, error) {
	antwort := d.flaeche().DecklisteAusgeben(ctx, zahl64(deckID))

	return ausMCP(map[string]any{"liste": antwort.Liste}, antwort.Fehler)
}
