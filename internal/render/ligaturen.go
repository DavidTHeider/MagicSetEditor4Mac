package render

import (
	"encoding/binary"
	"slices"
)

// **Ligaturen: was Core Text zu einer Glyphe zusammenzieht** —
// #1805.
//
// # Wofür
//
// MSE2 setzt unter macOS über wxWidgets und Core Text, und wxWidgets stellt
// dabei kein Merkmal um: `kCTLigatureAttributeName` kommt in seinem Quelltext
// nicht vor. Core Text setzt damit die **Standardligaturen** der Schrift —
// Beleren Bold und Beleren Small Caps führen je zwei, `ffi` und `ffl`, und im
// Referenzbild von `Armored Griffin` trägt das zweite `i` keinen eigenen
// Punkt. 91 von 25.102 Kartennamen des Kanons tragen eine der beiden Folgen,
// und ohne die Ligatur stand der Name bei 16 Punkt um 1,5 bis 2,4 Punkt zu
// breit (`docs/SCHRIFTUNTERSUCHUNG.md`).
//
// **Kein Verhalten von MSE2**, sondern eines von Core Text, das MSE2 unter
// macOS erbt — die Messlatte ist die Referenz, die MSE2 auf diesem Rechner
// zeichnet.
//
// # Was gelesen wird und was nicht
//
// **Das Merkmal `liga`, Lookup-Art 4, auch hinter einer Erweiterung (Art 7).**
// Welche Merkmale Core Text von selbst anwendet, ist gemessen, nicht
// angenommen: `liga` ja, `calt` nicht — `Ky`, `K)`, `n x` und `A&B` bekommen in
// Beleren keine Kontextalternative, obwohl ihr `calt` sie nennt
// (`scripts/coretext/glyphen.swift`, 2026-09-23).
//
// **Nicht gelesen werden `rlig`, `clig` und `ccmp`.** Im Pack führen sie nur
// Schriften, die sie für arabische Präsentationsformen oder für das
// Zusammensetzen kombinierender Marken brauchen — beides setzt dieser
// Rasterer ohnehin nicht ([0016](../../docs/entscheidungen/0016-keine-unicode-segmentierung.md)).
// Keine der acht Schriften des Kanons führt eines davon.
//
// **Und keine Lookup-Flags**: ein Lookup, der Marken überspringen soll, tut es
// hier nicht. Beide Beleren-Schnitte setzen das Flag auf null, nachgesehen am
// 2026-09-23. Wie beim Kerning wird kein Schriftsystem gewählt: gelesen wird
// jedes `liga`-Merkmal der Schrift.

// ligaturen sind die Lookups des `liga`-Merkmals einer Schrift, in der
// Reihenfolge der Lookup-Liste.
//
// **Ein Wert und kein Zeiger**, wie [kernung]: eine Schrift ohne Ligaturen
// trägt keine Lookups, und das kostet nichts.
type ligaturen struct {
	nachschlaege []ligaturnachschlag
}

// ligaturnachschlag ist ein Lookup: seine Untertabellen in ihrer Reihenfolge.
type ligaturnachschlag []ligatursaetze

// ligatursaetze sind die Ligaturen einer Untertabelle — je erster Glyphe in
// der Reihenfolge, in der die Schrift sie nennt.
type ligatursaetze map[uint16][]ligatur

// ligatur ist eine Ersetzung: die Glyphen nach der ersten und die Glyphe, die
// für alle zusammen steht.
type ligatur struct {
	folgende []uint16
	glyphe   uint16
}

// anwenden setzt eine Glyphenfolge, wie die Schrift es verlangt. Es gibt je
// Glyphe der Ausgabe, wie viele der Eingabe sie deckt.
//
// **Die Lookups nacheinander, jeder über die ganze Folge**, und ein späterer
// sieht, was ein früherer ersetzt hat. **Je Stelle gilt die erste
// Untertabelle, die ersetzt** — eine, die die Glyphe deckt und keine ihrer
// Ligaturen findet, lässt die nächste zu —, und **je Satz die erste Ligatur,
// die passt**: die Schrift ordnet sie nach Vorrang, nicht nach Länge. Dieselbe
// Regel wie HarfBuzz (`LigatureSet::apply`, `hb-ot-layout-gsub-table.hh`).
func (l ligaturen) anwenden(glyphen []uint16) (aus []uint16, deckt []int) {
	aus = slices.Clone(glyphen)
	deckt = make([]int, len(glyphen))

	for i := range deckt {
		deckt[i] = 1
	}

	for _, nachschlag := range l.nachschlaege {
		neu, neuDeckt := make([]uint16, 0, len(aus)), make([]int, 0, len(aus))

		for stelle := 0; stelle < len(aus); {
			glyphe, laenge := nachschlag.an(aus[stelle:])

			summe := 0
			for _, anzahl := range deckt[stelle : stelle+laenge] {
				summe += anzahl
			}

			neu, neuDeckt = append(neu, glyphe), append(neuDeckt, summe)
			stelle += laenge
		}

		aus, deckt = neu, neuDeckt
	}

	return aus, deckt
}

