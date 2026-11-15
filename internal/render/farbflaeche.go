package render

// **Die Form einer Farbfläche** — runde Ecken und ein Rahmen statt einer
// Füllung.
//
//	bool clip = style().left_width < style().width && …;
//	if (clip) {
//	  wxRegion r = dc.trRectToRegion(dc.getInternalRect());
//	  r.Subtract(dc.trRectToRegion(RealRect(style().left_width, style().top_width,
//	    style().width - style().left_width - style().right_width,
//	    style().height - style().top_width - style().bottom_width)));
//	  dc.getDC().SetClippingRegion(r);
//	}
//	dc.DrawRoundedRectangle(dc.getInternalRect(), style().radius);
//
// `ColorValueViewer::draw`, `src/render/value/color.cpp`, Zeilen 47–64, Fassung
// f9d9356 vom 2021-07-06.
//
// **Der schwarze Kartenrand ist ein Farbfeld.** `magic-m15.mse-style` gibt
// `border color` die Breiten 17/17/17/18 und einen Radius von 18
// (#270).

import "image"

// inFlaechenform sagt, ob ein Bildpunkt zur Form dieser Fläche gehört.
//
// **Ohne Kantenglättung.** MSE malt mit einem `wxRegion` und einem Pinsel; die
// Kanten sind dort so hart wie hier. Eine geglättete Rundung wäre schöner und
// wäre nicht dasselbe Bild — der Goldbildvergleich soll den Unterschied zu MSE
// zeigen und nicht unseren Geschmack.
func inFlaechenform(ebene Ebene, rahmen image.Rectangle, x, y int, vergroesserung float64) bool {
	if ebene.Rand.Rahmt(ebene.Rahmen.Breite, ebene.Rahmen.Hoehe) {
		innen := image.Rect(
			rahmen.Min.X+stellen(ebene.Rand.Links*vergroesserung),
			rahmen.Min.Y+stellen(ebene.Rand.Oben*vergroesserung),
			rahmen.Max.X-stellen(ebene.Rand.Rechts*vergroesserung),
			rahmen.Max.Y-stellen(ebene.Rand.Unten*vergroesserung),
		)

		if (image.Point{X: x, Y: y}).In(innen) {
			return false
		}
	}

	return inRundemRechteck(rahmen, ebene.Ecke*vergroesserung, x, y)
}

// inRundemRechteck prüft die vier Eckviertel.
//
// **Nur die Ecken werden geprüft**, und das ist die Form selbst: außerhalb der
// Eckquadrate ist ein rundes Rechteck ein Rechteck.
func inRundemRechteck(rahmen image.Rectangle, radius float64, x, y int) bool {
	if radius <= 0 {
		return true
	}

	ecke := stellen(radius)

	// **Ein Radius über die halbe Seite hinaus wird nicht abgewiesen.** wx
	// zeichnet dann eine Linse; hier wird der Radius auf die halbe Seite
	// begrenzt, und der Unterschied steht in `docs/FORMATE.md`.
	ecke = min(ecke, rahmen.Dx()/2, rahmen.Dy()/2)
	if ecke <= 0 {
		return true
	}

	mitte := image.Point{X: rahmen.Min.X + ecke, Y: rahmen.Min.Y + ecke}

	switch {
	case x >= rahmen.Max.X-ecke:
		mitte.X = rahmen.Max.X - ecke - 1

	case x >= rahmen.Min.X+ecke:
		return true
	}

	switch {
	case y >= rahmen.Max.Y-ecke:
		mitte.Y = rahmen.Max.Y - ecke - 1

	case y >= rahmen.Min.Y+ecke:
		return true
	}

	abstandX := float64(x - mitte.X)
	abstandY := float64(y - mitte.Y)

	return abstandX*abstandX+abstandY*abstandY <= float64(ecke*ecke)
}

// stellen rundet eine Länge in Vorlagenpunkten auf ganze Bildpunkte.
func stellen(laenge float64) int {
	if laenge <= 0 {
		return 0
	}

	return int(laenge + 0.5)
}
