package render_test

// `enlarge` legt einen durchsichtigen Rand um ein Bild
// (#472).
//
//	int dw = int(w * border_size), dh = int(h * border_size);
//	int w2 = w + dw + dw,     h2 = h + dh + dh;
//	Image larger(w2,h2);
//	larger.InitAlpha();
//	memset(larger.GetAlpha(),0,w2*h2); // blank
//
// `src/gfx/generated_image.cpp`, Zeilen 281–286, Fassung f9d9356 vom
// 2021-07-06.
//
// **37 Aufrufe im vollen Pack**, 35 in `M15-Magic-Pack`.

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func vergroessert(anteil float64, kind render.Bildbaum) render.Bildbaum {
	return render.Bildbaum{
		Knotenart: "enlarge",
		Zahlen:    map[string]float64{"border_size": anteil},
		Kinder:    []render.Bildkind{{Name: "input", Baum: kind}},
	}
}

func TestEnlargeLegtEinenRandUmDasBild(t *testing.T) {
	// Ein Bild von 10 × 10, voll deckend.
	quelle := maskenquelle{bilder: map[string]image.Image{
		"bild.png": maskenflaeche(10, 10, color.NRGBA{R: 200, G: 100, B: 50, A: 255}),
	}}

	t.Run("ein Zehntel Rand", func(t *testing.T) {
		plan := &render.Plan{Breite: 12, Hoehe: 12, Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Rahmen:  render.Rechteck{Breite: 12, Hoehe: 12},
			Bild:    vergroessert(0.1, render.Blatt("bild.png")),
			Deckung: 1,
		}}}

		bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 0 {
			t.Fatalf("Befunde: %v", befunde)
		}

		// Aus 10 × 10 wird 12 × 12, und die Ebene ist genau 12 × 12 groß —
		// der Rand liegt also außen und ist durchsichtig.
		if farbe := bild.NRGBAAt(0, 0); farbe.A != 0 {
			t.Errorf("die Ecke ist %v, erwartet durchsichtig", farbe)
		}

		if farbe := bild.NRGBAAt(6, 6); farbe.R != 200 || farbe.A != 255 {
			t.Errorf("die Mitte ist %v, erwartet das Bild", farbe)
		}
	})

	t.Run("ein zu kleiner Rand fällt weg", func(t *testing.T) {
		// **Abgeschnitten, nicht gerundet**: `int(10 * 0.04)` ist null.
		plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Rahmen:  render.Rechteck{Breite: 10, Hoehe: 10},
			Bild:    vergroessert(0.04, render.Blatt("bild.png")),
			Deckung: 1,
		}}}

		bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 0 {
			t.Fatalf("Befunde: %v", befunde)
		}

		if farbe := bild.NRGBAAt(0, 0); farbe.A != 255 {
			t.Errorf("die Ecke ist %v, erwartet das Bild ohne Rand", farbe)
		}
	})

	t.Run("ohne border_size ein Befund", func(t *testing.T) {
		plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{{
			Art:    render.ArtBild,
			Rahmen: render.Rechteck{Breite: 10, Hoehe: 10},
			Bild: render.Bildbaum{
				Knotenart: "enlarge",
				Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("bild.png")}},
			},
			Deckung: 1,
		}}}

		_, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 1 {
			t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
		}
	})

	t.Run("ein fehlendes Kind", func(t *testing.T) {
		plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{{
			Art:    render.ArtBild,
			Rahmen: render.Rechteck{Breite: 10, Hoehe: 10},
			Bild: render.Bildbaum{
				Knotenart: "enlarge",
				Zahlen:    map[string]float64{"border_size": 0.1},
				Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("fehlt.png")}},
			},
			Deckung: 1,
		}}}

		_, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 1 {
			t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
		}
	})
}

