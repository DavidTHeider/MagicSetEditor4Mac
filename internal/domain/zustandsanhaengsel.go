package domain

// **Was an nichts hängt, und was gar nicht hängen darf** — Werkstück M5 der
// Regelmaschine
// (#1607).
//
// # Drei Regeln, zwei Ausgänge
//
//   - **CR 704.5m**: eine **Aura** an einem unzulässigen Ding oder an gar
//     nichts geht in den Friedhof ihres Eigners.
//   - **CR 704.5n**: eine **Ausrüstung** oder Befestigung an einem
//     unzulässigen Permanent oder an einem Spieler **löst sich** — und bleibt
//     auf dem Schlachtfeld.
//   - **CR 704.5p**: alles **andere**, was hängt, obwohl es nicht hängen
//     darf, löst sich ebenfalls und bleibt liegen.
//
// **Der Unterschied ist der ganze Punkt**: eine Aura **ist** ihre Bindung und
// stirbt ohne sie; eine Ausrüstung liegt weiter da.
//
// # Woran „unzulässig" hier gemessen wird
//
// **An dem, was dieses Programm weiß**: hängt es an einer Kennung, die es
// nicht gibt, oder an etwas, das nicht auf dem Schlachtfeld liegt. **Was es
// nicht weiß**, ist, ob die Aura dieses Ding *verzaubern darf* — das steht in
// ihrem Untertyp („Aura — Enchant creature", CR 702.5) und damit in einer
// Fähigkeit, die niemand ausführt (`M8`,
// #1610).
// Die Grenze steht in der Regelmatrix.
//
// # Und an einen Spieler hängt hier nichts
//
// **[Spielobjekt.HaengtAn] trägt eine Objektkennung.** Ein Anhängsel an einem
// **Spieler** (CR 704.5n, 704.5p) lässt sich damit nicht ausdrücken; wer es
// bauen will, gibt der Beziehung ein zweites Feld. Das ist eine Grenze und
// keine Auslassung (§10).

import "strings"

// UntertypAura ist der Untertyp, den CR 704.5m meint.
//
// **Der Untertyp und nicht der Kartentyp**: CR 205.3h führt „Aura" als
// Untertyp einer Verzauberung.
const UntertypAura = "aura"

// UntertypAusruestung ist der Untertyp aus CR 704.5n.
const UntertypAusruestung = "equipment"

// UntertypBefestigung ist der zweite aus CR 704.5n.
const UntertypBefestigung = "fortification"

// anhaengselaktionen wendet CR 704.5m, 704.5n und 704.5p an.
func (p Partie) anhaengselaktionen() (Partie, bool) {
	geblieben := make([]Spielobjekt, 0, len(p.Objekte))

	etwas := false

	for _, objekt := range p.Objekte {
		if objekt.HaengtAn == "" || objekt.Zone != ZoneTisch ||
			p.haengtZulaessig(objekt) {
			geblieben = append(geblieben, objekt)

			continue
		}

		etwas = true

		if p.istAura(objekt) {
			// **CR 704.5m** — sie ist ihre Bindung und stirbt ohne sie.
			objekt = NeuesObjekt(objekt, p.Zug.Nummer)
			objekt.Zone = ZoneFriedhof
			objekt.Beherrscher = objekt.Eigner
		}

		// **CR 704.5n und CR 704.5p** — gelöst, und liegen bleibt es. Für
		// die Aura steht die Zeile auch da: im Friedhof hängt nichts mehr
		// (CR 400.7).
		objekt.HaengtAn = ""

		geblieben = append(geblieben, objekt)
	}

	if !etwas {
		return p, false
	}

	p.Objekte = geblieben

	return p, true
}

// haengtZulaessig sagt, ob das, woran etwas hängt, es tragen kann.
//
// **Zwei Bedingungen, und beide sind Tatsachen und keine Auslegung**: es gibt
// das Ding, und es liegt auf dem Schlachtfeld (CR 110.1 — nur Permanente
// tragen etwas).
func (p Partie) haengtZulaessig(objekt Spielobjekt) bool {
	for _, anderes := range p.Objekte {
		if anderes.Kennung == objekt.HaengtAn {
			return anderes.Zone == ZoneTisch
		}
	}

	return false
}

// istAura sagt, ob dieses Objekt eine Aura ist — CR 205.3h.
func (p Partie) istAura(objekt Spielobjekt) bool {
	eigen, da := p.Eigenschaften(objekt)
	if !da {
		return false
	}

	for _, unter := range eigen.Typzeile.Untertypen {
		if strings.EqualFold(unter, UntertypAura) {
			return true
		}
	}

	return false
}
