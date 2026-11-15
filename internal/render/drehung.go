package render

import (
	"image"
	"math"
	"strconv"
)

// ebeneGedreht zeichnet eine Ebene in ihrem eigenen gedrehten Kontext.
//
// # Wie MSE es macht
//
//	Rotation ValueViewer::getRotation() const {
//	  return Rotation(deg_to_rad(getStyle()->angle), bounding_box, 1.0, getStretch());
//	}
//
// `src/render/value/viewer.cpp`, Zeilen 44–46. Jedes Feld bekommt einen
// eigenen Zeichenkontext; Text, Bilder und Symbole laufen darin, als wäre
// nichts gedreht. `Rotation::tr` rechnet erst beim Setzen um
// (`src/util/rotation.cpp`, Zeilen 37–42):
//
//	return RealPoint(c * x + s * y + origin.x,
//	                -s * x + c * y + origin.y);
//
// Fassung f9d9356 vom 2021-07-06.
//
// # Wie es hier gebaut ist
//
// **Ein Puffer und eine Umrechnung beim Kopieren.** Die Ebene wird in ihrer
// **ungedrehten** Größe an den Ursprung gezeichnet und danach Punkt für Punkt
// an ihre gedrehte Stelle gelegt. Das trifft MSEs Bauform genauer als eine
// Drehung in jeder einzelnen Zeichenroutine — und billiger ist es auch: die
// Umrechnung steht an **einer** Stelle statt in vier.
//
// **Was das kostet, gehört ausgesprochen**: ein Bildpuffer je gedrehter Ebene.
// Ungedrehte Ebenen gehen weiter geradeaus durch, und das sind fast alle.
func (r *Rasterer) ebeneGedreht(
	ziel *image.NRGBA, ebene Ebene, maske *maskenbild, vergroesserung float64,
	guete Guete,
) []string {
	if ebene.Winkel == 0 {
		return r.ebeneZeichnen(ziel, ebene, maske, vergroesserung, guete)
	}

	// **Nur die geraden Winkel.** MSE unterscheidet sie selbst
	// (`is_straight(angle)`); schräge brauchen eine Neuabtastung, und im
	// vollen Pack stehen sie viermal gegen 142 gerade
	// (#486).
	viertel, gerade := geraderWinkel(ebene.Winkel)
	if !gerade {
		return append([]string{"ein Winkel von " + strconv.FormatFloat(ebene.Winkel, 'g', -1, 64) +
			" Grad ist nicht gerade — gezeichnet wird ungedreht"},
			r.ebeneZeichnen(ziel, ebene, maske, vergroesserung, guete)...)
	}

	aussen := skaliert(ebene.Rahmen, vergroesserung)
	if aussen.Empty() {
		return nil
	}

	// **Der Puffer trägt die ungedrehte Ebene**, an den Ursprung gerückt.
	puffer := image.NewNRGBA(image.Rect(0, 0, aussen.Dx(), aussen.Dy()))

	innen := ebene
	innen.Rahmen.Links, innen.Rahmen.Oben = 0, 0

	maengel := r.ebeneZeichnen(puffer, innen, maske, vergroesserung, guete)

	for by := range aussen.Dy() {
		for bx := range aussen.Dx() {
			roh := nrgbaAn(puffer, bx, by)
			if roh.A == 0 {
				continue
			}

			x, y := gedrehteStelle(viertel, aussen.Min.X, aussen.Min.Y, bx, by)
			if !(image.Point{X: x, Y: y}).In(ziel.Bounds()) {
				continue
			}

			// **Deckung 1**: was die Ebene an Deckung und Maske mitbringt,
			// steht schon im Puffer. Sie hier ein zweites Mal anzuwenden
			// verdoppelte sie.
			r.punkt(ziel, x, y, Farbe{R: roh.R, G: roh.G, B: roh.B, A: roh.A}, 1, nil)
		}
	}

	return maengel
}

// geraderWinkel sagt, ob ein Winkel ein Vielfaches von 90 Grad ist, und
// welches — 0, 1, 2 oder 3 Viertel gegen den Uhrzeigersinn.
//
// **Mit Toleranz.** Ein Winkel kommt aus einem Skript und darf `270.0000001`
// heißen; ein Vergleich auf Gleichheit machte daraus einen schrägen.
func geraderWinkel(grad float64) (viertel int, gerade bool) {
	umlauf := math.Mod(math.Mod(grad, 360)+360, 360)

	gerundet := math.Round(umlauf / 90)
	if math.Abs(umlauf-gerundet*90) > 0.001 {
		return 0, false
	}

	return int(gerundet) % 4, true
}

// gedrehteStelle rechnet einen Punkt des ungedrehten Feldes an seine Stelle
// auf der Karte.
//
// Aus `Rotation::tr` für die vier geraden Winkel:
//
//	  0°  → (links + bx, oben + by)
//	 90°  → (links + by, oben − bx)
//	180°  → (links − bx, oben − by)
//	270°  → (links − by, oben + bx)
func gedrehteStelle(viertel, links, oben, bx, by int) (x, y int) {
	switch viertel {
	case 1:
		return links + by, oben - bx
	case 2:
		return links - bx, oben - by
	case 3:
		return links - by, oben + bx
	default:
		return links + bx, oben + by
	}
}
