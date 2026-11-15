package domain

// **Ein Kartenbild über die Leitung** — P9
// (#2377),
// Befund C-12, Vermerk 0198.
//
// # Warum nach einem Objekt gefragt wird und nicht nach einer Karte
//
// **Weil eine Kartenkennung eine Zeile eines Bestands ist.** Der Fragende
// hat keine — er sieht die Bildkennung — und wer antwortet, entscheidet am
// Objekt, ob der Fragende es sehen darf. Eine Frage nach einer Zeile wäre eine
// Frage, die jede Karte des fremden Bestands herausgäbe.
//
// # Warum eine Antwort ein Bild **oder** einen Grund trägt
//
// **Weil beides eine Antwort ist.** Ein verdecktes Objekt, ein echter Druck,
// ein Deck aus einem dritten Bestand: der Fragende soll wissen, dass er keines
// bekommt, statt zu warten.

import "fmt"

// Bildgrenze ist die längste Kennung, die ein Bildwunsch nennen darf.
const Bildgrenze = 128

// Bildwunsch fragt nach der Seite eines Objekts in einer Güte.
type Bildwunsch struct {
	Objekt string `json:"objekt"`
	Seite  string `json:"seite"`
	Guete  string `json:"guete"`

	// Bild ist die Bildkennung, unter der der Fragende das Bild ablegt
	// (Nachlauf zum Zweitblick über Phase 7, R2) — wer zeichnet, prüft, dass
	// sie die des Objekts ist. **Freiwillig**: leer prüft niemand.
	Bild string `json:"bild,omitempty"`
}

// Pruefe hält die Gestalt.
func (w Bildwunsch) Pruefe() error {
	for _, teil := range []struct{ name, wert string }{
		{"Objekt", w.Objekt}, {"Seite", w.Seite}, {"Güte", w.Guete},
	} {
		if teil.wert == "" || len(teil.wert) > Bildgrenze {
			return fmt.Errorf("%w: ein Bildwunsch braucht %s, kürzer als %d "+
				"Zeichen", ErrValidation, teil.name, Bildgrenze)
		}
	}

	if len(w.Bild) > Bildgrenze {
		return fmt.Errorf("%w: eine Bildkennung ist kürzer als %d Zeichen",
			ErrValidation, Bildgrenze)
	}

	return nil
}

// Bildantwort ist die Antwort auf einen Bildwunsch.
type Bildantwort struct {
	Wunsch Bildwunsch `json:"wunsch"`

	// Png ist das Bild — leer, wenn ein Grund dasteht.
	Png    []byte `json:"png,omitempty"`
	Breite int    `json:"breite,omitempty"`
	Hoehe  int    `json:"hoehe,omitempty"`

	// Zweiseitig sagt, ob sich die Karte umdrehen lässt — wer ihr Bild
	// bekommt, darf das wissen.
	Zweiseitig bool `json:"zweiseitig,omitempty"`

	// Grund sagt, warum es kein Bild gibt.
	Grund string `json:"grund,omitempty"`
}

// Pruefe hält: ein Bild mit Maßen oder ein Grund — nie beides, nie keins.
func (a Bildantwort) Pruefe() error {
	if fehler := a.Wunsch.Pruefe(); fehler != nil {
		return fehler
	}

	switch {
	case len(a.Png) > 0 && a.Grund != "":
		return fmt.Errorf("%w: eine Bildantwort trägt ein Bild oder einen "+
			"Grund, nicht beides", ErrValidation)
	case len(a.Png) == 0 && a.Grund == "":
		return fmt.Errorf("%w: eine Bildantwort ohne Bild und ohne Grund "+
			"sagt nichts", ErrValidation)
	case len(a.Png) > 0 && (a.Breite <= 0 || a.Hoehe <= 0):
		return fmt.Errorf("%w: ein Bild ohne Maße", ErrValidation)
	}

	return nil
}
