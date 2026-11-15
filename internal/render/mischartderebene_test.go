package render_test

// **Die Mischart einer Ebene kommt aus ihrem Bildbaum**
// (#470).
//
// `set_combine` setzt sie; dieses Programm reichte das Bild durch und warf sie
// weg. Über `/Users/david/full-magic-pack/data` steht der Aufruf **698-mal**
// in 41 Textdateien aus 34 Paketen — 667-mal mit `shadow`, 17-mal mit
// `multiply`.
//
// # Die Regel ist ein reiner Baumdurchlauf
//
// MSE fragt sie am Baum, ohne ein Bild zu erzeugen
// (`src/gfx/generated_image.hpp` und `.cpp`, Fassung f9d9356 vom 2021-07-06):
//
//	GeneratedImage::combine()      → COMBINE_DEFAULT   (hpp:51)
//	SimpleFilterImage::combine()   → image->combine()  (hpp:79)
//	LinearBlendImage::combine()    → image1->combine() (cpp:104)
//	MaskedBlendImage::combine()    → light->combine()  (cpp:122)
//	CombineBlendImage::combine()   → image1->combine() (cpp:139)
//	SetCombineImage::combine()     → seine eigene      (cpp:178)
//
// `SimpleFilterImage` ist die Oberklasse von `set_mask`, `set_alpha`,
// `set_combine`, `saturate`, `invert_image`, `recolor_image`, den beiden
// Spiegelungen, `rotate`, `enlarge`, `crop` und `drop_shadow` (hpp:153–308).

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func gesetzt(art string, kind render.Bildbaum) render.Bildbaum {
	return render.Bildbaum{
		Knotenart: "set_combine",
		Texte:     map[string]string{"combine": art},
		Kinder:    []render.Bildkind{{Name: "input", Baum: kind}},
	}
}

func TestDieMischartKommtAusDemBildbaum(t *testing.T) {
	blatt := render.Blatt("rahmen.png")

	faelle := []struct {
		name     string
		baum     render.Bildbaum
		erwartet string
	}{
		{"ein Blatt hat keine", blatt, ""},
		{"leer hat keine", render.Bildbaum{}, ""},
		{"set_combine setzt sie", gesetzt("shadow", blatt), "shadow"},

		{
			// `SimpleFilterImage::combine()` reicht durch (hpp:79).
			name: "ein einfacher Filter reicht sie durch",
			baum: render.Bildbaum{
				Knotenart: "recolor_image",
				Kinder: []render.Bildkind{
					{Name: "input", Baum: gesetzt("multiply", blatt)},
				},
			},
			erwartet: "multiply",
		},
		{
			// `MaskedBlendImage::combine()` nimmt **`light`** (cpp:122).
			name: "masked_blend nimmt die helle Seite",
			baum: render.Bildbaum{
				Knotenart: "masked_blend",
				Kinder: []render.Bildkind{
					{Name: "light", Baum: gesetzt("glow", blatt)},
					{Name: "dark", Baum: gesetzt("multiply", blatt)},
					{Name: "mask", Baum: blatt},
				},
			},
			erwartet: "glow",
		},
		{
			// **Und `dark` allein zählt nicht** — die Gegenprobe dazu.
			name: "masked_blend übergeht die dunkle Seite",
			baum: render.Bildbaum{
				Knotenart: "masked_blend",
				Kinder: []render.Bildkind{
					{Name: "light", Baum: blatt},
					{Name: "dark", Baum: gesetzt("multiply", blatt)},
					{Name: "mask", Baum: blatt},
				},
			},
			erwartet: "",
		},
		{
			// `LinearBlendImage::combine()` nimmt `image1` (cpp:104).
			name: "linear_blend nimmt das erste Bild",
			baum: render.Bildbaum{
				Knotenart: "linear_blend",
				Kinder: []render.Bildkind{
					{Name: "image1", Baum: gesetzt("shadow", blatt)},
					{Name: "image2", Baum: blatt},
				},
			},
			erwartet: "shadow",
		},
		{
			// `CombineBlendImage::combine()` ebenso (cpp:139).
			name: "combine_blend nimmt das erste Bild",
			baum: render.Bildbaum{
				Knotenart: "combine_blend",
				Kinder: []render.Bildkind{
					{Name: "image1", Baum: gesetzt("add", blatt)},
					{Name: "image2", Baum: blatt},
				},
			},
			erwartet: "add",
		},
		{
			// **Ein Knoten, den MSE nicht durchreicht**: `symbol_variation`
			// erbt `GeneratedImage::combine()` und gibt die Vorgabe.
			name: "ein anderer Knoten gibt die Vorgabe",
			baum: render.Bildbaum{
				Knotenart: "symbol_variation",
				Kinder: []render.Bildkind{
					{Name: "symbol", Baum: gesetzt("shadow", blatt)},
				},
			},
			erwartet: "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if art := fall.baum.Mischart(); art != fall.erwartet {
				t.Errorf("Mischart = %q, erwartet %q", art, fall.erwartet)
			}
		})
	}
}