// an gibt die Glyphe, die an dieser Stelle steht, und wie viele Glyphen sie
// verbraucht — ohne Ligatur die erste, und eine.
func (n ligaturnachschlag) an(folge []uint16) (glyphe uint16, verbraucht int) {
	for _, saetze := range n {
		for _, kandidat := range saetze[folge[0]] {
			if len(kandidat.folgende) < len(folge) &&
				slices.Equal(kandidat.folgende, folge[1:1+len(kandidat.folgende)]) {
				return kandidat.glyphe, 1 + len(kandidat.folgende)
			}
		}
	}

	return folge[0], 1
}

// ligaturenAus liest die Ligaturen einer Schriftdatei — nur die, deren Glyphe
// es in der Schrift gibt: eine andere ließe sich nicht zeichnen.
func ligaturenAus(roh []byte, glyphenzahl int) ligaturen {
	gsub, da := tabellenverzeichnisAus(roh)["GSUB"]
	if !da || int(gsub)+10 > len(roh) {
		return ligaturen{}
	}

	merkmale := int(gsub) + int(zwei(roh, int(gsub)+6))
	liste := int(gsub) + int(zwei(roh, int(gsub)+8))

	var aus []ligaturnachschlag

	for _, nummer := range nachschlaegeDes(roh, merkmale, "liga") {
		if nachschlag := ligaturnachschlagAus(roh, liste, nummer, glyphenzahl); len(nachschlag) > 0 {
			aus = append(aus, nachschlag)
		}
	}

	return ligaturen{nachschlaege: aus}
}

// ligaturnachschlagAus liest die Untertabellen eines Lookups der Art 4 — auch
// hinter einer Erweiterung.
func ligaturnachschlagAus(roh []byte, liste int, nummer uint16, glyphenzahl int) ligaturnachschlag {
	lookup := liste + int(zwei(roh, liste+2+int(nummer)*2))
	if lookup+6 > len(roh) {
		return nil
	}

	art := zwei(roh, lookup)

	var aus ligaturnachschlag

	for i := range int(zwei(roh, lookup+4)) {
		tabelle := lookup + int(zwei(roh, lookup+6+i*2))

		// **Die Erweiterung trägt ihre Art und einen Versatz aus vier Byte** —
		// und nur die Art 4 ist hier eine Ligatur.
		if art == 7 {
			if tabelle+8 > len(roh) || zwei(roh, tabelle+2) != 4 {
				continue
			}

			tabelle += int(binary.BigEndian.Uint32(roh[tabelle+4:]))
		} else if art != 4 {
			return nil
		}

		if saetze := ligatursaetzeAus(roh, tabelle, glyphenzahl); len(saetze) > 0 {
			aus = append(aus, saetze)
		}
	}

	return aus
}

// ligatursaetzeAus liest eine `LigatureSubst`-Untertabelle in Format 1.
func ligatursaetzeAus(roh []byte, tabelle, glyphenzahl int) ligatursaetze {
	saetze := ligatursaetze{}

	if tabelle+6 > len(roh) || zwei(roh, tabelle) != 1 {
		return saetze
	}

	anzahl := int(zwei(roh, tabelle+4))

	for i, erste := range glyphenDerDeckung(roh, tabelle+int(zwei(roh, tabelle+2)), anzahl) {
		satz := tabelle + int(zwei(roh, tabelle+6+i*2))

		for j := range int(zwei(roh, satz)) {
			eintrag := satz + int(zwei(roh, satz+2+j*2))
			bestandteile := int(zwei(roh, eintrag+2))

			if bestandteile == 0 || eintrag+4+(bestandteile-1)*2 > len(roh) ||
				int(zwei(roh, eintrag)) >= glyphenzahl {
				continue
			}

			folgende := make([]uint16, bestandteile-1)
			for k := range folgende {
				folgende[k] = zwei(roh, eintrag+4+k*2)
			}

			saetze[erste] = append(saetze[erste], ligatur{folgende: folgende, glyphe: zwei(roh, eintrag)})
		}
	}

	return saetze
}
