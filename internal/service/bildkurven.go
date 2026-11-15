// **Ein Bild nachzeichnen: aus Ecken werden Kurven** —
// #1340,
// Werkstück F von Phase 4.11.
//
// Portiert aus `convert_to_curves`, `remove_points`, `straighten` und
// `merge_lines` (`src/data/format/image_to_symbol.cpp`, Zeilen 338–470,
// Fassung `f9d9356` vom 2021-07-06).

package service

import (
	"math"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// zuKurven macht aus jeder Strecke eine Bézierkurve.
//
// Portiert aus `convert_to_curves` (Zeilen 338–352). **Ein Drittel der
// Strecke** ist MSEs Griffweite; sie ist die Länge, bei der eine Kurve durch
// drei gleich weit stehende Punkte wie eine Gerade aussieht.
func zuKurven(punkte []msepaket.Kontrollpunkt) {
	for i := range punkte {
		hier := &punkte[i]
		naechster := punktZyklisch(punkte, i+1)

		hier.KurveDanach = true
		naechster.KurveDavor = true

		hier.GriffDanach = mal(minus(naechster.Stelle, hier.Stelle), 1.0/3.0)
		naechster.GriffDavor = mal(minus(hier.Stelle, naechster.Stelle), 1.0/3.0)
	}

	// **Und danach die Sperre einlösen** — MSEs `onUpdateLock()`
	// (`src/data/symbol.cpp`, Zeilen 63–75, Fassung `f9d9356` vom 2021-07-06).
	// Ohne diesen zweiten Durchgang
	// zeigten die beiden Griffe eines `direction`-Punktes in beliebige
	// Richtungen, und die Kurve knickte dort, wo sie glatt sein soll.
	for i := range punkte {
		sperreEinloesen(&punkte[i])
	}
}

// sperreEinloesen richtet die beiden Griffe nach der Sperre aus.
//
// **Nur `direction` tut hier etwas**: die Griffe bekommen dieselbe Achse und
// behalten ihre Länge. `size` käme aus dem Editor und entsteht beim
// Nachzeichnen nicht; `free` lässt beide, wie sie sind.
func sperreEinloesen(punkt *msepaket.Kontrollpunkt) {
	if punkt.Sperre != sperreRichtung {
		return
	}

	achse := genormt(minus(punkt.GriffDavor, punkt.GriffDanach))
	punkt.GriffDavor = mal(achse, laenge(punkt.GriffDavor))
	punkt.GriffDanach = mal(achse, -laenge(punkt.GriffDanach))
}

// punkteEntfernen nimmt so lange den billigsten Punkt heraus, bis es teuer wird.
//
// Portiert aus `remove_points` (Zeilen 396–415). **Die Schwelle `0.0002` ist
// MSEs**, und sie ist eine Fläche: die Kosten sind Abstand mal Länge, also
// ungefähr das, was die Form beim Entfernen gewinnt oder verliert.
func punkteEntfernen(punkte []msepaket.Kontrollpunkt) []msepaket.Kontrollpunkt {
	const schwellwert = 0.0002

	for len(punkte) > 3 {
		bester := -1
		bestenfalls := math.Inf(1)

		for i := range punkte {
			if kosten := entfernungskosten(punkte, i); kosten < bestenfalls {
				bestenfalls, bester = kosten, i
			}
		}

		if bestenfalls > schwellwert || bester < 0 {
			break
		}

		punkte = punktHeraus(punkte, bester)
	}

	return punkte
}

// entfernungskosten sagt, was das Entfernen eines Punktes die Form kostet.
//
// Portiert aus `cost_of_point_removal` (Zeilen 417–439). **Ecken werden nie
// entfernt** — `if (cur.lock != LOCK_DIR) return 1e100`. Eine Ecke ist die
// Aussage über die Form, die eine Glättung nicht treffen darf.
func entfernungskosten(punkte []msepaket.Kontrollpunkt, i int) float64 {
	hier := punktZyklisch(punkte, i)
	if hier.Sperre != sperreRichtung {
		return math.Inf(1)
	}

	vorher := punktZyklisch(punkte, i-1)
	nachher := punktZyklisch(punkte, i+1)

	// **Die kleine Zugabe verhindert eine Teilung durch null** — MSEs
	// `+ 0.00001` (Zeilen 425–426), nicht meine Zutat.
	davorLaenge := laenge(hier.GriffDavor) + 0.00001
	danachLaenge := laenge(hier.GriffDanach) + 0.00001
	gesamt := davorLaenge + danachLaenge

	erster := mal(vorher.GriffDanach, gesamt/davorLaenge)
	zweiter := mal(nachher.GriffDavor, gesamt/danachLaenge)

	auf := bezierstelle(
		vorher.Stelle,
		plus(vorher.Stelle, erster),
		plus(nachher.Stelle, zweiter),
		nachher.Stelle,
		davorLaenge/gesamt,
	)

	return laenge(minus(hier.Stelle, auf)) * laenge(minus(vorher.Stelle, nachher.Stelle))
}

// punktHeraus entfernt einen Punkt und streckt die Griffe der Nachbarn.
//
// Portiert aus `remove_point` (Zeilen 441–456) — dieselbe Rechnung wie in den
// Kosten, damit die Form nach dem Entfernen dort verläuft, wo die Kosten es
// versprochen haben.
func punktHeraus(punkte []msepaket.Kontrollpunkt, i int) []msepaket.Kontrollpunkt {
	hier := punktZyklisch(punkte, i)
	vorher := punktZyklisch(punkte, i-1)
	nachher := punktZyklisch(punkte, i+1)

	davorLaenge := laenge(hier.GriffDavor) + 0.00001
	danachLaenge := laenge(hier.GriffDanach) + 0.00001
	gesamt := davorLaenge + danachLaenge

	vorher.GriffDanach = mal(vorher.GriffDanach, gesamt/davorLaenge)
	nachher.GriffDavor = mal(nachher.GriffDavor, gesamt/danachLaenge)

	return append(punkte[:i], punkte[i+1:]...)
}

// bezierstelle gibt den Punkt einer kubischen Bézierkurve bei `t`.
func bezierstelle(a, b, c, d msepaket.Vektor, t float64) msepaket.Vektor {
	u := 1 - t

	return plus(
		plus(mal(a, u*u*u), mal(b, 3*u*u*t)),
		plus(mal(c, 3*u*t*t), mal(d, t*t*t)),
	)
}

// begradigen macht aus fast geraden Kurven Geraden.
//
// Portiert aus `straighten` (Zeilen 355–370). **Gemessen wird die Fläche unter
// den Griffen**, nicht der Winkel: `|a×ab| + |b×ab| < 0.2`.
func begradigen(punkte []msepaket.Kontrollpunkt) {
	const schwellwert = 0.2

	for i := range punkte {
		hier := &punkte[i]
		naechster := punktZyklisch(punkte, i+1)

		ab := genormt(minus(naechster.Stelle, hier.Stelle))
		aa := genormt(hier.GriffDanach)
		bb := genormt(naechster.GriffDavor)

		if math.Abs(kreuzprodukt(aa, ab))+math.Abs(kreuzprodukt(bb, ab)) >= schwellwert {
			continue
		}

		hier.KurveDanach = false
		naechster.KurveDavor = false
		hier.GriffDanach = msepaket.Vektor{}
		naechster.GriffDavor = msepaket.Vektor{}
		hier.Sperre = sperreFrei
		naechster.Sperre = sperreFrei
	}
}

// geradenVerschmelzen entfernt Punkte, die auf einer Geraden liegen.
//
// Portiert aus `merge_lines` (Zeilen 373–388). **Das Maß ist Winkel mal
// Streckenquadrat**: ein winziger Knick auf einer langen Strecke bleibt, ein
// deutlicher auf einer sehr kurzen fällt weg — dort sieht ihn ohnehin niemand.
func geradenVerschmelzen(punkte []msepaket.Kontrollpunkt) []msepaket.Kontrollpunkt {
	for i := 0; i < len(punkte); i++ {
		if len(punkte) <= 3 {
			break
		}

		hier := punktZyklisch(punkte, i)
		if hier.KurveDavor != hier.KurveDanach {
			continue
		}

		a := punktZyklisch(punkte, i-1).Stelle
		b := hier.Stelle
		c := punktZyklisch(punkte, i+1).Stelle

		ab := genormt(minus(a, b))
		bc := genormt(minus(b, c))

		winkel := math.Abs(math.Atan2(ab.X, ab.Y) - math.Atan2(bc.X, bc.Y))
		ac := minus(a, c)

		if winkel*(ac.X*ac.X+ac.Y*ac.Y) < 0.0001 {
			punkte = append(punkte[:i], punkte[i+1:]...)
			i--
		}
	}

	return punkte
}

// formVereinfachen ist MSEs `simplify_symbol_shape` (Zeilen 462–472).
//
// **Die Reihenfolge ist die Aussage.** Dreimal glätten steht dort so; zweimal
// gäbe eine kantigere, viermal eine weichere Form — und weder das eine noch
// das andere ist das, was MSE zeichnet.
func formVereinfachen(teil msepaket.Symbolteil) msepaket.Symbolteil {
	punkte := teil.Punkte
	if len(punkte) < 4 {
		return teil
	}

	eckenMarkieren(punkte)

	punkte = eckenVerschmelzen(punkte)

	for range 3 {
		glaetten(punkte)
	}

	zuKurven(punkte)

	punkte = punkteEntfernen(punkte)

	begradigen(punkte)

	punkte = geradenVerschmelzen(punkte)

	teil.Punkte = punkte

	return teil
}
