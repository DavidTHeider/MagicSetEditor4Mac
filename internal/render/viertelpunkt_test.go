package render_test

// **Eine Schriftgröße liegt auf dem Raster von ¼ Punkt** — #2302.
//
// MSE2 setzt jede Schrift als wxFont mit ganzzahliger Punktgröße, gemessen
// und gezeichnet in vierfacher Auflösung: `to_int(4 · Größe · Maßstab)`
// (`Font::toWxFont`, `src/data/font.cpp`, Zeile 84; `text_scaling = 4`,
// `src/gfx/resample_text.cpp`, Zeile 22; `RotatedDC::SetFont`,
// `src/util/rotation.cpp`, Zeile 326 — Fassung f9d9356 vom 2021-07-06). Ein
// verkleinerter Text steht dort also bis zu ⅛ Punkt kleiner oder größer als
// gerechnet — bei 14 Punkt fast ein Prozent, und genau so viel enger stand
// Clockwork Beast bei MSE2 da.

import (
	"math"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEineSchriftgroesseLiegtAufDemViertelpunktraster(t *testing.T) {
	// breite ist die der Zeile „ab" bei halbem Geviert je Zeichen — so groß
	// wie die Schrift, und selbst auf ¼ Punkt gerastert.
	for name, fall := range map[string]struct{ groesse, erwartet, breite float64 }{
		"abgerundet":  {groesse: 13.3, erwartet: 13.25, breite: 13.25},
		"aufgerundet": {groesse: 13.4, erwartet: 13.5, breite: 13.5},
		"genau":       {groesse: 14, erwartet: 14, breite: 14},
		// `to_int` addiert 0,4999995 und nicht ½ (`src/util/vector2d.hpp`,
		// Zeile 20, Fassung f9d9356 vom 2021-07-06): die genaue Mitte rundet
		// ab.
		"genau auf der Mitte": {groesse: 13.375, erwartet: 13.25, breite: 13.25},
		// **Unter ⅛ Punkt bliebe null** — wx nähme dann seine Vorgabe, und
		// eine Schrift der Größe null zeichnet hier nichts. Sie bleibt, wie
		// sie ist; eine Vorlage, die so klein setzt, gibt es nicht.
		"unter einem Achtelpunkt": {groesse: 0.1, erwartet: 0.1, breite: 0.25},
	} {
		t.Run(name, func(t *testing.T) {
			satz := render.Satz{Groesse: fall.groesse}

			ergebnis := render.SetzeAbschnitte(satz, render.Rechteck{Breite: 1000, Hoehe: 100},
				[]render.Abschnitt{{Text: "ab"}}, render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

			if ergebnis.Groesse != fall.erwartet {
				t.Errorf("gesetzt mit %g, erwartet %g", ergebnis.Groesse, fall.erwartet)
			}

			// **Gemessen wird mit derselben Größe**, nicht mit der gerechneten.
			if breite := ergebnis.Zeilen[0].Breite; breite != fall.breite {
				t.Errorf("die Zeile ist %g breit, erwartet %g", breite, fall.breite)
			}
		})
	}
}

// **Auch die Maßstabssuche setzt auf dem Raster.**
func TestEinVerkleinerterTextLiegtAufDemViertelpunktraster(t *testing.T) {
	satz := render.Satz{Groesse: 14, Kleinstgroesse: 6}

	ergebnis := render.SetzeAbschnitte(satz, render.Rechteck{Breite: 60, Hoehe: 40},
		[]render.Abschnitt{{Text: "abcde fghij klmno pqrst uvwxy"}},
		render.Schriftsatz{Aufrecht: gleichbreit{breite: 0.5}}, nil)

	if ergebnis.Groesse >= 14 {
		t.Fatalf("nicht verkleinert: %g", ergebnis.Groesse)
	}

	if viertel := ergebnis.Groesse * 4; viertel != math.Round(viertel) {
		t.Errorf("verkleinert auf %g, das liegt nicht auf ¼ Punkt", ergebnis.Groesse)
	}
}
