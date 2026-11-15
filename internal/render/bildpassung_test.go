package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Wahlbild wird eingepasst, nicht gestreckt.**
//
//	} else if (options.preserve_aspect == ASPECT_FIT) {
//	  int w, h;
//	  if (iw * options.height > ih * options.width) { // too much height requested
//	    w = options.width;  h = options.width * ih / iw;
//	  } else {
//	    w = options.height * iw / ih; h = options.height;
//	  }
//	  image = resample(image, w, h);
//	}
//
// `conform_image`, `src/gfx/generated_image.cpp`, Zeilen 49–59, und danach
//
//	RealRect rect(align_in_rect(style.alignment, size, dc.getInternalRect()), size);
//
// `draw_choice_viewer`, `src/render/value/choice.cpp`, Zeile 83 — Fassung
// f9d9356 vom 2021-07-06.
//
// **Daran hängt das Seltenheitssymbol.** `magic-modules.mse-include/rarities`
// gibt dem Feld `alignment: middle right` und `render style: image`; gestreckt
// wurde daraus ein breiter Balken quer über die Typzeile
// (#270).
func TestEinBildWirdEingepasstUndAusgerichtet(t *testing.T) {
	// Ein quadratisches Bild in einem doppelt so breiten Feld.
	quelle := bilder{"/probe.mse-style/quadrat.png": vollflaeche(20, 20)}

	plan := &render.Plan{Breite: 40, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Breite: 40, Hoehe: 20},
		Bild:    render.Blatt("/probe.mse-style/quadrat.png"),
		Deckung: 1,
		Passung: &render.Bildpassung{Ausrichtung: render.RechtsMitte},
	}}}

	bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) > 0 {
		t.Fatalf("zeichnen: %v, %v", err, befunde)
	}

	// **Rechts steht das Quadrat**, links bleibt es leer.
	if bild.NRGBAAt(35, 10).A == 0 {
		t.Error("rechts wurde nichts gezeichnet")
	}

	if bild.NRGBAAt(5, 10).A != 0 {
		t.Error("links wurde gezeichnet — das Bild ist gestreckt")
	}

	// **Und die Breite ist die Höhe**: 20 Punkte, also alles links von 20 leer.
	if bild.NRGBAAt(19, 10).A != 0 {
		t.Error("das Bild reicht über seine eingepasste Breite hinaus")
	}
}

// **Ein hohes Bild in einem flachen Feld wird an der Höhe eingepasst.**
//
// Der andere Zweig von `conform_image` — `iw * height > ih * width` ist dann
// falsch, und die Höhe gibt den Maßstab.
func TestEinHohesBildWirdAnDerHoeheEingepasst(t *testing.T) {
	quelle := bilder{"/probe.mse-style/hoch.png": vollflaeche(10, 40)}

	plan := &render.Plan{Breite: 40, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Breite: 40, Hoehe: 20},
		Bild:    render.Blatt("/probe.mse-style/hoch.png"),
		Deckung: 1,
		Passung: &render.Bildpassung{Ausrichtung: render.LinksOben},
	}}}

	bild, _, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	// 10 × 40 in 40 × 20 heißt 5 × 20 — links oben.
	if bild.NRGBAAt(2, 10).A == 0 {
		t.Error("links wurde nichts gezeichnet")
	}

	if bild.NRGBAAt(7, 10).A != 0 {
		t.Error("das Bild ist breiter als eingepasst")
	}
}

// vollflaeche ist ein deckend rotes Bild der gegebenen Größe.
func vollflaeche(breite, hoehe int) image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	for y := range hoehe {
		for x := range breite {
			bild.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}

	return bild
}

// **Ein leeres Bild wird nicht eingepasst.** Es gibt kein Seitenverhältnis, und
// eine Division dadurch wäre der Absturz statt des leeren Rechtecks.
func TestEinLeeresBildWirdNichtEingepasst(t *testing.T) {
	quelle := bilder{"/probe.mse-style/leer.png": vollflaeche(0, 0)}

	plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Breite: 10, Hoehe: 10},
		Bild:    render.Blatt("/probe.mse-style/leer.png"),
		Deckung: 1,
		Passung: &render.Bildpassung{Ausrichtung: render.MitteMitte},
	}}}

	bild, _, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if bild.NRGBAAt(5, 5).A != 0 {
		t.Error("aus einem leeren Bild wurde etwas gezeichnet")
	}
}

// **Und ein Rahmen ohne Fläche bleibt, wie er ist.** Eine Ebene mit der Breite
// null zeichnet nichts; die Einpassung darf daran nichts ändern, sondern gibt
// den leeren Rahmen zurück.
func TestEinLeererRahmenBleibtLeer(t *testing.T) {
	quelle := bilder{"/probe.mse-style/quadrat.png": vollflaeche(20, 20)}

	plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Links: 3, Oben: 3, Breite: 0, Hoehe: 4},
		Bild:    render.Blatt("/probe.mse-style/quadrat.png"),
		Deckung: 1,
		Passung: &render.Bildpassung{Ausrichtung: render.MitteMitte},
	}}}

	bild, _, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if bild.NRGBAAt(3, 4).A != 0 {
		t.Error("ein Rahmen ohne Fläche hat gezeichnet")
	}
}

// **Ein breites Bild in einem hohen Feld wird an der Breite eingepasst.**
//
//	if (iw * options.height > ih * options.width) { // too much height requested
//	  w = options.width;  h = options.width * ih / iw;
//	}
//
// Der erste Zweig von `conform_image` (`src/gfx/generated_image.cpp`, Zeilen
// 51–53, Fassung f9d9356 vom 2021-07-06).
func TestEinBreitesBildWirdAnDerBreiteEingepasst(t *testing.T) {
	quelle := bilder{"/probe.mse-style/breit.png": vollflaeche(40, 10)}

	plan := &render.Plan{Breite: 20, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Breite: 20, Hoehe: 20},
		Bild:    render.Blatt("/probe.mse-style/breit.png"),
		Deckung: 1,
		Passung: &render.Bildpassung{Ausrichtung: render.LinksOben},
	}}}

	bild, _, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	// 40 × 10 in 20 × 20 heißt 20 × 5 — links oben.
	if bild.NRGBAAt(10, 2).A == 0 {
		t.Error("oben wurde nichts gezeichnet")
	}

	if bild.NRGBAAt(10, 8).A != 0 {
		t.Error("das Bild ist höher als eingepasst")
	}
}
