package render

import (
	"encoding/binary"
	"slices"
)

// **Kerning: was zwischen zwei Glyphen dazukommt** —
// #752.
//
// # Wofür
//
// MSE misst eine Vorsilbe **als Ganzes**:
//
//	dc.GetTextExtent(text, &w, &h);
//
// `RotatedDC::GetTextExtent`, `src/util/rotation.cpp`, Zeile 339, Fassung
// f9d9356 vom 2021-07-06. wx wendet dabei die Paartabelle der Schrift an, und
// es zeichnet ganze Zeichenketten — Messung und Satz kernen also beide.
//
// **Gemessen, was das kostet**: `beleren-bold` — die Schrift **jedes**
// Kartennamens — führt 1.983 Einzelpaare mit Vorschubkorrektur und dahinter
// eine Klassentabelle. `Force of Will` kommt auf −606 Einheiten von 2048 je
// Geviert, −426 davon aus den Einzelpaaren; bei 16 Punkt sind das 4,7 Punkt.
//
// **Und dass MSE sie anwendet, ist am Bild belegt**: acht Karten aus `White`
// mit und ohne Kerning gegen MSE2s Ausgabe gehalten — wo gekernt wird, halbiert
// sich der Abstand (33,6 gegen 73,3 im Feld bei `2 Mace`).
//
// **Der Regeltext war nie betroffen**: `mplantin.ttf` trägt gar keine
// Kerningtabelle, und deshalb sah die Messung vom 2026-08-20 sauber aus.
//
// # Was gelesen wird und was nicht
//
// **`PairPos` in beiden Formaten** — Paarlisten je Glyphe (Format 1) und
// Klassenpaare (Format 2, `kernklassen.go`) — und die alte **`kern`-Tabelle**
// (Format 0). Beleren, Beleren Small Caps und `relay-medium` benutzen `GPOS`,
// `matrixb.ttf` die alte Tabelle.
//
// **Nicht gelesen werden Erweiterungen** (Lookup-Art 9): keine Schrift des
// Packs benutzt sie für ihr Kerning, nachgesehen am 2026-09-23 an allen sechs,
// die `GPOS` tragen. **Und kein Schriftsystem wird gewählt**: gelesen wird
// jedes `kern`-Merkmal der Schrift. In allen sechs verweisen die
// Schriftsysteme auf **ein** Merkmal; eine Schrift mit verschiedenen
// `kern`-Lookups je Schriftsystem bekäme deren Vereinigung.

// kernung beantwortet, was zwischen zwei Glyphen dazukommt.
//
// **Ein Wert und kein Zeiger**: eine Schrift, die nicht kernt, trägt keine
// Lookups, und das kostet nichts.
type kernung struct {
	// nachschlaege sind die Lookups des `kern`-Merkmals — oder die
	// Untertabellen der alten `kern`-Tabelle — in ihrer Reihenfolge.
	nachschlaege []kernnachschlag
}

// kernnachschlag ist ein Lookup: seine Untertabellen in ihrer Reihenfolge.
type kernnachschlag []kernuntertabelle

// kernuntertabelle ist eine Untertabelle eines Lookups — Einzelpaare oder
// Klassen. Sie sagt, ob sie für ein Paar gilt, und mit welcher Verschiebung
// in Schrifteinheiten.
type kernuntertabelle interface {
	wert(erste, zweite uint16) (int16, bool)
}

// einzelpaare sind die Paare einer Untertabelle in Format 1 oder der alten
// `kern`-Tabelle — **samt denen mit null**: auch sie gelten, und in Format 1
// sperren sie die Klasse dahinter.
type einzelpaare map[[2]uint16]int16

func (p einzelpaare) wert(erste, zweite uint16) (int16, bool) {
	wert, da := p[[2]uint16{erste, zweite}]

	return wert, da
}

// fuer gibt die Verschiebung zwischen zwei Glyphen — 0, wenn es keine gibt.
//
// **Je Lookup gilt die erste Untertabelle, die das Paar kennt, und die
// Lookups addieren sich** — dieselbe Regel wie HarfBuzz und Core Text; ohne
// sie bekäme ein Paar, das Beleren ausdrücklich nicht kernt, den Wert seiner
// Klasse. Gezählt wird in `int`, damit viele Lookups nicht überlaufen.
func (k kernung) fuer(erste, zweite uint16) int {
	summe := 0

	for _, nachschlag := range k.nachschlaege {
		for _, tabelle := range nachschlag {
			if wert, gilt := tabelle.wert(erste, zweite); gilt {
				summe += int(wert)

				break
			}
		}
	}

	return summe
}

