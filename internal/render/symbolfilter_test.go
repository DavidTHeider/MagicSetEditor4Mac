package render_test

import (
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Erst der Filter macht aus der Maske ein Bild.**
//
//	if      (point == SYMBOL_INSIDE) return fill_color;
//	else if (point == SYMBOL_BORDER) return border_color;
//	else                             return Color(0,0,0,0);
//
// `SolidFillSymbolFilter::color`, `src/render/symbol/filter.cpp`, Zeilen
// 90–94, Fassung f9d9356 vom 2021-07-06. **Außen ist durchsichtig** — ein
// Symbol steht auf einer Karte und nicht auf einem Kasten
// (#323).
func TestDerVollfilterGibtZweiFarbenUndDurchsicht(t *testing.T) {
	maske := render.RastereSymbol(
		[]render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)}, 100, 100, 0.05, false)

	bild := render.FaerbeSymbol(maske, render.Vollfilter{
		Fuellfarbe: color.NRGBA{R: 200, G: 0, B: 0, A: 255},
		Randfarbe:  color.NRGBA{R: 0, G: 0, B: 0, A: 255},
	})

	faelle := []struct {
		name     string
		x, y     int
		erwartet color.NRGBA
	}{
		{name: "innen", x: 50, y: 50, erwartet: color.NRGBA{R: 200, A: 255}},
		{name: "am Rand", x: 23, y: 50, erwartet: color.NRGBA{A: 255}},
		{name: "außen", x: 2, y: 2, erwartet: color.NRGBA{}},
	}

	for _, fall := range faelle {
		haben := bild.NRGBAAt(fall.x, fall.y)
		if haben != fall.erwartet {
			t.Errorf("%s (%d,%d): %+v, erwartet %+v",
				fall.name, fall.x, fall.y, haben, fall.erwartet)
		}
	}
}

// **Ein Farbverlauf mischt zwei Farben über die Fläche.**
//
//	double t = fabs((x-cx)*(ex-cx) + (y-cy)*(ey-cy)) / len;
//
// `LinearGradientSymbolFilter::t`, Zeilen 158–161. **Der Anteil geht vom
// Mittelpunkt aus und wird abgeschnitten** — `min(1, max(0, t))`.
func TestDerVerlaufsfilterMischtUeberDieFlaeche(t *testing.T) {
	maske := render.RastereSymbol(
		[]render.Symbolteil{viereck(0, 0, 1, 1)}, 100, 100, 0, false)

	// Von der linken Kante zur rechten: bei x = 0 ganz die erste Farbe, bei
	// x = 1 ganz die zweite.
	bild := render.FaerbeSymbol(maske, render.Verlaufsfilter{
		Fuellfarbe1: color.NRGBA{A: 255},
		Fuellfarbe2: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Randfarbe1:  color.NRGBA{A: 255},
		Randfarbe2:  color.NRGBA{A: 255},
		MitteX:      0, MitteY: 0.5, EndeX: 1, EndeY: 0.5,
	})

	links := bild.NRGBAAt(1, 50)
	rechts := bild.NRGBAAt(98, 50)
	mitte := bild.NRGBAAt(50, 50)

	if links.R > 10 {
		t.Errorf("links ist %d, erwartet fast schwarz", links.R)
	}

	if rechts.R < 245 {
		t.Errorf("rechts ist %d, erwartet fast weiß", rechts.R)
	}

	if mitte.R < 110 || mitte.R > 145 {
		t.Errorf("die Mitte ist %d, erwartet ungefähr die Hälfte", mitte.R)
	}
}

// **Der Kreisverlauf geht von der Mitte nach außen.**
//
//	return sqrt((sqr(x - 0.5) + sqr(y - 0.5)) * 2);
//
// `RadialGradientSymbolFilter::t`, `src/render/symbol/filter.cpp`, Zeilen
// 184–186, Fassung f9d9356 vom 2021-07-06. **Ohne Abschneiden**: in der Ecke
// wird der Anteil genau eins, und weiter reicht die Fläche nicht.
func TestDerKreisfilterGehtVonDerMitteNachAussen(t *testing.T) {
	maske := render.RastereSymbol(
		[]render.Symbolteil{viereck(0, 0, 1, 1)}, 100, 100, 0.05, false)

	bild := render.FaerbeSymbol(maske, render.Kreisfilter{
		Fuellfarbe1: color.NRGBA{A: 255},
		Fuellfarbe2: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Randfarbe1:  color.NRGBA{R: 100, A: 255},
		Randfarbe2:  color.NRGBA{R: 200, A: 255},
	})

	if mitte := bild.NRGBAAt(50, 50); mitte.R > 5 {
		t.Errorf("die Mitte ist %d, erwartet fast schwarz", mitte.R)
	}

	// **An der Kante der Mitte gegenüber**: t = √(0,45² · 2) = 0,636, und
	// das sind 162 von 255.
	if aussen := bild.NRGBAAt(95, 50); aussen.R < 158 || aussen.R > 166 {
		t.Errorf("außen ist %d, erwartet rund 162", aussen.R)
	}

	// **Auch der Rand verläuft** — er hat seine eigenen beiden Farben.
	if rand := bild.NRGBAAt(0, 50); rand.R < 100 || rand.R > 200 {
		t.Errorf("der Rand ist %d, erwartet zwischen seinen beiden Farben", rand.R)
	}
}

// **Jeder Filter gibt auch dem Rand und dem Außen seine Farbe.**
//
// Außen ist bei allen dreien dasselbe: `Color(0,0,0,0)` — nichts. Ein Filter,
// der dort etwas zeichnete, machte aus einem Symbol einen Kasten.
func TestJederFilterKenntRandUndAussen(t *testing.T) {
	maske := render.RastereSymbol(
		[]render.Symbolteil{viereck(0.25, 0.25, 0.75, 0.75)}, 100, 100, 0.05, false)

	filter := map[string]render.Symbolfilter{
		"Verlauf": render.Verlaufsfilter{
			Randfarbe1: color.NRGBA{R: 90, A: 255},
			Randfarbe2: color.NRGBA{R: 90, A: 255},
			MitteX:     0.5, MitteY: 0.5, EndeX: 1, EndeY: 1,
		},
		"Kreis": render.Kreisfilter{
			Randfarbe1: color.NRGBA{R: 90, A: 255},
			Randfarbe2: color.NRGBA{R: 90, A: 255},
		},

		// **Ein Verlauf ohne Länge**: MSE setzt `len = 1`, sonst teilte es
		// durch null (`LinearGradientSymbolFilter::color`, Zeilen 152–156).
		"Verlauf ohne Länge": render.Verlaufsfilter{
			Randfarbe1: color.NRGBA{R: 90, A: 255},
			Randfarbe2: color.NRGBA{R: 90, A: 255},
			MitteX:     0.5, MitteY: 0.5, EndeX: 0.5, EndeY: 0.5,
		},
	}

	for name, einer := range filter {
		t.Run(name, func(t *testing.T) {
			bild := render.FaerbeSymbol(maske, einer)

			if rand := bild.NRGBAAt(23, 50); rand.R != 90 {
				t.Errorf("der Rand ist %+v, erwartet die Randfarbe", rand)
			}

			if aussen := bild.NRGBAAt(2, 2); aussen != (color.NRGBA{}) {
				t.Errorf("außen ist %+v, erwartet nichts", aussen)
			}
		})
	}
}
