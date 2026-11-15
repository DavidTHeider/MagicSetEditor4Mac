package app

// **Die Rechnung hinter einer Entdeckungszusage** — QA-02
// (#1734).
//
// Ohne Build-Tag, damit diese Läufe in jeder Pipeline laufen: die Läufe
// hinter `fremdpakete` benutzen die Rechnung, prüfen aber nur, was an echten
// Sets herauskommt.

import (
	"math"
	"testing"
)

// **Der Gegenfall aus QA-02.AK-04**: 107 Karten, zwei davon betroffen, 93
// gezogen. Die Endlichkeitskorrektur versprach 99,9 Prozent; exakt sind es
// rund 98,3953.
func TestDieEntdeckungWirdHypergeometrischGerechnet(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		gesamt, betroffen, gezogen int
		erwartet                   float64
	}{
		"der Gegenfall aus dem Issue":         {107, 2, 93, 1 - 182.0/11342},
		"mit 104 Karten genug":                {107, 2, 104, 1 - 6.0/11342},
		"mit 103 Karten nicht":                {107, 2, 103, 1 - 12.0/11342},
		"ohne Betroffene nichts zu finden":    {107, 0, 50, 0},
		"wer alles zieht, findet sicher":      {107, 2, 107, 1},
		"wer mehr zieht als Unbetroffene":     {107, 2, 106, 1},
		"eine Karte, eine betroffen, gezogen": {1, 1, 1, 1},
		"nichts gezogen, nichts gefunden":     {107, 2, 0, 0},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gerechnet := findewahrscheinlichkeit(fall.gesamt, fall.betroffen, fall.gezogen)
			if math.Abs(gerechnet-fall.erwartet) > 1e-12 {
				t.Errorf("findewahrscheinlichkeit(%d, %d, %d) = %.10f, erwartet %.10f",
					fall.gesamt, fall.betroffen, fall.gezogen, gerechnet, fall.erwartet)
			}
		})
	}

	// **Auf vier Stellen, wie das Issue es nennt.**
	if gerundet := math.Round(findewahrscheinlichkeit(107, 2, 93)*1e6) / 1e4; gerundet != 98.3953 {
		t.Errorf("98,3953 %% erwartet, gerechnet %.4f %%", gerundet)
	}
}

// **Die nötige Stichprobe ist die kleinste, die es exakt schafft** — für den
// Gegenfall 104. **Hier stand bis zum 2026-09-22 eine Rechnung, die 93 gab**:
// `n∞ / (1 + (n∞ − 1)/N)` ist die Endlichkeitskorrektur für die **Varianz**
// einer Schätzung und keine Wahrscheinlichkeit, etwas zu finden (AK-02).
func TestDieNoetigeStichprobeIstDieKleinsteDieReicht(t *testing.T) {
	t.Parallel()

	if n := noetigeStichprobe(0.999, 0.01, 107); n != 104 {
		t.Errorf("für 107 Karten %d, erwartet 104", n)
	}

	// **Für jede Größe gilt**: die gefundene Zahl reicht, eine weniger nicht.
	for _, gesamt := range []int{1, 2, 10, 99, 100, 101, 107, 580, 1000, 3216, 3788, 25148} {
		n := noetigeStichprobe(0.999, 0.01, gesamt)
		betroffen := betroffeneBei(0.01, gesamt)

		if findewahrscheinlichkeit(gesamt, betroffen, n) < 0.999 {
			t.Errorf("N=%d: %d Karten reichen nicht", gesamt, n)
		}

		if n > 0 && findewahrscheinlichkeit(gesamt, betroffen, n-1) >= 0.999 {
			t.Errorf("N=%d: %d Karten reichten schon", gesamt, n-1)
		}
	}
}

// **Betroffen ist die kleinste ganze Zahl, die die Quote erreicht** — ein
// Prozent von 107 sind 1,07 Karten, und eine Karte wäre weniger als ein
// Prozent.
func TestBetroffenWirdAufgerundet(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		gesamt, erwartet int
	}{{107, 2}, {100, 1}, {101, 2}, {3788, 38}, {1, 1}, {0, 0}} {
		if gerechnet := betroffeneBei(0.01, fall.gesamt); gerechnet != fall.erwartet {
			t.Errorf("1 %% von %d: %d, erwartet %d", fall.gesamt, gerechnet, fall.erwartet)
		}
	}
}

// **Je Schicht und gemeinsam sind zwei Zusagen** (AK-03). Wer zehn Schichten
// je mit 99,9 Prozent zieht, trifft **in allen zugleich** nur mit dem Produkt
// — 99,00 Prozent, nicht 99,9.
func TestDieGemeinsameSicherheitIstDasProdukt(t *testing.T) {
	t.Parallel()

	zehn := make([]float64, 10)
	for i := range zehn {
		zehn[i] = 0.999
	}

	if gemeinsam := gemeinsameSicherheit(zehn); math.Abs(gemeinsam-math.Pow(0.999, 10)) > 1e-12 ||
		math.Round(gemeinsam*1e4)/100 != 99.00 {
		t.Errorf("gemeinsam %.6f, erwartet 0,999^10 ≈ 0,990045", gemeinsam)
	}

	if gemeinsameSicherheit(nil) != 1 {
		t.Error("ohne Schichten ist nichts zu verfehlen — erwartet 1")
	}
}

// **Die feste Kohorte bleibt, wie sie war** (AK-01) — sonst wäre sie keine
// Vergleichskohorte mehr, und jede Zahl über die Zeit verglich zwei
// verschiedene Mengen. Die Werte stehen so in `docs/MESSMETHODIK.md`.
func TestDieFesteKohorteBleibtWieSieWar(t *testing.T) {
	t.Parallel()

	for gesamt, erwartet := range map[int]int{107: 93, 3788: 582, 3689: 580, 25148: 670} {
		if n := kohortengroesse(gesamt); n != erwartet {
			t.Errorf("die feste Kohorte aus %d Karten hat %d, erwartet %d", gesamt, n, erwartet)
		}
	}
}
