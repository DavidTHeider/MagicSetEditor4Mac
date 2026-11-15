package render

import (
	"image"
	"image/color"
	"math"
)

// `drop_shadow` legt einen weichgezeichneten Schatten unter ein Bild
// (#472).
//
// **129 Aufrufe im vollen Pack**, 104 in `M15-Magic-Pack` — die häufigste
// Bildverknüpfung, die hier fehlte. Ein echter Aufruf sieht so aus:
//
//	drop_shadow(offset_x: -0.01, offset_y: 0.03, blur_radius: 0.02,
//	            alpha: 1, color: rgb(0,0,0), enlarged_image)
//
// **Meist zusammen mit `enlarge`**: der Schatten braucht Platz neben dem Bild,
// und den schafft der Rand.
//
// # Die Rechnung
//
//	int dw = int(w * offset_x), dh = int(h * offset_y);
//	int delta = dw + w * dh;
//	int sa = (int)(shadow_alpha * (1 << 16));
//	…
//	int shad = ((((255 - a)*sa)>>16) * shadow[p - delta]) / total;
//	int factor = max(1, a + shad);
//	data[3*p] = (a * data[3*p] + shad * shadow_color.Red()) / factor;
//	alpha[p] = a + shad;
//
// `src/gfx/generated_image.cpp`, Zeilen 370–401, Fassung f9d9356 vom
// 2021-07-06.
//
// **Der Schatten kommt aus dem Alphakanal**, nicht aus den Farben: was
// undurchsichtig ist, wirft Schatten. **Und er erscheint nur, wo das Bild
// selbst durchsichtig ist** — `(255 − a)`.
//
// **Ohne Alphakanal geschieht nichts** (Zeilen 373–376). Ein Bild, das überall
// deckt, hat keinen Rand, an dem ein Schatten sichtbar würde.
func (r *Rasterer) schlagschatten(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	rahmen := unten.Bounds()
	breite, hoehe := rahmen.Dx(), rahmen.Dy()

	if breite <= 0 || hoehe <= 0 {
		return unten, ""
	}

	// **Die Vorgaben sind null** — `SCRIPT_OPTIONAL_PARAM_`. Mit `alpha: 0`
	// ist der Schatten unsichtbar, und das ist die richtige Antwort auf einen
	// Aufruf ohne Angaben.
	schattenalpha := baum.Zahlen["alpha"]
	radius := baum.Zahlen["blur_radius"]

	schattenfarbe, hatFarbe := baum.Farben["color"]
	if !hatFarbe {
		schattenfarbe = Farbe{A: 255}
	}

	deckung := make([]uint8, breite*hoehe)
	for y := range hoehe {
		for x := range breite {
			deckung[x+y*breite] = nrgbaAn(unten, rahmen.Min.X+x, rahmen.Min.Y+y).A
		}
	}

	schatten, gesamt := weichgezeichnet(deckung, breite, hoehe, radius)
	if gesamt == 0 {
		return unten, ""
	}

	versatzX := int(float64(breite) * baum.Zahlen["offset_x"])
	versatzY := int(float64(hoehe) * baum.Zahlen["offset_y"])

	return mitSchatten(unten, deckung, schatten, gesamt,
		schattenalpha, schattenfarbe, versatzX, versatzY), ""
}

// mitSchatten legt den weichgezeichneten Schatten unter das Bild.
func mitSchatten(
	unten image.Image, deckung []uint8, schatten []int64, gesamt int64,
	schattenalpha float64, farbe Farbe, versatzX, versatzY int,
) *image.NRGBA {
	rahmen := unten.Bounds()
	breite, hoehe := rahmen.Dx(), rahmen.Dy()

	ziel := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	for y := range hoehe {
		for x := range breite {
			ziel.SetNRGBA(x, y, nrgbaAn(unten, rahmen.Min.X+x, rahmen.Min.Y+y))
		}
	}

	// `int sa = (int)(shadow_alpha * (1 << 16));`
	sa := int64(schattenalpha * (1 << 16))
	verschiebung := versatzX + breite*versatzY

	for y := max(0, versatzY); y < min(hoehe, hoehe+versatzY); y++ {
		for x := max(0, versatzX); x < min(breite, breite+versatzX); x++ {
			stelle := x + y*breite

			a := int64(deckung[stelle])
			schatte := (((255 - a) * sa) >> 16) * schatten[stelle-verschiebung] / gesamt

			// `int factor = max(1, a + shad);`
			teiler := max(int64(1), a+schatte)

			alt := ziel.NRGBAAt(x, y)
			ziel.SetNRGBA(x, y, color.NRGBA{
				R: kanalbyte((a*int64(alt.R) + schatte*int64(farbe.R)) / teiler),
				G: kanalbyte((a*int64(alt.G) + schatte*int64(farbe.G)) / teiler),
				B: kanalbyte((a*int64(alt.B) + schatte*int64(farbe.B)) / teiler),
				A: kanalbyte(a + schatte),
			})
		}
	}

	return ziel
}

