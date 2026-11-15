package app

// **Zwei Rechnungen für zwei Kohorten** — QA-02
// (#1734).
//
// Ohne Build-Tag, damit ihre Läufe (`stichprobenrechnung_test.go`) in jeder
// Pipeline laufen; benutzt werden sie von den Läufen hinter `fremdpakete`.
//
// # Die feste Kohorte
//
// **Sie ist die Vergleichskohorte** (AK-01): dieselben Karten in jedem Lauf,
// damit eine Zahl, die sich bewegt, eine Änderung am Programm ist und kein
// Rauschen. Ihre Größe kommt aus der Rechnung, die hier seit dem 2026-09-07
// steht — und **sie trägt keine Wahrscheinlichkeitszusage mehr**: die Formel
// ist die Endlichkeitskorrektur für die Varianz einer Schätzung, keine
// Entdeckungswahrscheinlichkeit, und die Auswahl ist systematisch statt
// zufällig. Geändert wird sie nicht, sonst verlöre sie ihren Zweck.
//
// # Die Zufallsstichprobe
//
// **Für eine Zusage wie „mit 99,9 Prozent entdeckt"** — exakt
// hypergeometrisch gerechnet (AK-02) und aus einem gespeicherten Plan
// gezogen.

import "math"

// kohortengroesse gibt die Zahl der Karten der **festen** Kohorte je Set.
//
//	n∞ = ln(1 − 0,999) / ln(1 − 0,01)        ≈ 687,3
//	n  = n∞ / (1 + (n∞ − 1) / N)             Endlichkeitskorrektur
//
// Für 107 Karten 93, für 3.788 Karten 582 — so hat jeder Lauf seit dem
// 2026-09-07 gezogen (Vanguard und Black in `README.md`, „Je Set").
func kohortengroesse(gesamt int) int {
	unendlich := math.Log(1-0.999) / math.Log(1-0.01)

	return int(math.Ceil(unendlich / (1 + (unendlich-1)/float64(gesamt))))
}

// findewahrscheinlichkeit gibt die Wahrscheinlichkeit, beim Ziehen von `gezogen` aus
// `gesamt` Karten **ohne Zurücklegen** mindestens eine von `betroffen` zu
// treffen — exakt hypergeometrisch (AK-02):
//
//	P(keine) = C(N−K, n) / C(N, n) = ∏_{i<K} (N−n−i) / (N−i)
//
// **Das Produkt läuft über K und nicht über n**: ein Prozent von 3.788 sind
// 38 Faktoren statt 580.
func findewahrscheinlichkeit(gesamt, betroffen, gezogen int) float64 {
	betroffen = min(betroffen, gesamt)
	if betroffen <= 0 || gezogen <= 0 {
		return 0
	}

	keine := 1.0

	for i := range betroffen {
		faktor := float64(gesamt-gezogen-i) / float64(gesamt-i)
		if faktor <= 0 {
			return 1
		}

		keine *= faktor
	}

	return 1 - keine
}

// betroffeneBei gibt die kleinste Zahl von Karten, die eine Quote erreicht —
// ein Prozent von 107 sind 1,07 Karten, und eine wäre weniger als ein
// Prozent.
//
// **Das Epsilon fängt die Darstellung**: 0,01 · 100 soll eins sein und nicht
// zwei, weil 0,01 im Rechner ein wenig mehr ist als ein Hundertstel.
func betroffeneBei(quote float64, gesamt int) int {
	return int(math.Ceil(quote*float64(gesamt) - 1e-9))
}

// noetigeStichprobe gibt die kleinste Zahl von Karten, die bei `gesamt`
// Karten jeden Fehler, der die `quote` erreicht, mit `vertrauen` trifft.
func noetigeStichprobe(vertrauen, quote float64, gesamt int) int {
	betroffen := betroffeneBei(quote, gesamt)

	for gezogen := 0; gezogen <= gesamt; gezogen++ {
		if findewahrscheinlichkeit(gesamt, betroffen, gezogen) >= vertrauen {
			return gezogen
		}
	}

	return gesamt
}

// gemeinsameSicherheit gibt die Wahrscheinlichkeit, **in jeder** Schicht zu
// treffen, wenn die Schichten unabhängig gezogen werden (AK-03). Zehn
// Schichten je 99,9 Prozent sind gemeinsam 99,00 Prozent.
func gemeinsameSicherheit(jeSchicht []float64) float64 {
	gemeinsam := 1.0
	for _, sicherheit := range jeSchicht {
		gemeinsam *= sicherheit
	}

	return gemeinsam
}