// **Die Mischart wirkt beim Auflegen**
// (#470).
//
// MSE nimmt bei `combine > NORMAL` den Zielausschnitt auf, rechnet je Kanal
// `f(Ziel, Quelle)` und legt das Ergebnis mit dem **Alpha der Quelle** auf:
//
//	void draw_combine_image(DC& dc, UInt x, UInt y, const Image& img, ImageCombine combine) {
//	  if (combine <= COMBINE_NORMAL) {
//	    dc.DrawBitmap(img, x, y);
//	  } else {
//	    …sourceDC.Blit(0, 0, …, &dc, x, y);
//	    Image source = sourceB.ConvertToImage();
//	    combine_image(source, img, combine);
//	    dc.DrawBitmap(source, x, y);
//	  }
//	}
//
// `src/gfx/combine_image.cpp`, Zeilen 142–157, Fassung f9d9356 vom
// 2021-07-06. `combine_image` kopiert dabei zuerst den Alphakanal von `b`
// nach `a` (ebd. Zeilen 108–112) — deshalb das Alpha der **Quelle**.
type festesbild struct{ farbe color.NRGBA }

func (f festesbild) Bild(string) (image.Image, error) {
	bild := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			bild.SetNRGBA(x, y, f.farbe)
		}
	}

	return bild, nil
}

func TestDieMischartWirktBeimAuflegen(t *testing.T) {
	// Unten liegt eine Fläche, oben ein Bild derselben Größe.
	const unten, oben = 200, 100

	faelle := []struct {
		name     string
		art      string
		erwartet uint8
	}{
		// Ohne Mischart ersetzt das Bild.
		{"ohne", "", oben},

		// `(a * b) / 255` = 200 · 100 / 255 = 78
		{"multiply", "multiply", 78},

		// `(b * a * a) / (255 * 255)` = 100 · 200 · 200 / 65025 = 61
		{"shadow", "shadow", 61},

		// `add` rechnet `min(255, a + b)`, also 255.
		{"add", "add", 255},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			bildbaum := render.Blatt("oben.png")
			if fall.art != "" {
				bildbaum = gesetzt(fall.art, bildbaum)
			}

			plan := &render.Plan{Breite: 4, Hoehe: 4, Ebenen: []render.Ebene{
				{
					Art:     render.ArtFlaeche,
					Rahmen:  render.Rechteck{Breite: 4, Hoehe: 4},
					Farbe:   render.Farbe{R: unten, G: unten, B: unten, A: 255},
					Deckung: 1,
				},
				{
					Art:      render.ArtBild,
					Rahmen:   render.Rechteck{Breite: 4, Hoehe: 4},
					Bild:     bildbaum,
					Mischart: fall.art,
					Deckung:  1,
				},
			}}

			bild, befunde, err := render.NeuerRasterer(
				festesbild{color.NRGBA{R: oben, G: oben, B: oben, A: 255}}).
				Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if len(befunde) != 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			ergebnis := bild.NRGBAAt(2, 2)
			if ergebnis.R != fall.erwartet {
				t.Errorf("R = %d, erwartet %d (die ganze Farbe: %v)",
					ergebnis.R, fall.erwartet, ergebnis)
			}
		})
	}
}

