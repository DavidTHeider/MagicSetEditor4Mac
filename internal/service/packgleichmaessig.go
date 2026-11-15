// Die gleichmäßige Verteilung eines Packs
// (#1072).
//
// **Eine eigene Datei**, weil sie eine eigene Regel trägt:
// `weighted_equal_divide` verteilt statt zu würfeln, und das ist der
// Unterschied zwischen „ein Booster hat meistens eine Rare" und „ein Booster
// hat eine Rare".

package service

import (
	"math/rand/v2"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gewichteterPosten ist ein Posten in der Verteilung.
type gewichteterPosten struct {
	gewicht float64
	anzahl  int

	// losentscheid bricht Gleichstände. MSE zieht ihn aus demselben
	// Zufallsgeber (`WeightedItem::tiebreaker`, `src/data/pack.cpp`, Zeilen (Fassung f9d9356 vom 2021-07-06)
	// 175–179) — ohne ihn bekäme bei gleichem Verhältnis immer derselbe
	// Posten den Zuschlag.
	losentscheid int
}

// gleichmaessig ist der `SELECT_EQUAL`-Zweig (`src/data/pack.cpp`, Zeilen
// 247–296, Fassung f9d9356 vom 2021-07-06).
//
// **Bei einer einzigen Kopie wird gewürfelt**, und der Quelltext sagt warum:
//
//	// somewhat of a hack to keep things fair: just pick at random
//	// otherwise we would end up picking the lowest weight item
func gleichmaessig(
	staende map[string]*packstand, stand *packstand, verlangt int,
	wuerfel *rand.Rand, gezogen []*domain.CardRow,
) []*domain.CardRow {
	if verlangt == 1 {
		return einZufaelliges(staende, stand, wuerfel, gezogen)
	}

	posten := make([]gewichteterPosten, 0, len(stand.art.Posten)+1)

	for _, eines := range stand.art.Posten {
		unten := staende[eines.Name]
		gewicht := eines.Gewicht

		switch {
		case stand.art.Auswahl == "equal proportional":
			gewicht = eines.Gewicht * unten.gewicht

		case stand.art.Auswahl == "equal nonempty" && unten.gewicht <= 0:
			gewicht = 0
		}

		posten = append(posten,
			gewichteterPosten{gewicht: gewicht, losentscheid: wuerfel.Int()})
	}

	// **Die Karten selbst sind der letzte Posten**, mit ihrer Zahl als
	// Gewicht (Zeile 271).
	posten = append(posten, gewichteterPosten{
		gewicht: float64(len(stand.karten)), losentscheid: wuerfel.Int(),
	})

	verteileGleichmaessig(posten, verlangt)

	for stelle, eines := range stand.art.Posten {
		staende[eines.Name].verlangt += eines.Anzahl * posten[stelle].anzahl
	}

	return eigeneKarten(stand, posten[len(posten)-1].anzahl, wuerfel, gezogen)
}

// eigeneKarten nimmt so viele eigene Karten, wie die Verteilung zugewiesen hat
// (Zeilen 281–295).
//
// **Erst jede Karte gleich oft, dann der Rest zufällig.** Wer alles zufällig
// zöge, gäbe bei fünf Karten und zehn Zügen nicht zweimal jede.
func eigeneKarten(
	stand *packstand, anzahl int, wuerfel *rand.Rand, gezogen []*domain.CardRow,
) []*domain.CardRow {
	if len(stand.karten) == 0 || anzahl <= 0 {
		return gezogen
	}

	for range anzahl / len(stand.karten) {
		gezogen = append(gezogen, stand.karten...)
	}

	for range anzahl % len(stand.karten) {
		gezogen = append(gezogen, stand.karten[wuerfel.IntN(len(stand.karten))])
	}

	return gezogen
}

// verteileGleichmaessig ist `weighted_equal_divide` (Zeilen 192–213).
//
// # Wie verteilt wird
//
//	// repeatedly pick the item that minimizes, after incrementing count:
//	//   max_wi  wi->count/wi->weight
//
// MSE benutzt dafür eine Vorrangwarteschlange. **Hier steht eine Schleife**,
// und der Grund ist die Größe: verteilt werden die Züge eines Boosters —
// vierzehn im echten Pack —, und eine Warteschlange für vierzehn Schritte
// wäre mehr Bauwerk als Nutzen. Die **Auswahlregel** ist dieselbe, samt
// Losentscheid.
func verteileGleichmaessig(posten []gewichteterPosten, gesamt int) {
	if len(posten) == 1 {
		posten[0].anzahl = gesamt

		return
	}

	for range gesamt {
		beste := -1

		for stelle := range posten {
			if posten[stelle].gewicht <= 0 {
				continue
			}

			if beste < 0 || besserDran(posten[stelle], posten[beste]) {
				beste = stelle
			}
		}

		if beste < 0 {
			return
		}

		posten[beste].anzahl++
	}
}

// besserDran ist `CompareWeightedItems`, umgedreht: MSE vergleicht „b wird
// bevorzugt", hier heißt es „a wird bevorzugt".
//
//	double delta = b->weight * (a->count + 1) - a->weight * (b->count + 1);
func besserDran(eines, anderes gewichteterPosten) bool {
	abstand := anderes.gewicht*float64(eines.anzahl+1) - eines.gewicht*float64(anderes.anzahl+1)

	switch {
	case abstand < 0:
		return true

	case abstand > 0:
		return false

	default:
		return eines.losentscheid > anderes.losentscheid
	}
}
