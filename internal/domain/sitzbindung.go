package domain

import (
	"fmt"
	"strings"
)

// Sitzbindung bindet ein Zugriffstoken an **einen** Sitz einer Partie —
// #1819,
// Befund 2 des Zweitblicks über Phase 6.1.
//
// # Wozu
//
// **Ein Werkzeug, das an einem Sitz spielt, soll nur diesen Sitz sehen.**
// Den Sitz einer Sicht nennt sonst der Rufer (G8, #1465) — das ist richtig,
// solange ein Mensch am Rechner sitzt, und eine Einladung zum Mitlesen der
// fremden Hand, sobald ein Sprachmodell über MCP gegen einen Menschen am
// selben Tisch spielt.
//
// # Was sie verlangt
//
// **Beides oder keines**: eine Partie ohne Sitz bände an nichts Bestimmtes,
// ein Sitz ohne Partie an einen Namen, den es in jeder Partie geben kann.
// Eine leere Bindung ist ein Token wie bisher.
type Sitzbindung struct {
	Partie string
	Sitz   string
}

// Leer sagt, ob das Token an keinen Sitz gebunden ist.
func (b Sitzbindung) Leer() bool {
	return b.Partie == "" && b.Sitz == ""
}

// Validate weist eine halbe Bindung ab.
func (b Sitzbindung) Validate() error {
	partie, sitz := strings.TrimSpace(b.Partie), strings.TrimSpace(b.Sitz)

	switch {
	case partie == "" && sitz != "":
		return NewFieldError("partie", SchluesselPflichtfeld,
			"Ein Sitz gehört zu einer Partie — nenne sie mit.")
	case partie != "" && sitz == "":
		return NewFieldError("sitz", SchluesselPflichtfeld,
			"Eine Bindung an eine Partie nennt den Sitz, an dem das Token spielt.")
	case partie != b.Partie || sitz != b.Sitz:
		return NewFieldError("sitz", SchluesselValidierung,
			"Partie und Sitz tragen keine Leerzeichen am Rand.")
	}

	return nil
}

// Erlaubt sagt, ob ein Ruf über diese Partie und diesen Sitz gehen darf.
//
// **Ein ungebundenes Token darf alles, was es bisher durfte** — die Bindung
// verengt, sie gibt nichts dazu.
func (b Sitzbindung) Erlaubt(partie, sitz string) error {
	if b.Leer() {
		return nil
	}

	if partie != b.Partie {
		return fmt.Errorf("%w: dieses Token spielt in einer anderen Partie", ErrErlaubnis)
	}

	if sitz != b.Sitz {
		return fmt.Errorf("%w: dieses Token sieht nur den Sitz %q", ErrErlaubnis, b.Sitz)
	}

	return nil
}

// ErlaubtPartie sagt, ob ein Ruf über diese Partie gehen darf — für die
// Wege, die keinen Sitz nennen.
func (b Sitzbindung) ErlaubtPartie(partie string) error {
	if b.Leer() || partie == b.Partie {
		return nil
	}

	return fmt.Errorf("%w: dieses Token spielt in einer anderen Partie", ErrErlaubnis)
}

// ErlaubtVollenZustand sagt, ob der **ganze** Zustand hinausgehen darf — die
// Ausfuhr, die Kopie, das Löschen. **Ein gebundenes Token darf das nie**:
// die Ausfuhr trägt jede Hand, und genau die soll es nicht sehen.
func (b Sitzbindung) ErlaubtVollenZustand() error {
	if b.Leer() {
		return nil
	}

	return fmt.Errorf("%w: ein an einen Sitz gebundenes Token bekommt nicht den ganzen "+
		"Zustand der Partie", ErrErlaubnis)
}
