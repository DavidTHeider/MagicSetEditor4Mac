package domain

// **Der Gast setzt sich mit seinem Deck** — P9b
// (#2377).
//
// # Warum genau eines von zweien
//
// **Ein eigenes Deck** bringt der Gast als Abzug mit; seine Kartenkennungen
// zeigen in **seinen** Bestand, und die Bilder zeichnet er selbst (Vermerk
// 0198). **Eine Leihe des Wirts** nennt er nur bei ihrer Kennung: der Wirt
// nimmt dann den Abzug, den er selbst ausgestellt und unterschrieben hat. Ein
// mitgeschickter Abzug, der behauptete, aus dem Bestand des Wirts zu stammen,
// brächte den Wirt dazu, beliebige Zeilen seines Bestands zu zeichnen — und
// sie dem Gast zu zeigen.

import "fmt"

// Mitgebracht ist, was ein eingeladener Gast beim Hinsetzen mitbringt.
type Mitgebracht struct {
	Deck  *Partiedeck `json:"deck,omitempty"`
	Leihe string      `json:"leihe,omitempty"`

	// Spiel ist das Spiel, zu dem das Deck gehört — ein Abzug nennt es
	// nicht, und ein Deck eines anderen Spiels gehört an keinen Tisch.
	Spiel string `json:"spiel,omitempty"`
}

// Pruefe sagt, ob es genau eines ist — und ob es eines ist, das trägt.
func (m Mitgebracht) Pruefe() error {
	switch {
	case m.Deck != nil && m.Leihe != "":
		return fmt.Errorf("%w: ein Deck oder eine Leihe, nicht beides",
			ErrValidation)
	case m.Deck != nil && m.Spiel == "":
		return fmt.Errorf("%w: ein mitgebrachtes Deck nennt sein Spiel",
			ErrValidation)
	case m.Deck != nil:
		return m.Deck.Pruefe()
	case m.Leihe == "":
		return fmt.Errorf("%w: mitgebracht ist weder ein Deck noch eine Leihe",
			ErrValidation)
	case len(m.Leihe) > Bildgrenze:
		return fmt.Errorf("%w: eine Leihe nennt höchstens %d Zeichen",
			ErrValidation, Bildgrenze)
	}

	return nil
}

// Platznehmen setzt den Eingeladenen mit seinem Deck an seinen Sitz.
//
// **Die Bibliothek kommt von außen**, wie beim Anlegen: welche Objekte ein
// Abzug ergibt, rechnet der Aufbau in `service`, und eine zweite Rechnung hier
// liefe beim ersten neuen Feld auseinander.
func (p Partie) Platznehmen(
	sitz, peer string, deck Partiedeck, bibliothek []Spielobjekt,
) (Partie, error) {
	gemeint := p.SitzVon(sitz)

	switch {
	case gemeint == nil:
		return Partie{}, fmt.Errorf("%w: %q sitzt nicht an diesem Tisch",
			ErrValidation, sitz)
	case gemeint.Eingeladen != peer:
		return Partie{}, fmt.Errorf(
			"%w: an diesem Sitz ist jemand anderes eingeladen", ErrConflict)
	case !gemeint.Wartet():
		return Partie{}, fmt.Errorf(
			"%w: an diesem Sitz liegt schon ein Deck", ErrConflict)
	}

	if fehler := deck.Pruefe(); fehler != nil {
		return Partie{}, fehler
	}

	danach := p
	danach.Sitze = append([]Sitz(nil), p.Sitze...)
	danach.SitzVon(sitz).Deck = deck
	danach.Objekte = append(append([]Spielobjekt{}, p.Objekte...), bibliothek...)

	return danach, nil
}
