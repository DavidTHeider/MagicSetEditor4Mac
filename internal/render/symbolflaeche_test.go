package render_test

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Setsymbol wird in der Auflösung gerastert, die das Ziel verlangt** —
// #735.
//
//	int size = max(100, 3*max(opt.width,opt.height));
//	if (opt.width <= 1 || opt.height <= 1) {
//	  return render_symbol(…, size, size);
//	} else {
//	  int width  = size * opt.width  / max(opt.width,opt.height);
//	  int height = size * opt.height / max(opt.width,opt.height);
//	  return render_symbol(…, width, height, false, true);
//	}
//
// `SymbolToImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 453–470,
// Fassung f9d9356 vom 2021-07-06.
//
// **Dreifach, damit die Verkleinerung sauber ist.** Wer 200 Punkte breit
// zeichnen will, rastert 600 und rechnet herunter — nicht 128 und rechnet
// hoch. Bis zum 2026-08-21 stand hier die feste Zahl 128.
//
// **`<= 1`, nicht `== 0`.** Ein Maß von einem Punkt trägt kein
// Seitenverhältnis; MSE fällt dort auf das Quadrat zurück.
func TestDieFlaecheEinesSetsymbols(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name          string
		gefragt       image.Point
		breite, hoehe int
		kleiner       bool
	}{
		{
			// **Ohne Maß gilt die Untergrenze** — `max(100, 0)`.
			name: "ohne Maß", gefragt: image.Point{},
			breite: 100, hoehe: 100,
		},
		{
			name: "ein Punkt ist kein Maß", gefragt: image.Point{X: 1, Y: 1},
			breite: 100, hoehe: 100,
		},
		{
			name: "nur eine Breite", gefragt: image.Point{X: 200},
			breite: 600, hoehe: 600,
		},
		{
			name: "nur eine Höhe", gefragt: image.Point{Y: 200},
			breite: 600, hoehe: 600,
		},
		{
			// **Mit beiden Maßen im verlangten Verhältnis**, und `allow_smaller`
			// dazu: die Zeichnung darf das Bild weiter verkleinern.
			name: "beide Maße", gefragt: image.Point{X: 200, Y: 100},
			breite: 600, hoehe: 300, kleiner: true,
		},
		{
			name: "beide Maße, hochkant", gefragt: image.Point{X: 100, Y: 300},
			breite: 300, hoehe: 900, kleiner: true,
		},
		{
			// **Ein Setsymbol auf einer Karte ist klein**, und die Untergrenze
			// greift: `magic-m15` gibt ihm 21 × 21.
			name: "ein Setsymbol auf einer Karte", gefragt: image.Point{X: 21, Y: 21},
			breite: 100, hoehe: 100, kleiner: true,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			breite, hoehe, kleiner := render.Symbolflaeche(fall.gefragt)
			if breite != fall.breite || hoehe != fall.hoehe || kleiner != fall.kleiner {
				t.Errorf("%d × %d, allow_smaller %v — erwartet %d × %d, %v",
					breite, hoehe, kleiner, fall.breite, fall.hoehe, fall.kleiner)
			}
		})
	}
}

// **Und der Rasterer benutzt sie.**
//
// **Der Test geht über die Naht und nicht über die Rechnung**: dass
// [render.Symbolflaeche] stimmt, sagt nichts darüber, ob `symbol_variation`
// sie fragt — bis zum 2026-08-21 tat es das nicht, und die Rechnung hätte
// danebengestanden, ohne dass ein Lauf rot wurde (§10).
func TestSymbolVariationRastertInDerVerlangtenAufloesung(t *testing.T) {
	t.Parallel()

	baum := render.Bildbaum{
		Knotenart: "symbol_variation",
		Symbol:    []render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)},
		Texte:     map[string]string{"fill_type": "solid"},
		Farben: map[string]render.Farbe{
			"fill_color":   {R: 210, G: 180, B: 60, A: 255},
			"border_color": {A: 255},
		},
	}

	faelle := []struct {
		name          string
		gefragt       image.Point
		breite, hoehe int
	}{
		{name: "ohne Maß", gefragt: image.Point{}, breite: 100, hoehe: 100},
		{name: "nur eine Breite", gefragt: image.Point{X: 200}, breite: 600, hoehe: 600},
		{
			// 600 × 300 verlangt, auf das Verhältnis der Zeichnung geklemmt.
			name: "beide Maße", gefragt: image.Point{X: 200, Y: 100},
			breite: 300, hoehe: 300,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			bild, befund := render.NeuerRasterer(bilder{}).
				AusBaumMitMassFuerTest(baum, fall.gefragt)
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			rahmen := bild.Bounds()
			if rahmen.Dx() != fall.breite || rahmen.Dy() != fall.hoehe {
				t.Errorf("%d × %d, erwartet %d × %d",
					rahmen.Dx(), rahmen.Dy(), fall.breite, fall.hoehe)
			}
		})
	}
}