// kernungAus liest das Kerning einer Schriftdatei.
//
// **Die alte `kern`-Tabelle nur ohne lesbares `GPOS`-Kerning.** Eine Schrift,
// die beides trägt, meint dasselbe zweimal — `vanilla whale.ttf` des Packs und
// vierzehn Systemschriften dieses Rechners führen in beiden dieselben Werte,
// und Core Text zählt sie einmal (nachgemessen am 2026-09-23). Welche Tabelle
// gilt, wenn sie verschieden sind, zeigt keine davon; festgelegt ist es nach
// HarfBuzz, das die alte Tabelle nur nimmt, wenn `GPOS` kein `kern`-Merkmal
// hat (`hb_ot_shape_planner_t::compile`, `src/hb-ot-shape.cc`, nachgelesen am
// 2026-09-23).
func kernungAus(roh []byte) kernung {
	tabellen := tabellenverzeichnisAus(roh)

	var nachschlaege []kernnachschlag

	if versatz, da := tabellen["GPOS"]; da {
		nachschlaege = ausGPOS(roh, versatz)
	}

	if versatz, da := tabellen["kern"]; da && len(nachschlaege) == 0 {
		nachschlaege = ausKerntabelle(roh, versatz)
	}

	return kernung{nachschlaege: nachschlaege}
}

// tabellenverzeichnisAus gibt je Kennung den Versatz der Tabelle.
func tabellenverzeichnisAus(roh []byte) map[string]uint32 {
	aus := map[string]uint32{}
	if len(roh) < 12 {
		return aus
	}

	for i := range int(zwei(roh, 4)) {
		eintrag := 12 + i*16
		if eintrag+16 > len(roh) {
			return aus
		}

		aus[string(roh[eintrag:eintrag+4])] = vier(roh, eintrag+8)
	}

	return aus
}

// ausGPOS liest die Lookups des `kern`-Merkmals.
func ausGPOS(roh []byte, gpos uint32) []kernnachschlag {
	if int(gpos)+10 > len(roh) {
		return nil
	}

	merkmale := int(gpos) + int(zwei(roh, int(gpos)+6))
	nachschlaege := int(gpos) + int(zwei(roh, int(gpos)+8))

	var aus []kernnachschlag

	for _, nummer := range nachschlaegeDes(roh, merkmale, "kern") {
		if nachschlag := untertabellen(roh, nachschlaege, nummer); len(nachschlag) > 0 {
			aus = append(aus, nachschlag)
		}
	}

	return aus
}

// nachschlaegeDes gibt die Nummern der Lookups eines Merkmals — `kern` hier,
// `liga` für die Ligaturen — **aufsteigend und jede einmal**.
//
// Mehrere Merkmalseinträge dürfen auf denselben Lookup zeigen, einer je
// Schriftsystem. Solange sich eine Karte merkte, was die Paare sagen, schadete
// das nicht; seit sich die Lookups addieren, kernte ein doppelter Eintrag
// doppelt. Angewandt werden sie in der Reihenfolge der Lookup-Liste.
func nachschlaegeDes(roh []byte, merkmale int, kennung string) []uint16 {
	var aus []uint16

	if merkmale+2 > len(roh) {
		return aus
	}

	for i := range int(zwei(roh, merkmale)) {
		eintrag := merkmale + 2 + i*6
		if eintrag+6 > len(roh) || string(roh[eintrag:eintrag+4]) != kennung {
			continue
		}

		merkmal := merkmale + int(zwei(roh, eintrag+4))
		if merkmal+4 > len(roh) {
			continue
		}

		for j := range int(zwei(roh, merkmal+2)) {
			stelle := merkmal + 4 + j*2
			if stelle+2 <= len(roh) {
				aus = append(aus, zwei(roh, stelle))
			}
		}
	}

	slices.Sort(aus)

	return slices.Compact(aus)
}

// untertabellen liest die `PairPos`-Untertabellen eines Lookups in ihrer
// Reihenfolge.
func untertabellen(roh []byte, nachschlaege int, nummer uint16) kernnachschlag {
	stelle := nachschlaege + 2 + int(nummer)*2
	if stelle+2 > len(roh) {
		return nil
	}

	lookup := nachschlaege + int(zwei(roh, stelle))
	if lookup+6 > len(roh) || zwei(roh, lookup) != 2 {
		return nil
	}

	var aus kernnachschlag

	for i := range int(zwei(roh, lookup+4)) {
		versatz := lookup + 6 + i*2
		if versatz+2 > len(roh) {
			return aus
		}

		tabelle := lookup + int(zwei(roh, versatz))

		switch zwei(roh, tabelle) {
		case 1:
			aus = append(aus, paarpos(roh, tabelle))

		case 2:
			if klassen, gelesen := klassenpaareAus(roh, tabelle); gelesen {
				aus = append(aus, klassen)
			}
		}
	}

	return aus
}

