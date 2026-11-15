package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Bild wird auf das Seitenverhältnis der Zeichnung geklemmt** —
// #735.
//
//	double ar  = symbol->aspectRatio();
//	double par = (double)width/height;
//	if (par > ar && (ar > 1 || (allow_smaller && height < width))) {
//	  width  = int(height * ar);
//	} else if (par < ar && (ar < 1 || (allow_smaller && width < height))) {
//	  height = int(width / ar);
//	}
//
// `render_symbol`, `src/render/symbol/viewer.cpp`, Zeilen 18–25, Fassung
// f9d9356 vom 2021-07-06.
//
// **Geklemmt wird die Fläche, nicht die Zeichnung.** Wer ein 400 × 100 großes
// Bild eines flachen Symbols verlangt, bekommt ein 200 × 100 großes — das
// Symbol wird nicht in die Breite gezogen, und der Aufrufer sieht an den Maßen,
// was er bekommen hat.
func TestDieFlaecheWirdAufDasSeitenverhaeltnisGeklemmt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name                  string
		teile                 []render.Symbolteil
		breite, hoehe         int
		kleiner               bool
		zielBreite, zielHoehe int
	}{
		{
			// ar = 2, par = 4 → die Breite folgt der Höhe.
			name:   "ein flaches Symbol in einem zu breiten Bild",
			teile:  []render.Symbolteil{viereck(0, 0.25, 1, 0.75)},
			breite: 400, hoehe: 100,
			zielBreite: 200, zielHoehe: 100,
		},
		{
			// ar = 0,5, par = 0,25 → die Höhe folgt der Breite.
			name:   "ein hohes Symbol in einem zu hohen Bild",
			teile:  []render.Symbolteil{viereck(0.25, 0, 0.75, 1)},
			breite: 100, hoehe: 400,
			zielBreite: 100, zielHoehe: 200,
		},
		{
			// **Ohne `allow_smaller` bleibt ein quadratisches Symbol
			// ungeklemmt.** `ar > 1` gilt nicht, und der Aufrufer hat nicht
			// gesagt, dass er ein kleineres Bild annimmt.
			name:   "ein quadratisches Symbol bleibt ungeklemmt",
			teile:  []render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)},
			breite: 200, hoehe: 100,
			zielBreite: 200, zielHoehe: 100,
		},
		{
			name:   "mit allow_smaller wird es geklemmt",
			teile:  []render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)},
			breite: 200, hoehe: 100, kleiner: true,
			zielBreite: 100, zielHoehe: 100,
		},
		{
			name:   "und quer dazu ebenso",
			teile:  []render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)},
			breite: 100, hoehe: 200, kleiner: true,
			zielBreite: 100, zielHoehe: 100,
		},
		{
			// **Ein quadratisches Bild bleibt quadratisch**, auch bei einem
			// flachen Symbol: `par > ar` gilt nicht.
			name:   "ein quadratisches Bild bleibt",
			teile:  []render.Symbolteil{viereck(0, 0.25, 1, 0.75)},
			breite: 100, hoehe: 100, kleiner: true,
			zielBreite: 100, zielHoehe: 100,
		},
		{
			// **Eine Fläche ohne Höhe wird nicht geklemmt.** MSE käme über die
			// Division auf `inf` und rechnete die Breite auf 0; hier bleibt
			// stehen, was hereinkam — beides ergibt ein leeres Bild, und
			// [Symbolflaeche] gibt diesen Fall gar nicht erst her.
			name:   "eine Fläche ohne Höhe",
			teile:  []render.Symbolteil{viereck(0, 0.25, 1, 0.75)},
			breite: 100, hoehe: 0, kleiner: true,
			zielBreite: 100, zielHoehe: 0,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			maske := render.RastereSymbol(fall.teile, fall.breite, fall.hoehe, 0, fall.kleiner)

			if maske.Breite != fall.zielBreite || maske.Hoehe != fall.zielHoehe {
				t.Errorf("%d × %d, erwartet %d × %d",
					maske.Breite, maske.Hoehe, fall.zielBreite, fall.zielHoehe)
			}
		})
	}
}

// **Der Randstrich wächst mit der kürzeren Seite, nicht mit dem Zoom.**
//
//	if (width > height) { viewer.setZoom(width);  viewer.border_radius *= (double)height / width; }
//	else                { viewer.setZoom(height); viewer.border_radius *= (double)width / height; }
//
// Dieselbe Stelle, Zeilen 26–34. Der Zoom nimmt die **längere** Seite, der
// Randanteil wird um das Verhältnis gekürzt — zusammen ergibt das
// `border_radius · kürzere Seite`.
//
// **Bis zum 2026-08-21 stand hier `randbreite · zoom`.** Bei einem
// quadratischen Bild ist das dasselbe, und quadratisch war jedes Bild, das
// dieses Programm bisher gerastert hat. Mit
// #735 kommen
// die anderen dazu.
func TestDerRandstrichWaechstMitDerKuerzerenSeite(t *testing.T) {
	t.Parallel()

	// Ein Quadrat von 0,25 bis 0,75. Bei 200 × 100 liegt seine linke Kante auf
	// x = 50; der Randstrich liegt mittig darauf.
	maske := render.RastereSymbol(
		[]render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)}, 200, 100, 0.05, false)

	if maske.Breite != 200 || maske.Hoehe != 100 {
		t.Fatalf("die Fläche ist %d × %d, erwartet 200 × 100", maske.Breite, maske.Hoehe)
	}

	// **Sichtbar ist nur die äußere Hälfte des Strichs** — das Innere wird
	// hinterher abgezogen (`combineBuffers`, Zeilen 108–118). Bei einer Breite
	// von 5 liegt sie auf x ∈ [47,5; 50).
	if bei := maske.Bei(48, 50); bei != render.SymbolRand {
		t.Errorf("bei x = 48 steht %v, erwartet Rand", bei)
	}

	// **Und weiter draußen ist nichts.** Mit der längeren Seite wäre der
	// Strich 10 breit und reichte bis x = 45 — das ist der Unterschied, um den
	// es hier geht.
	if bei := maske.Bei(46, 50); bei != render.SymbolAussen {
		t.Errorf("bei x = 46 steht %v, erwartet außen — der Strich ist zu breit", bei)
	}
}

// **Die Strichbreite ist ganzzahlig** —
// #735.
//
//	border->SetPen(wxPen(*wxWHITE, (int) rotation.trS(border_radius)));
//
// `drawSymbolShape`, `src/render/symbol/viewer.cpp`, Zeile 268, Fassung
// f9d9356 vom 2021-07-06. **`(int)` schneidet ab**: bei 0,135 auf 100 Punkten
// zeichnet MSE einen Strich von 13, nicht von 13,5.
//
// **Gemessen wird die Zahl der Randpunkte auf der mittleren Zeile**, links und
// rechts zusammen. Sichtbar ist nur die äußere Hälfte des Strichs, und ein
// halber Punkt mehr Breite ist genau ein Punkt mehr über beide Seiten.
func TestDieStrichbreiteIstGanzzahlig(t *testing.T) {
	t.Parallel()

	maske := render.RastereSymbol(
		[]render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)}, 100, 100, 0.135, false)

	randpunkte := 0

	for x := range 100 {
		if maske.Bei(x, 50) == render.SymbolRand {
			randpunkte++
		}
	}

	if randpunkte != 13 {
		t.Errorf("%d Randpunkte, erwartet 13 — mit 13,5 werden es 14", randpunkte)
	}
}
