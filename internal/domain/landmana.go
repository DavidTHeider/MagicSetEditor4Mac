package domain

// **Die einzige Manafähigkeit, die in der Regel steht** — Werkstück M3 der
// Regelmaschine
// (#1605).
//
// # Warum ausgerechnet diese eine
//
// **Weil CR 305.6 sie einem Land gibt, ohne dass Text darauf stehen muss.**
// Ein Land mit einem Grundlandtyp hat die Fähigkeit „{T}: Erzeuge [Mana]"
// **innewohnend** — sie steht im Regelwerk und nicht im Kartentext.
//
// **Jede andere Manafähigkeit steht als Text auf der Karte**, und einen Text
// auszuführen ist `M8`
// (#1610).
// Ohne diese eine gäbe es in einer Partie **kein Mana**, und `M3` wäre ein
// Vorrat, den niemand füllen kann — eine Zusage ohne Einlösung (§10).
//
// # Und sie geht nicht über den Stapel
//
// **CR 605.3b**: eine Manafähigkeit wird sofort verrechnet, lässt sich nicht
// abwehren und nicht beantworten. Hier heißt das: der Vorrat steht nach dem
// Ruf da, und es gibt nichts dazwischen.

import "fmt"

// grundlandmana ordnet die fünf Grundlandtypen ihrem Mana zu — CR 305.6.
//
// **Die Untertypen stehen englisch**, so wie sie auf der Karte und im
// MSE-Paket stehen (§9: was aus einem Paket kommt, wird nicht übersetzt).
var grundlandmana = map[string]Manaart{
	"Plains":   ManaWeiss,
	"Island":   ManaBlau,
	"Swamp":    ManaSchwarz,
	"Mountain": ManaRot,
	"Forest":   ManaGruen,
}

// LandFuerMana tappt ein Grundland und legt sein Mana in den Vorrat —
// CR 305.6, CR 605.3b.
//
// **Wer tappt, muss es beherrschen** (CR 110.2a): eine Fähigkeit aktiviert,
// wer das Objekt beherrscht.
func (p Partie) LandFuerMana(wer, objekt string) (Partie, error) {
	stelle, fehler := p.landstelle(wer, objekt)
	if fehler != nil {
		return p, fehler
	}

	art, fehler := p.grundlandart(p.Objekte[stelle])
	if fehler != nil {
		return p, fehler
	}

	// **Erst tappen, dann Mana** — die Kosten vor der Wirkung (CR 601.2h).
	// Die Reihenfolge ist hier ohne sichtbaren Unterschied; sie steht so,
	// weil ein späterer Fehlerweg sonst Mana ohne Kosten hinterließe.
	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)
	objekte[stelle].Getappt = true
	p.Objekte = objekte

	return p.ManaHinzu(wer, art, 1)
}

// landstelle findet das Land und prüft, ob es tappen darf.
func (p Partie) landstelle(wer, objekt string) (int, error) {
	for stelle := range p.Objekte {
		if p.Objekte[stelle].Kennung != objekt {
			continue
		}

		switch {
		case p.Objekte[stelle].Beherrscher != wer:
			return 0, fmt.Errorf(
				"%w: %q beherrscht das Objekt %q nicht (CR 110.2a)",
				ErrValidation, wer, objekt)
		case p.Objekte[stelle].Zone != ZoneTisch:
			return 0, fmt.Errorf(
				"%w: das Objekt %q liegt nicht auf dem Schlachtfeld",
				ErrValidation, objekt)
		case p.Objekte[stelle].Getappt:
			return 0, fmt.Errorf(
				"%w: das Objekt %q ist schon getappt (CR 118.3)",
				ErrValidation, objekt)
		}

		return stelle, nil
	}

	return 0, fmt.Errorf("%w: das Objekt %q gibt es in dieser Partie nicht",
		ErrNotFound, objekt)
}

// grundlandart sagt, welches Mana ein Objekt gibt — CR 305.6.
//
// **Zwei Bedingungen, und beide stehen in derselben Regel**: es muss den
// Kartentyp Land tragen **und** einen der fünf Grundlandtypen. Ein Land ohne
// Grundlandtyp hat keine innewohnende Fähigkeit; was es sonst kann, steht in
// seinem Text und wartet auf `M8`.
func (p Partie) grundlandart(objekt Spielobjekt) (Manaart, error) {
	eigenschaften, gefunden := p.Eigenschaften(objekt)
	if !gefunden {
		return "", fmt.Errorf(
			"%w: zu %q steht keine Karte im eingefrorenen Abzug",
			ErrNotFound, objekt.Kennung)
	}

	if !eigenschaften.Typzeile.Ist(TypLand) {
		return "", fmt.Errorf("%w: %q ist kein Land (CR 305.6)",
			ErrValidation, eigenschaften.Name)
	}

	for _, untertyp := range eigenschaften.Typzeile.Untertypen {
		if art, ist := grundlandmana[untertyp]; ist {
			return art, nil
		}
	}

	return "", fmt.Errorf(
		"%w: %q trägt keinen Grundlandtyp — was es sonst kann, steht in "+
			"seinem Text (CR 305.6)", ErrValidation, eigenschaften.Name)
}
