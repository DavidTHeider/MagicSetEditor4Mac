package render_test

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Symbol misst in ganzen Punkten** — gerechnet wie bei MSE2.
//
//	RealSize SymbolInFont::size(Package& pkg, double size) {
//	  …
//	  return wxSize(actual_size * (int) (size) / (int) (img_size));
//	}
//
// `src/data/symbol_font.cpp`, Zeilen 147–153, Fassung f9d9356 vom 2021-07-06.
// `wxSize` rechnet mit ganzen Zahlen: die Schriftgröße wird abgeschnitten,
// und die Division ist ganzzahlig. Ein Manasymbol aus `magic-mana-small` (160
// Punkte, `image font size: 135`) misst bei Schriftgröße 14 deshalb 16 und
// nicht 16,59 — und bei 13,8 nur 15, weil `(int) 13,8` 13 ist.
//
// **Was das kostet, wenn man es überliest**: eine Zeile mit einem Symbol
// stand um gut einen halben Punkt höher als bei MSE2, und jede Zeile darunter
// rückte mit — bei „Icatian Infantry" (White) zwei Punkte ab der zweiten
// Zeile.
func TestEinSymbolMisstInGanzenPunkten(t *testing.T) {
	faelle := []struct {
		name          string
		groesse       float64
		breite, hoehe float64
	}{
		// 160 · 14 / 135 = 16,59 → 16; dazu `horizontal space: 2`.
		{name: "Schriftgröße 14", groesse: 14, breite: 18, hoehe: 16},
		// (int) 13,8 = 13; 160 · 13 / 135 = 15,41 → 15.
		{name: "Schriftgröße 13,8", groesse: 13.8, breite: 17, hoehe: 15},
	}

	quelle := bildmasse{da: map[string]image.Point{"/1.png": {X: 160, Y: 160}}}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ergebnis := render.SetzeAbschnitte(
				render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
				render.Rechteck{Breite: 200, Hoehe: 40},
				[]render.Abschnitt{{
					Text:    "1",
					Symbole: []render.Symbol{{Bild: render.Blatt("/1.png")}},
					Symbolmasse: render.Symbolmasse{
						Groesse: fall.groesse, Bildgroesse: 135, AbstandWaagerecht: 2,
					},
				}},
				render.Schriftsatz{Aufrecht: einheitsschrift{}},
				quelle,
			)

			if len(ergebnis.Zeilen) != 1 || len(ergebnis.Zeilen[0].Stuecke) != 1 {
				t.Fatalf("gesetzt wurden %+v", ergebnis.Zeilen)
			}

			if breite := ergebnis.Zeilen[0].Stuecke[0].Breite; breite != fall.breite {
				t.Errorf("das Symbol ist %g breit, erwartet %g", breite, fall.breite)
			}

			// Die Schrift gibt 12; das Symbol hebt die Zeile.
			if hoehe := ergebnis.Zeilen[0].Hoehe; hoehe != fall.hoehe {
				t.Errorf("die Zeile ist %g hoch, erwartet %g", hoehe, fall.hoehe)
			}
		})
	}
}

// **Ein verkleinerter Text verkleinert seine Symbole mit.**
//
//	font.font->getCharInfo(dc, ctx, font.size * scale, …, out);
//
// `SymbolTextElement::getCharInfo`, `src/render/text/symbol.cpp`, Zeilen
// 22–26, Fassung f9d9356 vom 2021-07-06: `scale` ist derselbe Faktor, mit dem
// `FontTextElement` die Schrift verkleinert. Hier verkleinerte `scale down
// to:` nur die Schrift — ein Symbol blieb so groß, wie es war, und eine Zeile
// mit Symbol passte nie.
func TestEinVerkleinerterTextVerkleinertSeineSymbole(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{"/1.png": {X: 10, Y: 10}}}

	// Eine Zeile misst bei 10 Punkten 12 hoch; der Kasten lässt 6 — passend
	// erst um 5 Punkte, und nur, wenn das Symbol mitschrumpft (10 → 5).
	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10, Kleinstgroesse: 5, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 200, Hoehe: 6},
		[]render.Abschnitt{{
			Text:        "1",
			Symbole:     []render.Symbol{{Bild: render.Blatt("/1.png")}},
			Symbolmasse: render.Symbolmasse{Groesse: 10, Bildgroesse: 10},
		}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}},
		quelle,
	)

	// **Es passt** — ohne das mitschrumpfende Symbol bliebe die Zeile 10 hoch.
	if ergebnis.Hoehe > 6 || ergebnis.Groesse >= 6 {
		t.Errorf("gesetzt in %g Punkten und %g hoch, erwartet unter 6 Punkte und höchstens 6 hoch",
			ergebnis.Groesse, ergebnis.Hoehe)
	}

	if breite := ergebnis.Zeilen[0].Stuecke[0].Breite; breite != 5 {
		t.Errorf("das Symbol ist %g breit, erwartet 5", breite)
	}
}
