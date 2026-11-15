package domain

// **Wie ein Objekt für einen Sitz aussieht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Aus `partiesicht.go` herausgelöst am 2026-09-17**, als sie über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht die
// Projektion der **Partie**, hier die eines einzelnen **Objekts** — und die
// ist der Ort, an dem die Zusage „die Bibliothek ist auch für ihren Eigner
// verdeckt" wirklich eingelöst wird.

// fuerSicht gibt dieses Objekt, so weit `fuer` es sehen darf.
func (o Spielobjekt) fuerSicht(fuer string) Sichtobjekt {
	gezeigt := o.offen()

	// **Ein verdeckt liegendes Objekt ist auch in einer offenen Zone
	// geheim** — für jeden außer seinem Beherrscher. Die Zonentabelle allein
	// sähe eine verdeckte Kreatur auf dem Tisch als offen an, und dann stünde
	// da, was sie ist.
	seines := fuer != "" && o.Beherrscher == fuer

	if o.Verdeckt && !seines {
		return o.geheim(gezeigt)
	}

	switch o.Zone.Sicht() {
	case SichtOffen:
		return gezeigt

	case SichtBeherrscher:
		if seines {
			return gezeigt
		}
	case SichtVerdeckt:
	}

	return o.geheim(gezeigt)
}

// geheim nimmt aus einem Sichtobjekt, was der Empfänger nicht sehen darf.
//
// **Das Objekt bleibt, sein Inhalt geht.** Wer es ganz wegließe, ließe eine
// leere Hand entstehen — und niemand wüsste mehr, dass das Gegenüber
// überhaupt Karten hält.
func (o Spielobjekt) geheim(gezeigt Sichtobjekt) Sichtobjekt {
	gezeigt.Karte = ""
	gezeigt.Name = ""
	gezeigt.Geheim = true

	// **Und die Ordnung geht mit** — sie ist das eigentliche Geheimnis einer
	// Bibliothek: wer sie kennt, kennt den nächsten Zug, ohne eine einzige
	// Karte zu lesen.
	gezeigt.Ordnung = 0

	return gezeigt
}

// offen gibt dieses Objekt ungekürzt.
func (o Spielobjekt) offen() Sichtobjekt {
	return Sichtobjekt{
		Kennung:     o.Kennung,
		Art:         o.Art,
		Beherrscher: o.Beherrscher,
		Eigner:      o.Eigner,
		Zone:        o.Zone,
		Karte:       o.Karte,
		Name:        o.Name,
		Ordnung:     o.Ordnung,
		Getappt:     o.Getappt,
		Verdeckt:    o.Verdeckt,
		HaengtAn:    o.HaengtAn,
		Haufen:      o.Haufen,
		Zaehler:     o.Zaehler,
	}
}

// Sichteigenschaften sind die Eigenschaften einer Karte, wie der Tisch sie
// zeigt — CR 109.3 (M2, #1604).
//
// # Warum die **gelebten** Werte und nicht die gedruckten
//
// **Weil am Tisch die gelebten stehen.** Eine 2/3 mit zwei +1/+1-Marken ist
// 4/5 (CR 122.1a), und genau das ist die zweite Abnahme von M2. Was gedruckt
// war, steht weiter im Abzug; wer es braucht, liest es dort.
//
// # Und warum die Farbe hier ausgerechnet dasteht
//
// **Weil sie den Manakosten folgt** (CR 202.2) und nicht in einem Feld steht.
// Sie hier abzulegen heißt, dass es **eine** Herleitung gibt statt einer je
// Fläche.
type Sichteigenschaften struct {
	// Name ist der Name der Karte — CR 201.1.
	//
	// **Er steht hier und nicht an [Sichtobjekt.Name]**, und der Grund ist
	// eine Falle: [Spielobjekt.Name] trägt den Namen dessen, wo **keine**
	// Karte dahintersteht (ein Spielstein). Bei einer Karte ist er leer, und
	// am Tisch stand deshalb die Objektkennung statt „Goblin Guide" —
	// gesehen bei der ersten Sichtprüfung mit einem echten Deck am
	// 2026-09-17.
	Name string

	// Farben sind die Farben dieses Objekts — leer heißt **farblos**
	// (CR 105.2c), in der Reihenfolge des Regelwerks.
	Farben []Farbe

	// Manawert ist die Summe des Manas im Kosten — CR 202.3.
	Manawert int

	Supertypen []Supertyp
	Typen      []Kartentyp
	Untertypen []string

	// Staerke und Widerstandskraft **mit Marken** — CR 122.1a.
	//
	// `Steht` unterscheidet „keine" von „null" (CR 208.3): eine Nichtkreatur
	// hat keine Stärke.
	Staerke          Kartenwert
	Widerstandskraft Kartenwert

	Loyalitaet   Kartenwert
	Verteidigung Kartenwert

	// UnbekannteTypen sind die Wörter der Typzeile, die CR 205.2a und
	// CR 205.4a nicht kennen — §8: ein Befund ist kein Fehler.
	//
	// **Die Wörter und nicht die ganzen Befunde**: am Tisch hilft „was
	// versteht das Programm nicht", und Datei und Ort eines [Befund] sind
	// hier dieselbe Typzeile, an der es ohnehin steht.
	UnbekannteTypen []string
}

// sichteigenschaften baut die Sicht auf die Eigenschaften eines Objekts.
//
// **nil, wenn es keine gibt** — ein Spielstein hat keine Karte, und ein
// Verweis ins Leere ebenso wenig (§8).
func (p Partie) sichteigenschaften(objekt Spielobjekt) *Sichteigenschaften {
	eigen, da := p.Eigenschaften(objekt)
	if !da {
		return nil
	}

	staerke, widerstand := p.Werte(objekt)

	unbekannt := make([]string, 0, len(eigen.Befunde))
	for _, befund := range eigen.Befunde {
		unbekannt = append(unbekannt, befund.Ort)
	}

	return &Sichteigenschaften{
		Name:             eigen.Name,
		Farben:           eigen.Manakosten.Farben(),
		Manawert:         eigen.Manakosten.Manawert(),
		Supertypen:       eigen.Typzeile.Supertypen,
		Typen:            eigen.Typzeile.Typen,
		Untertypen:       eigen.Typzeile.Untertypen,
		Staerke:          staerke,
		Widerstandskraft: widerstand,
		Loyalitaet:       eigen.Loyalitaet,
		Verteidigung:     eigen.Verteidigung,
		UnbekannteTypen:  unbekannt,
	}
}
