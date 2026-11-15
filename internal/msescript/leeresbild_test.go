package msescript_test

// **Eine leere Zeichenkette und `nil` sind ein leeres Bild, kein Fehler**
// (#509).
//
//	GeneratedImageP toImage() const override {
//	  if (value.empty()) {
//	    return make_intrusive<BlankImage>();
//	  } else {
//	    return make_intrusive<PackagedImage>(value);
//	  }
//	}
//
// `ScriptString::toImage`, `src/script/value.cpp`, Zeilen 369–375, und für
// `nil` dasselbe in Zeile 450. Fassung f9d9356 vom 2021-07-06.
//
//	Image BlankImage::generate(const Options& opt) const {
//	  int w = max(1, opt.width >= 0  ? opt.width  : opt.height);
//	  int h = max(1, opt.height >= 0 ? opt.height : opt.width);
//	  Image img(w, h);
//	  img.InitAlpha();
//	  memset(img.GetAlpha(), 0, w * h);
//	  return img;
//	}
//
// `src/gfx/generated_image.cpp`, Zeilen 83–91 — **durchsichtig in der Größe,
// die gefragt ist.**
//
// # Woran es aufgefallen ist
//
// Der Lauf über alle 23.168 kanonischen Karten meldete „linear_blend bekam für
// image1 kein Bild, sondern Zeichenkette" — die Zeichenkette war **leer**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestEineLeereAngabeIstEinLeeresBild(t *testing.T) {
	faelle := []struct{ name, quelle string }{
		{"die leere Zeichenkette", `linear_blend(image1: "", image2: "b.png", x1: 0, y1: 0, x2: 1, y2: 0)`},
		{"nil", `linear_blend(image1: nil, image2: "b.png", x1: 0, y1: 0, x2: 1, y2: 0)`},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(
				t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			baum, istBild := wert.AlsBild()
			if !istBild {
				t.Fatalf("%v ist kein Bild", wert)
			}

			var gefunden bool

			for _, kind := range baum.Kinder {
				if kind.Name != "image1" {
					continue
				}

				gefunden = true

				if kind.Baum.Knotenart != "blank_image" {
					t.Errorf("image1 ist %q, erwartet blank_image", kind.Baum.Knotenart)
				}
			}

			if !gefunden {
				t.Fatal("image1 fehlt im Baum")
			}
		})
	}
}
