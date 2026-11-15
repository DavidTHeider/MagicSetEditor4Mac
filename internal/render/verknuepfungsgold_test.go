package render_test

// **Ein Goldbildvergleich je Bildverknüpfung**
// (#465).
//
// # Warum neben den vorhandenen Tests
//
// `blenden_test.go`, `deckung_test.go` und `geradefarbe_test.go` prüfen die
// **Regel** an einzelnen Bildpunkten: weiße Maske nimmt `light`, graue mischt.
// Das ist die richtige Frage für die Regel und die falsche für das Bild: ein
// Fehler, der die geprüften Punkte meidet, fällt dort nicht auf.
//
// Hier läuft jede Verknüpfung über ein **ganzes** Bild mit einem Verlauf, und
// das Ergebnis wird gegen ein Goldbild gehalten, das aus der **Regel** gerechnet
// ist — nicht aus der Umsetzung. Die Toleranz ist [render.Toleranz] mit ihrer
// Begründung; `Innerhalb` hält zusätzlich die Spitze.
//
// **Das Goldbild wird gebaut, nicht abgelegt.** So hält es `aussparung_test.go`
// schon, und Referenzbilder gehören nach §11 ohnehin nicht ins Repositorium.
//
// # Die fünfte Verknüpfung steht woanders
//
// `symbol_variation` hat ihren Vergleich in `symbolgold_test.go` — dort gegen
// ein **abgelegtes** Goldbild, samt Gegenprobe („ein Symbol ohne Rand liegt
// über der Toleranz"). Sie hier ein zweites Mal zu prüfen hieße, dieselbe
// Zusage an zwei Stellen zu führen, die auseinanderlaufen können.

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// verlaufsquelle gibt für jeden Pfad ein Bild mit einem eigenen Verlauf.
//
// **Ein Verlauf und keine Fläche**: über einer Fläche sähe eine Verschiebung
// oder eine Spiegelung genauso aus wie das Richtige.
type verlaufsquelle struct{ seite int }

func (v verlaufsquelle) Bild(pfad string) (image.Image, error) {
	bild := image.NewNRGBA(image.Rect(0, 0, v.seite, v.seite))

	for y := range v.seite {
		for x := range v.seite {
			bild.Set(x, y, v.punkt(pfad, x, y))
		}
	}

	return bild, nil
}

// punkt gibt die Farbe eines Punktes — je Pfad eine andere Regel.
func (v verlaufsquelle) punkt(pfad string, x, y int) color.NRGBA {
	anteil := func(n int) uint8 { return uint8(255 * n / max(v.seite-1, 1)) } //nolint:gosec // durch die Division begrenzt

	switch pfad {
	case "/hell.png":
		return color.NRGBA{R: anteil(x), G: 40, B: 40, A: 255}

	case "/dunkel.png":
		return color.NRGBA{R: 40, G: 40, B: anteil(y), A: 255}

	case "/maske.png":
		return color.NRGBA{R: anteil(x), G: anteil(x), B: anteil(x), A: 255}
	}

	return color.NRGBA{R: anteil(y), G: anteil(x), B: 128, A: 255}
}

const goldseite = 16

