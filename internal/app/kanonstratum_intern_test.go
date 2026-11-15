package app

// **Der volle Kanon ist eine geschichtete Stichprobe, keine Sammlung** — auf
// Anweisung des Auftraggebers vom 2026-09-07.
//
// # Der Befund
//
// Die Abstände aller zehn Sets wurden **zusammengeschüttet** und darüber
// Mittel, Median und Perzentile gerechnet. Das ist falsch, sobald die Sets
// verschieden stark gezogen werden — und das werden sie:
//
//	Archenemy    64 von    70  =  91,4 %
//	Black       582 von 3.788  =  15,4 %
//
// Eine Karte aus Archenemy vertritt im Kanon **1,09** Karten, eine aus Black
// **6,51**. Wer beide gleich zählt, gibt Archenemy das Sechsfache seines
// Gewichts. Der Mittelwert über den Topf schätzt damit nicht den Mittelwert
// des Kanons, sondern den einer Grundgesamtheit, die es nicht gibt.
//
// # Die Schichten
//
// **Jedes Set ist eine Schicht.** Das ist nicht willkürlich: sie sind nach
// Farbe und Sonderform getrennt, und genau daran hängt der Zeichenweg —
// Goldrahmen und Blendmasken bei `Multicolor`, Artefakt-Masken bei
// `Colorless`, Überformate bei `Planechase` und `Vanguard`. Eine Schichtung
// entlang dieser Grenze ist der Grund, warum die Stichprobe überhaupt etwas
// sagt.
//
// # Was hier gerechnet wird
//
// Mit `H` Schichten, `N_h` Karten in Schicht `h`, `n_h` gezogenen Karten,
// `N = Σ N_h` und `W_h = N_h / N`:
//
//	Schichtmittel      x̄_h = (1/n_h) · Σ x_hi
//	Schichtvarianz     s²_h = (1/(n_h−1)) · Σ (x_hi − x̄_h)²
//
//	Kanonmittel        x̄_str = Σ W_h · x̄_h
//
//	Varianz des        Var(x̄_str) = Σ W_h² · (s²_h / n_h) · (1 − n_h/N_h)
//	Schätzers                                              └─ Endlichkeits-
//	                                                          korrektur
//	Standardfehler     SE = √Var(x̄_str)
//	Vertrauensband     x̄_str ± 1,96 · SE            (95 %, normalapproximiert)
//
// **Die Endlichkeitskorrektur `(1 − n_h/N_h)` gehört dazu**: bei Archenemy
// sind 91 % der Schicht gezogen, und der Schätzer ist dort fast exakt. Ohne
// sie stünde ein Fehlerband, das es nicht gibt.
//
// # Und für Median, Quartile und Perzentile
//
// **Eine gewichtete empirische Verteilungsfunktion.** Jede gezogene Karte
// trägt das Gewicht `w_hi = N_h / n_h` — sie vertritt so viele Karten des
// Kanons. Sortiert man alle Karten nach ihrem Abstand und summiert die
// Gewichte, ist das `p`-Quantil der erste Wert, an dem die kumulierte
// Gewichtssumme `p · Σw` erreicht:
//
//	Q(p) = min { x_(k) : Σ_{i≤k} w_(i) ≥ p · Σ w_i }
//
// **Das ist dieselbe Nächste-Rang-Regel wie ungewichtet**, nur zählt jede
// Karte so oft, wie sie vertritt. Bei gleichen Anteilen fällt beides zusammen.
//
// # Was diese Zahlen nicht sind
//
//   - **Kein Vertrauensband um den Median.** Für Quantile aus einer
//     geschichteten Stichprobe gibt es keinen geschlossenen Ausdruck, der
//     ohne weitere Annahmen auskommt; ein Bootstrap wäre der Weg, und er
//     steht hier nicht.
//   - **Keine Aussage über Vorlagen, die in keiner Schicht vorkommen** (§10).

import (
	"math"
	"sort"
	"testing"
)

// schicht ist eine Schicht der geschichteten Stichprobe — hier: ein Set.
type schicht struct {
	name string

	// ganz ist N_h — die Karten der Schicht im Kanon.
	ganz int

	// werte sind die n_h gezogenen Abstände, **sortiert**.
	werte []float64
}

// kanonschaetzung ist der Schätzer über alle Schichten.
type kanonschaetzung struct {
	schichten      int
	ganz, gezogen  int
	mittel         float64
	varianzMittel  float64
	standardfehler float64
	unten, oben    float64

	// Die Verteilung über die **gewichtete** empirische Verteilungsfunktion.
	kleinster, groesster         float64
	q1, median, q3               float64
	p90, p95, p99                float64
	streuung, standardabweichung float64
}

