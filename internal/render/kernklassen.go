package render

// **Kerning über Klassen — `PairPos` Format 2** — QA-03.AK-02
// (#1656).
//
// Statt jedes Paar einzeln zu nennen, teilt die Schrift ihre Glyphen in
// Klassen und nennt je Klassenpaar einen Wert. `beleren-bold` — die Schrift
// jedes Kartennamens — legt so den Normalfall ab; die Einzelpaare davor sind
// die Ausnahmen.
//
// **Die Namensschrift braucht es.** Über
// alle 25.011 Kartennamen des Kanons ohne Ligatur gegen Core Text gehalten —
// den Weg, auf dem MSE2 unter macOS setzt; wxWidgets stellt dort kein Merkmal
// um (`docs/SCHRIFTUNTERSUCHUNG.md`) —: mit Format 1
// allein stimmte die Breite bei **2.287**, mit beiden Formaten bei **allen**,
// jede auf ein Tausendstel Punkt. Im Mittel fehlten 1,8 Punkt je Namen bei
// Größe 16, beim längsten 18,5.

// klassenpaare ist eine `PairPos`-Untertabelle in Format 2.
type klassenpaare struct {
	// deckung sind die ersten Glyphen, für die die Untertabelle gilt.
	deckung map[uint16]bool

	// erste und zweite ordnen den Glyphen ihre Klasse zu; wer fehlt, ist
	// Klasse null.
	erste, zweite map[uint16]uint16

	// zeilen und spalten ist die Zahl der Klassen je Seite.
	zeilen, spalten int

	// werte ist die Matrix der Vorschübe, Zeile für Zeile — leer, wenn das
	// Wertformat keinen Vorschub trägt.
	werte []int16
}

func (k klassenpaare) wert(erste, zweite uint16) (int16, bool) {
	if !k.deckung[erste] {
		return 0, false
	}

	zeile, spalte := int(k.erste[erste]), int(k.zweite[zweite])

	// **Eine Klasse jenseits der Matrix gilt nicht** — die Datei widerspricht
	// sich dann selbst, und eine Zahl aus ihr wäre geraten.
	if zeile >= k.zeilen || spalte >= k.spalten {
		return 0, false
	}

	if k.werte == nil {
		return 0, true
	}

	return k.werte[zeile*k.spalten+spalte], true
}

// klassenpaareAus liest eine `PairPos`-Untertabelle in Format 2.
//
// **Erst nachsehen, dann zuteilen.** Zwei Zahlen im Kopf bestimmen, wie groß
// die Matrix ist — bis zu 65.535 mal 65.535 Einträge. Sie wird nur gelesen,
// wenn die Datei sie auch trägt; sonst bestimmte eine beschädigte Schrift den
// Arbeitsspeicher.
func klassenpaareAus(roh []byte, tabelle int) (klassenpaare, bool) {
	if tabelle+16 > len(roh) {
		return klassenpaare{}, false
	}

	format1, format2 := zwei(roh, tabelle+4), zwei(roh, tabelle+6)
	zeilen, spalten := int(zwei(roh, tabelle+12)), int(zwei(roh, tabelle+14))
	schritt := felder(format1) + felder(format2)

	if zeilen*spalten*schritt > len(roh)-tabelle-16 {
		return klassenpaare{}, false
	}

	klassen := klassenpaare{
		deckung: map[uint16]bool{},
		erste:   klassenAus(roh, tabelle+int(zwei(roh, tabelle+8))),
		zweite:  klassenAus(roh, tabelle+int(zwei(roh, tabelle+10))),
		zeilen:  zeilen,
		spalten: spalten,
	}

	for _, glyphe := range glyphenDerDeckung(roh, tabelle+int(zwei(roh, tabelle+2)), glyphenraum) {
		klassen.deckung[glyphe] = true
	}

	if format1&0x0004 != 0 {
		vorlauf := felder(format1 & 0x0003)
		klassen.werte = make([]int16, zeilen*spalten)

		for i := range klassen.werte {
			klassen.werte[i] = int16(zwei(roh, tabelle+16+i*schritt+vorlauf)) /* #nosec G115 -- Schriftdaten sind vorzeichenbehaftet */
		}
	}

	return klassen, true
}

// klassenAus liest eine `ClassDef`-Tabelle, Format 1 und 2.
//
// **Mit denselben Grenzen wie [glyphenDerDeckung]**: gezählt wird in `int`,
// und mehr Schritte, als es Glyphennummern gibt, macht keine Tabelle.
func klassenAus(roh []byte, tabelle int) map[uint16]uint16 {
	aus := map[uint16]uint16{}

	switch zwei(roh, tabelle) {
	case 1:
		erste := int(zwei(roh, tabelle+2))

		for i := range int(zwei(roh, tabelle+4)) {
			stelle := tabelle + 6 + i*2
			if stelle+2 > len(roh) || erste+i >= glyphenraum {
				return aus
			}

			aus[uint16(erste+i)] = zwei(roh, stelle) /* #nosec G115 -- unter glyphenraum */
		}

	case 2:
		schritte := 0

		for i := range int(zwei(roh, tabelle+2)) {
			stelle := tabelle + 4 + i*6
			if stelle+6 > len(roh) {
				return aus
			}

			for glyphe := int(zwei(roh, stelle)); glyphe <= int(zwei(roh, stelle+2)); glyphe++ {
				schritte++
				if schritte > glyphenraum {
					return aus
				}

				aus[uint16(glyphe)] = zwei(roh, stelle+4) /* #nosec G115 -- aus zwei Byte gelesen */
			}
		}
	}

	return aus
}