// paarpos liest eine `PairPos`-Untertabelle in Format 1.
//
// **Ein Paar ohne Vorschub im Wertformat gilt trotzdem** — mit null. Es
// verschiebt die Grundlinie oder nichts, aber es ist das Paar, das die Schrift
// für diese zwei Glyphen nennt, und keine Untertabelle dahinter kommt zum Zug.
func paarpos(roh []byte, tabelle int) einzelpaare {
	paare := einzelpaare{}

	if tabelle+10 > len(roh) {
		return paare
	}

	format1, format2 := zwei(roh, tabelle+4), zwei(roh, tabelle+6)
	vorschub := format1&0x0004 != 0

	// **Die Deckung nur so weit, wie es Paarlisten gibt** — eine längere
	// zeigte mit ihren übrigen Glyphen auf Versätze, die niemand geschrieben
	// hat.
	deckung := glyphenDerDeckung(roh, tabelle+int(zwei(roh, tabelle+2)), int(zwei(roh, tabelle+8)))
	schritt := 2 + felder(format1) + felder(format2)
	vorlauf := felder(format1 & 0x0003)

	for i, erste := range deckung {
		// **Ohne eigene Grenzprüfung**: [zwei] gibt außerhalb null, und ein
		// Versatz von null zeigt auf den Tabellenkopf — dessen erste zwei Byte
		// sind das Format, also 1, und die eine Paarzeile dahinter fällt unter
		// die Prüfung darunter. Eine zweite Prüfung hier wäre ein Zweig, den
		// keine Eingabe betritt (§3).
		liste := tabelle + int(zwei(roh, tabelle+10+i*2))

		for j := range int(zwei(roh, liste)) {
			eintrag := liste + 2 + j*schritt
			if eintrag+2+vorlauf+2 > len(roh) {
				break
			}

			var wert int16
			if vorschub {
				wert = int16(zwei(roh, eintrag+2+vorlauf)) /* #nosec G115 -- Schriftdaten sind vorzeichenbehaftet */
			}

			paare[[2]uint16{erste, zwei(roh, eintrag)}] = wert
		}
	}

	return paare
}

// glyphenraum ist, wie viele Glyphennummern es gibt.
const glyphenraum = 1 << 16

// glyphenDerDeckung liest eine `Coverage`-Tabelle, Format 1 und 2 — höchstens
// so viele Glyphen, wie der Aufrufer braucht.
//
// **Gezählt wird in `int`.** Mit einer 16-Bit-Zahl sprang die Schleife über
// einen Bereich, der bei `0xFFFF` endet, auf null zurück und hörte nie auf —
// eine beschädigte Schriftdatei hing den Rasterer auf. Und die Obergrenze hält
// eine Datei, die denselben Bereich tausendmal nennt, davon ab, die Zeit zu
// bestimmen.
func glyphenDerDeckung(roh []byte, tabelle, hoechstens int) []uint16 {
	var aus []uint16

	if tabelle+4 > len(roh) {
		return aus
	}

	anzahl := int(zwei(roh, tabelle+2))

	switch zwei(roh, tabelle) {
	case 1:
		for i := range min(anzahl, hoechstens) {
			stelle := tabelle + 4 + i*2
			if stelle+2 > len(roh) {
				return aus
			}

			aus = append(aus, zwei(roh, stelle))
		}

	case 2:
		for i := range anzahl {
			stelle := tabelle + 4 + i*6
			if stelle+6 > len(roh) {
				return aus
			}

			for glyphe := int(zwei(roh, stelle)); glyphe <= int(zwei(roh, stelle+2)); glyphe++ {
				if len(aus) >= hoechstens {
					return aus
				}

				aus = append(aus, uint16(glyphe)) /* #nosec G115 -- aus zwei Byte gelesen */
			}
		}
	}

	return aus
}

// felder gibt, wie viele Byte ein `ValueRecord` mit diesem Format belegt.
func felder(format uint16) int {
	anzahl := 0

	for bit := range 16 {
		if format&(1<<bit) != 0 {
			anzahl++
		}
	}

	return anzahl * 2
}

func zwei(roh []byte, stelle int) uint16 {
	if stelle+2 > len(roh) {
		return 0
	}

	return binary.BigEndian.Uint16(roh[stelle:])
}

// vier liest vier Byte.
//
// **Ohne Grenzprüfung, anders als [zwei]**: der einzige Aufrufer
// ([tabellenverzeichnisAus]) prüft den ganzen Eintrag von sechzehn Byte, bevor
// er hierher kommt. Eine zweite Prüfung wäre ein Zweig, den kein Aufruf
// betreten kann — und damit eine Behauptung darüber, was im Ernstfall geschieht
// (§3).
func vier(roh []byte, stelle int) uint32 {
	return binary.BigEndian.Uint32(roh[stelle:])
}
