package render

import (
	"image"
	"image/color"
)

// **23 Mischarten, und jede ist eine Zeile.**
//
//	COMBINE_FUN(COMBINE_ADD,         top(a + b))
//	COMBINE_FUN(COMBINE_MULTIPLY,    (a * b) / 255)
//	COMBINE_FUN(COMBINE_SCREEN,      255 - (((255 - a) * (255 - b)) / 255))
//	…
//
// `src/gfx/combine_image.cpp`, Zeilen 57–85, Fassung f9d9356 vom 2021-07-06.
// MSEs Quelle dafür steht dort: „Based on
// http://www.pegtop.net/delphi/articles/blendmodes/".
//
// **`soft light` gibt `b` zurück** — dieselbe Formel wie `normal`. Das ist
// MSEs Zeile, keine Auslassung dieses Programms
// (#270).

// mischarten ordnet den Namen aus `combine:` ihre Rechnung zu.
//
// **Über den Namen und nicht über eine Zahl.** MSEs `parse_enum` liest
// denselben Text, und eine eigene Nummerierung daneben liefe auseinander.
var mischarten = map[string]func(a, b int) int{
	"normal":     func(_, b int) int { return b },
	"add":        func(a, b int) int { return min(255, a+b) },
	"subtract":   func(a, b int) int { return max(0, a-b) },
	"stamp":      func(a, b int) int { return kanalwert(a - 2*b + 256) },
	"difference": func(a, b int) int { return abs(a - b) },
	"negation":   func(a, b int) int { return 255 - abs(255-a-b) },
	"multiply":   func(a, b int) int { return a * b / 255 },
	"darken":     func(a, b int) int { return min(a, b) },
	"lighten":    func(a, b int) int { return max(a, b) },
	"screen":     func(a, b int) int { return 255 - (255-a)*(255-b)/255 },

	// **`>> 7` und nicht `/ 128`.** MSE schiebt, und bei negativen Zahlen ist
	// das nicht dasselbe — hier kann keine auftreten, und die Schreibweise
	// bleibt trotzdem die seine.
	"overlay":    func(a, b int) int { return ueberlagert(a, b, a) },
	"hard light": func(a, b int) int { return ueberlagert(a, b, b) },

	// **`soft light` ist `b`** — MSEs Zeile, nicht unsere Vereinfachung.
	"soft light": func(_, b int) int { return b },

	"color dodge": func(a, b int) int {
		if b == 255 {
			return 255
		}

		return min(255, a*255/(255-b))
	},
	"color burn": func(a, b int) int {
		if b == 0 {
			return 0
		}

		return max(0, 255-(255-a)*255/b)
	},
	"reflect": func(a, b int) int {
		if b == 255 {
			return 255
		}

		return min(255, a*a/(255-b))
	},
	"glow": func(a, b int) int {
		if a == 255 {
			return 255
		}

		return min(255, b*b/(255-a))
	},
	"freeze": func(a, b int) int {
		if b == 0 {
			return 0
		}

		return max(0, 255-(255-a)*(255-a)/b)
	},
	"heat": func(a, b int) int {
		if a == 0 {
			return 0
		}

		return max(0, 255-(255-b)*(255-b)/a)
	},

	"and":    func(a, b int) int { return a & b },
	"or":     func(a, b int) int { return a | b },
	"xor":    func(a, b int) int { return a ^ b },
	"shadow": func(a, b int) int { return b * a * a / (255 * 255) },

	"symmetric overlay": func(a, b int) int {
		return (ueberlagert(a, b, a) + ueberlagert(b, a, b)) / 2
	},
}

// ueberlagert ist die gemeinsame Formel von `overlay` und `hard light`.
//
// **Sie unterscheiden sich nur darin, wer entscheidet**: `overlay` fragt `a`,
// `hard light` fragt `b`.
func ueberlagert(a, b, entscheider int) int {
	if entscheider < 128 {
		return a * b >> 7
	}

	return 255 - ((255 - a) * (255 - b) >> 7)
}

// kanalwert ist MSEs `col` — es begrenzt auf 0…255.
func kanalwert(wert int) int { return max(0, min(255, wert)) }

func abs(wert int) int {
	if wert < 0 {
		return -wert
	}

	return wert
}