// TestJedeVerknuepfungHaeltIhrGoldbild vergleicht über das ganze Bild.
func TestJedeVerknuepfungHaeltIhrGoldbild(t *testing.T) {
	quelle := verlaufsquelle{seite: goldseite}

	for _, fall := range []struct {
		name string
		baum render.Bildbaum
		gold func(x, y int) color.NRGBA
	}{
		{
			// **`masked_blend`**: wo die Maske hell ist, gilt `light`.
			name: "masked_blend",
			baum: render.Bildbaum{Knotenart: "masked_blend", Kinder: []render.Bildkind{
				{Name: "light", Baum: render.Blatt("/hell.png")},
				{Name: "dark", Baum: render.Blatt("/dunkel.png")},
				{Name: "mask", Baum: render.Blatt("/maske.png")},
			}},
			gold: func(x, y int) color.NRGBA {
				hell := quelle.punkt("/hell.png", x, y)
				dunkel := quelle.punkt("/dunkel.png", x, y)
				maske := quelle.punkt("/maske.png", x, y)

				return gemischt(hell, dunkel, float64(maske.R)/255)
			},
		},
		{
			// **`linear_blend`** mischt entlang einer Geraden von `image1`
			// nach `image2`.
			//
			// **Das Goldbild rechnet MSEs Formel nach, nicht unsere.** Sie
			// steht in `src/gfx/blend_image.cpp`, Zeilen 24–58, Fassung
			// f9d9356 vom 2021-07-06:
			//
			//	const int fixed = 1<<16;
			//	double a = fixed / (sqr(width) * sqr(x1-x2) + sqr(height) * sqr(y1-y2));
			//	int xm = to_int( (x2 - x1) * width  * a );
			//	int ym = to_int( (y2 - y1) * height * a );
			//	int d  = to_int( - (x1 * width * xm + y1 * height * ym) );
			//	…
			//	int mult = x * xm + y * ym + d;
			//	if (mult < 0)      mult = 0;
			//	if (mult > fixed)  mult = fixed;
			//	data1[0] = data1[0] + mult * (data2[0] - data1[0]) / fixed;
			//
			// **Das ist der Punkt dieses Falls**: nähme das Goldbild die
			// Formel aus `linearblende.go`, prüfte es die Umsetzung gegen
			// sich selbst und wäre grün, auch wenn beide falsch sind (§10).
			//
			// **Drei Kanäle.** Das Alpha des ersten Bildes bleibt — wx führt
			// es getrennt von den Daten.
			name: "linear_blend",
			baum: render.Bildbaum{
				Knotenart: "linear_blend",
				Kinder: []render.Bildkind{
					{Name: "image1", Baum: render.Blatt("/hell.png")},
					{Name: "image2", Baum: render.Blatt("/dunkel.png")},
				},
				Zahlen: map[string]float64{"x1": 0.2, "y1": 0.1, "x2": 0.9, "y2": 0.8},
			},
			gold: nachMSEsGerade(quelle, 0.2, 0.1, 0.9, 0.8),
		},
		{
			// **`recolor_image`** zerlegt jeden Punkt in Rot-, Grün-, Blau-
			// und Weißanteil und setzt ihn aus vier Zielfarben neu zusammen.
			//
			// **Auch dieses Goldbild rechnet MSEs Formel nach**
			// (`src/gfx/image_effects.cpp`, Zeilen 76–92, Fassung f9d9356 vom
			// 2021-07-06):
			//
			//	int lo = min(x.r,min(x.g,x.b));
			//	int nr = x.r - lo; … int nw = lo;
			//	int total = max(255, nr+ng+nb+nw);
			//	return RGB((nr*cr.r + ng*cg.r + nb*cb.r + nw*cw.r) / total, …);
			//
			// **`total` ist ein `max`, keine Summe.** Für ein Bild, dessen
			// Anteile sich zu weniger als 255 addieren, wird durch 255
			// geteilt — die Zeile darüber sagt warum.
			name: "recolor_image",
			baum: render.Bildbaum{
				Knotenart: "recolor_image",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/bild.png")}},
				Farben: map[string]render.Farbe{
					"red":   {R: 200, G: 30, B: 10, A: 255},
					"green": {R: 20, G: 190, B: 40, A: 255},
					"blue":  {R: 15, G: 45, B: 210, A: 255},
					"white": {R: 240, G: 235, B: 225, A: 255},
				},
			},
			gold: nachMSEsUmfaerbung(quelle,
				render.Farbe{R: 200, G: 30, B: 10, A: 255},
				render.Farbe{R: 20, G: 190, B: 40, A: 255},
				render.Farbe{R: 15, G: 45, B: 210, A: 255},
				render.Farbe{R: 240, G: 235, B: 225, A: 255}),
		},
		{
			// **`set_alpha`**: die Deckung wird gesetzt, die Farbe bleibt.
			name: "set_alpha",
			baum: render.Bildbaum{
				Knotenart: "set_alpha",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/bild.png")}},
				Zahlen:    map[string]float64{"alpha": 0.5},
			},
			gold: func(x, y int) color.NRGBA {
				punkt := quelle.punkt("/bild.png", x, y)

				// **Abgeschnitten, nicht gerundet**: MSE rechnet
				// `Byte(alpha * 255)`, und 0,5 × 255 = 127,5 ergibt **127**
				// (`src/gfx/blend_image.cpp`, Zeile 108, Fassung f9d9356 vom
				// 2021-07-06).
				punkt.A = uint8(int(punkt.A) * 127 / 255) //nolint:gosec // beide ≤ 255

				return punkt
			},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			neu, befund := render.NeuerRasterer(quelle).AusBaumFuerTest(fall.baum)
			if befund != "" {
				t.Fatalf("ausführen: %s", befund)
			}

			gold := gebaut(fall.gold)

			vergleich, err := render.Vergleiche(gold, neu)
			if err != nil {
				t.Fatalf("vergleichen: %v", err)
			}

			if !vergleich.Innerhalb() {
				t.Errorf("%s: %s", fall.name, vergleich)
			}

			// **Die Gegenprobe im selben Test** (§10): ein Goldbild, das um
			// einen Punkt verschoben ist, muss auffallen — sonst sagte der
			// Vergleich nur, dass irgendetwas gleich ist.
			verschoben := gebaut(func(x, y int) color.NRGBA {
				return fall.gold(min(x+1, goldseite-1), y)
			})

			gegen, err := render.Vergleiche(verschoben, neu)
			if err != nil {
				t.Fatalf("gegenvergleichen: %v", err)
			}

			if gegen.Innerhalb() {
				t.Errorf("%s: eine Verschiebung um einen Punkt fällt nicht auf: %s",
					fall.name, gegen)
			}
		})
	}
}

