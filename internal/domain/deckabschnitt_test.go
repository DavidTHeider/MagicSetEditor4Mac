package domain_test

// **Welche Überschrift welchen Bereich meint** — Werkstück `T8` von Phase
// 5.6 (#1473).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jede Überschrift, die ein Deckbaukasten schreibt, hat eine Antwort.**
//
// **Und drei davon lauten „gehört nicht ins Deck".** Ein `maybeboard` ist
// eine Merkliste, `tokens` sind Spielmarken, `about` ist Text — wer sie ins
// Hauptdeck legte, machte aus einer Notiz eine Karte im Deck.
func TestJedeUeberschriftHatEinenBereich(t *testing.T) {
	t.Parallel()

	for abschnitt, fall := range map[string]struct {
		bereich domain.Deckbereich
		gehoert bool
	}{
		"":           {domain.BereichHauptdeck, true},
		"deck":       {domain.BereichHauptdeck, true},
		"commander":  {domain.BereichHauptdeck, true},
		"companion":  {domain.BereichHauptdeck, true},
		"sideboard":  {domain.BereichSideboard, true},
		"maybeboard": {"", false},
		"tokens":     {"", false},
		"about":      {"", false},
	} {
		t.Run("»"+abschnitt+"«", func(t *testing.T) {
			t.Parallel()

			bereich, gehoert := domain.DeckbereichAus(abschnitt)
			if gehoert != fall.gehoert {
				t.Fatalf("gehört ins Deck: %v, erwartet %v", gehoert, fall.gehoert)
			}

			if bereich != fall.bereich {
				t.Errorf("Bereich %q, erwartet %q", bereich, fall.bereich)
			}
		})
	}
}

// **Eine unbekannte Überschrift ist das Hauptdeck.**
//
// **Und nicht „gehört nicht dazu".** Wer eine Liste mit einer selbst
// geschriebenen Überschrift einfügt — `Creatures`, `Lands` —, meint damit
// sein Deck und keine Merkliste; wer sie ausließe, ließe die halbe Liste
// aus. **Ausgeschlossen wird nur, was benannt ist**, und die Namen stehen im
// Lauf darüber.
func TestEineUnbekannteUeberschriftIstDasHauptdeck(t *testing.T) {
	t.Parallel()

	for _, abschnitt := range []string{"creatures", "lands", "spells"} {
		bereich, gehoert := domain.DeckbereichAus(abschnitt)
		if !gehoert || bereich != domain.BereichHauptdeck {
			t.Errorf("%q gibt %q/%v", abschnitt, bereich, gehoert)
		}
	}
}
