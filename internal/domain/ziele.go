package domain

// **Ziele** — Werkstück M4 der Regelmaschine
// (#1606).
//
// # Zweimal geprüft, und das ist der Punkt
//
// **Beim Ankündigen** (CR 601.2c, CR 115.1) und **beim Verrechnen noch
// einmal** (CR 608.2b). Dazwischen kann alles geschehen: das Ziel wechselt
// die Zone, es verlässt das Spiel, es ändert seine Eigenschaften. Ein Zauber,
// dessen Ziele **alle** unzulässig geworden sind, wird nicht verrechnet — er
// verpufft.
//
// # Was hier nicht geprüft wird
//
// **Ob der Zauber überhaupt ein Ziel verlangt, und wie viele.** Das steht in
// seinem Text, und einen Text zu lesen ist `M8`
// (#1610).
// Geprüft wird, ob die **genannten** Ziele zulässig sind — nicht, ob sie die
// richtigen sind. Der Unterschied steht als Grenze in der Regelmatrix.

import "fmt"

// Ziel ist ein angesagtes Ziel samt der Zone, in der es damals lag.
//
// # Warum die Zone dazugehört
//
// **Weil CR 608.2b sie beim Verrechnen vergleicht**: ein Ziel, das nicht mehr
// dort liegt, wo es beim Ansagen lag, ist unzulässig.
//
// **Und weil sie zum Zauber gehört und nicht zum Ziel.** Zwei Zauber auf
// dasselbe Objekt haben verschiedene Momente; eine Angabe am Ziel ließe den
// zweiten die des ersten überschreiben.
//
// **Leer bei einem Sitz** — er hat keine Zone. Was ihn unzulässig macht, ist
// das Ausscheiden (CR 800.4a), und das nimmt ihn aus [Partie.Sitze].
type Ziel struct {
	Kennung string
	Zone    Zone
}

// zielePruefen prüft die genannten Ziele beim Ankündigen — CR 115.
//
// **Sie gibt die Ziele samt Zone zurück**: geprüft und festgehalten ist
// derselbe Durchgang, und ein zweiter über dieselbe Liste könnte auseinander
// laufen.
func (p Partie) zielePruefen(ziele []string) ([]Ziel, error) {
	gesehen := make(map[string]bool, len(ziele))
	gefasst := make([]Ziel, 0, len(ziele))

	for _, kennung := range ziele {
		zone, steht := p.zielzone(kennung)
		if !steht {
			return nil, fmt.Errorf(
				"%w: %q ist weder ein Sitz noch ein Objekt dieser Partie "+
					"(CR 115.1)", ErrValidation, kennung)
		}

		// **Dasselbe Ziel nicht zweimal** — CR 115.3. Die Regel erlaubt es
		// für **je ein** Vorkommen des Wortes „Ziel"; wie oft es auf der
		// Karte steht, weiß ohne `M8` niemand, und die strengere Lesart ist
		// die, die keinen unmöglichen Zustand zulässt.
		if gesehen[kennung] {
			return nil, fmt.Errorf("%w: %q steht zweimal als Ziel (CR 115.3)",
				ErrValidation, kennung)
		}

		gesehen[kennung] = true
		gefasst = append(gefasst, Ziel{Kennung: kennung, Zone: zone})
	}

	return gefasst, nil
}

// zielzone gibt die Zone eines Ziels — leer bei einem Sitz.
func (p Partie) zielzone(kennung string) (Zone, bool) {
	for _, sitz := range p.Sitze {
		if sitz.Kennung == kennung {
			return "", true
		}
	}

	for _, objekt := range p.Objekte {
		if objekt.Kennung == kennung {
			return objekt.Zone, true
		}
	}

	return "", false
}

// zieleNochZulaessig sagt, ob beim Verrechnen noch eines zulässig ist —
// CR 608.2b.
//
// **Ohne Ziele ist die Antwort ja.** Ein Zauber ohne Ziel kann nicht daran
// scheitern; wer hier `false` gäbe, ließe jeden Blitz verpuffen.
//
// # Woran ein Ziel unzulässig wird
//
// **An der Zone, in der es stand.** CR 608.2b nennt sie ausdrücklich. Die
// anderen Gründe — geänderte Eigenschaften, Schutz, Hexenmeisterschaft —
// stehen in Kartentexten und brauchen `M8`.
func (p Partie) zieleNochZulaessig(zauber Spielobjekt) bool {
	if len(zauber.Ziele) == 0 {
		return true
	}

	for _, ziel := range zauber.Ziele {
		if p.zielZulaessig(ziel) {
			return true
		}
	}

	return false
}

// zielZulaessig prüft ein einzelnes Ziel gegen die Zone von damals.
func (p Partie) zielZulaessig(ziel Ziel) bool {
	for _, sitz := range p.Sitze {
		if sitz.Kennung == ziel.Kennung {
			return true
		}
	}

	for _, objekt := range p.Objekte {
		if objekt.Kennung == ziel.Kennung {
			return objekt.Zone == ziel.Zone
		}
	}

	return false
}
