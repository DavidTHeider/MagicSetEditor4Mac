package app

// **Die Plätze eines Decks und seine Formatwahl** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `deck.go` unter der Zeilengrenze aus §2 bleibt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// DeckplatzAntwort trägt einen gelegten Platz.
type DeckplatzAntwort struct {
	Platz  *DeckplatzDTO `json:"platz"`
	Fehler *Fehler       `json:"fehler"`
}

// DeckplatzLegen legt einen Platz ins Deck oder schreibt seine Menge fort.
//
// **Ein Weg für beide Herkünfte**: `karteID` nennt eine eigene Karte,
// `printing` einen echten Druck. Welche von beiden gilt, entscheidet der
// Dienst — hier wird nur umgerechnet.
//
// **Eine leere `karteID` ist die Null und kein Fehler.** Das Fenster schickt
// für einen Druck keine Kartenkennung mit, und ein Befund darüber wäre einer
// über eine Angabe, die niemand gemacht hat.
func (a *App) DeckplatzLegen(
	ctx context.Context, deckID, karteID, printing string, menge int, bereich string,
) DeckplatzAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return DeckplatzAntwort{Fehler: befund}
	}

	platz := domain.Deckplatz{
		Printing: printing,
		Menge:    menge,
		Bereich:  domain.Deckbereich(bereich),
	}

	if karteID != "" {
		karte, kartenbefund := benannteKennungAus("karteID", karteID)
		if kartenbefund != nil {
			return DeckplatzAntwort{Fehler: kartenbefund}
		}

		platz.KarteID = karte
	}

	gelegt, err := a.decks.PlatzLegen(ctx, kennung, platz)
	if err != nil {
		return DeckplatzAntwort{Fehler: Uebersetze(ctx, err)}
	}

	dto := alsDeckplatz(gelegt)

	return DeckplatzAntwort{Platz: &dto}
}

// DeckplatzNehmen nimmt einen Platz aus dem Deck.
func (a *App) DeckplatzNehmen(ctx context.Context, deckID, platzID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	platz, platzbefund := benannteKennungAus("platzID", platzID)
	if platzbefund != nil {
		return SchreibAntwort{Fehler: platzbefund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.decks.PlatzNehmen(ctx, kennung, platz))}
}

// DeckplatzVerschieben legt einen Platz in den anderen Bereich.
//
// **Ein Ruf und nicht zwei** — Befund 7 des Zweitblicks über Phase 5.1
// (#1405).
// Das Fenster hat den Wechsel bis zum 2026-09-10 aus `DeckplatzLegen` und
// `DeckplatzNehmen` zusammengesetzt; scheiterte der zweite, lag dieselbe
// Karte in **beiden** Bereichen, und §7 Regel 3 nimmt nach einem
// gescheiterten Schreiben nichts zurück.
func (a *App) DeckplatzVerschieben(
	ctx context.Context, deckID, platzID, bereich string,
) DeckplatzAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return DeckplatzAntwort{Fehler: befund}
	}

	platz, platzbefund := benannteKennungAus("platzID", platzID)
	if platzbefund != nil {
		return DeckplatzAntwort{Fehler: platzbefund}
	}

	verschoben, err := a.decks.PlatzVerschieben(
		ctx, kennung, platz, domain.Deckbereich(bereich))
	if err != nil {
		return DeckplatzAntwort{Fehler: Uebersetze(ctx, err)}
	}

	dto := alsDeckplatz(verschoben)

	return DeckplatzAntwort{Platz: &dto}
}

// DeckFormatWaehlen stellt ein Deck auf ein Format — **leer nimmt es
// herunter**.
//
// Ein Deck ohne Format ist **ungeprüft**, nicht falsch.
func (a *App) DeckFormatWaehlen(ctx context.Context, deckID, formatID string) DeckAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return DeckAntwort{Fehler: befund}
	}

	var format int64

	if formatID != "" {
		gewaehlt, formatbefund := benannteKennungAus("formatID", formatID)
		if formatbefund != nil {
			return DeckAntwort{Fehler: formatbefund}
		}

		format = gewaehlt
	}

	deck, err := a.decks.FormatWaehlen(ctx, kennung, format)
	if err != nil {
		return DeckAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return einDeck(deck)
}
