package render

import "bytes"

// Bauteile für die Kerningtests — Schriftdateien aus einzelnen Tabellen und
// der Blick auf die Einzelpaare einer [kernung].

// schriftMitNachschlaegen baut ein GPOS mit einem `kern`-Merkmal, das auf
// jeden der Lookups verweist; je Lookup die Untertabellen in ihrer
// Reihenfolge.
func schriftMitNachschlaegen(nachschlaege [][][]byte) []byte {
	merkmale := make([]byte, 12+2*len(nachschlaege))
	setze(merkmale, 0, 1)
	copy(merkmale[2:], "kern")
	setze(merkmale, 6, 8)
	setze(merkmale, 10, len(nachschlaege))

	for i := range nachschlaege {
		setze(merkmale, 12+2*i, i)
	}

	liste := make([]byte, 2+2*len(nachschlaege))
	setze(liste, 0, len(nachschlaege))

	var koerper []byte

	for i, untertabellen := range nachschlaege {
		setze(liste, 2+2*i, len(liste)+len(koerper))

		lookup := make([]byte, 6+2*len(untertabellen))
		setze(lookup, 0, 2)
		setze(lookup, 4, len(untertabellen))

		var daten []byte

		for j, tabelle := range untertabellen {
			setze(lookup, 6+2*j, len(lookup)+len(daten))
			daten = append(daten, tabelle...)
		}

		koerper = append(koerper, lookup...)
		koerper = append(koerper, daten...)
	}

	kopf := make([]byte, 10)
	setze(kopf, 0, 1)
	setze(kopf, 6, len(kopf))
	setze(kopf, 8, len(kopf)+len(merkmale))

	return mitTabelle("GPOS", bytes.Join([][]byte{kopf, merkmale, liste, koerper}, nil))
}

// mitTabelle legt eine Tabelle als einzige in eine Schriftdatei.
func mitTabelle(kennung string, tabelle []byte) []byte {
	verzeichnis := make([]byte, 12+16, 12+16+len(tabelle))
	verzeichnis[1] = 1
	setze(verzeichnis, 4, 1)
	copy(verzeichnis[12:], kennung)
	setze(verzeichnis, 22, len(verzeichnis))
	setze(verzeichnis, 26, len(tabelle))

	return append(verzeichnis, tabelle...)
}

// mitZweiterTabelle legt die einzige Tabelle einer zweiten Schriftdatei zu der
// einer ersten — zwei Einträge im Verzeichnis, die Tabellen dahinter.
func mitZweiterTabelle(erste, zweite []byte) []byte {
	const eintrag = 12

	ersteTabelle, zweiteTabelle := erste[eintrag+16:], zweite[eintrag+16:]

	verzeichnis := make([]byte, 12+32, 12+32+len(ersteTabelle)+len(zweiteTabelle))
	verzeichnis[1] = 1
	setze(verzeichnis, 4, 2)

	copy(verzeichnis[12:], erste[eintrag:eintrag+4])
	setze(verzeichnis, 22, len(verzeichnis))
	setze(verzeichnis, 26, len(ersteTabelle))

	copy(verzeichnis[28:], zweite[eintrag:eintrag+4])
	setze(verzeichnis, 38, len(verzeichnis)+len(ersteTabelle))
	setze(verzeichnis, 42, len(zweiteTabelle))

	return bytes.Join([][]byte{verzeichnis, ersteTabelle, zweiteTabelle}, nil)
}

// paareVon gibt die Einzelpaare einer Kernung, die etwas verschieben — je Paar
// der Wert, der gilt: die erste Untertabelle eines Lookups, die es kennt.
func paareVon(kern kernung) map[[2]uint16]int16 {
	aus := map[[2]uint16]int16{}

	for _, nachschlag := range kern.nachschlaege {
		for _, tabelle := range nachschlag {
			paare, einzeln := tabelle.(einzelpaare)
			if !einzeln {
				continue
			}

			for paar := range paare {
				if wert := kern.fuer(paar[0], paar[1]); wert != 0 {
					aus[paar] = int16(wert) //nolint:gosec // Testdaten
				}
			}
		}
	}

	return aus
}
