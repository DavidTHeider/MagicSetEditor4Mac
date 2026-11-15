package render

// **Das Schachbrett eines leeren Bildfelds** — `draw_checker`:
//
//	dc.SetBrush(*wxWHITE_BRUSH);
//	dc.DrawRectangle(rect);
//	dc.SetBrush(Color(235,235,235));
//	const double checker_size = 10;
//	int odd = 0;
//	for (double y = 0 ; y < rect.height ; y += checker_size) {
//	  for (double x = odd * checker_size ; x < rect.width ; x += checker_size * 2) {
//
// `src/gui/util.cpp`, Zeilen 114–132, Fassung f9d9356 vom 2021-07-06 — grau
// beginnt oben links, und die Felder sind **zehn Bildpunkte des Ziels**:
// `imagePlaceholder` zeichnet in ein Bild von `w × h` Bildpunkten mit dem
// Maßstab eins (`src/render/value/image.cpp`, Zeilen 102–110).

import (
	"image/color"
	"testing"
)

func TestDasSchachbrettHatZehnBildpunkteJeFeld(t *testing.T) {
	faelle := []struct {
		name           string
		vergroesserung float64
	}{
		{"einfach", 1},
		{"doppelt", 2},
	}

	grau := color.NRGBA{R: 235, G: 235, B: 235, A: 255}
	weiss := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := &Plan{Breite: 100, Hoehe: 100, Ebenen: []Ebene{{
				Art:     ArtBild,
				Rahmen:  Rechteck{Links: 10, Oben: 10, Breite: 60, Hoehe: 40},
				Bild:    Bildbaum{Knotenart: Platzhalterknoten},
				Deckung: 1,
			}}}

			bild, befunde, err := NeuerRasterer(nil).Zeichne(t.Context(), plan, fall.vergroesserung, GueteMittel)
			if err != nil || len(befunde) != 0 {
				t.Fatalf("zeichnen: %v %v", err, befunde)
			}

			ecke := int(10 * fall.vergroesserung)

			for _, probe := range []struct {
				x, y     int
				erwartet color.NRGBA
			}{
				{0, 0, grau},
				{9, 9, grau},
				{10, 0, weiss},
				{0, 10, weiss},
				{10, 10, grau},
				{20, 0, grau},
				{25, 15, weiss},
			} {
				if punkt := bild.NRGBAAt(ecke+probe.x, ecke+probe.y); punkt != probe.erwartet {
					t.Errorf("(%d, %d) ist %v, erwartet %v", probe.x, probe.y, punkt, probe.erwartet)
				}
			}
		})
	}
}
