package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Farbfeld zeichnet ein Rechteck mit runden Ecken — und oft nur dessen
// Rand.**
//
//	bool clip = style().left_width < style().width  && style().right_width  < style().width &&
//	      style().top_width  < style().height && style().bottom_width < style().height;
//	if (clip) {
//	  wxRegion r = dc.trRectToRegion(dc.getInternalRect());
//	  r.Subtract(dc.trRectToRegion(RealRect(
//	    style().left_width, style().top_width,
//	    style().width  - style().left_width - style().right_width,
//	    style().height - style().top_width  - style().bottom_width)));
//	  dc.getDC().SetClippingRegion(r);
//	}
//	dc.DrawRoundedRectangle(dc.getInternalRect(), style().radius);
//
// `ColorValueViewer::draw`, `src/render/value/color.cpp`, Zeilen 47–64, Fassung
// f9d9356 vom 2021-07-06.
//
// **Daran hängt der schwarze Kartenrand.** `magic-m15.mse-style` gibt dem Feld
// `border color` die Breiten 17/17/17/18 und einen Radius von 18: ein Rahmen,
// kein Klotz (#270).
func TestEineFlaecheZeichnetRundeEckenUndEinenRand(t *testing.T) {
	const seite = 40

	flaeche := func(ecke float64, rand *render.Randbreiten) *render.Plan {
		return &render.Plan{Breite: seite, Hoehe: seite, Ebenen: []render.Ebene{{
			Art:     render.ArtFlaeche,
			Rahmen:  render.Rechteck{Breite: seite, Hoehe: seite},
			Farbe:   render.Farbe{R: 255, A: 255},
			Deckung: 1,
			Ecke:    ecke,
			Rand:    rand,
		}}}
	}

	gemalt := func(plan *render.Plan, x, y int) bool {
		t.Helper()

		bild, _, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), plan, 1, render.GueteMittel)
		if err != nil {
			t.Fatalf("zeichnen: %v", err)
		}

		return bild.NRGBAAt(x, y).A > 0
	}

	t.Run("runde Ecken lassen den Zwickel frei", func(t *testing.T) {
		plan := flaeche(10, nil)

		if gemalt(plan, 0, 0) {
			t.Error("die Ecke wurde gemalt")
		}

		if !gemalt(plan, seite/2, seite/2) {
			t.Error("die Mitte blieb leer")
		}

		// **Auf der Kante liegt Farbe** — die Rundung schneidet nur die Ecke.
		if !gemalt(plan, seite/2, 0) {
			t.Error("die Kantenmitte blieb leer")
		}
	})

	// **Ein Radius unter einem halben Bildpunkt ist keiner.** `radius: 0.4`
	// stünde sonst für eine Ecke, die kein Bildpunkt breit ist — und die Prüfung
	// auf den Kreis schnitte den Punkt (0,0) weg, den MSE malt.
	t.Run("ein winziger Radius schneidet nichts", func(t *testing.T) {
		if !gemalt(flaeche(0.4, nil), 0, 0) {
			t.Error("die Ecke wurde weggeschnitten")
		}
	})

	t.Run("die Breiten machen daraus einen Rahmen", func(t *testing.T) {
		plan := flaeche(0, &render.Randbreiten{Links: 5, Oben: 5, Rechts: 5, Unten: 5})

		if !gemalt(plan, 2, seite/2) {
			t.Error("der linke Rand blieb leer")
		}

		if gemalt(plan, seite/2, seite/2) {
			t.Error("die Mitte wurde gemalt")
		}
	})

	// **Eine Breite von null ist keine Rahmung dieser Seite.** `stellen(0)` ist
	// null, und das innere Rechteck reicht dann bis an die Kante.
	t.Run("eine Breite von null lässt die Kante frei", func(t *testing.T) {
		plan := flaeche(0, &render.Randbreiten{Links: 0, Oben: 5, Rechts: 5, Unten: 5})

		if gemalt(plan, 0, seite/2) {
			t.Error("die linke Kante wurde gemalt")
		}
	})

	t.Run("eine Breite über das Feld hinaus füllt ganz", func(t *testing.T) {
		// **MSEs Vorgabe ist 100000**, und die Bedingung `left_width <
		// width` ist dann falsch: ohne eigene Breiten wird gefüllt und nicht
		// gerahmt.
		plan := flaeche(0, &render.Randbreiten{
			Links: 100000, Oben: 100000, Rechts: 100000, Unten: 100000,
		})

		if !gemalt(plan, seite/2, seite/2) {
			t.Error("die Mitte blieb leer")
		}
	})
}