// gebaut malt ein Goldbild aus einer Regel.
//
// **`NRGBA` und nicht `RGBA`, auch bei der Farbe.** Gos `image.RGBA` **und**
// `color.RGBA` sind **vormultipliziert**: `color.RGBA{R: 200, A: 128}` meint
// dort nicht „zu 50 % durchsichtiges Rot", sondern einen Wert, dessen Rotanteil
// größer ist als seine Deckung. Der erste Anlauf schrieb genau das, und
// `set_alpha` wich an jedem Punkt um 127 ab — der Fehler lag in der
// Vorrichtung, nicht in der Verknüpfung.
func gebaut(regel func(x, y int) color.NRGBA) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, goldseite, goldseite))

	for y := range goldseite {
		for x := range goldseite {
			bild.Set(x, y, regel(x, y))
		}
	}

	return bild
}

// nachMSEsGerade baut das Goldbild für `linear_blend` aus MSEs Formel.
//
// **Ganzzahlig gerechnet wie dort**, samt `to_int` — es schneidet ab, und der
// Rest der Rechnung hängt daran.
func nachMSEsGerade(
	quelle verlaufsquelle, x1, y1, x2, y2 float64,
) func(x, y int) color.NRGBA {
	const fest = 1 << 16

	breite, hoehe := float64(goldseite), float64(goldseite)

	a := fest / (breite*breite*(x1-x2)*(x1-x2) + hoehe*hoehe*(y1-y2)*(y1-y2))
	xm := int((x2 - x1) * breite * a)
	ym := int((y2 - y1) * hoehe * a)
	d := int(-(x1*breite*float64(xm) + y1*hoehe*float64(ym)))

	return func(x, y int) color.NRGBA {
		mass := max(0, min(fest, x*xm+y*ym+d))

		eins := quelle.punkt("/hell.png", x, y)
		zwei := quelle.punkt("/dunkel.png", x, y)

		kanal := func(a, b uint8) uint8 {
			return uint8(int(a) + mass*(int(b)-int(a))/fest) //nolint:gosec // Ergebnis in [0,255]
		}

		return color.NRGBA{
			R: kanal(eins.R, zwei.R), G: kanal(eins.G, zwei.G),
			B: kanal(eins.B, zwei.B),
			// **Das Alpha des ersten Bildes bleibt.**
			A: eins.A,
		}
	}
}

// nachMSEsUmfaerbung baut das Goldbild für `recolor_image` aus MSEs Formel.
func nachMSEsUmfaerbung(
	quelle verlaufsquelle, rot, gruen, blau, weiss render.Farbe,
) func(x, y int) color.NRGBA {
	return func(x, y int) color.NRGBA {
		punkt := quelle.punkt("/bild.png", x, y)

		tief := int(min(punkt.R, min(punkt.G, punkt.B)))
		nr, ng := int(punkt.R)-tief, int(punkt.G)-tief
		nb, nw := int(punkt.B)-tief, tief

		gesamt := max(255, nr+ng+nb+nw)

		kanal := func(cr, cg, cb, cw uint8) uint8 {
			return uint8((nr*int(cr) + ng*int(cg) + nb*int(cb) + nw*int(cw)) / gesamt) //nolint:gosec // durch gesamt begrenzt
		}

		return color.NRGBA{
			R: kanal(rot.R, gruen.R, blau.R, weiss.R),
			G: kanal(rot.G, gruen.G, blau.G, weiss.G),
			B: kanal(rot.B, gruen.B, blau.B, weiss.B),
			// **Das Alpha bleibt** — MSE fasst nur die drei Farbkanäle an.
			A: punkt.A,
		}
	}
}

// gemischt mischt zwei Farben nach dem Anteil der Maske.
func gemischt(hell, dunkel color.NRGBA, anteil float64) color.NRGBA {
	kanal := func(h, d uint8) uint8 {
		return uint8(float64(d) + (float64(h)-float64(d))*anteil) //nolint:gosec // Anteil in [0,1]
	}

	return color.NRGBA{
		R: kanal(hell.R, dunkel.R), G: kanal(hell.G, dunkel.G),
		B: kanal(hell.B, dunkel.B), A: kanal(hell.A, dunkel.A),
	}
}
