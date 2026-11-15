package app

// **Die Werte des Kanonlaufs je Karte** — geschrieben, gelesen und gegen einen
// früheren Lauf gehalten (F02,
// #2244).
// Warum, steht in `kanonjekarte_intern_test.go`.

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

// kartenwert ist der Abstand **einer** gezogenen Karte zu MSE2.
//
// **Die Stelle gehört zum Schlüssel**: zwei Karten eines Sets können denselben
// Namen tragen, und die Stichprobe zieht ohne Zufall — dieselbe Stelle ist im
// nächsten Lauf dieselbe Karte.
type kartenwert struct {
	Set     string
	Stelle  int
	Name    string
	Abstand float64
	// Vorher ist der Abstand im früheren Lauf — nur im Vergleich gesetzt.
	Vorher float64
}

func (w kartenwert) schluessel() string {
	return w.Set + "\t" + strconv.Itoa(w.Stelle) + "\t" + w.Name
}

// schreibeKartenwerte schreibt je Karte eine Zeile: Set, Stelle, Name (in
// Anführungszeichen, damit ein Tab im Namen keine Spalte wird) und Abstand.
func schreibeKartenwerte(ziel io.Writer, werte []kartenwert) error {
	puffer := bufio.NewWriter(ziel)

	for _, wert := range werte {
		if _, err := fmt.Fprintf(puffer, "%s\t%d\t%s\t%s\n", wert.Set, wert.Stelle,
			strconv.Quote(wert.Name), strconv.FormatFloat(wert.Abstand, 'g', -1, 64)); err != nil {
			return err
		}
	}

	return puffer.Flush()
}

// liesKartenwerte liest, was [schreibeKartenwerte] schrieb — nach Schlüssel.
func liesKartenwerte(quelle io.Reader) (map[string]float64, error) {
	werte := map[string]float64{}
	zeilen := bufio.NewScanner(quelle)

	for nummer := 1; zeilen.Scan(); nummer++ {
		teile := strings.Split(zeilen.Text(), "\t")
		if len(teile) != 4 {
			return nil, fmt.Errorf("Zeile %d: %d Spalten statt 4", nummer, len(teile))
		}

		stelle, err := strconv.Atoi(teile[1])
		if err != nil {
			return nil, fmt.Errorf("Zeile %d: Stelle: %w", nummer, err)
		}

		name, err := strconv.Unquote(teile[2])
		if err != nil {
			return nil, fmt.Errorf("Zeile %d: Name: %w", nummer, err)
		}

		abstand, err := strconv.ParseFloat(teile[3], 64)
		if err != nil {
			return nil, fmt.Errorf("Zeile %d: Abstand: %w", nummer, err)
		}

		werte[kartenwert{Set: teile[0], Stelle: stelle, Name: name}.schluessel()] = abstand
	}

	return werte, zeilen.Err()
}

// kartenvergleich zählt, wie sich jede Karte gegen den früheren Lauf bewegt.
type kartenvergleich struct {
	verglichen, besser, gleich int
	// ohneVorher und ohneNachher sind Karten, die nur in einem Lauf stehen —
	// eine andere Stichprobe oder ein Ausfall, und beides gehört genannt.
	ohneVorher, ohneNachher int
	// schlechter sind die Karten über der Toleranz, die schlimmste zuerst.
	schlechter []kartenwert
}

// vergleicheKartenwerte hält einen Lauf gegen einen früheren. **Gleich** ist,
// was höchstens um die Toleranz wandert, in beide Richtungen.
func vergleicheKartenwerte(vorher map[string]float64, nachher []kartenwert, toleranz float64) kartenvergleich {
	var ergebnis kartenvergleich

	gesehen := map[string]bool{}

	for _, wert := range nachher {
		alt, da := vorher[wert.schluessel()]
		if !da {
			ergebnis.ohneVorher++

			continue
		}

		gesehen[wert.schluessel()] = true
		ergebnis.verglichen++

		switch unterschied := wert.Abstand - alt; {
		case math.Abs(unterschied) <= toleranz:
			ergebnis.gleich++
		case unterschied < 0:
			ergebnis.besser++
		default:
			wert.Vorher = alt
			ergebnis.schlechter = append(ergebnis.schlechter, wert)
		}
	}

	ergebnis.ohneNachher = len(vorher) - len(gesehen)

	sort.SliceStable(ergebnis.schlechter, func(i, j int) bool {
		a, b := ergebnis.schlechter[i], ergebnis.schlechter[j]

		return a.Abstand-a.Vorher > b.Abstand-b.Vorher
	})

	return ergebnis
}

// abweichungen sagt, warum ein Vergleich nicht besteht — leer, wenn er besteht.
//
// **Erst die Stichprobe, dann die Zahlen** (Befund 2 des Zweitblicks über
// E/F): eine Toleranz schützt nur Karten, die beide Läufe ziehen. Ohne eine
// einzige vergleichbare Karte oder mit einer, die nur ein Lauf kennt, ist der
// Vergleich keiner.
func (v kartenvergleich) abweichungen(toleranz float64, quelle string) []string {
	var gemeldet []string

	if v.verglichen == 0 {
		gemeldet = append(gemeldet, fmt.Sprintf("keine Karte ist mit %s vergleichbar", quelle))
	}

	if v.ohneVorher > 0 || v.ohneNachher > 0 {
		gemeldet = append(gemeldet, fmt.Sprintf(
			"die Stichprobe ist nicht dieselbe wie in %s: %d ohne früheren Wert, %d nicht mehr gemessen",
			quelle, v.ohneVorher, v.ohneNachher))
	}

	if len(v.schlechter) > 0 {
		gemeldet = append(gemeldet, fmt.Sprintf("%d Karten sind um mehr als %.2f schlechter als in %s",
			len(v.schlechter), toleranz, quelle))
	}

	return gemeldet
}
