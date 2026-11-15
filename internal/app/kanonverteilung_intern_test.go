package app

// **Die Verteilung der Abstände, nicht nur ihr Mittel** — auf Anweisung des
// Auftraggebers vom 2026-09-07, vor der Freigabe von 0.4.0.
//
// # Warum ein Mittel allein zu wenig sagt
//
// **Ein Mittelwert über 680 Karten verschweigt, wie sie sich verteilen.**
// Zwei Läufe mit demselben Mittel können bedeuten: „alle Karten liegen dicht
// beieinander" oder „die Hälfte ist perfekt, die andere Hälfte deutet daneben".
// Für die Frage, wie weit dieses Programm von MSE2 entfernt ist, sind das zwei
// verschiedene Antworten.
//
// **Der Median trennt sie**, die Quartile zeigen die Streuung, und die oberen
// Perzentile sagen, wie schlimm der schlechte Rand ist — ein Maximum allein
// kann eine einzelne Karte sein.
//
// # Was hier gerechnet wird und wie
//
// **Die Streuung als Stichprobenvarianz** (Nenner `n−1`, nicht `n`): die
// gemessenen Karten sind eine **Stichprobe** aus dem, was die Vorlagen
// zeichnen können, keine Vollerhebung. Wer durch `n` teilt, unterschätzt die
// Streuung der Grundgesamtheit systematisch.
//
// **Perzentile ohne Interpolation** — die nächstgelegene Rangzahl. Bei 680
// Werten liegt zwischen zwei Nachbarn kein Unterschied, der eine Entscheidung
// trägt; eine interpolierte Zahl sähe genauer aus, als die Messung ist.
//
// # Was diese Zahlen **nicht** sagen
//
//   - **Nichts über Vorlagen, die in der Stichprobe nicht vorkommen.** Die
//     Reichweite einer Messung ist die Reichweite ihrer Aussage (§10).
//   - **Nichts über die Wahrnehmung.** Der Abstand ist ein Pixelmaß; ob ein
//     Mensch zwei Karten unterscheiden könnte, sagt er nicht.

import (
	"math"
	"sort"
	"testing"
)

// verteilung fasst eine sortierte Messreihe zusammen.
type verteilung struct {
	anzahl                       int
	mittel, streuung, abweichung float64
	kleinster, groesster         float64
	q1, median, q3               float64
	p90, p95, p99                float64
}

// verteilungAus rechnet die Kennzahlen einer **sortierten** Reihe.
//
// **Sortiert wird vom Aufrufer**, weil er die Reihe ohnehin schon sortiert hat
// — ein zweites Sortieren wäre Arbeit ohne Aussage.
func verteilungAus(sortiert []float64) verteilung {
	if len(sortiert) == 0 {
		return verteilung{}
	}

	var summe float64
	for _, wert := range sortiert {
		summe += wert
	}

	mittel := summe / float64(len(sortiert))

	// **Stichprobenvarianz** — der Grund steht im Kopf dieser Datei.
	var quadrate float64
	for _, wert := range sortiert {
		quadrate += (wert - mittel) * (wert - mittel)
	}

	streuung := 0.0
	if len(sortiert) > 1 {
		streuung = quadrate / float64(len(sortiert)-1)
	}

	return verteilung{
		anzahl:     len(sortiert),
		mittel:     mittel,
		streuung:   streuung,
		abweichung: math.Sqrt(streuung),
		kleinster:  sortiert[0],
		groesster:  sortiert[len(sortiert)-1],
		q1:         perzentil(sortiert, 25),
		median:     perzentil(sortiert, 50),
		q3:         perzentil(sortiert, 75),
		p90:        perzentil(sortiert, 90),
		p95:        perzentil(sortiert, 95),
		p99:        perzentil(sortiert, 99),
	}
}

