package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Das Polster schneidet den Kasten kleiner, bevor gesetzt wird.**
//
// MSE zieht es von der Fläche ab und beginnt die erste Zeile darunter:
//
//	line.top = style.padding_top;
//	lineLeft  = … + style.padding_left;
//	lineRight = … - style.padding_right;
//	RealSize s = add_diagonal(dc.getInternalSize(),
//	    -RealSize(padding_left + padding_right, padding_top + padding_bottom));
//
// `src/render/text/viewer.cpp`, Zeilen 567, 736–740 und 745–748, Fassung
// f9d9356 vom 2021-07-06. Im Pack steht `padding top:` 279-mal
// (#345).
func TestDasPolsterSchneidetDenKasten(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{
			Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
			Polster: render.Polster{Links: 6, Rechts: 4, Oben: 3},
		},
		render.Rechteck{Links: 100, Oben: 50, Breite: 60, Hoehe: 100},
		[]render.Abschnitt{{Text: "aaaa bbbb cccc"}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	// **Die Zeile beginnt hinter dem linken Polster.**
	if links := gesetzt.Zeilen[0].Links; links != 106 {
		t.Errorf("die Zeile beginnt bei %g, erwartet 106", links)
	}

	// **Und unter dem oberen.**
	if oben := gesetzt.Zeilen[0].Oben; oben != 53 {
		t.Errorf("die Zeile sitzt bei %g, erwartet 53", oben)
	}

	// **Umgebrochen wird auf 60 − 6 − 4 = 50**, also nach zwei Wörtern à 20
	// plus Lücke: „aaaa bbbb" misst 45, mit „ cccc" wären es 70.
	if len(gesetzt.Zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet zwei: %+v", len(gesetzt.Zeilen), gesetzt.Zeilen)
	}
}

// **Die Zeilenhöhe hängt daran, wie eine Zeile endet.**
//
//	double line_height_multiplier = line.break_after == LineBreak::HARD ? style.line_height_hard
//	                              : line.break_after == LineBreak::LINE ? style.line_height_line
//	                              :                                       style.line_height_soft;
//	line.top += line.line_height * line_height_multiplier;
//
// `src/render/text/viewer.cpp`, Zeilen 678–681. `magic-m15` schreibt 1,2 für
// den harten Umbruch, 1,5 für `<line>` und **0,9** für den weichen — ein
// umgebrochener Regeltext steht dort enger als zwei Absätze. Die drei Angaben
// stehen je rund 290-mal im Pack.
//
// Fassung f9d9356 vom 2021-07-06.
func TestDieZeilenhoeheHaengtAmUmbruch(t *testing.T) {
	satz := render.Satz{
		Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben,
		Zeilenfaktoren: render.Zeilenfaktoren{Weich: 0.9, Hart: 1.2, Linie: 1.5},
	}

	// Die Zeilenhöhe ist 10 × 1,2 = 12 (einheitsschrift).
	gesetzt := render.SetzeAbschnitte(satz,
		render.Rechteck{Breite: 30, Hoehe: 200},
		[]render.Abschnitt{
			{Text: "aaa bbb"},                // bricht weich um
			{Text: "\nccc", Blockende: true}, // harter Umbruch, dann Blockende
			{Text: "\nddd"},
		},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) != 4 {
		t.Fatalf("%d Zeilen, erwartet vier: %+v", len(gesetzt.Zeilen), gesetzt.Zeilen)
	}

	// Zeile 0 („aaa") endet weich       → +12 × 0,9 = 10,8
	// Zeile 1 („bbb") endet an `<line>` → +12 × 1,5 = 18    (28,8)
	// Zeile 2 („ccc") endet hart        → +12 × 1,2 = 14,4  (43,2)
	erwartet := []float64{0, 10.8, 28.8, 43.2}

	for stelle, soll := range erwartet {
		if ist := gesetzt.Zeilen[stelle].Oben; ist != soll {
			t.Errorf("Zeile %d sitzt bei %g, erwartet %g", stelle, ist, soll)
		}
	}
}

// **Ohne Angabe bleibt es bei eins.**
//
// `TextStyle` setzt die drei Faktoren auf 1,0 (`src/data/field/text.cpp`,
// Zeilen 62–64). Ein Nullwert im Verbund darf den Zug nicht auf eine Zeile
// stauchen.
//
// Fassung f9d9356 vom 2021-07-06.
func TestOhneZeilenfaktorenGiltEins(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		render.Rechteck{Breite: 100, Hoehe: 200},
		[]render.Abschnitt{{Text: "aaa\nbbb"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, nil)

	if len(gesetzt.Zeilen) != 2 || gesetzt.Zeilen[1].Oben != 12 {
		t.Errorf("die zweite Zeile sitzt bei %+v, erwartet 12", gesetzt.Zeilen)
	}
}
