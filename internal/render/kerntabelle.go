package render

// **Die alte `kern`-Tabelle** — der Weg vor `GPOS`, und für `matrixb.ttf` des
// Packs der einzige (#752).
// Wann sie gilt, entscheidet [kernungAus].

// ausKerntabelle liest die alte `kern`-Tabelle, Untertabellenformat 0 — **je
// Untertabelle ein Lookup**: ihre Werte addieren sich.
//
// **`matrixb.ttf` hat kein GPOS** und trägt seine Paare nur hier.
func ausKerntabelle(roh []byte, kern uint32) []kernnachschlag {
	var aus []kernnachschlag

	if int(kern)+4 > len(roh) {
		return aus
	}

	stelle := int(kern) + 4

	for range int(zwei(roh, int(kern)+2)) {
		if stelle+6 > len(roh) {
			return aus
		}

		laenge := int(zwei(roh, stelle+2))
		deckung := zwei(roh, stelle+4)

		// **Nur waagerechtes Kerning ohne Nachbearbeitung** — Bit 0 gesetzt,
		// Bit 1 (senkrecht) und Bit 2 (kreuzend) nicht.
		if deckung&0x0001 != 0 && deckung&0x0006 == 0 && deckung>>8 == 0 {
			aus = append(aus, kernnachschlag{paarliste(roh, stelle+6)})
		}

		if laenge == 0 {
			return aus
		}

		stelle += laenge
	}

	return aus
}

// paarliste liest die Paare einer `kern`-Untertabelle im Format 0.
func paarliste(roh []byte, kopf int) einzelpaare {
	paare := einzelpaare{}

	if kopf+8 > len(roh) {
		return paare
	}

	for i := range int(zwei(roh, kopf)) {
		eintrag := kopf + 8 + i*6
		if eintrag+6 > len(roh) {
			return paare
		}

		paare[[2]uint16{zwei(roh, eintrag), zwei(roh, eintrag+2)}] = int16(zwei(roh, eintrag+4)) /* #nosec G115 -- Schriftdaten sind vorzeichenbehaftet */
	}

	return paare
}
