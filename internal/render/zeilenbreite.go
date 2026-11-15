package render

import "strings"

// harteStuecke schneidet die weichen Stücke am **Ende** einer Zeile ab.
//
//	size_t end_or_soft; ///< Index just beyond the last non-soft character
//	double width() const { return positions[end_or_soft-start] - positions.front(); }
//
// `src/render/text/viewer.cpp`, Zeilen 17–35, Fassung f9d9356 vom 2021-07-06.
//
// **Nur am Ende, nicht mittendrin.** MSE merkt sich die Stelle **hinter dem
// letzten** nicht weichen Zeichen (`if (!c.soft) word_end_or_soft = i + 1`,
// Zeile 610) — ein weiches Stück zwischen zwei harten zählt also mit. Nur der
// Schwanz fällt weg, und genau der ist der Trenner, den `sub_type_filter`
// anhängt (#345).
func harteStuecke(stuecke []Stueck) []Stueck {
	ende := len(stuecke)
	for ende > 0 && stuecke[ende-1].Weich {
		ende--
	}

	return stuecke[:ende]
}

// breiteDerStuecke misst eine Zeile über ihre Stücke hinweg.
func breiteDerStuecke(stuecke []Stueck, satz Satz) float64 {
	breite, buchstaben := 0.0, 0

	for _, stueck := range stuecke {
		breite += stueck.Breite

		if len(stueck.Symbole) == 0 {
			buchstaben += buchstabenIn(stueck.Text)
		}
	}

	// **Die Laufweite zwischen den Stücken kommt dazu.** [breiteVon] zählt sie
	// je Stück und lässt sie am Ende weg — über zwei Stücke fehlte sie an der
	// Naht. Gezählt wird deshalb über die ganze Zeile: so viele Zwischenräume,
	// wie es Buchstaben minus eins gibt, minus die schon gezählten.
	fehlend := max(buchstaben-1, 0)

	for _, stueck := range stuecke {
		if len(stueck.Symbole) == 0 {
			fehlend -= max(buchstabenIn(stueck.Text)-1, 0)
		}
	}

	return breite + float64(max(fehlend, 0))*satz.Laufweite
}

// stueckbreite misst ein Stück — in Buchstaben oder in Symbolen.
//
// **Jedes Symbol einzeln.** Sie sind verschieden breit: `magic-mana-large`
// führt Bilder von 165 und von 176 Bildpunkten Breite, und einzelne Symbole
// bringen eine eigene `image font size` mit
// (#330).
//
// **Sie ist gemessen, nicht gerechnet** (#598). [teileAus] setzt sie beim
// Zerlegen, weil die Vorsilbe ihren Abschnitt kennen muss; hier steht nur noch
// die Rechnung für die Symbole, die keine Vorsilbe haben.
func symbolbreite(stueck Stueck, bilder Bildquelle) float64 {
	breite := 0.0

	for i := range stueck.Symbole {
		symbol := &stueck.Symbole[i]

		breite += stueck.Symbolmasse.kasten(symbol, bilder).Breite
	}

	return breite
}

// textAus fügt die Stücke einer Zeile zusammen.
func textAus(stuecke []Stueck) string {
	var bau strings.Builder

	for _, stueck := range stuecke {
		bau.WriteString(stueck.Text)
	}

	return bau.String()
}

// klartext ist der Zug ohne seine Schnitte.
func klartext(abschnitte []Abschnitt) string {
	var bau strings.Builder

	for _, abschnitt := range abschnitte {
		bau.WriteString(abschnitt.Text)
	}

	return bau.String()
}