// schaetzeKanon rechnet den geschichteten Schätzer.
func schaetzeKanon(schichten []schicht) kanonschaetzung {
	var (
		ergebnis kanonschaetzung
		ganz     int
	)

	for _, s := range schichten {
		if len(s.werte) == 0 {
			continue
		}

		ganz += s.ganz
		ergebnis.gezogen += len(s.werte)
		ergebnis.schichten++
	}

	if ganz == 0 {
		return ergebnis
	}

	ergebnis.ganz = ganz

	// **Mittel und Varianz des Schätzers** — die Gleichungen stehen im Kopf.
	for _, s := range schichten {
		if len(s.werte) == 0 {
			continue
		}

		anteil := float64(s.ganz) / float64(ganz)
		je := verteilungAus(s.werte)

		ergebnis.mittel += anteil * je.mittel

		gezogen := float64(len(s.werte))
		korrektur := 1 - gezogen/float64(s.ganz)
		ergebnis.varianzMittel += anteil * anteil * (je.streuung / gezogen) * korrektur
	}

	ergebnis.standardfehler = math.Sqrt(ergebnis.varianzMittel)
	ergebnis.unten = ergebnis.mittel - 1.96*ergebnis.standardfehler
	ergebnis.oben = ergebnis.mittel + 1.96*ergebnis.standardfehler

	ergebnis.verteilungAusGewichten(schichten)

	return ergebnis
}

// gewichteterWert ist ein Abstand samt der Zahl der Karten, die er vertritt.
type gewichteterWert struct {
	wert, gewicht float64
}

// verteilungAusGewichten füllt Quantile und Streuung über die gewichtete
// empirische Verteilungsfunktion.
func (k *kanonschaetzung) verteilungAusGewichten(schichten []schicht) {
	var alle []gewichteterWert

	for _, s := range schichten {
		if len(s.werte) == 0 {
			continue
		}

		// **`w = N_h / n_h`** — so viele Karten vertritt eine gezogene.
		gewicht := float64(s.ganz) / float64(len(s.werte))
		for _, wert := range s.werte {
			alle = append(alle, gewichteterWert{wert: wert, gewicht: gewicht})
		}
	}

	if len(alle) == 0 {
		return
	}

	sort.Slice(alle, func(i, j int) bool { return alle[i].wert < alle[j].wert })

	k.kleinster = alle[0].wert
	k.groesster = alle[len(alle)-1].wert
	k.q1 = gewichtetesQuantil(alle, 0.25)
	k.median = gewichtetesQuantil(alle, 0.50)
	k.q3 = gewichtetesQuantil(alle, 0.75)
	k.p90 = gewichtetesQuantil(alle, 0.90)
	k.p95 = gewichtetesQuantil(alle, 0.95)
	k.p99 = gewichtetesQuantil(alle, 0.99)

	// **Die gewichtete Streuung der Karten** — nicht die des Schätzers.
	//
	// Sie sagt, wie weit die Karten des Kanons auseinanderliegen;
	// `varianzMittel` sagt, wie genau ihr Mittel geschätzt ist. Zwei Fragen,
	// zwei Zahlen, und sie zu verwechseln ist der häufigste Fehler mit einer
	// Stichprobe.
	//
	//	s²_w = Σ w_i (x_i − x̄_w)² / (Σ w_i − 1)
	var summeGewicht, summeWert float64
	for _, g := range alle {
		summeGewicht += g.gewicht
		summeWert += g.gewicht * g.wert
	}

	mittelGewichtet := summeWert / summeGewicht

	var quadrate float64
	for _, g := range alle {
		abweichung := g.wert - mittelGewichtet
		quadrate += g.gewicht * abweichung * abweichung
	}

	if summeGewicht > 1 {
		k.streuung = quadrate / (summeGewicht - 1)
		k.standardabweichung = math.Sqrt(k.streuung)
	}
}

// gewichtetesQuantil gibt Q(p) über die gewichtete Verteilungsfunktion.
//
// **Die Werte müssen sortiert sein** — der Aufrufer tut es.
func gewichtetesQuantil(sortiert []gewichteterWert, anteil float64) float64 {
	if len(sortiert) == 0 {
		return 0
	}

	var summe float64
	for _, g := range sortiert {
		summe += g.gewicht
	}

	schwelle := anteil * summe

	var bisher float64
	for _, g := range sortiert {
		bisher += g.gewicht
		if bisher >= schwelle {
			return g.wert
		}
	}

	return sortiert[len(sortiert)-1].wert
}

