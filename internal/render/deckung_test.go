package render_test

import (
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`set_alpha` setzt die Durchsichtigkeit eines Bildes.**
//
//	void set_alpha(Image& img, double alpha) {
//	  Byte b_alpha = Byte(alpha * 255);
//	  if (!img.HasAlpha()) {
//	    img.InitAlpha();
//	    memset(img.GetAlpha(), b_alpha, …);
//	  } else {
//	    for (…) im[i] = (im[i] * b_alpha) / 255;
//	  }
//	}
//
// `src/gfx/blend_image.cpp`, Zeilen 107–119, Fassung f9d9356 vom 2021-07-06.
//
// **Ein Bild ohne Alpha bekommt den Wert gesetzt, eines mit Alpha ihn
// multipliziert** — die vorhandene Form bleibt also erhalten. 239 Befunde über
// die 190 Stile von `M15-Magic-Pack`
// (#270).
func TestSetAlphaSetztDieDurchsichtigkeit(t *testing.T) {
	faelle := []struct {
		name     string
		quelle   color.RGBA
		alpha    float64
		erwartet uint8
	}{
		{name: "ganz deckend bleibt deckend", quelle: color.RGBA{R: 255, A: 255}, alpha: 1, erwartet: 255},
		{name: "halb", quelle: color.RGBA{R: 255, A: 255}, alpha: 0.5, erwartet: 127},

		// **Multipliziert, nicht überschrieben**: 128 · 127 / 255 = 63.
		{name: "auf ein halbdurchsichtiges Bild", quelle: color.RGBA{R: 255, A: 128}, alpha: 0.5, erwartet: 63},
		{name: "ganz durchsichtig", quelle: color.RGBA{R: 255, A: 255}, alpha: 0, erwartet: 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
				"/eins.png": fall.quelle,
			}})

			bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
				Knotenart: "set_alpha",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
				Zahlen:    map[string]float64{"alpha": fall.alpha},
			})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			if _, _, _, alpha := bild.At(0, 0).RGBA(); alpha>>8 != uint32(fall.erwartet) {
				t.Errorf("Alpha %d, erwartet %d", alpha>>8, fall.erwartet)
			}
		})
	}
}

// **Ohne `alpha:` ist es ein Befund.**
//
// `SCRIPT_PARAM(double, alpha)` — ein Pflichtargument. Eine Vorlage ohne es ist
// falsch geschrieben, und ein durchgereichtes Bild verdeckte das (§8).
func TestSetAlphaOhneAlphaIstEinBefund(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 255, A: 255},
	}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "set_alpha",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
	})

	if !strings.Contains(befund, "alpha") {
		t.Errorf("Befund %q nennt das fehlende Argument nicht", befund)
	}

	if bild == nil {
		t.Error("ohne `alpha:` wurde gar nichts gezeichnet — erwartet das Ausgangsbild")
	}
}

// **`set_combine` ändert das Bild nicht.**
//
//	Image SetCombineImage::generate(const Options& opt) const {
//	  return image->generate(opt);
//	}
//
// `src/gfx/generated_image.cpp`, Zeilen 175–177. Es setzt die **Mischart** der
// Ebene, und die trägt der Plan an der Ebene, nicht am Bild — hier reicht das
// Bild durch, damit die Kette darüber weiterläuft.
//
// Fassung f9d9356 vom 2021-07-06.
func TestSetCombineReichtDasBildDurch(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 200, G: 100, B: 50, A: 255},
	}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "set_combine",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
		Texte:     map[string]string{"combine": "add"},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	rot, gruen, blau, _ := bild.At(0, 0).RGBA()
	if rot>>8 != 200 || gruen>>8 != 100 || blau>>8 != 50 {
		t.Errorf("(%d,%d,%d), erwartet das Ausgangsbild", rot>>8, gruen>>8, blau>>8)
	}
}

// **Ein unlesbares Ausgangsbild bricht `set_alpha` ab.**
//
// Kein halber Stand: wo nichts zu setzen ist, wird nichts gezeichnet, und der
// Befund nennt die Datei (§8).
func TestSetAlphaOhneAusgangsbild(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "set_alpha",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/fehlt.png")}},
		Zahlen:    map[string]float64{"alpha": 0.5},
	})

	if bild != nil || !strings.Contains(befund, "/fehlt.png") {
		t.Errorf("Bild %v, Befund %q", bild != nil, befund)
	}
}
