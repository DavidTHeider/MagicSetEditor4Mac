package render_test

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`combine_blend` mischt zwei Bilder in einer von 23 Arten.**
//
//	COMBINE_FUN(COMBINE_ADD,         top(a + b))
//	COMBINE_FUN(COMBINE_MULTIPLY,    (a * b) / 255)
//	COMBINE_FUN(COMBINE_SCREEN,      255 - (((255 - a) * (255 - b)) / 255))
//	COMBINE_FUN(COMBINE_STAMP,       col(a - 2 * b + 256))
//	…
//
// `src/gfx/combine_image.cpp`, Zeilen 57–85, Fassung f9d9356 vom 2021-07-06 —
// „Based on http://www.pegtop.net/delphi/articles/blendmodes/", so steht es dort.
//
// **`soft light` gibt `b` zurück.** Das ist keine Auslassung dieses Programms,
// sondern MSEs Zeile: `COMBINE_FUN(COMBINE_SOFT_LIGHT, b)` — dieselbe Formel wie
// `normal` (#270).
func TestCombineBlendMischtInDreiundzwanzigArten(t *testing.T) {
	// **a = 200, b = 100**, und jede Zahl ist aus MSEs Zeile gerechnet:
	// `negation` ist 255 − |255 − 200 − 100| = 210, `screen`
	// 255 − (55·155)/255 = 222, `overlay` 255 − ((55·155) >> 7) = 189.
	faelle := []struct {
		art      string
		erwartet uint8
	}{
		{"normal", 100},
		{"add", 255},
		{"subtract", 100},
		{"stamp", 255},
		{"difference", 100},
		{"negation", 210},
		{"multiply", 78},
		{"darken", 100},
		{"lighten", 200},
		{"screen", 222},
		{"overlay", 189},
		{"hard light", 156},
		{"soft light", 100},
		{"and", 64},
		{"or", 236},
		{"xor", 172},
		{"shadow", 61},
	}

	for _, fall := range faelle {
		t.Run(fall.art, func(t *testing.T) {
			bild, befund := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
				"/a.png": {R: 200, G: 200, B: 200, A: 255},
				"/b.png": {R: 100, G: 100, B: 100, A: 255},
			}}).AusBaumFuerTest(render.Bildbaum{
				Knotenart: "combine_blend",
				Kinder: []render.Bildkind{
					{Name: "image1", Baum: render.Blatt("/a.png")},
					{Name: "image2", Baum: render.Blatt("/b.png")},
				},
				Texte: map[string]string{"combine": fall.art},
			})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			if rot, _, _, _ := bild.At(0, 0).RGBA(); rot>>8 != uint32(fall.erwartet) {
				t.Errorf("%s ergibt %d, erwartet %d", fall.art, rot>>8, fall.erwartet)
			}
		})
	}
}

// **Eine unbekannte Mischart ist ein Befund.**
//
// MSEs `parse_enum` meldet sie; eine stillschweigend als `normal` behandelte
// Angabe verdeckte einen Tippfehler in der Vorlage (§8).
func TestEineUnbekannteMischartIstEinBefund(t *testing.T) {
	_, befund := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/a.png": {R: 200, A: 255}, "/b.png": {R: 100, A: 255},
	}}).AusBaumFuerTest(render.Bildbaum{
		Knotenart: "combine_blend",
		Kinder: []render.Bildkind{
			{Name: "image1", Baum: render.Blatt("/a.png")},
			{Name: "image2", Baum: render.Blatt("/b.png")},
		},
		Texte: map[string]string{"combine": "gibtsnicht"},
	})

	if !strings.Contains(befund, "gibtsnicht") {
		t.Errorf("der Befund lautet %q und nennt die Angabe nicht", befund)
	}
}

