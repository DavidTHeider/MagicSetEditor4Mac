package render_test

// **Fehlt ein Zeichen, wird auf eine andere Schrift ausgewichen**
// (#515).
//
// MSE weicht aus, weil wxWidgets es tut: ein Stil nennt eine Schrift, und
// fehlt darin ein Zeichen, sucht die Plattform eine andere — unter macOS Core
// Text. Dieses Programm rastert selbst und hatte diesen Weg nicht.
//
// # Woran es aufgefallen ist
//
// Über alle 23.168 kanonischen Karten fehlte **ein** Zeichen: `ǵ` (U+01F5),
// und es fehlte 159-mal — der größte einzelne Posten des Laufs.

import (
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// luecke ist eine Schrift, der genau ein Zeichen fehlt.
type luecke struct {
	einheitsschrift

	ohne rune
}

func (l luecke) Maske(zeichen rune, groesse float64) (image.Image, image.Point, bool) {
	if zeichen == l.ohne {
		return nil, image.Point{}, false
	}

	return l.einheitsschrift.Maske(zeichen, groesse)
}

func TestEinFehlendesZeichenWeichtAus(t *testing.T) {
	const fehlt = 'ǵ'

	plan := func() *render.Plan {
		return &render.Plan{Breite: 20, Hoehe: 12, Ebenen: []render.Ebene{{
			Art:        render.ArtText,
			Rahmen:     render.Rechteck{Breite: 20, Hoehe: 12},
			Abschnitte: []render.Abschnitt{{Text: string(fehlt)}},
			Farbe:      render.Farbe{R: 255, A: 255},
			Satz: render.Satz{
				Schrift: "Probe", Groesse: 6, Ausrichtung: render.LinksOben,
			},
			Deckung: 1,
		}}}
	}

	faelle := []struct {
		name       string
		mitErsatz  bool
		erwartetOK bool
	}{
		{"ohne Ersatz fehlt es", false, false},
		{"mit Ersatz steht es da", true, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			quelle := schriften{"Probe": luecke{ohne: fehlt}}

			rasterer := render.NeuerRasterer(bildmasse{}).MitSchriften(quelle)
			if fall.mitErsatz {
				rasterer = rasterer.MitErsatzschrift(
					func(rune) (render.Glyphenschrift, bool) {
						return einheitsschrift{}, true
					})
			}

			bild, befunde, err := rasterer.Zeichne(t.Context(), plan(), 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			var gemalt int

			for y := range 12 {
				for x := range 20 {
					if bild.NRGBAAt(x, y).A > 0 {
						gemalt++
					}
				}
			}

			if fall.erwartetOK {
				if gemalt == 0 {
					t.Error("nichts gezeichnet — der Ersatz greift nicht")
				}

				if len(befunde) != 0 {
					t.Errorf("Befunde: %v — ein ausgewichenes Zeichen fehlt nicht", befunde)
				}

				return
			}

			if gemalt != 0 {
				t.Error("gezeichnet, obwohl das Zeichen fehlt")
			}

			if len(befunde) != 1 {
				t.Errorf("%d Befunde, erwartet einer", len(befunde))
			}
		})
	}
}
