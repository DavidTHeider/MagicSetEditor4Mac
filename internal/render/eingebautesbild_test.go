package render_test

import (
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Die zwei eingebauten Bilder, die im vollen Pack wirklich vorkommen.**
//
// Gezählt am 2026-09-03 über `/Users/david/full-magic-pack`: `built_in_image`
// steht **neunmal in vier Dateien**, und jedes Mal mit `bool_yes` oder
// `bool_no` — die Häkchen an einem Ja-Nein-Feld. Andere Namen kommen nicht
// vor.
//
//	SCRIPT_FUNCTION(built_in_image) {
//	  SCRIPT_PARAM_C(String, input);
//	  return make_intrusive<BuiltInImage>(input);
//	}
//
// `src/script/functions/image.cpp`, Zeilen 204–207, Fassung f9d9356 vom
// 2021-07-06. `BuiltInImage::generate` ruft `load_resource_image(name)`
// (`src/gfx/generated_image.cpp`, Zeile 436) — MSEs eigene Ressourcen.
func TestDieEingebautenBilderStehenBereit(t *testing.T) {
	for _, fall := range []struct {
		name  string
		mitte color.NRGBA
	}{
		{"bool_yes", color.NRGBA{R: 0, G: 128, B: 0, A: 255}},
		{"bool_no", color.NRGBA{R: 192, G: 0, B: 0, A: 255}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			bild, befund := render.NeuerRasterer(gemaltequelle{bild: zweifarbig()}).
				AusBaumFuerTest(render.Bildbaum{
					Knotenart: "built_in_image",
					Texte:     map[string]string{"name": fall.name},
				})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			if rahmen := bild.Bounds(); rahmen.Dx() != 16 || rahmen.Dy() != 16 {
				t.Fatalf("das Bild ist %dx%d, erwartet 16x16", rahmen.Dx(), rahmen.Dy())
			}

			// Beide Zeichnungen laufen durch die Mitte.
			hier := color.NRGBAModel.Convert(bild.At(8, 8))
			if hier != color.Color(fall.mitte) {
				t.Errorf("die Mitte ist %+v, erwartet %+v", hier, fall.mitte)
			}

			// Und beide lassen die Ecken frei.
			if ecke := color.NRGBAModel.Convert(bild.At(0, 0)); ecke != color.Color(color.NRGBA{}) {
				t.Errorf("die linke obere Ecke ist nicht durchsichtig: %+v", ecke)
			}
		})
	}
}

// **Ein unbekannter Name wird benannt, nicht erfunden.**
//
// MSE hat 59 Ressourcen, und die übrigen 57 sind Bedienelemente seines
// eigenen Fensters — Knöpfe, Griffe, Mauszeiger. Sie hier nachzuzeichnen
// hieße, das Aussehen eines fremden Fensters in einen Rasterer zu legen; wer
// eine davon aufruft, bekommt deshalb einen Befund und ein leeres Bild.
func TestEinUnbekanntesEingebautesBildWirdBenannt(t *testing.T) {
	bild, befund := render.NeuerRasterer(gemaltequelle{bild: zweifarbig()}).
		AusBaumFuerTest(render.Bildbaum{
			Knotenart: "built_in_image",
			Texte:     map[string]string{"name": "btn_ok_hover"},
		})
	if bild == nil {
		t.Fatal("ohne Bild")
	}

	if !strings.Contains(befund, "btn_ok_hover") {
		t.Errorf("der Befund lautet %q, erwartet wurde der Name darin", befund)
	}
}
