// Die Attrappen, mit denen die SVG-Tests arbeiten.
//
// **In einer eigenen Datei**, weil sie von jedem Lauf hier gebraucht werden
// und die Frage, *was* geprüft wird, von der Frage trennen, *womit*.
package svg_test

import (
	"errors"
	"image"
	"image/color"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// probeplan gibt einen Plan mit je einer Ebene jeder Art.
func probeplan() *render.Plan {
	return &render.Plan{
		Breite: 375,
		Hoehe:  523,
		Ebenen: []render.Ebene{
			{
				Art:     render.ArtFlaeche,
				Feld:    "grund",
				Rahmen:  render.Rechteck{Links: 0, Oben: 0, Breite: 375, Hoehe: 523},
				Farbe:   render.Farbe{R: 255, G: 255, B: 255, A: 255},
				Deckung: 1,
			},
			{
				Art:     render.ArtBild,
				Feld:    "rahmen",
				Rahmen:  render.Rechteck{Links: 0, Oben: 0, Breite: 375, Hoehe: 523},
				Bild:    render.Blatt("/probe.mse-style/rahmen.png"),
				Deckung: 1,
			},
			{
				Art:    render.ArtText,
				Feld:   "name",
				Rahmen: render.Rechteck{Links: 24, Oben: 24, Breite: 327, Hoehe: 28},
				Farbe:  render.Farbe{A: 255},
				Abschnitte: []render.Abschnitt{
					{Text: "Feuerball"},
					{Text: " (kursiv)", Kursiv: true},
					{Text: " (fett)", Fett: true},
				},
				Satz:    render.Satz{Schrift: "Helvetica", Groesse: 20},
				Deckung: 1,
			},
		},
	}
}

// baumprobe gibt für jeden Baum ein winziges Bild — oder nichts.
type baumprobe struct {
	fehlt bool

	// ohneFlaeche gibt ein Bild ohne Ausdehnung.
	//
	// **Das ist der einzige Weg, den Kodierer scheitern zu lassen** —
	// nachgemessen am 2026-08-22: `png.Encode` gibt für 0×0 `invalid image
	// size`. Ohne diesen Fall bliebe der Fehlerweg unbetreten, und damit
	// wäre er eine Behauptung darüber, was im Ernstfall geschieht (§3).
	ohneFlaeche bool

	// warnt gibt ein Bild **und** einen Befund — so wie eine noch nicht
	// gebaute Verknüpfung, die ihr Ausgangsbild liefert.
	warnt bool
}

func (b baumprobe) BildAusBaum(render.Bildbaum, int, int) (
	bild *image.NRGBA, befund string,
) {
	switch {
	case b.fehlt:
		return nil, "das Bild gibt es nicht"

	case b.ohneFlaeche:
		return image.NewNRGBA(image.Rect(0, 0, 0, 0)), ""
	}

	bild = image.NewNRGBA(image.Rect(0, 0, 2, 2))
	bild.SetNRGBA(0, 0, color.NRGBA{R: 200, A: 255})

	if b.warnt {
		return bild, "die Verknüpfung ist hier nicht gebaut"
	}

	return bild, ""
}

// schriftprobe ist eine Breitentabelle ohne Glyphen.
//
// **Ohne Glyphen genügt**: gesetzt wird mit jeder [render.Schrift], und der
// SVG-Schreiber zeichnet keine — der Betrachter tut es.
type schriftprobe struct{}

func (schriftprobe) Breite(rune) float64  { return 0.5 }
func (schriftprobe) Zeilenhoehe() float64 { return 1.2 }

// suchprobe gibt für jeden Namen dieselbe Schrift — oder einen Fehler.
type suchprobe struct{ fehlt bool }

func (s suchprobe) Suche(string, bool) (render.Schrift, error) {
	if s.fehlt {
		return nil, errors.New("die Schrift gibt es nicht")
	}

	return schriftprobe{}, nil
}

// mitAllem gibt die Quellen, mit denen alles geschrieben werden kann.
func mitAllem() svg.Quellen {
	return svg.Quellen{Baeume: baumprobe{}, Schriften: suchprobe{}}
}
