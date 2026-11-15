package render

import "encoding/binary"

// typometrik ist die Metrik aus der Tabelle `OS/2` einer Schriftdatei.
//
// # Wozu sie hier gelesen wird
//
// **Weil `x/image/font/sfnt` sie nicht herausgibt** und weil es Schriften
// gibt, deren `hhea` leer ist: `MATRIX.TTF` aus `Full-Magic-Pack` nennt dort
// asc = 0, desc = 0, gap = 0 — und `err = nil` dazu.
//
// wxWidgets fällt in diesem Fall auf die **typografischen** Werte aus `OS/2`
// zurück, nicht auf den Tintenkasten. Gemessen am 2026-08-20 über zwanzig
// Punktgrößen: für `MATRIX.TTF` trifft `ceil(0,860 · Punktgröße)` die Höhe und
// `ceil(0,633 · Punktgröße)` die Grundlinie, und genau diese beiden Werte
// stehen in seiner `OS/2`-Tabelle
// (#597).
//
// # Warum nicht `win`
//
// **`OS/2` führt zwei Metriken**, und die andere ist der Tintenkasten: bei
// `MATRIX.TTF` gibt `usWinAscent + usWinDescent` genau die 957, die der
// bisherige Rückfall über die Ränder ergab. Sie ist es nicht — gemessen.
type typometrik struct {
	Zeilenhoehe float64
	Grundlinie  float64
}

// Die Stellen in den Tabellen, alle aus der OpenType-Spezifikation.
const (
	tabellenkopf    = 12 // vor dem ersten Tabelleneintrag
	tabelleneintrag = 16
	emStelle        = 18 // `head.unitsPerEm`
	typoStelle      = 68 // `OS/2.sTypoAscender`, dann Descender und LineGap
)

// typoAus liest die typografische Metrik aus den rohen Bytes einer Schrift.
//
// **`da` ist falsch, wenn die Datei sie nicht hergibt** — kein Fehler, sondern
// die Auskunft, dass der nächste Rückfall dran ist.
func typoAus(roh []byte) (metrik typometrik, da bool) {
	kopf, os2 := tabellen(roh)
	if len(kopf) < emStelle+2 || len(os2) < typoStelle+6 {
		return typometrik{}, false
	}

	einheit := float64(binary.BigEndian.Uint16(kopf[emStelle:]))
	if einheit <= 0 {
		return typometrik{}, false
	}

	aufstieg := vorzeichenbehaftet(os2[typoStelle:])
	abstieg := vorzeichenbehaftet(os2[typoStelle+2:])
	zwischenraum := vorzeichenbehaftet(os2[typoStelle+4:])

	hoehe := aufstieg - abstieg + zwischenraum
	if hoehe <= 0 || aufstieg <= 0 {
		return typometrik{}, false
	}

	return typometrik{
		Zeilenhoehe: hoehe / einheit,
		Grundlinie:  aufstieg / einheit,
	}, true
}

// tabellen sucht `head` und `OS/2` im Tabellenverzeichnis.
//
// **Eine Sammlung (`ttcf`) wird nicht gelesen.** Ihr Kopf ist ein anderer, und
// welche der enthaltenen Schriften gemeint wäre, weiß diese Stelle nicht —
// dann bleibt es beim nächsten Rückfall.
func tabellen(roh []byte) (kopf, os2 []byte) {
	if len(roh) < tabellenkopf {
		return nil, nil
	}

	anzahl := int(binary.BigEndian.Uint16(roh[4:]))
	if len(roh) < tabellenkopf+anzahl*tabelleneintrag {
		return nil, nil
	}

	for i := range anzahl {
		eintrag := roh[tabellenkopf+i*tabelleneintrag:]

		stelle := int(binary.BigEndian.Uint32(eintrag[8:]))
		laenge := int(binary.BigEndian.Uint32(eintrag[12:]))

		if stelle < 0 || laenge < 0 || stelle+laenge > len(roh) {
			continue
		}

		switch string(eintrag[0:4]) {
		case "head":
			kopf = roh[stelle : stelle+laenge]
		case "OS/2":
			os2 = roh[stelle : stelle+laenge]
		}
	}

	return kopf, os2
}

// vorzeichenbehaftet liest die ersten zwei Bytes als `int16`.
//
// **Die drei typo-Werte sind in der OpenType-Spezifikation `int16`**, und der
// Descender ist dort regelmäßig negativ — bei `MATRIX.TTF` −167. Wer sie als
// `uint16` läse, bekäme 65.369 und daraus eine Zeilenhöhe von 65 em.
func vorzeichenbehaftet(b []byte) float64 {
	return float64(int16(binary.BigEndian.Uint16(b))) // #nosec G115 -- int16 ist die Form in der Spezifikation
}
