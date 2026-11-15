package render_test

import (
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`symbol_variation` rastert eine Vektorzeichnung in den Bildbaum.**
//
//	return make_intrusive<SymbolToImage>(value, filename, value->last_update, v);
//
// `src/script/functions/image.cpp`, Zeile 148, Fassung f9d9356 vom
// 2021-07-06 — eine der fünf Bildquellen aus `generated_image.hpp`.
//
// **Die Zeichnung steht im Plan, nicht im Rasterer.** Wer die `.mse-symbol`
// liest, ist `internal/service`; hier kommt die Gestalt an, wie [RastereSymbol]
// sie braucht (Vermerk 0022,
// #323).
func TestSymbolVariationRastertDieZeichnung(t *testing.T) {
	rasterer := render.NeuerRasterer(bilder{})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "symbol_variation",
		Symbol:    []render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)},
		Zahlen:    map[string]float64{"border_radius": 0.05},
		Texte:     map[string]string{"fill_type": "solid"},
		Farben: map[string]render.Farbe{
			"fill_color":   {R: 210, G: 180, B: 60, A: 255},
			"border_color": {A: 255},
		},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	rahmen := bild.Bounds()
	if rahmen.Dx() < 2 || rahmen.Dy() < 2 {
		t.Fatalf("das Bild ist %dx%d", rahmen.Dx(), rahmen.Dy())
	}

	mitte := color.NRGBAModel.Convert(bild.At(rahmen.Dx()/2, rahmen.Dy()/2))
	if mitte != (color.NRGBA{R: 210, G: 180, B: 60, A: 255}) {
		t.Errorf("die Mitte ist %+v, erwartet die Füllfarbe", mitte)
	}

	// **Außerhalb ist durchsichtig** — ein Symbol steht auf einer Karte.
	ecke := color.NRGBAModel.Convert(bild.At(1, 1))
	if ecke != (color.NRGBA{}) {
		t.Errorf("die Ecke ist %+v, erwartet nichts", ecke)
	}
}

// **Ohne Zeichnung ist es ein Befund.**
//
// Ein Set ohne Symbol ist der Normalfall — aber dann steht auch kein Baum da.
// Kommt einer ohne Gestalt an, hat der Planbauer die Datei nicht lesen können,
// und das gehört gesagt (§8).
func TestSymbolVariationOhneZeichnungIstEinBefund(t *testing.T) {
	_, befund := render.NeuerRasterer(bilder{}).AusBaumFuerTest(render.Bildbaum{
		Knotenart: "symbol_variation",
		Texte:     map[string]string{"symbol": "/probe.mse-set/fehlt.mse-symbol"},
	})

	if !strings.Contains(befund, "fehlt.mse-symbol") {
		t.Errorf("der Befund lautet %q und nennt die Datei nicht", befund)
	}
}

// **Die drei Fülltypen entscheiden über den Filter.**
//
//	if (fill_type == _("solid") || fill_type.empty()) …
//	else if (fill_type == _("linear gradient")) …
//	else if (fill_type == _("radial gradient")) …
//
// `src/script/functions/image.cpp`, Zeilen 172–195. **Ohne Angabe gilt
// `solid`.**
//
// Fassung f9d9356 vom 2021-07-06.
func TestDieDreiFuelltypenEinerFassung(t *testing.T) {
	for _, art := range []string{"", "solid", "linear gradient", "radial gradient"} {
		t.Run(art, func(t *testing.T) {
			bild, befund := render.NeuerRasterer(bilder{}).AusBaumFuerTest(render.Bildbaum{
				Knotenart: "symbol_variation",
				Symbol:    []render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)},
				Texte:     map[string]string{"fill_type": art},
				Farben: map[string]render.Farbe{
					"fill_color": {R: 200, A: 255}, "border_color": {A: 255},
					"fill_color_1": {R: 200, A: 255}, "fill_color_2": {B: 200, A: 255},
					"border_color_1": {A: 255}, "border_color_2": {A: 255},
				},
			})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			rahmen := bild.Bounds()
			if _, _, _, alpha := bild.At(rahmen.Dx()/2, rahmen.Dy()/2).RGBA(); alpha == 0 {
				t.Error("die Mitte ist durchsichtig — es wurde nichts gefüllt")
			}
		})
	}
}