// **`set_mask` nimmt das Alpha aus dem **roten** Kanal der Maske.**
//
//	void set_alpha(Image& img, const Image& img_alpha) {
//	  …
//	  im[i] = (im[i] * al[i*3]) / 255;
//	}
//
// `src/gfx/blend_image.cpp`, Zeilen 79–87, Fassung f9d9356 vom 2021-07-06.
// `al[i*3]` ist der **rote** Kanal — nicht die Helligkeit. Bei einer grauen
// Maske ist beides gleich, bei einer farbigen nicht.
func TestSetMaskNimmtDasAlphaAusDemRotenKanal(t *testing.T) {
	bild, befund := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/bild.png": {R: 200, G: 200, B: 200, A: 255},

		// **Rot halb, grün ganz** — die Helligkeit wäre eine andere Zahl.
		"/maske.png": {R: 128, G: 255, B: 0, A: 255},
	}}).AusBaumFuerTest(render.Bildbaum{
		Knotenart: "set_mask",
		Kinder: []render.Bildkind{
			{Name: "image", Baum: render.Blatt("/bild.png")},
			{Name: "mask", Baum: render.Blatt("/maske.png")},
		},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if _, _, _, alpha := bild.At(0, 0).RGBA(); alpha>>8 != 128 {
		t.Errorf("das Alpha ist %d, erwartet 128", alpha>>8)
	}
}

// **Die sechs Mischarten mit einer Division prüfen ihren Nenner.**
//
//	COMBINE_FUN(COMBINE_COLOR_DODGE, b == 255 ? 255 : top(a * 255 / (255 - b)))
//	COMBINE_FUN(COMBINE_COLOR_BURN,  b == 0   ? 0   : bot(255 - (255-a) * 255 / b))
//	COMBINE_FUN(COMBINE_REFLECT,     b == 255 ? 255 : top(a * a / (255 - b)))
//	COMBINE_FUN(COMBINE_GLOW,        a == 255 ? 255 : top(b * b / (255 - a)))
//	COMBINE_FUN(COMBINE_FREEZE,      b == 0 ? 0 : bot(255 - (255-a) * (255-a) / b))
//	COMBINE_FUN(COMBINE_HEAT,        a == 0 ? 0 : bot(255 - (255-b) * (255-b) / a))
//
// `src/gfx/combine_image.cpp`, Zeilen 68–83, Fassung f9d9356 vom 2021-07-06.
// **Jede prüft, und jede antwortet anders** — `color dodge` gibt 255,
// `color burn` null. Ein gemeinsamer Sonderfall wäre bei der Hälfte falsch.
func TestDieMischartenMitDivisionPruefenIhrenNenner(t *testing.T) {
	faelle := []struct {
		art      string
		a, b     uint8
		erwartet uint8
	}{
		{"color dodge", 200, 255, 255},
		{"color dodge", 100, 0, 100},
		{"color burn", 200, 0, 0},
		{"color burn", 200, 255, 200},
		{"reflect", 200, 255, 255},
		{"reflect", 100, 0, 39},
		{"glow", 255, 100, 255},
		{"glow", 0, 100, 39},
		{"freeze", 200, 0, 0},
		{"freeze", 200, 255, 244},
		{"heat", 0, 200, 0},
		{"heat", 255, 200, 244},
		{"symmetric overlay", 200, 100, 172},
	}

	for _, fall := range faelle {
		t.Run(fall.art, func(t *testing.T) {
			bild, befund := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
				"/a.png": {R: fall.a, G: fall.a, B: fall.a, A: 255},
				"/b.png": {R: fall.b, G: fall.b, B: fall.b, A: 255},
			}}).AusBaumFuerTest(render.Bildbaum{
				Knotenart: "combine_blend",
				Kinder: []render.Bildkind{
					{Name: "image1", Baum: render.Blatt("/a.png")},
					{Name: "image2", Baum: render.Blatt("/b.png")},
				},
				Texte: map[string]string{"combine": fall.art},
			})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			if rot, _, _, _ := bild.At(0, 0).RGBA(); rot>>8 != uint32(fall.erwartet) {
				t.Errorf("%s(%d,%d) ergibt %d, erwartet %d",
					fall.art, fall.a, fall.b, rot>>8, fall.erwartet)
			}
		})
	}
}

