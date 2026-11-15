package app

// **Was ein Gast von einer Hand sieht** — #2480: eine Karte, die verborgen
// wird, ist für ihn ein neues Objekt (Vermerk 0201), und gezählt wird, was
// in der Zone liegt, statt nach einem Namen zu suchen.

// objektIn sucht ein Objekt in der Sicht eines Gastes — unter dem Namen, den
// er kennt.
func objektIn(sicht *PartieDTO, kennung string) (ObjektDTO, bool) {
	if sicht == nil {
		return ObjektDTO{}, false
	}

	for _, eines := range sicht.Objekte {
		if eines.Kennung == kennung {
			return eines, true
		}
	}

	return ObjektDTO{}, false
}

// verborgenIn zählt die Objekte eines Sitzes in einer Zone — und ob eines
// davon offen liegt.
func verborgenIn(sicht *PartieDTO, wer, zone string) (anzahl int, offen bool) {
	if sicht == nil {
		return 0, false
	}

	for _, eines := range sicht.Objekte {
		if eines.Eigner == wer && eines.Zone == zone {
			anzahl++
			offen = offen || !eines.Geheim || eines.Karte != ""
		}
	}

	return anzahl, offen
}
