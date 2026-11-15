package render_test

// **`<soft-line>` bricht weich um**
// (#431).
//
//	LineBreak line_break = line > 0      ? LineBreak::LINE :
//	                       soft_line > 0 ? LineBreak::SOFT : LineBreak::HARD;
//
// `src/render/text/element.cpp`, Zeilen 235–236, Fassung f9d9356 vom
// 2021-07-06. Ein `\n` innerhalb eines `<soft-line>` ist damit **kein** harter
// Umbruch — und `addParagraphs` übergeht ihn zusätzlich ganz (ebd. Zeile 274),
// womit er auch keinen Absatz schließt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEineWeicheZeileBrichtWeichUm(t *testing.T) {
	// **Zwei Faktoren, die weit auseinanderliegen**: sonst sagte der Test
	// nicht, welcher genommen wurde.
	satz := render.Satz{
		Groesse: 10, Ausrichtung: render.LinksOben,
		Zeilenfaktoren: render.Zeilenfaktoren{Weich: 0.5, Hart: 2},
	}
	rahmen := render.Rechteck{Breite: 100, Hoehe: 100}

	hart := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "a\nb"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	weich := render.SetzeAbschnitte(satz, rahmen,
		[]render.Abschnitt{{Text: "a\nb", WeicheZeile: true}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(hart.Zeilen) != 2 || len(weich.Zeilen) != 2 {
		t.Fatalf("%d harte und %d weiche Zeilen, erwartet je zwei",
			len(hart.Zeilen), len(weich.Zeilen))
	}

	// Die Zeilenhöhe ist 10 × 1,2 = 12; hart ergibt 24, weich 6.
	if hart.Zeilen[1].Oben != 24 {
		t.Errorf("hart: zweite Zeile bei %g, erwartet 24", hart.Zeilen[1].Oben)
	}

	if weich.Zeilen[1].Oben != 6 {
		t.Errorf("weich: zweite Zeile bei %g, erwartet 6", weich.Zeilen[1].Oben)
	}

	// **Und der Absatz bleibt einer.** `addParagraphs` kehrt bei
	// `soft_line > 0` sofort zurück; der Umbruch schließt nichts.
	if len(weich.Absaetze) != 1 {
		t.Errorf("%d Absätze, erwartet einer", len(weich.Absaetze))
	}

	if len(hart.Absaetze) != 2 {
		t.Errorf("hart: %d Absätze, erwartet zwei", len(hart.Absaetze))
	}
}
