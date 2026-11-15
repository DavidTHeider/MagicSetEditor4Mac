package domain_test

// **Der Stand eines Decks** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Wogegen er steht
//
// **Gegen eine Einfuhr, die einen Stand übernimmt, den sie nie gesehen hat.**
// Zwischen Vorschau und Bestätigung kann jemand eine Karte hineinlegen,
// herausnehmen oder verschieben; wer dann trotzdem schreibt, legt eine
// Deckliste über eine andere. Das Issue sagt es: „Deckimport bei parallel
// geändertem Deck scheitert kontrolliert."
//
// # Warum [domain.Deck.Fassung] dafür nicht reicht
//
// **Weil sie die Plätze ausdrücklich nicht deckt** — es steht an ihr:
// `PlatzSetzen` schreibt in einem Zug und hebt keine Fassung. Ein Riegel, der
// nur sie prüfte, hielte die halbe Zusage und machte die andere unsichtbar
// (§10).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// probedeck ist ein Deck mit zwei Plätzen.
func probedeck() domain.Deck {
	return domain.Deck{
		ID: 7, Fassung: 3, Name: "Probe", Spiel: "magic",
		Plaetze: []domain.Deckplatz{
			{
				ID: 1, Printing: "11111111-1111-1111-1111-111111111111",
				Menge: 4, Bereich: domain.BereichHauptdeck,
			},
			{ID: 2, KarteID: 9, Menge: 2, Bereich: domain.BereichSideboard},
		},
	}
}

// **Derselbe Stand gibt denselben Abdruck.**
func TestDerDeckstandIstFuerDenselbenStandDerselbe(t *testing.T) {
	t.Parallel()

	erster, zweiter := domain.Deckstand(probedeck()), domain.Deckstand(probedeck())

	if erster != zweiter {
		t.Errorf("zweimal dasselbe Deck gibt %q und %q", erster, zweiter)
	}

	if erster == "" {
		t.Error("der Abdruck ist leer")
	}
}

// **Und die Reihenfolge der Plätze ändert ihn nicht.**
//
// **Weil sie keine Eigenschaft des Decks ist**, sondern eine der Abfrage. Wer
// sie mitzählte, ließe eine Einfuhr scheitern, weil jemand anders sortiert
// hat.
func TestDieReihenfolgeDerPlaetzeAendertDenDeckstandNicht(t *testing.T) {
	t.Parallel()

	gedreht := probedeck()
	gedreht.Plaetze = []domain.Deckplatz{gedreht.Plaetze[1], gedreht.Plaetze[0]}

	if domain.Deckstand(gedreht) != domain.Deckstand(probedeck()) {
		t.Error("eine andere Reihenfolge gibt einen anderen Abdruck")
	}
}

// **Jede Änderung, die zählt, ändert ihn.**
//
// **Die Gegenprobe zum Test darüber**: ein Abdruck, der immer gleich bliebe,
// wäre auch bei gleicher Reihenfolge gleich — und sagte nichts (§10).
func TestJedeAenderungAmDeckAendertDenStand(t *testing.T) {
	t.Parallel()

	for name, aendern := range map[string]func(*domain.Deck){
		"eine höhere Fassung": func(d *domain.Deck) { d.Fassung++ },
		"ein Platz mehr": func(d *domain.Deck) {
			d.Plaetze = append(d.Plaetze, domain.Deckplatz{ID: 3, KarteID: 4, Menge: 1, Bereich: domain.BereichHauptdeck})
		},
		"ein Platz weniger":   func(d *domain.Deck) { d.Plaetze = d.Plaetze[:1] },
		"eine andere Menge":   func(d *domain.Deck) { d.Plaetze[0].Menge = 3 },
		"ein anderer Bereich": func(d *domain.Deck) { d.Plaetze[0].Bereich = domain.BereichSideboard },
		"ein anderer Druck":   func(d *domain.Deck) { d.Plaetze[0].Printing = "22222222-2222-2222-2222-222222222222" },
		"eine andere Karte":   func(d *domain.Deck) { d.Plaetze[1].KarteID = 10 },
		"eine andere Kennung": func(d *domain.Deck) { d.Plaetze[1].ID = 99 },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			geaendert := probedeck()
			aendern(&geaendert)

			if domain.Deckstand(geaendert) == domain.Deckstand(probedeck()) {
				t.Errorf("%s ändert den Abdruck nicht", name)
			}
		})
	}
}

// **Was ihn nicht ändert, steht auch da.**
//
// **Der Name und die Notiz gehören zur Fassung** und nicht zu den Plätzen;
// wer sie mitzählte, ließe eine Einfuhr scheitern, weil jemand das Deck
// umbenannt hat — und dagegen steht schon [domain.Deck.Fassung].
//
// **Und die Bezeichnung eines Platzes kommt von außen** (§16, Regel 5): ohne
// Netz bleibt sie leer, mit Netz steht sie da. Ein Abdruck, der davon
// abhinge, wäre bei jedem zweiten Aufruf ein anderer.
func TestWasNichtZumStandGehoertAendertIhnNicht(t *testing.T) {
	t.Parallel()

	for name, aendern := range map[string]func(*domain.Deck){
		"eine Bezeichnung": func(d *domain.Deck) { d.Plaetze[0].Bezeichnung = "Blitzschlag" },
		"Manakosten":       func(d *domain.Deck) { d.Plaetze[0].Manakosten = "{R}" },
		"gezählte Karten":  func(d *domain.Deck) { d.Hauptdeck, d.Sideboard = 40, 15 },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			geaendert := probedeck()
			aendern(&geaendert)

			if domain.Deckstand(geaendert) != domain.Deckstand(probedeck()) {
				t.Errorf("%s ändert den Abdruck", name)
			}
		})
	}
}