// **Ohne Angabe gilt `normal`**, und was fehlt oder nicht passt, wird benannt.
func TestDieZweibildverknuepfungenMeldenIhreGrenzfaelle(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/a.png": {R: 200, A: 255}, "/b.png": {R: 100, A: 255},
	}})

	zwei := func(art string, kinder ...render.Bildkind) render.Bildbaum {
		return render.Bildbaum{Knotenart: art, Kinder: kinder}
	}

	// **Ohne `combine:` gilt `normal`** — MSEs `COMBINE_DEFAULT`.
	bild, befund := rasterer.AusBaumFuerTest(zwei("combine_blend",
		render.Bildkind{Name: "image1", Baum: render.Blatt("/a.png")},
		render.Bildkind{Name: "image2", Baum: render.Blatt("/b.png")}))
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if rot, _, _, _ := bild.At(0, 0).RGBA(); rot>>8 != 100 {
		t.Errorf("ohne Angabe ergibt es %d, erwartet das obere Bild", rot>>8)
	}

	faelle := []struct {
		name  string
		baum  render.Bildbaum
		nennt string
	}{
		{
			name: "combine_blend ohne zweites Bild",
			baum: zwei("combine_blend",
				render.Bildkind{Name: "image1", Baum: render.Blatt("/a.png")}),
			nennt: "image2",
		},
		{
			name: "set_mask ohne Maske",
			baum: zwei("set_mask",
				render.Bildkind{Name: "image", Baum: render.Blatt("/a.png")}),
			nennt: "mask",
		},
		{
			name: "combine_blend mit unlesbarem erstem Bild",
			baum: zwei("combine_blend",
				render.Bildkind{Name: "image1", Baum: render.Blatt("/fehlt.png")},
				render.Bildkind{Name: "image2", Baum: render.Blatt("/b.png")}),
			nennt: "/fehlt.png",
		},
		{
			name: "combine_blend mit unlesbarem zweitem Bild",
			baum: zwei("combine_blend",
				render.Bildkind{Name: "image1", Baum: render.Blatt("/a.png")},
				render.Bildkind{Name: "image2", Baum: render.Blatt("/fehlt.png")}),
			nennt: "/fehlt.png",
		},
		{
			name: "set_mask mit unlesbarem Bild",
			baum: zwei("set_mask",
				render.Bildkind{Name: "image", Baum: render.Blatt("/fehlt.png")},
				render.Bildkind{Name: "mask", Baum: render.Blatt("/b.png")}),
			nennt: "/fehlt.png",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, befund := rasterer.AusBaumFuerTest(fall.baum)
			if !strings.Contains(befund, fall.nennt) {
				t.Errorf("der Befund lautet %q und nennt %q nicht", befund, fall.nennt)
			}
		})
	}
}

// **Zwei Bilder verschiedener Größe sind ein Befund.**
//
//	assert(a.GetWidth()  == b.GetWidth());
//	assert(a.GetHeight() == b.GetHeight());
//
// `combine_image`, `src/gfx/combine_image.cpp`, Zeilen 103–105, Fassung f9d9356
// vom 2021-07-06. MSE **behauptet** die gleiche Größe; hier wird sie geprüft,
// denn eine Behauptung über fremde Daten ist keine (§8).
//
// **`set_mask` prüft das nicht** — dort skaliert MSE die Maske
// (`resample(img_alpha, …)`), und deshalb tut es das hier auch.
func TestZweiBilderVerschiedenerGroesseSindEinBefund(t *testing.T) {
	rasterer := render.NeuerRasterer(zweigroessen{})

	_, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "combine_blend",
		Kinder: []render.Bildkind{
			{Name: "image1", Baum: render.Blatt("/gross.png")},
			{Name: "image2", Baum: render.Blatt("/klein.png")},
		},
		Texte: map[string]string{"combine": "add"},
	})

	if !strings.Contains(befund, "Größe") {
		t.Errorf("der Befund lautet %q", befund)
	}

	// **Und `set_mask` nimmt sie trotzdem.**
	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "set_mask",
		Kinder: []render.Bildkind{
			{Name: "image", Baum: render.Blatt("/gross.png")},
			{Name: "mask", Baum: render.Blatt("/klein.png")},
		},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if bild.Bounds().Dx() != 8 {
		t.Errorf("das Bild ist %d breit, erwartet die Größe des Bildes", bild.Bounds().Dx())
	}
}

// zweigroessen gibt zwei Bilder verschiedener Größe.
type zweigroessen struct{}

func (zweigroessen) Bild(pfad string) (image.Image, error) {
	seite := 8
	if pfad == "/klein.png" {
		seite = 4
	}

	bild := image.NewNRGBA(image.Rect(0, 0, seite, seite))
	for y := range seite {
		for x := range seite {
			bild.SetNRGBA(x, y, color.NRGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	return bild, nil
}
