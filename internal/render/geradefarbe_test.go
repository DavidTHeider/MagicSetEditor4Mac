package render_test

// **Ungewichtet lesen, auch wo ein Bildpunkt durchsichtig ist.**

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein durchsichtiger Bildpunkt behält seine Farbe.**
//
// # Warum das für die Blenden zählt
//
// MSE hält RGB und Alpha getrennt (`wxImage`); `mask_blend` und `linear_blend`
// laufen über `GetData()` und sehen **nur RGB**, unabhängig vom Alpha. Ein
// vollständig durchsichtiger Punkt trägt dort trotzdem eine Farbe.
//
// Der Rasterer las bis dahin über `punktfarbe`, das an solchen Stellen schwarz
// gibt — richtig fürs Überlegen, falsch fürs Mischen: an Kanten mit
// durchsichtigen Rändern entstünden Farbsäume. Angestrichen vom Zweitblick über
// die Phasen B bis D.
func TestEinDurchsichtigerBildpunktBehaeltSeineFarbe(t *testing.T) {
	bild := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	bild.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 0})

	rasterer := render.NeuerRasterer(einBild{bild: bild})

	// Gemischt gegen sich selbst mit weißer Maske: heraus kommt `light`.
	gemischt, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "masked_blend", Kinder: []render.Bildkind{
			{Name: "light", Baum: render.Blatt("/a.png")},
			{Name: "dark", Baum: render.Blatt("/a.png")},
			{Name: "mask", Baum: render.Blatt("/a.png")},
		},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	// **Das Alpha bleibt null**, und die Farbe darunter ist nicht schwarz
	// geworden.
	farbe, passt := gemischt.(*image.NRGBA)
	if !passt {
		t.Fatalf("das Ergebnis ist %T, erwartet *image.NRGBA", gemischt)
	}

	stelle := farbe.PixOffset(0, 0)
	if farbe.Pix[stelle] == 0 && farbe.Pix[stelle+1] == 0 && farbe.Pix[stelle+2] == 0 {
		t.Error("der durchsichtige Bildpunkt ist schwarz geworden")
	}
}

// **Ein vorgewichtetes Bild gibt seine Farbe bei Alpha null nicht mehr her.**
//
// Das steht hier als Zusage über die **Grenze**: `image.RGBA` speichert
// vorgewichtet, dort ist die Farbe verloren. Die Pakete liefern PNG und JPEG,
// und `image/png` gibt für Bilder mit Alpha ein `*image.NRGBA` — der Fall, auf
// den es ankommt, ist gedeckt.
func TestEinVorgewichtetesBildVerliertSeineFarbe(t *testing.T) {
	bild := image.NewRGBA(image.Rect(0, 0, 2, 2))
	bild.SetRGBA(0, 0, color.RGBA{})

	rasterer := render.NeuerRasterer(einBild{bild: bild})

	gemischt, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "masked_blend", Kinder: []render.Bildkind{
			{Name: "light", Baum: render.Blatt("/a.png")},
			{Name: "dark", Baum: render.Blatt("/a.png")},
			{Name: "mask", Baum: render.Blatt("/a.png")},
		},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if r, _, _, _ := gemischt.At(0, 0).RGBA(); r != 0 {
		t.Errorf("Rot ist %d, erwartet 0 — bei einem vorgewichteten Bild ist die "+
			"Farbe bei Alpha null verloren", r>>8)
	}
}

// einBild gibt zu jedem Pfad dasselbe Bild.
type einBild struct{ bild image.Image }

func (e einBild) Bild(string) (image.Image, error) { return e.bild, nil }

// **Auch ein palettiertes und ein 16-Bit-Bild behalten ihre Farbe.**
//
// `image.Decode` gibt für ein palettiertes PNG mit Alpha ein `*image.Paletted`
// zurück, dessen Palette `color.NRGBA` führt — und das Pack enthält solche
// Bilder. Der erste Anlauf prüfte auf `*image.NRGBA` und fiel für alles andere
// zurück; angestrichen vom Nachlauf des Zweitblicks: „nur das Beispiel
// geschlossen".
func TestAuchPalettierteUndTiefeBilderBehaltenIhreFarbe(t *testing.T) {
	paletted := image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{
		color.NRGBA{R: 200, G: 100, B: 50, A: 0},
	})

	tief := image.NewNRGBA64(image.Rect(0, 0, 2, 2))
	tief.SetNRGBA64(0, 0, color.NRGBA64{R: 200 << 8, G: 100 << 8, B: 50 << 8})

	for name, bild := range map[string]image.Image{"palettiert": paletted, "16 Bit": tief} {
		t.Run(name, func(t *testing.T) {
			rasterer := render.NeuerRasterer(einBild{bild: bild})

			gemischt, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
				Knotenart: "masked_blend", Kinder: []render.Bildkind{
					{Name: "light", Baum: render.Blatt("/a.png")},
					{Name: "dark", Baum: render.Blatt("/a.png")},
					{Name: "mask", Baum: render.Blatt("/a.png")},
				},
			})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			farbe, passt := gemischt.(*image.NRGBA)
			if !passt {
				t.Fatalf("das Ergebnis ist %T", gemischt)
			}

			stelle := farbe.PixOffset(0, 0)
			if farbe.Pix[stelle] == 0 {
				t.Error("der durchsichtige Bildpunkt ist schwarz geworden")
			}
		})
	}
}
