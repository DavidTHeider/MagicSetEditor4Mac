package app

import "context"

// **Die Bindungen für Decks und Formate** — Werkstück 7 von Phase 4.6
// (#1293).
//
// Sie hängen den Kontext des Fensters an und reichen weiter; entschieden wird
// eine Schicht tiefer (§2).

func (b *Bindungen) Decks(ctx context.Context, spiel string) DecksAntwort {
	return b.app(ctx).Decks(b.ruf(ctx), spiel)
}

func (b *Bindungen) Deck(ctx context.Context, deckID string) DeckAntwort {
	return b.app(ctx).Deck(b.ruf(ctx), deckID)
}

func (b *Bindungen) DeckAnlegen(ctx context.Context, name, spiel, notiz string) DeckAntwort {
	return b.app(ctx).DeckAnlegen(b.ruf(ctx), name, spiel, notiz)
}

func (b *Bindungen) DeckUmbenennen(ctx context.Context, deckID, fassung, name, notiz string) DeckAntwort {
	return b.app(ctx).DeckUmbenennen(b.ruf(ctx), deckID, fassung, name, notiz)
}

func (b *Bindungen) DeckKopieren(ctx context.Context, deckID, name string) DeckAntwort {
	return b.app(ctx).DeckKopieren(b.ruf(ctx), deckID, name)
}

func (b *Bindungen) DeckFehlkartenWuenschen(ctx context.Context, deckID string) FehlkartenwunschAntwort {
	return b.app(ctx).DeckFehlkartenWuenschen(b.ruf(ctx), deckID)
}

// DeckLoeschen nimmt ein Deck samt seinen Plätzen heraus.
//
// **Ohne Rückfrage hier** — die gehört ins Fenster, wo jemand sie sieht.
func (b *Bindungen) DeckLoeschen(ctx context.Context, deckID string) SchreibAntwort {
	return b.app(ctx).DeckLoeschen(b.ruf(ctx), deckID)
}

func (b *Bindungen) DeckplatzLegen(
	ctx context.Context,
	deckID, karteID, printing string, menge int, bereich string,
) DeckplatzAntwort {
	return b.app(ctx).DeckplatzLegen(
		b.ruf(ctx), deckID, karteID, printing, menge, bereich)
}

func (b *Bindungen) DeckplatzNehmen(ctx context.Context, deckID, platzID string) SchreibAntwort {
	return b.app(ctx).DeckplatzNehmen(b.ruf(ctx), deckID, platzID)
}

// DeckplatzVerschieben legt einen Platz in den anderen Bereich — in einem Zug.
func (b *Bindungen) DeckplatzVerschieben(ctx context.Context, deckID, platzID, bereich string) DeckplatzAntwort {
	return b.app(ctx).DeckplatzVerschieben(b.ruf(ctx), deckID, platzID, bereich)
}

// DeckFormatWaehlen stellt ein Deck auf ein Format — leer nimmt es herunter.
func (b *Bindungen) DeckFormatWaehlen(ctx context.Context, deckID, formatID string) DeckAntwort {
	return b.app(ctx).DeckFormatWaehlen(b.ruf(ctx), deckID, formatID)
}

// DeckPruefen beurteilt ein Deck gegen sein gewähltes Format.
func (b *Bindungen) DeckPruefen(ctx context.Context, deckID string) DeckurteilAntwort {
	return b.app(ctx).DeckPruefen(b.ruf(ctx), deckID)
}

// DeckAbgleichen sagt, was von einem Deck in der Sammlung liegt.
func (b *Bindungen) DeckAbgleichen(ctx context.Context, deckID string) HabenstandAntwort {
	return b.app(ctx).DeckAbgleichen(b.ruf(ctx), deckID)
}

// DeckZuordnungen gibt den Bedarf eines Decks samt seinen Zuordnungen (U9, #1420).
func (b *Bindungen) DeckZuordnungen(ctx context.Context, deckID string) ZuordnungsstandAntwort {
	return b.app(ctx).DeckZuordnungen(b.ruf(ctx), deckID)
}

// DeckZuordnen legt Exemplare der Sammlung in einen Deckplatz.
func (b *Bindungen) DeckZuordnen(ctx context.Context, deckID, platzID, exemplarID string, menge int) ZuordnungAntwort {
	return b.app(ctx).DeckZuordnen(b.ruf(ctx), deckID, platzID, exemplarID, menge)
}

// DeckZuordnungLoesen nimmt eine Zuordnung aus ihrem Deck.
func (b *Bindungen) DeckZuordnungLoesen(ctx context.Context, deckID, zuordnungID string) SchreibAntwort {
	return b.app(ctx).DeckZuordnungLoesen(b.ruf(ctx), deckID, zuordnungID)
}

// DecklisteAusgeben gibt ein Deck als Liste (NET-04, #1708).
func (b *Bindungen) DecklisteAusgeben(ctx context.Context, deckID string) DecklisteAntwort {
	return b.app(ctx).DecklisteAusgeben(b.ruf(ctx), deckID)
}

// DeckFehlkarten gibt, was zu besorgen wäre — als Liste.
func (b *Bindungen) DeckFehlkarten(ctx context.Context, deckID string) FehlkartenAntwort {
	return b.app(ctx).DeckFehlkarten(b.ruf(ctx), deckID)
}

// ExemplarDecks sagt, welche Decks ein Exemplar benutzen — der Rückweg.
func (b *Bindungen) ExemplarDecks(ctx context.Context, exemplarID string) DeckbezuegeAntwort {
	return b.app(ctx).ExemplarDecks(b.ruf(ctx), exemplarID)
}

// Bauarten gibt die drei Bauarten samt ihren nachgelesenen Zahlen.
func (b *Bindungen) Bauarten(ctx context.Context) BauartenAntwort {
	return b.app(ctx).Bauarten()
}

func (b *Bindungen) Formate(ctx context.Context, spiel string) FormateAntwort {
	return b.app(ctx).Formate(b.ruf(ctx), spiel)
}

// FormatAusBauart legt ein Format mit den Zahlen einer Bauart an.
func (b *Bindungen) FormatAusBauart(
	ctx context.Context,
	bauart, name, spiel, scryfallname string,
) FormatAntwort {
	return b.app(ctx).FormatAusBauart(b.ruf(ctx), bauart, name, spiel, scryfallname)
}

// FormatSetzen legt ein Format an oder schreibt es fort — die Kennung
// entscheidet.
func (b *Bindungen) FormatSetzen(ctx context.Context, format FormatDTO) FormatAntwort {
	return b.app(ctx).FormatSetzen(b.ruf(ctx), format)
}

func (b *Bindungen) FormatLoeschen(ctx context.Context, formatID string) SchreibAntwort {
	return b.app(ctx).FormatLoeschen(b.ruf(ctx), formatID)
}