// **Der geschichtete Schätzer wird an einem Fall geprüft, den man nachrechnen
// kann.**
//
// # Der Aufbau
//
// Zwei Schichten, absichtlich sehr ungleich gezogen:
//
//	A: N=100, alle Werte 10, n=10   → W=100/1100, x̄=10, s²=0
//	B: N=1000, alle Werte 20, n=10  → W=1000/1100, x̄=20, s²=0
//
// **Der ungewichtete Topf gäbe 15** — zehn Zehner und zehn Zwanziger. Der
// geschichtete Schätzer gibt
//
//	x̄_str = (100/1100)·10 + (1000/1100)·20 = (1000 + 20000)/1100 = 19,0909…
//
// **Und der Median ebenso**: gewichtet vertritt jede A-Karte 10 Karten, jede
// B-Karte 100. Von 1100 Karten tragen 1000 den Wert 20 — der Median ist 20,
// nicht 15.
//
// Das ist der ganze Befund in einem Beispiel: über 4 Punkte Unterschied im
// Mittel und 5 im Median, allein durch die Gewichtung.
func TestDerGeschichteteSchaetzerGewichtetNachSchichtgroesse(t *testing.T) {
	t.Parallel()

	schichten := []schicht{
		{name: "A", ganz: 100, werte: gleicheWerte(10, 10)},
		{name: "B", ganz: 1000, werte: gleicheWerte(20, 10)},
	}

	geschaetzt := schaetzeKanon(schichten)

	if erwartet := (100.0*10 + 1000.0*20) / 1100; math.Abs(geschaetzt.mittel-erwartet) > 1e-9 {
		t.Errorf("Kanonmittel %.6f, erwartet %.6f", geschaetzt.mittel, erwartet)
	}

	// **Der ungewichtete Topf läge bei 15** — der Unterschied ist der Punkt.
	if geschaetzt.mittel < 19 {
		t.Errorf("Kanonmittel %.4f — das sieht nach dem ungewichteten Topf aus",
			geschaetzt.mittel)
	}

	if geschaetzt.median != 20 {
		t.Errorf("gewichteter Median %.4f, erwartet 20", geschaetzt.median)
	}

	if geschaetzt.ganz != 1100 || geschaetzt.gezogen != 20 {
		t.Errorf("N=%d, n=%d — erwartet 1100 und 20", geschaetzt.ganz, geschaetzt.gezogen)
	}
}

// **Die Endlichkeitskorrektur macht eine fast vollständige Schicht exakt.**
//
// Wird eine Schicht **ganz** gezogen, ist ihr Mittel kein Schätzer mehr,
// sondern die Wahrheit — und trägt keinen Fehler bei. Ohne `(1 − n_h/N_h)`
// stünde dort ein Fehlerband, das es nicht gibt.
func TestEineGanzGezogeneSchichtTraegtKeinenFehlerBei(t *testing.T) {
	t.Parallel()

	// **Streuende Werte**, sonst wäre `s²_h = 0` und die Aussage leer: die
	// Korrektur wäre nicht zu unterscheiden von „es gibt nichts zu streuen".
	streuend := []float64{10, 12, 14, 16, 18}

	ganz := schaetzeKanon([]schicht{{name: "A", ganz: 5, werte: streuend}})
	if ganz.standardfehler != 0 {
		t.Errorf("Standardfehler %.6f bei vollständiger Erhebung — erwartet 0",
			ganz.standardfehler)
	}

	// **Und eine halb gezogene trägt einen.**
	halb := schaetzeKanon([]schicht{{name: "A", ganz: 10, werte: streuend}})
	if halb.standardfehler <= 0 {
		t.Errorf("Standardfehler %.6f bei halber Stichprobe — erwartet über 0",
			halb.standardfehler)
	}
}

// **Und ohne Schichten gibt es Nullen statt einer Division durch null.**
func TestOhneSchichtenGibtEsNullen(t *testing.T) {
	t.Parallel()

	if leer := schaetzeKanon(nil); leer.ganz != 0 || leer.mittel != 0 {
		t.Errorf("%+v — erwartet der Nullverbund", leer)
	}

	if leer := schaetzeKanon([]schicht{{name: "A", ganz: 5}}); leer.ganz != 0 {
		t.Errorf("eine Schicht ohne Werte zählt mit: %+v", leer)
	}
}

// gleicheWerte gibt `anzahl` Mal denselben Wert.
func gleicheWerte(wert float64, anzahl int) []float64 {
	werte := make([]float64, anzahl)
	for i := range werte {
		werte[i] = wert
	}

	return werte
}
