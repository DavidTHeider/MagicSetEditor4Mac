package render

import (
	"image"
)

// **`set_alpha` setzt die Durchsichtigkeit, `set_combine` die Mischart.**
//
//	void set_alpha(Image& img, double alpha) {
//	  Byte b_alpha = Byte(alpha * 255);
//	  if (!img.HasAlpha()) {
//	    img.InitAlpha();
//	    memset(img.GetAlpha(), b_alpha, img.GetWidth() * img.GetHeight());
//	  } else {
//	    Byte *im = img.GetAlpha();
//	    for (size_t i = 0 ; i < size ; ++i) im[i] = (im[i] * b_alpha) / 255;
//	  }
//	}
//
// `src/gfx/blend_image.cpp`, Zeilen 107–119, Fassung f9d9356 vom 2021-07-06.
//
// Zusammen 239 Befunde über die 190 Stile von `M15-Magic-Pack`
// (#270).

// setzeDeckung führt `set_alpha` aus.
//
// **Ein Bild ohne Alpha bekommt den Wert gesetzt, eines mit Alpha ihn
// multipliziert.** Der Unterschied ist nicht kosmetisch: die zweite Form
// erhält die Form eines freigestellten Rahmens, die erste machte daraus ein
// Rechteck.
//
// **Go kennt kein „ohne Alpha".** Jedes gelesene Bild trägt hier einen
// Alphakanal; ein PNG ohne einen liest sich als durchweg 255. Damit fällt der
// erste Zweig mit dem zweiten zusammen — `255 · a / 255 = a` —, und das ist
// keine Näherung, sondern dieselbe Rechnung.
func (r *Rasterer) setzeDeckung(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	unten, tiefer := r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
	if unten == nil {
		return nil, tiefer
	}

	anteil, da := baum.Zahlen["alpha"]
	if !da {
		return unten, "set_alpha braucht ein Argument `alpha:`"
	}

	// **Abgeschnitten, nicht gerundet** — `Byte(alpha * 255)`.
	gesetzt := uint8(max(0, min(255, int(anteil*255)))) //nolint:gosec // durch min/max begrenzt

	rahmen := unten.Bounds()
	ziel := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			farbe := nrgbaAn(unten, rahmen.Min.X+x, rahmen.Min.Y+y)
			// #nosec G115 -- beide Faktoren sind höchstens 255, das Ergebnis also auch
			farbe.A = uint8(int(farbe.A) * int(gesetzt) / 255) //nolint:gosec // dito

			ziel.SetNRGBA(x, y, farbe)
		}
	}

	return ziel, ""
}

// reicheDurch führt `set_combine` aus — es ändert am Bild nichts.
//
//	Image SetCombineImage::generate(const Options& opt) const {
//	  return image->generate(opt);
//	}
//
// `src/gfx/generated_image.cpp`, Zeilen 175–177.
//
// **Was es setzt, ist die Mischart der Ebene** (`ImageCombine combine() const`).
// Die trägt der Plan an der Ebene und nicht am Bild; hier reicht das Bild
// durch, damit die Kette darüber weiterläuft. Wer die Mischart aus dem Bildbaum
// an die Ebene bringen will, findet hier ihre Stelle — heute liest sie niemand.
//
// Fassung f9d9356 vom 2021-07-06.
func (r *Rasterer) reicheDurch(baum Bildbaum, tiefe int) (quelle image.Image, befund string) {
	return r.ausBaumTief(baum.Kinder[0].Baum, tiefe+1)
}