// mischeArten führt `combine_blend` aus.
//
//	Image img = image1->generate(opt);
//	combine_image(img, image2->generate(opt), image_combine);
//
// `CombineBlendImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 134–138.
//
// **Das Alpha kommt vom zweiten Bild** — `memcpy(a.GetAlpha(), b.GetAlpha(), …)`
// (`combine_image.cpp`, Zeilen 106–110), und das ist mehr als eine Feinheit:
// die Form des Ergebnisses ist die des **oberen** Bildes.
//
// Fassung f9d9356 vom 2021-07-06.
func (r *Rasterer) mischeArten(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	erstes, zweites, fehlend := zweiBilder(r, baum, tiefe, "image1", "image2")
	if fehlend != "" {
		return nil, fehlend
	}

	art := baum.Texte["combine"]
	if art == "" {
		art = "normal"
	}

	rechnung, bekannt := mischarten[art]
	if !bekannt {
		return erstes, "combine_blend kennt die Mischart " + art + " nicht"
	}

	if erstes.Bounds() != zweites.Bounds() {
		return nil, "combine_blend braucht zwei Bilder derselben Größe"
	}

	rahmen := erstes.Bounds()
	ziel := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			unten := nrgbaAn(erstes, rahmen.Min.X+x, rahmen.Min.Y+y)
			oben := nrgbaAn(zweites, rahmen.Min.X+x, rahmen.Min.Y+y)

			ziel.SetNRGBA(x, y, color.NRGBA{
				R: gemischterKanal(rechnung, unten.R, oben.R),
				G: gemischterKanal(rechnung, unten.G, oben.G),
				B: gemischterKanal(rechnung, unten.B, oben.B),
				A: oben.A,
			})
		}
	}

	return ziel, ""
}

// gemischterKanal rechnet einen Kanal und begrenzt ihn.
func gemischterKanal(rechnung func(a, b int) int, unten, oben uint8) uint8 {
	// #nosec G115 -- kanalwert begrenzt auf 0…255
	return uint8(kanalwert(rechnung(int(unten), int(oben)))) //nolint:gosec // dito
}

// setzeMaske führt `set_mask` aus.
//
//	Image img = image->generate(opt);
//	set_alpha(img, mask->generate(opt));
//
// `SetMaskImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 151–155, und
// `set_alpha`, `src/gfx/blend_image.cpp`, Zeilen 79–87:
//
//	im[i] = (im[i] * al[i*3]) / 255;
//
// **`al[i*3]` ist der rote Kanal**, nicht die Helligkeit. Bei einer grauen
// Maske ist beides gleich, bei einer farbigen nicht — derselbe Unterschied wie
// bei `masked_blend`, und dort hat er schon einmal Geld gekostet.
//
// Fassung f9d9356 vom 2021-07-06.
func (r *Rasterer) setzeMaske(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	bild, maske, fehlend := zweiBilder(r, baum, tiefe, "image", "mask")
	if fehlend != "" {
		return nil, fehlend
	}

	rahmen := bild.Bounds()
	ziel := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			farbe := nrgbaAn(bild, rahmen.Min.X+x, rahmen.Min.Y+y)
			anteil := maskenrot(maske, x, y, rahmen.Dx(), rahmen.Dy())

			// #nosec G115 -- beide Faktoren höchstens 255, das Ergebnis also auch
			farbe.A = uint8(int(farbe.A) * anteil / 255) //nolint:gosec // dito

			ziel.SetNRGBA(x, y, farbe)
		}
	}

	return ziel, ""
}

// maskenrot liest den roten Kanal der Maske an der Stelle, die dem Punkt
// (`x`, `y`) eines Bildes von `breite` × `hoehe` entspricht.
//
// **Skaliert auf die Bildgröße**, wie MSE:
//
//	Image img_alpha_resampled = resample(img_alpha, img.GetWidth(), img.GetHeight());
//	…
//	im[i] = (im[i] * al[i*3]) / 255;
//
// `src/gfx/blend_image.cpp`, Zeilen 79–87, Fassung f9d9356 vom 2021-07-06.
// `al[i*3]` ist der rote Kanal.
//
// **Was weiter abweicht:** hier wird der nächste Punkt genommen, MSE mittelt.
// Der Unterschied ist eine harte gegen eine weiche Kante an der Maskengrenze.
func maskenrot(maske image.Image, x, y, breite, hoehe int) int {
	rahmen := maske.Bounds()

	return int(nrgbaAn(maske,
		rahmen.Min.X+x*rahmen.Dx()/max(1, breite),
		rahmen.Min.Y+y*rahmen.Dy()/max(1, hoehe)).R)
}

// zweiBilder holt zwei benannte Kinder und nennt, welches fehlt.
func zweiBilder(
	r *Rasterer, baum Bildbaum, tiefe int, ersterName, zweiterName string,
) (erstes, zweites image.Image, fehlend string) {
	ersterBaum, ersterDa := baum.Kind(ersterName)
	zweiterBaum, zweiterDa := baum.Kind(zweiterName)

	if !ersterDa || !zweiterDa {
		return nil, nil, baum.Knotenart + " braucht `" + ersterName + ":` und `" +
			zweiterName + ":`"
	}

	erstes, befund := r.ausBaumTief(ersterBaum, tiefe+1)
	if erstes == nil {
		return nil, nil, befund
	}

	zweites, befund = r.ausBaumTief(zweiterBaum, tiefe+1)
	if zweites == nil {
		return nil, nil, befund
	}

	return erstes, zweites, ""
}
