// **Die Bindungen der Decklisteneinfuhr** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// **Eine eigene Datei neben `deckbindungen.go`**, weil sie eine andere Quelle
// braucht: das Deck kommt aus dem Bestand, die Deutung aus Scryfall. Der
// Schnitt folgt der Frage, nicht der Zeilenzahl (§2).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Deckeinfuhrquelle ist, was das Fenster für eine Deckliste braucht.
//
// **Als Schnittstelle und nicht als `*Deckeinfuhrnaht`** (§2): so kommt keine
// Bindung ohne Zuhause zum Absturz, und ein Lauf braucht kein Netz.
type Deckeinfuhrquelle interface {
	Vorschau(ctx context.Context, deckID int64, text string) (*Deckvorschau, error)
	Uebernehmen(ctx context.Context, deckID int64, stand string,
		zeilen []Deckeinfuhrzeile) (*Deckeinfuhrbericht, error)
}

// MitDeckeinfuhr setzt die Naht.
func (a *App) MitDeckeinfuhr(quelle Deckeinfuhrquelle) { a.deckeinfuhr = quelle }

// Deckvorschauantwort trägt eine Vorschau oder einen Fehler.
type Deckvorschauantwort struct {
	Vorschau *Deckvorschau `json:"vorschau"`
	Fehler   *Fehler       `json:"fehler"`
}

// Deckeinfuhrantwort trägt einen Bericht oder einen Fehler.
type Deckeinfuhrantwort struct {
	Bericht *Deckeinfuhrbericht `json:"bericht"`
	Fehler  *Fehler             `json:"fehler"`
}

// DecklisteVorschau liest eine Deckliste und zeigt, was ins Deck käme.
//
// **Sie schreibt nichts** (§16, Regel 6). Was daraus wird, entscheidet, wer
// hinsieht — „Counterspell" ohne Setangabe passt auf 72 Drucke, und geraten
// wird keiner.
func (a *App) DecklisteVorschau(
	ctx context.Context, deckID int64, text string,
) Deckvorschauantwort {
	ctx = a.imProfil(ctx)

	if a.deckeinfuhr == nil {
		return Deckvorschauantwort{Fehler: Uebersetze(ctx, keineDeckeinfuhr())}
	}

	vorschau, err := a.deckeinfuhr.Vorschau(ctx, deckID, text)
	if err != nil {
		return Deckvorschauantwort{Fehler: Uebersetze(ctx, err)}
	}

	return Deckvorschauantwort{Vorschau: vorschau}
}

// DecklisteUebernehmen legt die bestätigten Zeilen ins Deck.
//
// **`stand` kommt aus der Vorschau** und ist die Bedingung (§7, Regel 2): hat
// sich das Deck seitdem geändert, wird nichts gelegt.
//
// **Es entsteht dabei kein Besitz.** Ein Deck verweist auf Karten; wer sie
// haben will, trägt sie in die Sammlung ein — über `ListeEinlesen`, und das
// ist ein anderer Ruf mit einer anderen Wirkung.
func (a *App) DecklisteUebernehmen(
	ctx context.Context, deckID int64, stand string, zeilen []Deckeinfuhrzeile,
) Deckeinfuhrantwort {
	ctx = a.imProfil(ctx)

	if a.deckeinfuhr == nil {
		return Deckeinfuhrantwort{Fehler: Uebersetze(ctx, keineDeckeinfuhr())}
	}

	bericht, err := a.deckeinfuhr.Uebernehmen(ctx, deckID, stand, zeilen)
	if err != nil {
		return Deckeinfuhrantwort{Fehler: Uebersetze(ctx, err)}
	}

	return Deckeinfuhrantwort{Bericht: bericht}
}

// DecklisteVorschau zeigt, was aus einer Deckliste würde.
func (b *Bindungen) DecklisteVorschau(ctx context.Context, deckID int64, text string) Deckvorschauantwort {
	return b.app(ctx).DecklisteVorschau(b.ruf(ctx), deckID, text)
}

// DecklisteUebernehmen legt die bestätigten Zeilen ins Deck.
func (b *Bindungen) DecklisteUebernehmen(
	ctx context.Context,
	deckID int64, stand string, zeilen []Deckeinfuhrzeile,
) Deckeinfuhrantwort {
	return b.app(ctx).DecklisteUebernehmen(b.ruf(ctx), deckID, stand, zeilen)
}

// keineDeckeinfuhr ist der Fehler ohne eingerichtete Naht.
//
// **Sie kann fehlen**, und das ist kein Programmfehler: sie braucht ein
// Zuhause für den Zwischenspeicher von Scryfall. Die Bindung sagt es dann
// selbst, statt dass die Seite leer bleibt (§8).
func keineDeckeinfuhr() error {
	return fmt.Errorf("%w: die Decklisteneinfuhr ist nicht eingerichtet — sie "+
		"braucht ein Zuhause für den Zwischenspeicher von Scryfall",
		domain.ErrValidation)
}
