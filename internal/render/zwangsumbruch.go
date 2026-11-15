package render

import "unicode/utf8"

// zwangsgebrochen bricht ein Wort, das allein breiter ist als seine Zeile,
// vor dem ersten Zeichen, das übersteht
// (Befund 3 des Nachlaufs über E/F).
//
// MSE prüft den Umbruch Zeichen für Zeichen; beginnt das Wort die Zeile und
// reicht es über `lineRight` hinaus, wird es dort gebrochen, und das
// überstehende Zeichen beginnt die nächste Zeile. **Ist es das einzige auf
// der Zeile, bleibt es stehen** — `line.positions.size() > 2` zählt die
// Startstelle mit (`TextViewer::prepareLinesAtScale`,
// `src/render/text/viewer.cpp`, Zeilen 614–633 und 651–658, Fassung f9d9356
// vom 2021-07-06).
//
// **Ein Symbolzug ist ein Zeichen**: er bleibt ganz, wie überall beim
// Umbruch ([teileAus]). Bei MSE ist jedes Symbol ein eigenes Zeichen; ein
// Zug aus mehreren, der allein nicht in eine Zeile passt, steht hier über.
//
// rest ist leer, wenn sich nichts brechen lässt — bei einem Wort aus einem
// einzigen Zeichen.
func zwangsgebrochen(
	wort []Stueck, frei float64, satz Satz, schriften Schriftsatz,
) (kopf, rest []Stueck) {
	stellen := zeichenstellen(wort)

	// ueber ist das erste Zeichen, mit dem der Kopf übersteht. Das ganze
	// Wort steht über, sonst wäre es nicht hier; das letzte muss deshalb
	// nicht mehr gemessen werden.
	ueber := 0
	for ueber < len(stellen)-1 {
		versuch, _ := geteilt(wort, stellen[ueber], satz, schriften)
		if breiteDerStuecke(versuch, satz) > frei {
			break
		}

		ueber++
	}

	return geteilt(wort, stellen[max(ueber-1, 0)], satz, schriften)
}

// zeichenstelle nennt das Ende eines Zeichens: im wievielten Stück, und bis
// zu welchem Byte seines Texts.
type zeichenstelle struct{ stueck, ende int }

// zeichenstellen zählt die Zeichen eines Worts auf, über die Stücke hinweg.
func zeichenstellen(wort []Stueck) []zeichenstelle {
	var stellen []zeichenstelle

	for nummer, stueck := range wort {
		if len(stueck.Symbole) > 0 {
			stellen = append(stellen, zeichenstelle{stueck: nummer, ende: len(stueck.Text)})

			continue
		}

		for ende := 0; ende < len(stueck.Text); {
			_, laenge := utf8.DecodeRuneInString(stueck.Text[ende:])
			ende += laenge
			stellen = append(stellen, zeichenstelle{stueck: nummer, ende: ende})
		}
	}

	return stellen
}

// geteilt schneidet ein Wort hinter einem Zeichen.
//
// **Der Kopf wird gemessen, der Rest ist der Unterschied**: so bleibt die
// Summe der beiden die Breite des Stücks, wie sie im Abschnitt gemessen
// wurde ([teileAus]).
func geteilt(
	wort []Stueck, stelle zeichenstelle, satz Satz, schriften Schriftsatz,
) (kopf, rest []Stueck) {
	stueck := wort[stelle.stueck]
	kopf = append([]Stueck{}, wort[:stelle.stueck]...)

	if stelle.ende == len(stueck.Text) {
		return append(kopf, stueck), append([]Stueck{}, wort[stelle.stueck+1:]...)
	}

	lauf := vorsilbe{ersatz: schriften.Ersatz}
	vorn, hinten := stueck, stueck
	vorn.Text, hinten.Text = stueck.Text[:stelle.ende], stueck.Text[stelle.ende:]
	vorn.Breite = lauf.nimm(vorn.Text, satz, schriften.fuer(stueck.Kursiv, stueck.Fett))
	hinten.Breite = max(stueck.Breite-vorn.Breite, 0)

	return append(kopf, vorn), append([]Stueck{hinten}, wort[stelle.stueck+1:]...)
}