// perzentil gibt den Wert an der nächstgelegenen Rangzahl.
//
// **Ohne Interpolation**, und der Grund steht im Kopf: eine interpolierte Zahl
// sähe genauer aus, als die Messung ist.
//
// # Die Nächste-Rang-Methode, und was daran definiert ist
//
// **Der kleinste Wert, unter dem mindestens p Prozent der Reihe liegen** — die
// Rangzahl ist `ceil(p/100 · n)`, eins-basiert. Bei zehn Werten trifft der
// Median damit den fünften, P90 den neunten.
//
// **Der erste Anlauf rechnete `round(p/100 · n − 0,5)`** und traf bei glatten
// Halben daneben: Gos `math.Round` rundet von der Null weg, und der Median lag
// damit auf dem **sechsten** von zehn Werten. Der Test daneben hat es gefangen
// — er rechnet eine Reihe von 1 bis 10 nach, deren Kennzahlen von Hand
// nachzuprüfen sind.
func perzentil(sortiert []float64, anteil int) float64 {
	if len(sortiert) == 0 {
		return 0
	}

	rang := int(math.Ceil(float64(anteil) / 100 * float64(len(sortiert))))

	// **Der Rang ist eins-basiert, der Index null-basiert.**
	stelle := rang - 1
	if stelle < 0 {
		stelle = 0
	}

	if stelle >= len(sortiert) {
		stelle = len(sortiert) - 1
	}

	return sortiert[stelle]
}

// **Die Rechnung selbst wird geprüft** — an einer Reihe, deren Kennzahlen von
// Hand nachzurechnen sind.
//
// **Ohne Build-Tag**, damit sie in der Pipeline läuft: der Kanonlauf braucht
// fremde Pakete, diese Rechnung nicht. Eine Statistik, die nur hinter einem
// Tag geprüft wird, ist eine Statistik, die niemand prüft.
func TestDieVerteilungRechnetRichtig(t *testing.T) {
	t.Parallel()

	// Zehn Werte, 1 bis 10 — jede Kennzahl ist von Hand zu prüfen.
	reihe := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sort.Float64s(reihe)

	gerechnet := verteilungAus(reihe)

	for _, fall := range []struct {
		name     string
		hat, ist float64
	}{
		{"Mittel", gerechnet.mittel, 5.5},
		{"Minimum", gerechnet.kleinster, 1},
		{"Maximum", gerechnet.groesster, 10},

		// **Stichprobenvarianz**: Σ(x−x̄)² = 82,5, geteilt durch n−1 = 9.
		{"Streuung", gerechnet.streuung, 82.5 / 9},

		// **Nächste Rangzahl**: ceil(0,25·10) = 3 → das dritte Element.
		{"Q1", gerechnet.q1, 3},
		{"Median", gerechnet.median, 5},
		{"Q3", gerechnet.q3, 8},
		{"P90", gerechnet.p90, 9},
		{"P99", gerechnet.p99, 10},
	} {
		if math.Abs(fall.hat-fall.ist) > 1e-9 {
			t.Errorf("%s ist %.4f, erwartet %.4f", fall.name, fall.hat, fall.ist)
		}
	}
}

// **Und eine leere Reihe gibt Nullen statt einer Division durch null.**
func TestEineLeereReiheGibtNullen(t *testing.T) {
	t.Parallel()

	if leer := verteilungAus(nil); leer.anzahl != 0 || leer.mittel != 0 {
		t.Errorf("%+v — erwartet war der Nullverbund", leer)
	}

	if wert := perzentil(nil, 50); wert != 0 {
		t.Errorf("perzentil auf einer leeren Reihe gab %v", wert)
	}
}

// **Ein einziger Wert hat keine Streuung** — und keine Division durch null.
func TestEinEinzelnerWertHatKeineStreuung(t *testing.T) {
	t.Parallel()

	eine := verteilungAus([]float64{7})

	if eine.streuung != 0 || eine.mittel != 7 || eine.median != 7 {
		t.Errorf("%+v — erwartet Mittel 7, Median 7, Streuung 0", eine)
	}
}
