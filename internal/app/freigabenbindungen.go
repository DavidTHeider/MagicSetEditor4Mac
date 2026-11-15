// **Die Bindungen der Freigaben und der Deckübernahme** — Werkstück 11 und 12
// von Phase 4.12
// (#1351).

package app

import "context"

// Freigaben gibt die Freigaben eines Profils samt der eigenen Rolle.
//
// **Beides in einer Antwort**, weil die Seite an der Rolle entscheidet, was
// sie überhaupt anbietet (§7, Regel 2).
func (b *Bindungen) Freigaben(ctx context.Context, profilID string) FreigabenAntwort {
	return b.app(ctx).Freigaben(b.ruf(ctx), profilID)
}

// FreigabeSetzen legt eine Freigabe an oder schreibt ihre Rolle fort.
func (b *Bindungen) FreigabeSetzen(ctx context.Context, angaben FreigabeangabenDTO) FreigabenAntwort {
	return b.app(ctx).FreigabeSetzen(b.ruf(ctx), angaben)
}

// FreigabeLoeschen zieht eine Freigabe zurück.
func (b *Bindungen) FreigabeLoeschen(ctx context.Context, profilID, nutzerID string) FreigabenAntwort {
	return b.app(ctx).FreigabeLoeschen(b.ruf(ctx), profilID, nutzerID)
}

// DeckUebernehmen holt ein Deck aus einem fremden Profil in das eigene.
//
// **Hier und nicht bei den Deckbindungen** — Werkstück 12 (#1351): sie fasst
// **zwei** Profile an und verlangt dieselbe Rolle wie der Wechsel. Sie gehört
// zu den Freigaben, nicht zum Deckbau.
func (b *Bindungen) DeckUebernehmen(ctx context.Context, quellProfil, quellDeck string) Deckuebernahmeantwort {
	return b.app(ctx).DeckUebernehmen(b.ruf(ctx), quellProfil, quellDeck)
}

// DeckeZurUebernahme nennt die Decks eines fremden Profils — zur Wahl, hinter
// derselben Erlaubnis wie [Bindungen.DeckUebernehmen] (UX-08.AK-04, #1680).
func (b *Bindungen) DeckeZurUebernahme(ctx context.Context, quellProfil string) DecksAntwort {
	return b.app(ctx).DeckeZurUebernahme(b.ruf(ctx), quellProfil)
}
