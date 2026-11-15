package render

import (
	"image"
	"strconv"
)

// drehe dreht das Ausgangsbild um einen geraden Winkel.
//
// # Wie MSE es macht
//
//	Image rotate_image(const Image& image, Radians angle) {
//	  double a = constrain_radians(angle);
//	  if (is_rad0(a))   return image;
//	  if (is_rad90(a))  return rotate_image_impl<Rotate90deg> (image);
//	  …
//	  else return image.Rotate(angle, wxPoint(0,0));
//	}
//
// `src/gfx/rotate_image.cpp`, Zeilen 86–95, Fassung f9d9356 vom 2021-07-06.
// Die drei Strukturen darüber (Zeilen 49–82) rechnen je einen Zielversatz
// aus Quellkoordinaten; 90° und 270° legen dabei die Seitenlängen um.
//
// # Was hier anders ist
//
// **Der schräge Winkel wird benannt statt neu abgetastet** — dieselbe Antwort
// wie bei einer schräg gestellten Ebene ([Rasterer.ebeneGedreht]). MSE fällt
// dafür auf `wxImage::Rotate` zurück, eine Neuabtastung mit eigenem
// Randverhalten; im vollen Pack steht `rotate(` **null**-mal, und eine
// Neuabtastung ohne einen einzigen Aufrufer wäre ungeprüfte Fläche, die sich
// gegen nichts vergleichen ließe.
//
// **Kennzeichnung: adapted** (§15).
func (r *Rasterer) drehe(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	grad, da := baum.Zahlen["angle"]
	if !da {
		return unten, "rotate braucht `angle:`"
	}

	viertel, gerade := geraderWinkel(grad)
	if !gerade {
		return unten, "ein Winkel von " + strconv.FormatFloat(grad, 'g', -1, 64) +
			" Grad ist nicht gerade — gezeichnet wird ungedreht"
	}

	if viertel == 0 {
		return unten, ""
	}

	rahmen := unten.Bounds()
	breite, hoehe := rahmen.Dx(), rahmen.Dy()

	zielBreite, zielHoehe := hoehe, breite
	if viertel == 2 {
		zielBreite, zielHoehe = breite, hoehe
	}

	ziel := image.NewNRGBA(image.Rect(0, 0, zielBreite, zielHoehe))

	for y := range hoehe {
		for x := range breite {
			zx, zy := gedrehterPunkt(viertel, x, y, breite, hoehe)
			ziel.SetNRGBA(zx, zy, nrgbaAn(unten, rahmen.Min.X+x, rahmen.Min.Y+y))
		}
	}

	return ziel, ""
}

// gedrehterPunkt rechnet einen Quellpunkt an seine Stelle im gedrehten Bild.
//
// Die drei Fälle stehen als `offset`-Funktionen in `Rotate90deg`,
// `Rotate180deg` und `Rotate270deg` (`src/gfx/rotate_image.cpp`, Zeilen 49–82,
// Fassung f9d9356 vom 2021-07-06). Dort liefern sie einen Versatz in ein
// flaches Feld; hier stehen dieselben Zahlen als Punkt, weil `image.NRGBA`
// ihn so nimmt.
func gedrehterPunkt(viertel, x, y, breite, hoehe int) (zx, zy int) {
	switch viertel {
	case 1:
		return y, breite - x - 1
	case 2:
		return breite - x - 1, hoehe - y - 1
	default:
		return hoehe - y - 1, x
	}
}
