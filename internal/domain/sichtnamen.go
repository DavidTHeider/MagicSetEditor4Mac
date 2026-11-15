package domain

// **Wie eine Sicht ihre Verweise nennt** — #2480, Vermerk 0201.
//
// Eine Kennung steht nicht nur am Objekt: ein Anhängsel nennt, woran es
// hängt, ein Haufen seinen Anker, ein Kampf Angreifer und Blocker, und der
// Commanderschaden zählt je Commander. **Jede davon geht unter der
// Sichtkennung hinaus**, sonst stünde die Kennung im Verweis, nachdem sie am
// Objekt verschwunden ist.

import (
	"cmp"
	"slices"
)

// sichtnamen ordnet jeder Kennung ihre Sichtkennung zu.
type sichtnamen map[string]string

// sichtnamen gibt die Namen aller Objekte dieser Partie.
func (p Partie) sichtnamen() sichtnamen {
	namen := make(sichtnamen, len(p.Objekte))
	for _, objekt := range p.Objekte {
		namen[objekt.Kennung] = p.Sichtkennung(objekt)
	}

	return namen
}

// von gibt die Sichtkennung einer Kennung — und was kein Objekt nennt, etwa
// einen Sitz oder nichts, unverändert.
func (n sichtnamen) von(kennung string) string {
	if name, da := n[kennung]; da {
		return name
	}

	return kennung
}

// kampf gibt einen Kampf unter Sichtkennungen — **eine neue Tabelle**, die
// Partie bleibt, was sie ist.
func (n sichtnamen) kampf(kampf Kampfstand) Kampfstand {
	angriffe := make([]Angriff, len(kampf.Angriffe))
	for stelle, angriff := range kampf.Angriffe {
		angriffe[stelle] = Angriff{Kreatur: n.von(angriff.Kreatur), Ziel: n.von(angriff.Ziel)}
	}

	bloecke := make([]Block, len(kampf.Bloecke))
	for stelle, block := range kampf.Bloecke {
		bloecke[stelle] = Block{Kreatur: n.von(block.Kreatur), Angreifer: n.von(block.Angreifer)}
	}

	kampf.Angriffe, kampf.Bloecke = angriffe, bloecke

	return kampf
}

// schaden gibt den Commanderschaden unter Sichtkennungen.
func (n sichtnamen) schaden(roh map[string]int) map[string]int {
	kopie := make(map[string]int, len(roh))
	for wessen, wieviel := range roh {
		kopie[n.von(wessen)] = wieviel
	}

	return kopie
}

// geordnet stellt die offenen Objekte in ihrer Reihenfolge voran und die
// verborgenen **nach ihrer Sichtkennung** dahinter — in der Reihenfolge des
// Speichers stünden Geschwister eines Deckplatzes nebeneinander.
func geordnet(objekte []Sichtobjekt) []Sichtobjekt {
	offen := slices.DeleteFunc(slices.Clone(objekte), func(o Sichtobjekt) bool { return o.Geheim })
	verborgen := slices.DeleteFunc(slices.Clone(objekte), func(o Sichtobjekt) bool { return !o.Geheim })

	slices.SortFunc(verborgen, func(a, b Sichtobjekt) int { return cmp.Compare(a.Kennung, b.Kennung) })

	return append(offen, verborgen...)
}
