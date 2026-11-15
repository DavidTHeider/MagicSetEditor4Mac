package render

// **`shrink`: ein zu breiter Zug wird waagerecht gestaucht.**
//
// Die Ansage steht 676-mal in den 190 Stilen von `M15-Magic-Pack`, fast immer
// am Namen und an der Typzeile. Sie steht hier und nicht in `satz.go`, weil sie
// eine eigene Frage beantwortet: `satz.go` sagt, **wie** ein Zug umbricht und
// sich ausrichtet — hier steht, **wie schmal** er dafür wird.
//
// # Gestaucht, nicht verkleinert
//
// MSE lässt die Schriftgröße stehen und staucht die **Fläche** — ein langer
// Kartenname steht dort in voller Höhe und schmal
// (#319). Die
// Größe zu senken, bis der Zug hineinpasst, wäre ein anderes Modell und kein
// Näherungsfehler.

// stauchung gibt den waagerechten Faktor eines Zugs.
//
//	factor = (width - padding_left - padding_right) / layout->width
//
// `src/data/field/text.cpp`, Zeilen 73–81, Fassung f9d9356 vom 2021-07-06.
// **Das Polster ist schon abgezogen**: [SetzeAbschnitte] reicht den Rahmen
// **innen** herein (`rahmen = innen`), also `width - padding_left -
// padding_right`. Hier stand bis zum 2026-10-12, `msepaket` lese das Polster
// nicht — das stimmte längst nicht mehr (Zweitblick über die Folgen vor
// 0.8.0, Befund 13).
//
// Mit [Satz.NurBeiUeberlauf] wird nur gestaucht, nie gedehnt. Ohne
// [Satz.Stauchen] bleibt es bei 1.
func stauchung(satz Satz, rahmen Rechteck, breite float64) float64 {
	if !satz.Stauchen || rahmen.Breite <= 0 || satz.Groesse <= 0 {
		return 1
	}

	if breite <= 0 {
		return 1
	}

	faktor := rahmen.Breite / breite
	if satz.NurBeiUeberlauf && faktor >= 1 {
		return 1
	}

	return faktor
}
