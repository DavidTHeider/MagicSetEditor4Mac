package render

import (
	"math"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// **Quartz' Schriftglättung, nachgebildet** —
// #1806.
//
// # Was Quartz tut
//
// Ein Bitmapkontext, wie wxWidgets ihn für MSE2 anlegt, glättet Schrift —
// ohne Angabe gezeichnet ist gleich mit `setShouldSmoothFonts(true)`, und
// wxWidgets schaltet nichts um. Die Glättung ist **keine Gammakurve, sondern
// eine Verdickung des Umrisses**: quer durch einen Stamm verschieben sich
// beide Kanten um denselben Bruchteil eines Bildpunkts nach außen, und der
// Verlauf dazwischen bleibt, wie er war (`docs/SCHRIFTUNTERSUCHUNG.md`).
//
// # Wie weit
//
// **Je Seite min(Geviert/64; 0,3) Bildpunkte**, waagerecht wie senkrecht —
// gemessen am 2026-09-23 mit `scripts/coretext/glaettung.swift` an MPlantin,
// Beleren Bold und Relay-Medium von 6 bis 128 Bildpunkten je Geviert, die
// senkrechte Richtung an einem um 90° gedrehten Stamm. Bis 19 Bildpunkte
// wächst sie mit der Größe (Verhältnis zu Geviert/64 zwischen 0,99 und 1,03,
// wo der Stamm dicker als ein Bildpunkt ist), darüber bleibt sie bei 0,30.
// MSE2 zeichnet Text vierfach — eine Regeltextzeile in 8 Punkt hat 32
// Bildpunkte je Geviert —, auf der Karte gilt also fast immer die Obergrenze.
//
// **Gemessen und nicht nachgelesen.** Quartz legt seine Rechnung nicht offen;
// die Gleichung ist die, die die Messung trägt, und wie weit sie trägt, steht
// in [0130](../../docs/entscheidungen/0130-quartz-glaettung-nachgebildet.md).

// quartzGrenze ist die Verdickung je Seite, über die Quartz nicht hinausgeht,
// in Bildpunkten.
const quartzGrenze = 0.3

// quartzVerdickung gibt, um wie viele Bildpunkte Quartz einen Umriss je Seite
// verdickt, bei einer Geviertgröße von `ppem` Bildpunkten.
func quartzVerdickung(ppem float64) float64 {
	return min(ppem/64, quartzGrenze)
}

// verdickt verschiebt jeden Punkt eines Umrisses nach außen, so dass jede
// gerade Kante um `um` Bildpunkte wandert.
//
// **Das Gerüst ist FreeTypes** `FT_Outline_EmboldenXY` (`src/base/ftoutln.c`,
// nachgelesen am 2026-09-23): jeder Punkt — auch ein Stützpunkt einer Kurve —
// wandert dorthin, wo sich seine zwei verschobenen Kanten schneiden. Bei einer
// Wende über rund 160 Grad bleibt er stehen, und an kurzen Kanten wird die
// Verschiebung begrenzt, damit die Kante nicht umklappt. **Anders als
// FreeType** rückt eine schräge Kante weiter als eine gerade ([gehrung]), und
// der Umriss rückt danach nicht als Ganzes nach rechts oben: Quartz verdickt
// nach beiden Seiten gleich, gemessen an beiden Kanten eines Stamms.
//
// **Senkrecht wächst er nur nach oben.** Quer durch einen waagerechten Stamm
// rückt die Oberkante um das Doppelte, und die Unterkante bleibt stehen —
// gemessen am Bindestrich von Relay-Medium bei 56 Bildpunkten: oben 81 → 235,
// unten 114 → 114. Der Umriss wird deshalb nach der Verdickung um `um`
// angehoben; das ist FreeTypes Richtung in der Senkrechten, in der
// Waagerechten nicht.
//
// **Welche Seite außen ist, sagt die Fläche des ganzen Umrisses** — eine
// innere Kontur läuft andersherum, und dieselbe Regel zieht ein Loch
// zusammen, während sie die äußere Kontur weitet.
func verdickt(umriss sfnt.Segments, um float64) sfnt.Segments {
	aus := append(sfnt.Segments(nil), umriss...)
	if um == 0 {
		return aus
	}

	// Im Glyphenraum von `sfnt` wächst y nach unten; eine äußere Kontur mit
	// positiver Fläche läuft dort im Uhrzeigersinn, und ihr Außen liegt links
	// der Laufrichtung.
	richtung := 1.0
	if umlauf(umriss) > 0 {
		richtung = -1
	}

	for anfang := 0; anfang < len(aus); {
		ende := anfang + 1
		for ende < len(aus) && aus[ende].Op != sfnt.SegmentOpMoveTo {
			ende++
		}

		verdickeKontur(aus[anfang:ende], um*richtung, um)
		anfang = ende
	}

	return aus
}

// punkt ist ein Punkt in Bildpunkten.
type punkt struct{ x, y float64 }

func punktAus(p fixed.Point26_6) punkt { return punkt{float64(p.X) / 64, float64(p.Y) / 64} }

// verdickeKontur verschiebt die Punkte einer Kontur an Ort und Stelle. Das
// Vorzeichen von `um` sagt, welche Seite außen ist; `hoch` hebt die ganze
// Kontur danach an.
func verdickeKontur(kontur sfnt.Segments, um, hoch float64) {
	// Die Punkte der Kontur in ihrer Reihenfolge, samt der Stelle, an der
	// jeder steht; der letzte fällt weg, wenn er den Kreis schließt.
	type stelle struct{ stueck, arg int }

	var stellen []stelle

	for i, stueck := range kontur {
		for arg := range argumente(stueck.Op) {
			stellen = append(stellen, stelle{i, arg})
		}
	}

	punkte := make([]punkt, len(stellen))
	for i, s := range stellen {
		punkte[i] = punktAus(kontur[s.stueck].Args[s.arg])
	}

	anzahl := len(punkte)
	if anzahl > 1 && punkte[anzahl-1] == punkte[0] {
		anzahl--
	}

	verschiebung := make([]punkt, anzahl)

	for i := range anzahl {
		verschiebung[i] = gehrung(punkte, anzahl, i, um)
	}

	for i, s := range stellen {
		v := verschiebung[i%max(anzahl, 1)]
		p := punkte[i]

		kontur[s.stueck].Args[s.arg] = fixed.Point26_6{
			X: fixed.Int26_6(math.Round((p.x + v.x) * 64)),
			Y: fixed.Int26_6(math.Round((p.y + v.y - hoch) * 64)),
		}
	}
}

// gehrung gibt, wohin ein Punkt wandert — [verdickt] sagt, nach welcher
// Regel.
//
// **Jede Kante rückt um `um·(|nₓ| + |n_y|)`**, nicht um `um`: Quartz verdickt
// achsparallel, wie eine Minkowski-Summe mit einem Quadrat. Gemessen an einem
// gedrehten Stamm, dessen Tinte bei 10, 20 und 30 Grad um das 1,17-, 1,29- und
// 1,38-fache dessen wächst, was sie ungedreht zulegt — `|cos| + |sin|` gibt
// 1,16, 1,28 und 1,37. Der Punkt liegt dann dort, wo sich seine zwei
// verschobenen Kanten schneiden.
func gehrung(punkte []punkt, anzahl, i int, um float64) punkt {
	vorher, lVorher, da := nachbar(punkte, anzahl, i, -1)
	nachher, lNachher, auch := nachbar(punkte, anzahl, i, +1)

	if !da || !auch {
		return punkt{}
	}

	// `d` ist eins plus der Kosinus der Wende; unter 1/16 dreht der Umriss um
	// mehr als rund 160 Grad (FreeType: `d > -0xF000L`), und die Gehrung ginge
	// ins Unendliche.
	d := 1 + vorher.x*nachher.x + vorher.y*nachher.y
	if d <= 1.0/16 {
		return punkt{}
	}

	// Die Normalen beider Kanten, nach außen, und wie weit jede rückt.
	seite := math.Copysign(1, um)
	nVorher := punkt{-vorher.y * seite, vorher.x * seite}
	nNachher := punkt{-nachher.y * seite, nachher.x * seite}
	oVorher := math.Abs(um) * (math.Abs(nVorher.x) + math.Abs(nVorher.y))
	oNachher := math.Abs(um) * (math.Abs(nNachher.x) + math.Abs(nNachher.y))

	// **Fast gleichlaufende Kanten schneiden sich nicht** — dort rückt der
	// Punkt mit ihnen, um das Mittel beider Wege.
	det := nVorher.x*nNachher.y - nVorher.y*nNachher.x
	if math.Abs(det) < 1e-9 {
		mittel := (oVorher + oNachher) / 2

		return punkt{nVorher.x * mittel, nVorher.y * mittel}
	}

	weg := punkt{
		(oVorher*nNachher.y - oNachher*nVorher.y) / det,
		(nVorher.x*oNachher - nNachher.x*oVorher) / det,
	}

	// **Kurze Kanten begrenzen die Verschiebung** (FreeType), sonst klappte
	// eine Kante um, die kürzer ist als die Gehrung lang.
	if laenge, kuerzer := math.Hypot(weg.x, weg.y), min(lVorher, lNachher); laenge > kuerzer {
		weg = punkt{weg.x * kuerzer / laenge, weg.y * kuerzer / laenge}
	}

	return weg
}

// nachbar gibt die Richtung zum nächsten verschiedenen Punkt davor oder
// danach — als Einheitsvektor in Laufrichtung — und den Abstand dorthin.
func nachbar(punkte []punkt, anzahl, i, schritt int) (punkt, float64, bool) {
	for k := 1; k < anzahl; k++ {
		j := ((i+schritt*k)%anzahl + anzahl) % anzahl
		dx, dy := punkte[j].x-punkte[i].x, punkte[j].y-punkte[i].y

		if laenge := math.Hypot(dx, dy); laenge > 0 {
			if schritt < 0 {
				dx, dy = -dx, -dy
			}

			return punkt{dx / laenge, dy / laenge}, laenge, true
		}
	}

	return punkt{}, 0, false
}

// argumente gibt, wie viele Punkte ein Stück trägt.
func argumente(art sfnt.SegmentOp) int {
	switch art {
	case sfnt.SegmentOpQuadTo:
		return 2
	case sfnt.SegmentOpCubeTo:
		return 3
	default:
		return 1
	}
}

// umlauf gibt die doppelte, vorzeichenbehaftete Fläche des Umrisses über
// seine Stützpolygone — ihr Vorzeichen ist die Umlaufrichtung der äußeren
// Konturen.
func umlauf(umriss sfnt.Segments) float64 {
	flaeche := 0.0

	var anfang, letzter punkt

	for _, stueck := range umriss {
		if stueck.Op == sfnt.SegmentOpMoveTo {
			flaeche += letzter.x*anfang.y - anfang.x*letzter.y
			anfang = punktAus(stueck.Args[0])
			letzter = anfang

			continue
		}

		for arg := range argumente(stueck.Op) {
			p := punktAus(stueck.Args[arg])
			flaeche += letzter.x*p.y - p.x*letzter.y
			letzter = p
		}
	}

	return flaeche + letzter.x*anfang.y - anfang.x*letzter.y
}
