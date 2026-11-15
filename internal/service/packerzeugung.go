// Das Erzeugen einer einzelnen Packart
// (#1072).
//
// **Von `packzieher.go` getrennt**, weil der Schnitt einer Frage folgt: dort
// steht, **welche** Packarten es gibt und in welcher Reihenfolge sie
// drankommen, hier, **wie** eine von ihnen zieht.

package service

import (
	"math/rand/v2"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// erzeuge zieht aus einer Packart — `PackInstance::generate`
// (`src/data/pack.cpp`, Zeilen 215–312, Fassung f9d9356 vom 2021-07-06).
func (z *Packzieher) erzeuge(
	staende map[string]*packstand, stand *packstand,
	wuerfel *rand.Rand, gezogen []*domain.CardRow,
) []*domain.CardRow {
	verlangt := stand.verlangt
	stand.verlangt = 0

	if verlangt == 0 {
		return gezogen
	}

	switch stand.art.Auswahl {
	case "all":
		return alleZiehen(staende, stand, verlangt, gezogen)

	case "replace", "proportional", "nonempty":
		for range verlangt {
			gezogen = einZufaelliges(staende, stand, wuerfel, gezogen)
		}

		return gezogen

	case "no replace":
		return ohneZuruecklegen(stand, verlangt, wuerfel, gezogen)

	case "equal", "equal proportional", "equal nonempty":
		return gleichmaessig(staende, stand, verlangt, wuerfel, gezogen)

	case "first":
		return dasErste(staende, stand, verlangt, gezogen)

	default:
		// **`auto` ohne Filter und ohne Posten zieht nichts** — MSE hat dort
		// keinen Zweig, und ein erfundener wäre eine Behauptung.
		return gezogen
	}
}

// alleZiehen nimmt jede Karte und verlangt jeden Posten (`generate_all`,
// Zeilen 316–328).
func alleZiehen(
	staende map[string]*packstand, stand *packstand, kopien int,
	gezogen []*domain.CardRow,
) []*domain.CardRow {
	for range kopien {
		gezogen = append(gezogen, stand.karten...)
	}

	// **Ohne Wache auf `nil`.** `staende` legt für jede genannte Packart einen
	// Stand an oder scheitert; ein Zweig für „gibt es nicht" wäre eine Zeile,
	// die kein Aufruf betreten kann (§3).
	for _, posten := range stand.art.Posten {
		staende[posten.Name].verlangt += kopien * posten.Anzahl
	}

	return gezogen
}

// einZufaelliges zieht **eines** — `generate_one_random` (Zeilen 330–357).
//
// **Der Zufall trifft die Gesamtsumme**, nicht erst die Karten und dann die
// Posten: `r` läuft über beide, und wo es hinfällt, entscheidet.
func einZufaelliges(
	staende map[string]*packstand, stand *packstand,
	wuerfel *rand.Rand, gezogen []*domain.CardRow,
) []*domain.CardRow {
	if stand.gewicht <= 0 {
		return gezogen
	}

	rest := wuerfel.Float64() * stand.gewicht

	if rest < float64(len(stand.karten)) {
		return append(gezogen, stand.karten[int(rest)])
	}

	rest -= float64(len(stand.karten))

	for _, posten := range stand.art.Posten {
		unten := staende[posten.Name]

		switch stand.art.Auswahl {
		case "proportional", "equal proportional":
			rest -= posten.Gewicht * unten.gewicht

		case "nonempty", "equal nonempty":
			if unten.gewicht > 0 {
				rest -= posten.Gewicht
			}

		default:
			rest -= posten.Gewicht
		}

		if rest < 0 {
			unten.verlangt += posten.Anzahl

			break
		}
	}

	return gezogen
}

// ohneZuruecklegen mischt und schneidet ab (Zeilen 230–246).
//
//	// to prevent us from being too predictable for small sets,
//	// periodically reshuffle
//
// **Die Stapelgröße ist die halbe Kartenzahl, aufgerundet.** Wer alles auf
// einmal nähme, gäbe bei einem kleinen Set immer dieselbe Reihenfolge.
func ohneZuruecklegen(
	stand *packstand, verlangt int, wuerfel *rand.Rand, gezogen []*domain.CardRow,
) []*domain.CardRow {
	if len(stand.karten) == 0 {
		return gezogen
	}

	haufen := append([]*domain.CardRow(nil), stand.karten...)
	proStapel := (len(haufen) + 1) / 2

	for rest := verlangt; rest > 0; rest -= proStapel {
		wuerfel.Shuffle(len(haufen), func(eins, zwei int) {
			haufen[eins], haufen[zwei] = haufen[zwei], haufen[eins]
		})

		gezogen = append(gezogen, haufen[:min(rest, proStapel)]...)
	}

	return gezogen
}

// dasErste nimmt die eigene Karte, sonst den ersten Posten, der etwas hat
// (Zeilen 297–311).
func dasErste(
	staende map[string]*packstand, stand *packstand, verlangt int,
	gezogen []*domain.CardRow,
) []*domain.CardRow {
	if len(stand.karten) > 0 {
		for range verlangt {
			gezogen = append(gezogen, stand.karten[0])
		}

		return gezogen
	}

	for _, posten := range stand.art.Posten {
		unten := staende[posten.Name]
		if unten.gewicht <= 0 {
			continue
		}

		unten.verlangt += verlangt * posten.Anzahl

		break
	}

	return gezogen
}
