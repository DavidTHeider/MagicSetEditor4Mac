// **Die gezählten Deckantworten** — Werkstück 7 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, damit `vorrichtungzaehler_test.go` unter der
// Zeilengrenze aus §2 bleibt. Jede Methode zählt ihren Ruf und reicht durch —
// ohne sie sagte der Wächter „0mal gerufen" über einen Weg, der geht.

package rest_test

import "context"

func (z *zaehlender) Decks(ctx context.Context, spiel string) (any, error) {
	z.zaehle()

	return z.williger.Decks(ctx, spiel)
}

func (z *zaehlender) Deck(ctx context.Context, deckID int64) (any, error) {
	z.zaehle()

	return z.williger.Deck(ctx, deckID)
}

func (z *zaehlender) DeckAnlegen(ctx context.Context, name, spiel, notiz string) (
	any, error,
) {
	z.zaehle()

	return z.williger.DeckAnlegen(ctx, name, spiel, notiz)
}

func (z *zaehlender) DeckUmbenennen(
	ctx context.Context, deckID, fassung int64, name, notiz string,
) (any, error) {
	z.zaehle()

	return z.williger.DeckUmbenennen(ctx, deckID, fassung, name, notiz)
}

func (z *zaehlender) DeckFehlkartenWuenschen(ctx context.Context, deckID int64) (any, error) {
	z.zaehle()

	return z.williger.DeckFehlkartenWuenschen(ctx, deckID)
}

func (z *zaehlender) DeckAusSet(ctx context.Context, setID int64, name string, menge int) (any, error) {
	z.zaehle()

	return z.williger.DeckAusSet(ctx, setID, name, menge)
}

func (z *zaehlender) DeckKopieren(ctx context.Context, deckID int64, name string) (
	any, error,
) {
	z.zaehle()

	return z.williger.DeckKopieren(ctx, deckID, name)
}

func (z *zaehlender) DeckLoeschen(ctx context.Context, deckID int64) error {
	z.zaehle()

	return z.williger.DeckLoeschen(ctx, deckID)
}

func (z *zaehlender) DeckplatzLegen(
	ctx context.Context, deckID, karteID int64, printing string, menge int, bereich string,
) (any, error) {
	z.zaehle()

	return z.williger.DeckplatzLegen(ctx, deckID, karteID, printing, menge, bereich)
}

func (z *zaehlender) DeckplatzNehmen(ctx context.Context, deckID, platzID int64) error {
	z.zaehle()

	return z.williger.DeckplatzNehmen(ctx, deckID, platzID)
}

func (z *zaehlender) DeckplatzVerschieben(
	ctx context.Context, deckID, platzID int64, bereich string,
) (any, error) {
	z.zaehle()

	return z.williger.DeckplatzVerschieben(ctx, deckID, platzID, bereich)
}

func (z *zaehlender) DeckFormatWaehlen(ctx context.Context, deckID, formatID int64) (
	any, error,
) {
	z.zaehle()

	return z.williger.DeckFormatWaehlen(ctx, deckID, formatID)
}

func (z *zaehlender) DeckPruefen(ctx context.Context, deckID int64) (any, error) {
	z.zaehle()

	return z.williger.DeckPruefen(ctx, deckID)
}

func (z *zaehlender) DeckAbgleichen(ctx context.Context, deckID int64) (any, error) {
	z.zaehle()

	return z.williger.DeckAbgleichen(ctx, deckID)
}

func (z *zaehlender) DeckZuordnungen(ctx context.Context, deckID int64) (any, error) {
	z.zaehle()

	return z.williger.DeckZuordnungen(ctx, deckID)
}

func (z *zaehlender) DeckZuordnen(
	ctx context.Context, deckID, platzID, exemplarID int64, menge int,
) (any, error) {
	z.zaehle()

	return z.williger.DeckZuordnen(ctx, deckID, platzID, exemplarID, menge)
}

func (z *zaehlender) DeckZuordnungLoesen(ctx context.Context, deckID, zuordnungID int64) error {
	z.zaehle()

	return z.williger.DeckZuordnungLoesen(ctx, deckID, zuordnungID)
}

func (z *zaehlender) Deckliste(ctx context.Context, deckID int64) (any, error) {
	z.zaehle()

	return z.williger.Deckliste(ctx, deckID)
}

func (z *zaehlender) DeckFehlkarten(ctx context.Context, deckID int64) (any, error) {
	z.zaehle()

	return z.williger.DeckFehlkarten(ctx, deckID)
}

func (z *zaehlender) ExemplarDecks(ctx context.Context, exemplarID int64) (any, error) {
	z.zaehle()

	return z.williger.ExemplarDecks(ctx, exemplarID)
}

func (z *zaehlender) Bauarten(ctx context.Context) (any, error) {
	z.zaehle()

	return z.williger.Bauarten(ctx)
}

func (z *zaehlender) Formate(ctx context.Context, spiel string) (any, error) {
	z.zaehle()

	return z.williger.Formate(ctx, spiel)
}

func (z *zaehlender) FormatAusBauart(
	ctx context.Context, bauart, name, spiel, scryfallname string,
) (any, error) {
	z.zaehle()

	return z.williger.FormatAusBauart(ctx, bauart, name, spiel, scryfallname)
}

func (z *zaehlender) FormatSetzen(ctx context.Context, format map[string]any) (
	any, error,
) {
	z.zaehle()

	return z.williger.FormatSetzen(ctx, format)
}

func (z *zaehlender) FormatLoeschen(ctx context.Context, formatID int64) error {
	z.zaehle()

	return z.williger.FormatLoeschen(ctx, formatID)
}
