// **Das Spielpaket** — welche Felder eine Karte und ein Set haben.
//
// **Aus `karte.go` herausgelöst**, als es über 400 Zeilen ging (§2). Der
// Schnitt folgt der Frage: dort steht, was ein **Feld** und eine **Karte**
// sind, hier, was ein **Spiel** ist, das beides erklärt.

package domain

import "strings"

// Game ist ein Spielpaket: es sagt, welche Felder eine Karte hat.
type Game struct {
	Name   string
	Fields []Field

	// SetFields sind die `set field:`-Blöcke — was ein **Set** dieses Spiels
	// an Angaben führt: Titel, Symbol, Copyright, Setcode.
	//
	// **Getrennt von [Game.Fields] und nicht darunter gemischt**: ein Set hat
	// ein Feld `title`, eine Karte hat eines `name`, und beide heißen im
	// Skript verschieden (`set.title` gegen `card.name`). Eine gemeinsame
	// Liste ließe die Suche nach `title` an einer Karte fündig werden.
	SetFields []Field

	// Skript ist das `init script:` des Spielpakets, samt seinen Modulen.
	//
	// **Es steht hier, weil die Feldskripte darauf aufsetzen.** `full name`
	// heißt `script: card_full_name()`, `cmc` heißt `script:
	// cmc(card.casting_cost)` — beides Funktionen, die das Spiel selbst
	// anlegt. Ohne sie rechnet keines der beiden
	// (#1238).
	//
	// **Ein Text und kein gerechneter Zustand**: wer ihn braucht, wertet ihn
	// in seiner eigenen Umgebung aus. `domain` kennt keinen Interpreter (§2).
	Skript string
}

// Field sucht eine Felddefinition. Der zweite Rückgabewert sagt, ob es sie gibt.
func (g Game) Field(name string) (Field, bool) {
	for i := range g.Fields {
		if g.Fields[i].Name == name {
			return g.Fields[i], true
		}
	}

	return Field{}, false
}

// Validate prüft ein Spiel samt seiner Felder.
func (g Game) Validate() error {
	if strings.TrimSpace(g.Name) == "" {
		return NewFieldError("name", SchluesselPflichtfeld, "Ein Spiel braucht einen Namen.")
	}

	gesehen := make(map[string]bool, len(g.Fields))

	for i := range g.Fields {
		feld := &g.Fields[i]

		if err := feld.Validate(); err != nil {
			return err
		}

		if gesehen[feld.Name] {
			return NewFieldError("name", SchluesselDoppelt,
				"Das Feld %q ist zweimal definiert.", feld.Name)
		}

		gesehen[feld.Name] = true
	}

	return nil
}

// Style ist ein Stilpaket: es sagt, wie eine Karte aussieht.
type Style struct {
	Name   string
	Game   string
	Width  int
	Height int
	DPI    int
}

// Validate prüft einen Stil.
func (s Style) Validate() error {
	switch {
	case strings.TrimSpace(s.Name) == "":
		return NewFieldError("name", SchluesselPflichtfeld, "Ein Stil braucht einen Namen.")
	case strings.TrimSpace(s.Game) == "":
		return NewFieldError("game", SchluesselPflichtfeld,
			"Der Stil %q nennt kein Spiel.", s.Name)
	case s.Width <= 0 || s.Height <= 0:
		return NewFieldError("size", SchluesselValidierung,
			"Der Stil %q hat keine brauchbaren Kartenmaße (%d × %d).", s.Name, s.Width, s.Height)
	case s.DPI <= 0:
		return NewFieldError("dpi", SchluesselValidierung,
			"Der Stil %q nennt keine Auflösung.", s.Name)
	}

	return nil
}
