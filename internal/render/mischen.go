package render

import "math"

// mischen legt eine Farbe über eine andere — „source over".
//
// # Warum nicht premultipliziert
//
// Go führt `image.RGBA` **mit vorgewichtetem Alpha**. Das ist zum Zeichnen
// schneller und beim Rechnen die häufigste Fehlerquelle: eine halbdurchsichtige
// Farbe, die jemand ohne Vorgewicht hineinschreibt, wird beim nächsten Schritt
// ein zweites Mal gewichtet und ist zu dunkel. Auf einer Karte sieht das aus
// wie „der Rahmen ist etwas kräftiger als in MSE" — und niemand sucht dafür
// nach einer Multiplikation.
//
// Hier wird deshalb mit `image.NRGBA` gerechnet: Alpha steht daneben, nicht in
// den Farbkanälen.
//
// Die Formel, mit `a` in 0…1:
//
//	aus   = aOben + aUnten·(1 − aOben)
//	farbe = (fOben·aOben + fUnten·aUnten·(1 − aOben)) / aus
func mischen(unten, oben Farbe) Farbe {
	obenA := float64(oben.A) / 255
	untenA := float64(unten.A) / 255

	ausA := obenA + untenA*(1-obenA)
	if ausA <= 0 {
		return Farbe{}
	}

	kanal := func(fOben, fUnten uint8) uint8 {
		wert := (float64(fOben)*obenA + float64(fUnten)*untenA*(1-obenA)) / ausA

		return alsAnteil(wert)
	}

	return Farbe{
		R: kanal(oben.R, unten.R),
		G: kanal(oben.G, unten.G),
		B: kanal(oben.B, unten.B),
		A: alsAnteil(ausA * 255),
	}
}

// gedaempft nimmt einer Farbe Deckung — für `Ebene.Deckung` und für die Maske.
func gedaempft(farbe Farbe, anteil float64) Farbe {
	farbe.A = alsAnteil(float64(farbe.A) * begrenzt01(anteil))

	return farbe
}

// alsAnteil rundet auf das nächste Byte und begrenzt auf 0…255.
//
// **Gerundet, nicht abgeschnitten.** Ein Abschneiden verliert bei jeder
// Mischung bis zu einem halben Wert; über zwölf Ebenen sind das bis zu sechs
// Stufen Helligkeit, und die sieht man im Vergleich zweier Karten.
func alsAnteil(wert float64) uint8 {
	switch {
	case wert <= 0 || math.IsNaN(wert):
		return 0

	case wert >= 255:
		return 255
	}

	return uint8(math.Round(wert))
}

func begrenzt01(wert float64) float64 {
	switch {
	case wert <= 0 || math.IsNaN(wert):
		return 0

	case wert >= 1:
		return 1
	}

	return wert
}

// farbeAn gibt die Farbe eines Verlaufs an einer Stelle zwischen 0 und 1.
//
// Zwischen zwei Stufen wird linear gemischt — **in den Farbkanälen, nicht über
// die Mischformel oben.** Ein Verlauf ist kein Übereinanderlegen: bei halbem
// Weg zwischen zwei undurchsichtigen Farben soll ihre Mitte herauskommen, nicht
// die obere.
func farbeAn(verlauf *Verlauf, stelle float64) Farbe {
	stufen := verlauf.Stufen
	stelle = begrenzt01(stelle)

	if stelle <= stufen[0].Stelle {
		return stufen[0].Farbe
	}

	for i := 1; i < len(stufen); i++ {
		if stelle > stufen[i].Stelle {
			continue
		}

		// **Ohne Prüfung auf eine Spanne von null.** Sie kann hier nicht
		// auftreten: die Schleife ist nur dann hier, wenn `stelle` größer als
		// die vorige Stufe und höchstens so groß wie diese ist — dann liegen
		// die beiden Stufen auseinander. Ein Zweig, den kein Test betreten
		// kann, ist eine Behauptung darüber, was im Ernstfall geschieht.
		spanne := stufen[i].Stelle - stufen[i-1].Stelle

		return zwischen(stufen[i-1].Farbe, stufen[i].Farbe, (stelle-stufen[i-1].Stelle)/spanne)
	}

	return stufen[len(stufen)-1].Farbe
}

func zwischen(von, bis Farbe, anteil float64) Farbe {
	kanal := func(a, b uint8) uint8 {
		return alsAnteil(float64(a) + (float64(b)-float64(a))*anteil)
	}

	return Farbe{
		R: kanal(von.R, bis.R),
		G: kanal(von.G, bis.G),
		B: kanal(von.B, bis.B),
		A: kanal(von.A, bis.A),
	}
}
