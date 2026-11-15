package msescript

// `cycle(…)` und seine Kurzform `(…)` — die Ordnung, die den Rahmen zeichnet.
//
// # Herkunft
//
// `src/util/spec_sort.hpp` von twanvl/MagicSetEditor2, Stand 2026-08-16,
// beschreibt sie im Kopfkommentar; `cycle_sort` in `spec_sort.cpp:135` setzt
// sie um. Gelesen nach dem Lizenzwechsel auf GPL-2.0
// ([0019](../../docs/entscheidungen/0019-gpl-2-0-statt-mit.md)).
//
// # Die Regel
//
// Die genannten Zeichen bilden einen **Kreis im Uhrzeigersinn**. Ausgegeben
// wird die Drehung, die
//
//  1. den kürzesten Weg über den Kreis nimmt,
//  2. ihre **Lücken früh** hat — eine Lücke ist ein Zeichen der Ordnung, das in
//     der Eingabe fehlt,
//  3. bei Gleichstand am weitesten vorn im Ausdruck beginnt.
//
// Für Magic heißt das: `WUR` wird zu `RWU`, weil der Weg von R über W nach U
// kürzer ist als der von W über U nach R.
//
// Fassung f9d9356 vom 2021-07-06.

// zyklus ist eine Gruppe, deren Reihenfolge erst die Eingabe entscheidet.
type zyklus struct {
	// zeichen in der Reihenfolge des Kreises.
	zeichen []rune

	// rang ist der Rang des **ersten** Zeichens; die übrigen folgen darauf.
	rang int
}

// drehe gibt die Ränge dieser Gruppe für eine bestimmte Eingabe.
//
// **Der Rang hängt von der Eingabe ab**, und das ist der Unterschied zu jeder
// anderen Gruppe dieser Ordnung. Deshalb wird er beim Anwenden gerechnet und
// nicht beim Zerlegen.
func (z zyklus) drehe(vorkommen map[rune]int) map[rune]int {
	beginn := besterBeginn(z.zeichen, vorkommen)

	raenge := make(map[rune]int, len(z.zeichen))

	for schritt := range z.zeichen {
		raenge[z.zeichen[(beginn+schritt)%len(z.zeichen)]] = z.rang + schritt
	}

	return raenge
}

// besterBeginn sucht die Drehung mit dem kleinsten Wert.
//
// # Die Bewertung, portiert
//
// `cycle_sort` in `spec_sort.cpp:144` baut je Startpunkt eine Zahl:
//
//	size_t score = 0, mul = 1;
//	for (i …) { mul *= 2; if (counts[(start+i) % n]) { score = score*mul + 1; mul = 1; } }
//
// Gelesen: jede **Lücke** verdoppelt den Faktor, mit dem der bisherige Wert
// beim nächsten belegten Zeichen multipliziert wird. Frühe Lücken wiegen damit
// schwerer als späte, und der kleinste Wert gewinnt — das sind die Regeln 1
// und 2 zusammen. Regel 3 folgt aus `<` statt `<=`: bei Gleichstand bleibt der
// frühere Startpunkt stehen.
//
// **Als `uint64`, nicht als `int`.** Die Zahl verdoppelt sich je Lücke; bei
// einer Ordnung mit vielen Zeichen liefe ein vorzeichenbehafteter Wert über und
// würde negativ — und dann gewänne ausgerechnet die schlechteste Drehung. Kein
// Paket hat so viele, aber fremde Daten sind der Normalfall (§8).
func besterBeginn(zeichen []rune, vorkommen map[rune]int) int {
	beginn, bester := 0, ^uint64(0)

	for start := range zeichen {
		wert, faktor := uint64(0), uint64(1)

		for schritt := range zeichen {
			faktor *= 2

			if vorkommen[zeichen[(start+schritt)%len(zeichen)]] > 0 {
				wert = wert*faktor + 1
				faktor = 1
			}
		}

		if wert < bester {
			bester, beginn = wert, start
		}
	}

	return beginn
}