// **Die Ränder von `drop_shadow`.**
func TestDropShadowAnDenRaendern(t *testing.T) {
	quelle := maskenquelle{bilder: map[string]image.Image{
		"bild.png": maskenflaeche(8, 8, color.NRGBA{R: 200, A: 128}),
	}}

	schatten := func(zahlen map[string]float64, kind render.Bildbaum) render.Bildbaum {
		return render.Bildbaum{
			Knotenart: "drop_shadow", Zahlen: zahlen,
			Kinder: []render.Bildkind{{Name: "input", Baum: kind}},
		}
	}

	t.Run("ohne blur_radius geschieht nichts", func(t *testing.T) {
		// `sigma` wäre null, und die Rechnung teilte durch null.
		plan := &render.Plan{Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
			Art: render.ArtBild, Rahmen: render.Rechteck{Breite: 8, Hoehe: 8},
			Bild:    schatten(map[string]float64{"alpha": 1}, render.Blatt("bild.png")),
			Deckung: 1,
		}}}

		_, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 0 {
			t.Errorf("Befunde: %v", befunde)
		}
	})

	t.Run("ohne color ist der Schatten schwarz", func(t *testing.T) {
		// **Die Vorgabe steht im Code, nicht im Aufruf**: MSEs
		// `SCRIPT_OPTIONAL_PARAM_(Color, color)` gibt eine Nullfarbe.
		plan := &render.Plan{Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
			Art: render.ArtBild, Rahmen: render.Rechteck{Breite: 8, Hoehe: 8},
			Bild: schatten(map[string]float64{
				"alpha": 1, "blur_radius": 0.2, "offset_x": 0.25,
			}, render.Blatt("bild.png")),
			Deckung: 1,
		}}}

		_, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 0 {
			t.Errorf("Befunde: %v", befunde)
		}
	})

	t.Run("ein fehlendes Kind", func(t *testing.T) {
		plan := &render.Plan{Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
			Art: render.ArtBild, Rahmen: render.Rechteck{Breite: 8, Hoehe: 8},
			Bild: schatten(map[string]float64{"alpha": 1, "blur_radius": 0.2},
				render.Blatt("fehlt.png")),
			Deckung: 1,
		}}}

		_, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 1 {
			t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
		}
	})
}

// **Ein Bild ohne Fläche wirft keinen Schatten.**
//
// Es kommt vor: `crop` mit einer Breite von null gibt eines, und ein Paket von
// 2007 ist fremde Eingabe (§8).
func TestDropShadowAufEinemBildOhneFlaeche(t *testing.T) {
	quelle := maskenquelle{bilder: map[string]image.Image{
		"leer.png": image.NewNRGBA(image.Rect(0, 0, 0, 0)),
	}}

	plan := &render.Plan{Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
		Art:    render.ArtBild,
		Rahmen: render.Rechteck{Breite: 8, Hoehe: 8},
		Bild: render.Bildbaum{
			Knotenart: "drop_shadow",
			Zahlen:    map[string]float64{"alpha": 1, "blur_radius": 0.2},
			Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("leer.png")}},
		},
		Deckung: 1,
	}}}

	if _, _, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel); err != nil {
		t.Fatalf("zeichnen: %v", err)
	}
}

// **Ein leeres Bild ist durchsichtig**
// (#509).
//
//	Image BlankImage::generate(const Options& opt) const {
//	  …
//	  img.InitAlpha();
//	  memset(img.GetAlpha(), 0, w * h);
//	  return img;
//	}
//
// `src/gfx/generated_image.cpp`, Zeilen 83–91, Fassung f9d9356 vom
// 2021-07-06. Es entsteht aus der leeren Zeichenkette und aus `nil`, wo eine
// Bildfunktion ein Bild verlangt.
func TestEinLeeresBildIstDurchsichtig(t *testing.T) {
	quelle := maskenquelle{bilder: map[string]image.Image{
		"b.png": maskenflaeche(4, 4, color.NRGBA{R: 200, A: 255}),
	}}

	t.Run("als ganze Ebene", func(t *testing.T) {
		plan := &render.Plan{Breite: 4, Hoehe: 4, Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Rahmen:  render.Rechteck{Breite: 4, Hoehe: 4},
			Bild:    render.Bildbaum{Knotenart: "blank_image"},
			Deckung: 1,
		}}}

		bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 0 {
			t.Fatalf("Befunde: %v", befunde)
		}

		for y := range 4 {
			for x := range 4 {
				if bild.NRGBAAt(x, y).A != 0 {
					t.Fatalf("(%d,%d) ist nicht durchsichtig", x, y)
				}
			}
		}
	})

	// **Und es hat die gefragte Größe** — sonst meldete `linear_blend`
	// „die Bilder sind verschieden groß (1x1 gegen 4x4)".
	t.Run("in einer Blende", func(t *testing.T) {
		plan := &render.Plan{Breite: 4, Hoehe: 4, Ebenen: []render.Ebene{{
			Art:    render.ArtBild,
			Rahmen: render.Rechteck{Breite: 4, Hoehe: 4},
			Bild: render.Bildbaum{
				Knotenart: "linear_blend",
				Zahlen:    map[string]float64{"x1": 0, "y1": 0, "x2": 1, "y2": 0},
				Kinder: []render.Bildkind{
					{Name: "image1", Baum: render.Bildbaum{Knotenart: "blank_image"}},
					{Name: "image2", Baum: render.Blatt("b.png")},
				},
			},
			Deckung: 1,
		}}}

		_, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		if len(befunde) != 0 {
			t.Errorf("Befunde: %v — erwartet keine", befunde)
		}
	})
}