// **Eine unbekannte Mischart wird benannt und normal gezeichnet.**
//
// Sichtbar falsch schlägt unsichtbar fehlend (§8): das Bild kommt an, und der
// Befund reist mit.
func TestEineUnbekannteMischartWirdBenannt(t *testing.T) {
	plan := &render.Plan{Breite: 4, Hoehe: 4, Ebenen: []render.Ebene{{
		Art:      render.ArtBild,
		Rahmen:   render.Rechteck{Breite: 4, Hoehe: 4},
		Bild:     gesetzt("gibtesnicht", render.Blatt("oben.png")),
		Mischart: "gibtesnicht",
		Deckung:  1,
	}}}

	bild, befunde, err := render.NeuerRasterer(
		festesbild{color.NRGBA{R: 100, G: 100, B: 100, A: 255}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 ||
		!strings.Contains(befunde[0].Meldung, "gibtesnicht") {
		t.Fatalf("Befunde = %v, erwartet einer, der die Mischart nennt", befunde)
	}

	// **Und sie verdrängt einen Befund aus dem Baum nicht.** `recolor_image`
	// ohne Farben gibt sein Ausgangsbild samt Meldung; die Mischart stellt
	// ihre daneben, statt sie zu überschreiben — **als eigenen Befund** seit
	// QA-03 (#1656).
	mitBeidem := &render.Plan{Breite: 4, Hoehe: 4, Ebenen: []render.Ebene{{
		Art:    render.ArtBild,
		Rahmen: render.Rechteck{Breite: 4, Hoehe: 4},
		Bild: gesetzt("gibtesnicht", render.Bildbaum{
			Knotenart: "recolor_image",
			Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("oben.png")}},
		}),
		Mischart: "gibtesnicht",
		Deckung:  1,
	}}}

	_, zwei, err := render.NeuerRasterer(
		festesbild{color.NRGBA{R: 100, G: 100, B: 100, A: 255}}).
		Zeichne(t.Context(), mitBeidem, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(zwei) != 2 || !strings.Contains(zwei[0].Meldung, "recolor_image") ||
		!strings.Contains(zwei[1].Meldung, "gibtesnicht") {
		t.Errorf("Befunde = %v, erwartet zwei: den aus dem Baum, dann die Mischart", zwei)
	}

	// **Und gezeichnet ist trotzdem worden.**
	if got := bild.NRGBAAt(2, 2).R; got != 100 {
		t.Errorf("R = %d, erwartet 100 — das Bild ist nicht angekommen", got)
	}
}

// **Die Grenzfälle des Baumdurchlaufs.**
func TestDieMischartAnDenRaendern(t *testing.T) {
	blatt := render.Blatt("rahmen.png")

	t.Run("set_mask fragt sein Kind `image`", func(t *testing.T) {
		baum := render.Bildbaum{
			Knotenart: "set_mask",
			Kinder: []render.Bildkind{
				{Name: "image", Baum: gesetzt("shadow", blatt)},
				{Name: "mask", Baum: blatt},
			},
		}

		if art := baum.Mischart(); art != "shadow" {
			t.Errorf("Mischart = %q, erwartet \"shadow\"", art)
		}
	})

	t.Run("ein fehlendes Kind gibt nichts", func(t *testing.T) {
		baum := render.Bildbaum{
			Knotenart: "recolor_image",
			Kinder:    []render.Bildkind{{Name: "mask", Baum: blatt}},
		}

		if art := baum.Mischart(); art != "" {
			t.Errorf("Mischart = %q, erwartet leer", art)
		}
	})

	t.Run("ein Ring endet an der Tiefengrenze", func(t *testing.T) {
		// **Derselbe Vertrag wie bei [Bildbaum.Gebaut]**: ein Baum aus einem
		// fremden Paket darf tief sein, und ein Ring sähe genauso aus.
		tief := gesetzt("shadow", blatt)
		for range render.Baumtiefe + 2 {
			tief = render.Bildbaum{
				Knotenart: "recolor_image",
				Kinder:    []render.Bildkind{{Name: "input", Baum: tief}},
			}
		}

		if art := tief.Mischart(); art != "" {
			t.Errorf("Mischart = %q, erwartet leer — die Grenze greift nicht", art)
		}
	})
}