// kanalbyte klemmt eine Zwischenrechnung auf einen Kanalwert.
func kanalbyte(wert int64) uint8 {
	return uint8(min(int64(255), max(int64(0), wert))) //nolint:gosec // durch min/max begrenzt
}

// weichgezeichnet ist `gaussian_blur` — zweimal eindimensional statt einmal
// zweidimensional.
//
//	double sigma = radius * w;
//	double mult = (1 << 8) / (sqrt(2 * M_PI) * sigma);
//	double sigsqr2 = 1 / (2 * sigma * sigma);
//	int range = min(w, (int)(3*sigma));
//	for (int d = -range ; d <= range ; ++d) {
//	  UInt factor = (int)( mult * exp(-d * d * sigsqr2) );
//	  …
//	}
//	return total_x * total_y;
//
// `src/gfx/generated_image.cpp`, Zeilen 322–367, Fassung f9d9356 vom
// 2021-07-06.
//
// **Die Streuung hängt an der Bildgröße**, nicht an einer Punktzahl: `radius`
// ist ein Anteil, `sigma` ist `radius · breite` beziehungsweise `radius · höhe`.
// Ein Bild doppelter Größe bekommt denselben Schatten, nur größer.
//
// **Die Faktoren sind ganzzahlig und werden abgeschnitten** (`(int)(mult *
// exp(…))`). Das ist kein Versehen: die Summe `total` teilt am Ende wieder
// heraus, und mit Fließkomma käme eine andere Zahl heraus als bei MSE.
func weichgezeichnet(
	deckung []uint8, breite, hoehe int, radius float64,
) (schatten []int64, gesamt int64) {
	schatten = make([]int64, breite*hoehe)

	if radius <= 0 {
		return schatten, 0
	}

	waagerecht := make([]int64, breite*hoehe)
	gesamtX := kernfaltung(breite, radius, func(d int, faktor int64) {
		for y := range hoehe {
			for x := max(0, -d); x < min(breite, breite-d); x++ {
				waagerecht[x+y*breite] += int64(deckung[x+d+y*breite]) * faktor
			}
		}
	})

	gesamtY := kernfaltung(hoehe, radius, func(d int, faktor int64) {
		for y := max(0, -d); y < min(hoehe, hoehe-d); y++ {
			for x := range breite {
				schatten[x+y*breite] += waagerecht[x+(d+y)*breite] * faktor
			}
		}
	})

	// **`255 * total`**, denn `shadow` trägt Deckungswerte bis 255:
	// `UInt total = 255 * gaussian_blur(…)` (Zeile 381).
	return schatten, 255 * gesamtX * gesamtY
}

// kernfaltung läuft über den Gaußkern einer Achse und gibt die Summe der
// Faktoren.
func kernfaltung(laenge int, radius float64, anwenden func(d int, faktor int64)) int64 {
	sigma := radius * float64(laenge)
	mult := float64(int(1)<<8) / (math.Sqrt(2*math.Pi) * sigma)
	sigsqr2 := 1 / (2 * sigma * sigma)

	weite := min(laenge, int(3*sigma))

	var gesamt int64

	for d := -weite; d <= weite; d++ {
		faktor := int64(mult * math.Exp(-float64(d)*float64(d)*sigsqr2))
		gesamt += faktor

		if faktor > 0 {
			anwenden(d, faktor)
		}
	}

	return gesamt
}
