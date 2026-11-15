package render

import "slices"

// **Ein zu kurzer Zug wird auf die Kastenhöhe gedehnt.**
//
// MSE verteilt den freien Platz auf die Abstände **zwischen** den Zeilen, und
// zwar bis zu `line height … max` je Umbruchart. Die Rechnung steht in
// `alignParagraph` (`src/render/text/viewer.cpp`, Zeilen 780–815, Fassung
// f9d9356 vom 2021-07-06) und ist hier Zug um Zug nachgebildet — auch dort, wo
// sie überrascht; der Grund steht bei [gedehnt].

// gedehnt verteilt den freien Platz auf die Zeilenabstände.
//
// `oben` sind die Oberkanten, `hoehen` die Zeilenhöhen, `arten` die Umbruchart
// je Zeile. Zurück kommen die neuen Oberkanten und die neue Gesamthöhe.
//
// # Wie MSE es rechnet
//
// Für jede Art gibt es einen Spielraum `d = max(0, max − faktor)`. Die drei
// Spielräume werden sortiert und in Stufen abgearbeitet; in jeder Stufe
// bekommen die Zeilen, deren Art die Stufe noch tragen kann, denselben
// Zuschlag — begrenzt auf das, was der Kasten noch hergibt.
//
// **Verglichen wird der Spielraum mit dem Zuwachs der Stufe, nicht mit dem
// erreichten Stand** (`bool soft = d_soft >= stop;`). Eine Art kann damit über
// ihren eigenen Höchstwert hinaus wachsen, wenn die Stufen ungleich groß sind.
// Das ist nachgebildet und nicht behoben: die Parität zu MSE2 ist das Ziel,
// und ein „richtigerer" Zug säße anders als der, den ein Anwender kennt
// (#345).
func gedehnt(
	oben, hoehen []float64, arten []umbruchart, faktoren Zeilenfaktoren, kasten float64,
) (gedehnteOberkanten []float64, gedehnteHoehe float64) {
	// **Ohne Prüfung auf einen Zug ohne Zeilen.** [SetzeAbschnitte] gibt immer
	// mindestens eine — eine leere Zeile trennt zwei Absätze und bleibt
	// stehen. Ein Zweig dafür wäre einer, den kein Aufruf betreten kann (§3).
	hoehe := gesamthoehe(oben, hoehen)
	if hoehe >= kasten {
		return oben, hoehe
	}

	spielraum := map[umbruchart]float64{
		umbruchWeich: max(faktoren.WeichMax-oderEins(faktoren.Weich), 0),
		umbruchHart:  max(faktoren.HartMax-oderEins(faktoren.Hart), 0),
		umbruchLinie: max(faktoren.LinieMax-oderEins(faktoren.Linie), 0),
	}

	stufen := []float64{
		0, spielraum[umbruchWeich], spielraum[umbruchHart], spielraum[umbruchLinie],
	}
	slices.Sort(stufen[1:])

	for nummer := 1; nummer < len(stufen) && hoehe < kasten; nummer++ {
		zuwachs := stufen[nummer] - stufen[nummer-1]
		if zuwachs <= 0 {
			continue
		}

		summe := 0.0

		for stelle, art := range arten {
			if spielraum[art] >= zuwachs {
				summe += hoehen[stelle]
			}
		}

		if summe == 0 {
			break
		}

		je := min(zuwachs, (kasten-hoehe)/summe)

		zugabe := 0.0

		for stelle, art := range arten {
			oben[stelle] += zugabe

			if spielraum[art] >= zuwachs {
				zugabe += je * hoehen[stelle]
			}
		}

		hoehe += zugabe
	}

	return oben, hoehe
}

// gesamthoehe misst von der Oberkante der ersten bis zur Unterkante der letzten
// Zeile, **die nicht leer ist**.
//
//	// Find height of the text, don't count the last lines if they are empty
//	double height = 0;
//	for (size_t li = end_line - 1 ; li + 1 > start_line ; --li) {
//	  Line& l = lines[li];
//	  height = l.top + l.line_height;
//	  if (l.line_height) break; // not an empty line
//	}
//	height -= lines[start_line].top;
//
// `src/render/text/viewer.cpp`, Zeilen 771–778, Fassung f9d9356 vom
// 2021-07-06.
//
// # Was das Auslassen kostet, wenn man es nicht tut
//
// Das Kartenfeld `text` von `magic.mse-game` ist
//
//	combined_editor(field1: card.rule_text, separator: "<line>\n</line>", field2: card.flavor_text)
//
// und bei leerem Flavortext endet sein Wert auf `<line>\n</line>` und nichts
// dahinter. Ginge diese leere Zeile in die Höhe ein, säße der Block um **die
// halbe `<line>`-Sprunghöhe** zu hoch — 13 Bildpunkte bei
// `A-Cloister Gargoyle`, 12 bei `Choose Your Demise`. Betroffen wären **9.380
// von 25.148** kanonischen Karten
// (#528).
//
// **Nur die *letzten*.** Der Kommentar im Altcode sagt es wörtlich: „don't
// count the **last** lines if they are empty". Eine leere Zeile zwischen zwei
// Absätzen trägt ihre Höhe wie jede andere.
func gesamthoehe(oben, hoehen []float64) float64 {
	for letzte := len(oben) - 1; letzte > 0; letzte-- {
		if hoehen[letzte] != 0 {
			return oben[letzte] + hoehen[letzte] - oben[0]
		}
	}

	return oben[0] + hoehen[0] - oben[0]
}

// umbruchart sagt, womit eine Zeile endet.
//
// **Die letzte Zeile endet mit nichts.** Bei MSE ist `break_after` dort
// `LineBreak::NO` (`Line()`, Zeile 29) — hinter ihr kommt kein Abstand, der
// sich dehnen ließe.
type umbruchart int

const (
	umbruchKeiner umbruchart = iota
	umbruchWeich
	umbruchHart
	umbruchLinie
)

// umbruchartenVon liest die Umbruchart je Zeile.
func umbruchartenVon(roh []rohzeile) []umbruchart {
	arten := make([]umbruchart, len(roh))

	for stelle := range roh {
		switch {
		case stelle == len(roh)-1:
			arten[stelle] = umbruchKeiner

		case roh[stelle].blockende:
			arten[stelle] = umbruchLinie

		case roh[stelle].weich:
			arten[stelle] = umbruchWeich

		case roh[stelle].absatzende:
			arten[stelle] = umbruchHart

		default:
			arten[stelle] = umbruchWeich
		}
	}

	return arten
}

// faktorFuer gibt den Abstandsfaktor einer Umbruchart.
func (z Zeilenfaktoren) faktorFuer(art umbruchart) float64 {
	switch art {
	case umbruchLinie:
		return oderEins(z.Linie)

	case umbruchHart:
		return oderEins(z.Hart)

	case umbruchWeich, umbruchKeiner:
	}

	return oderEins(z.Weich)
}
