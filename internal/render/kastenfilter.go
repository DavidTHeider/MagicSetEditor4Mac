package render

import "image"

// kastenverschiebung ist `shift` aus MSEs Umrechnung: Festkomma mit 14
// Nachkommabits (`const int shift = 32-10-8`, `src/gfx/resample_image.cpp`,
// Zeile 18, Fassung f9d9356 vom 2021-07-06).
const kastenverschiebung = 32 - 10 - 8

// kanalbild ist der Rot- und der Deckungskanal eines Bilds, roh — so, wie
// wxWidgets sie hält (`GetData`, `GetAlpha`).
type kanalbild struct {
	breite, hoehe int
	rot, deckung  []uint32

	// mitDeckung sagt, ob das Bild Deckung trägt. Ohne sie rechnet MSE den
	// Rotkanal allein; mit ihr gewichtet es ihn mit der Deckung.
	mitDeckung bool
}

// kanaeleAus liest ein Bild in Kanäle.
func kanaeleAus(bild image.Image) kanalbild {
	feld := bild.Bounds()
	kanaele := kanalbild{
		breite: feld.Dx(), hoehe: feld.Dy(),
		rot:     make([]uint32, feld.Dx()*feld.Dy()),
		deckung: make([]uint32, feld.Dx()*feld.Dy()),
	}

	for y := range kanaele.hoehe {
		for x := range kanaele.breite {
			farbe := nrgbaAn(bild, feld.Min.X+x, feld.Min.Y+y)
			stelle := y*kanaele.breite + x
			kanaele.rot[stelle], kanaele.deckung[stelle] = uint32(farbe.R), uint32(farbe.A)
			kanaele.mitDeckung = kanaele.mitDeckung || farbe.A != 255
		}
	}

	return kanaele
}

// umgerechnet bringt die Kanäle auf breite mal hoehe — MSEs Kastenfilter:
// erst waagerecht, dann senkrecht, jeder Punkt der Quelle gibt einen festen
// Anteil an das Ziel ab, und der erste dazu den Rest, damit die Summe stimmt
// (`resample`, `resample_and_clip`, `src/gfx/resample_image.cpp`, Zeilen
// 102–153, Fassung f9d9356 vom 2021-07-06).
func (k kanalbild) umgerechnet(breite, hoehe int) kanalbild {
	if k.breite == breite && k.hoehe == hoehe {
		return k
	}

	waagerecht := k.leer(breite, k.hoehe)
	k.durchgang(waagerecht, k.breite, 1, breite, 1, k.hoehe, k.breite, breite)

	if k.hoehe == hoehe {
		return waagerecht
	}

	senkrecht := k.leer(breite, hoehe)
	waagerecht.durchgang(senkrecht, k.hoehe, breite, hoehe, breite, breite, 1, 1)

	return senkrecht
}

// leer gibt ein Ziel derselben Art.
func (k kanalbild) leer(breite, hoehe int) kanalbild {
	return kanalbild{
		breite: breite, hoehe: hoehe, mitDeckung: k.mitDeckung,
		rot: make([]uint32, breite*hoehe), deckung: make([]uint32, breite*hoehe),
	}
}

// durchgang rechnet in einer Richtung um — `resample_pass`, ebd. Zeilen
// 30–100: je Zeile wird so lange Quelle „gegessen", bis ein Zielpunkt voll
// ist. Mit Deckung wird der Rotkanal mit ihr gewichtet und durch ihre Summe
// geteilt; ohne Deckung um die Nachkommabits geschoben.
func (k kanalbild) durchgang(
	ziel kanalbild, laengeEin, schrittEin, laengeAus, schrittAus, zeilen,
	zeilenschrittEin, zeilenschrittAus int,
) {
	// #nosec G115 -- eine Länge ist höchstens 1 << 14 ([groessteMaske]), verschoben also unter 1 << 28
	anteil := uint32((laengeAus << kastenverschiebung) / laengeEin) //nolint:gosec // dito
	// #nosec G115 -- dito
	rest := uint32((laengeAus << kastenverschiebung) % laengeEin) //nolint:gosec // dito

	for zeile := range zeilen {
		ein, aus := zeile*zeilenschrittEin, zeile*zeilenschrittAus
		einRest := anteil + rest

		for range laengeAus {
			ausRest := uint32(1) << kastenverschiebung

			var summeRot, summeDeckung uint32

			essen := func(menge uint32) {
				deckung := uint32(1)
				if k.mitDeckung {
					deckung = k.deckung[ein]
				}

				summeRot += k.rot[ein] * menge * deckung
				summeDeckung += deckung * menge
			}

			for ausRest >= einRest {
				essen(einRest)
				ausRest -= einRest
				einRest = anteil
				ein += schrittEin
			}

			if ausRest > 0 {
				essen(ausRest)
				einRest -= ausRest
			}

			switch {
			case !k.mitDeckung:
				ziel.rot[aus] = summeRot >> kastenverschiebung
			case summeDeckung > 0:
				ziel.rot[aus] = summeRot / summeDeckung
				ziel.deckung[aus] = summeDeckung >> kastenverschiebung
			}

			aus += schrittAus
		}
	}
}
