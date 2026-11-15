package render_test

// **Zwei Maskenregeln, an MSE2 nachgelesen**
// (#471).
//
//	void set_alpha(Image& img, const Image& img_alpha) {
//	  Image img_alpha_resampled = resample(img_alpha, img.GetWidth(), img.GetHeight());
//	  …
//	  im[i] = (im[i] * al[i*3]) / 255;
//	}
//
// `src/gfx/blend_image.cpp`, Zeilen 79–87, Fassung f9d9356 vom 2021-07-06.
// Zwei Aussagen in vier Zeilen: die Maske wird **auf die Bildgröße gebracht**,
// und gelesen wird der **rote** Kanal (`al[i*3]`).
//
// Für die Ebenenmaske sagt MSE dasselbe noch deutlicher:
//
//	// Copy red chanel to alpha
//	Byte* from = img.GetData(), *to = alpha;
//	for (size_t i = 0 ; i < n ; ++i) {
//	  to[i] = from[3*i];
//	}
//
// `src/gfx/mask_image.cpp`, Zeilen 41–44.

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// maskenquelle gibt je Pfad ein Bild fester Größe und Farbe.
type maskenquelle struct{ bilder map[string]image.Image }

func (m maskenquelle) Bild(pfad string) (image.Image, error) {
	bild, da := m.bilder[pfad]
	if !da {
		return nil, errKeinBild
	}

	return bild, nil
}

// maskenflaeche baut ein einfarbiges Bild.
func maskenflaeche(breite, hoehe int, farbe color.NRGBA) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	for y := range hoehe {
		for x := range breite {
			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild
}

// einFleck baut ein schwarzes Bild mit **einem** weißen Punkt an (`fx`, `fy`).
//
// **Ein Fleck und keine Hälfte, und das ist der Punkt**: bei einer halbierten
// Maske ergibt eine fehlende Umrechnung dasselbe Bild, weil das Klemmen am
// Rand die letzte Spalte wiederholt — die Gegenprobe blieb damit grün, obwohl
// die Skalierung in X zurückgedreht war. Ein Fleck lässt sich nicht so
// verwechseln.
func einFleck(seite, fx, fy int) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, seite, seite))

	for y := range seite {
		for x := range seite {
			hell := uint8(0)
			if x == fx && y == fy {
				hell = 255
			}

			bild.SetNRGBA(x, y, color.NRGBA{R: hell, G: hell, B: hell, A: 255})
		}
	}

	return bild
}

// **Eine kleinere Maske deckt das ganze Bild ab.**
//
// Ohne die Umrechnung wirkte sie nur links oben — `maskenrot` teilte durch
// dieselbe Breite, mit der es multiplizierte, und die Zielgröße kam in der
// Rechnung gar nicht vor.
func TestEineKleinereMaskeWirdAufDieBildgroesseGebracht(t *testing.T) {
	quelle := maskenquelle{bilder: map[string]image.Image{
		"bild.png":  maskenflaeche(8, 8, color.NRGBA{R: 200, G: 200, B: 200, A: 255}),
		"maske.png": einFleck(4, 1, 1),
	}}

	baum := render.Bildbaum{
		Knotenart: "set_mask",
		Kinder: []render.Bildkind{
			{Name: "image", Baum: render.Blatt("bild.png")},
			{Name: "mask", Baum: render.Blatt("maske.png")},
		},
	}

	plan := &render.Plan{Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
		Art:     render.ArtBild,
		Rahmen:  render.Rechteck{Breite: 8, Hoehe: 8},
		Bild:    baum,
		Deckung: 1,
	}}}

	bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// Die 4×4-Maske ist schwarz bis auf den Punkt (1,1). Auf 8×8 gebracht
	// deckt sie das Feld (2..3, 2..3) und sonst nichts.
	//
	// **Beide Achsen werden geprüft**: (6,2) fällt nur auf, wenn X
	// umgerechnet wird, (2,6) nur bei Y — und (1,1) ist genau der Punkt, den
	// eine fehlende Umrechnung fälschlich deckte.
	faelle := []struct {
		x, y    int
		deckend bool
		wo      string
	}{
		{2, 2, true, "der Fleck"},
		{3, 3, true, "seine andere Ecke"},
		{1, 1, false, "wo die Maske ohne Umrechnung deckte"},
		{6, 2, false, "rechts daneben — die X-Achse"},
		{2, 6, false, "darunter — die Y-Achse"},
	}

	for _, fall := range faelle {
		alpha := bild.NRGBAAt(fall.x, fall.y).A
		if fall.deckend != (alpha > 128) {
			t.Errorf("%s (%d,%d): Alpha %d, erwartet %s",
				fall.wo, fall.x, fall.y, alpha,
				map[bool]string{true: "deckend", false: "durchsichtig"}[fall.deckend])
		}
	}
}

// **Die Ebenenmaske liest den roten Kanal, nicht die Helligkeit.**
//
// Bei einer echten Graustufenmaske ist beides dasselbe. Bei einer farbigen
// nicht — und fremde Pakete sind fremde Daten (§8).
func TestDieEbenenmaskeLiestDenRotenKanal(t *testing.T) {
	faelle := []struct {
		name     string
		maske    color.NRGBA
		erwartet uint8
	}{
		{"rot deckt ganz", color.NRGBA{R: 255, A: 255}, 255},
		{"grün deckt nicht", color.NRGBA{G: 255, A: 255}, 0},
		{"blau deckt nicht", color.NRGBA{B: 255, A: 255}, 0},
		{"weiß deckt ganz", color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 255},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			quelle := maskenquelle{bilder: map[string]image.Image{
				"bild.png":  maskenflaeche(4, 4, color.NRGBA{R: 200, G: 200, B: 200, A: 255}),
				"maske.png": maskenflaeche(4, 4, fall.maske),
			}}

			plan := &render.Plan{Breite: 4, Hoehe: 4, Ebenen: []render.Ebene{{
				Art:     render.ArtBild,
				Rahmen:  render.Rechteck{Breite: 4, Hoehe: 4},
				Bild:    render.Blatt("bild.png"),
				Maske:   render.Blatt("maske.png"),
				Deckung: 1,
			}}}

			bild, befunde, err := render.NeuerRasterer(quelle).Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if len(befunde) != 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if alpha := bild.NRGBAAt(2, 2).A; alpha != fall.erwartet {
				t.Errorf("Alpha = %d, erwartet %d", alpha, fall.erwartet)
			}
		})
	}
}
