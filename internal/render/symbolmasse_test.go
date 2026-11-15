package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// bildmasse ist eine Bildquelle, deren Bilder eine **genannte** Größe haben.
//
// **Nicht die 8 × 8 aus [bildquelle].** Die Frage dieser Datei ist, ob die
// Bildgröße durchschlägt; eine Quelle, die für jedes Bild dasselbe Quadrat
// liefert, könnte sie nicht stellen.
type bildmasse struct{ da map[string]image.Point }

func (b bildmasse) Bild(pfad string) (image.Image, error) {
	mass, gibtEs := b.da[pfad]
	if !gibtEs {
		return nil, errKeinBild
	}

	bild := image.NewRGBA(image.Rect(0, 0, mass.X, mass.Y))
	for y := range mass.Y {
		for x := range mass.X {
			bild.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	return bild, nil
}

// **Ein Symbol ist so groß wie sein Bild, umgerechnet — nicht quadratisch.**
//
// `SymbolInFont::size` rechnet die tatsächliche Bildgröße im Verhältnis
// `font_size / image font size` um und legt `horizontal space` dazu
// (`src/data/symbol_font.cpp`, Zeilen 147–153 und 384–401, Fassung f9d9356 vom
// 2021-07-06). Ein Bild von 20 × 10 bei `image font size: 10` und Schriftgröße
// 5 misst also 10 × 5; mit `horizontal space: 1` belegt es einen Kasten von
// 11 × 5.
//
// Bis zum 2026-08-17 rechnete dieses Programm ein Quadrat der Schriftgröße —
// hier also 5 × 5 statt 11 × 5
// (#330).
func TestEinSymbolkastenFolgtDemBildUndSeinemAbstand(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{"/breit.png": {X: 20, Y: 10}}}

	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 200, Hoehe: 40},
		[]render.Abschnitt{{
			Text:    "R",
			Symbole: []render.Symbol{{Bild: render.Blatt("/breit.png")}},
			Symbolmasse: render.Symbolmasse{
				Groesse: 5, Bildgroesse: 10, AbstandWaagerecht: 1,
			},
		}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}},
		quelle,
	)

	if len(ergebnis.Zeilen) != 1 || len(ergebnis.Zeilen[0].Stuecke) != 1 {
		t.Fatalf("gesetzt wurden %+v", ergebnis.Zeilen)
	}

	if breite := ergebnis.Zeilen[0].Stuecke[0].Breite; breite != 11 {
		t.Errorf("das Symbol ist %g breit, erwartet 11", breite)
	}
}

// **Die eigene `image font size` eines Symbols schlägt die der Schrift.**
//
// `magic-mana-large` nennt 135 und überschreibt sie an einzelnen Symbolen mit
// 120 oder 145 (`SymbolInFont::SymbolInFont`, Zeilen 111–113). Wer nur die
// Schrift läse, machte die Hybridsymbole um ein Zehntel zu groß.
func TestDieEigeneBildgroesseEinesSymbolsSchlaegtDieDerSchrift(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{"/s.png": {X: 20, Y: 20}}}

	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 200, Hoehe: 40},
		[]render.Abschnitt{{
			Text:        "R",
			Symbole:     []render.Symbol{{Bild: render.Blatt("/s.png"), Bildgroesse: 20}},
			Symbolmasse: render.Symbolmasse{Groesse: 5, Bildgroesse: 10},
		}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}},
		quelle,
	)

	// 20 × 5/20 = 5. Mit der Schriftangabe 10 wären es 20 × 5/10 = 10.
	if breite := ergebnis.Zeilen[0].Stuecke[0].Breite; breite != 5 {
		t.Errorf("das Symbol ist %g breit, erwartet 5", breite)
	}
}

// **Ein Symbol, das höher ist als die Zeile, macht die Zeile höher.**
//
// MSE nimmt die Zeilenhöhe aus den `CharInfo` der Elemente, und ein
// Symbolelement meldet seinen ganzen Kasten (`SymbolFont::getCharInfo`, Zeilen
// 384–393). Ohne das liefe ein Manasymbol im Regeltext in die Zeile darunter.
func TestEinHohesSymbolMachtDieZeileHoeher(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{"/hoch.png": {X: 10, Y: 40}}}

	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 200, Hoehe: 100},
		[]render.Abschnitt{{
			Text:        "R",
			Symbole:     []render.Symbol{{Bild: render.Blatt("/hoch.png")}},
			Symbolmasse: render.Symbolmasse{Groesse: 10, Bildgroesse: 10},
		}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}},
		quelle,
	)

	// Das Bild ist 40 hoch, umgerechnet 40 × 10/10 = 40. Die Schrift gäbe 10.
	if hoehe := ergebnis.Hoehe; hoehe != 40 {
		t.Errorf("die Zeile ist %g hoch, erwartet 40", hoehe)
	}
}

// **Ohne `image font size` bleibt es beim Quadrat.**
//
// MSE setzt `img_size = 1`, wo die Angabe fehlt, und zeichnete das Bild damit
// in Bildpunktgröße — 2.475 Punkte für ein Manasymbol. Dem folgt dieses
// Programm bewusst nicht; die Begründung steht in
// `docs/entscheidungen/0022-der-rasterer-misst-die-symbole.md`.
func TestOhneBildgroesseBleibtEsBeimQuadrat(t *testing.T) {
	quelle := bildmasse{da: map[string]image.Point{"/s.png": {X: 100, Y: 50}}}

	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 200, Hoehe: 40},
		[]render.Abschnitt{{
			Text:        "R",
			Symbole:     []render.Symbol{{Bild: render.Blatt("/s.png")}},
			Symbolmasse: render.Symbolmasse{Groesse: 7},
		}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}},
		quelle,
	)

	if breite := ergebnis.Zeilen[0].Stuecke[0].Breite; breite != 7 {
		t.Errorf("das Symbol ist %g breit, erwartet 7", breite)
	}
}

// **Ein Bild, das die Quelle nicht hergibt, bleibt ein Quadrat.**
//
// Gemessen wird, bevor gezeichnet wird; ein fehlendes Bild darf den Umbruch
// nicht auf null Breite setzen, sonst rutscht der ganze Rest der Zeile nach
// links und der Befund beim Zeichnen zeigt auf ein Bild, das dort gar nicht
// mehr steht.
func TestEinUnlesbaresSymbolbildBleibtEinQuadrat(t *testing.T) {
	ergebnis := render.SetzeAbschnitte(
		render.Satz{Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 200, Hoehe: 40},
		[]render.Abschnitt{{
			Text:        "R",
			Symbole:     []render.Symbol{{Bild: render.Blatt("/fehlt.png")}},
			Symbolmasse: render.Symbolmasse{Groesse: 6, Bildgroesse: 10},
		}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}},
		bildmasse{da: map[string]image.Point{}},
	)

	if breite := ergebnis.Zeilen[0].Stuecke[0].Breite; breite != 6 {
		t.Errorf("das Symbol ist %g breit, erwartet 6", breite)
	}
}
