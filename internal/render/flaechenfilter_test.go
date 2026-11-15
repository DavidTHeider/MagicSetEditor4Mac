package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// streifen baut ein Bild mit feinen waagerechten Linien.
//
// **Jede fünfte Zeile schwarz**, und die Fünf ist der Punkt: bei jeder vierten
// und einer Verkleinerung von acht liegt die Abtaststelle **immer** gleich zur
// Linie, und selbst die Punktabtastung liefert überall dasselbe Grau. Der erste
// Anlauf dieses Tests war deshalb grün, ohne etwas zu prüfen (§10).
//
// Das ist der Fall aus
// #113: eine
// Karte von 750 Punkten Breite auf eine Vorschau von 180 überspringt drei von
// vier Quellzeilen.
func streifen(hoehe int) *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, 8, hoehe))

	for y := range hoehe {
		farbe := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		if y%5 == 0 {
			farbe = color.NRGBA{A: 255}
		}

		for x := range 8 {
			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild
}

// **Beim starken Verkleinern bleibt das Grau gleichmäßig.**
//
// # Die Messung, die #113 verlangt
//
// Das Issue sagt: „Erst braucht es einen Fall, der das Flimmern zeigt." Hier
// ist er — und die Messgröße ist die **Streuung** der Zielzeilen.
//
// Mit Punktabtastung trifft jede Zielzeile eine Quellzeile: mal eine schwarze,
// mal eine weiße. Die Streuung ist maximal, und feine Linien flimmern oder
// verschwinden je nach Phase. Mit einem Flächenfilter mittelt jeder
// Zielbildpunkt über den ganzen Bereich, den er abdeckt — bei einem Viertel
// schwarzer Zeilen kommt überall dasselbe Grau heraus.
func TestFeineLinienFlimmernNichtBeimVerkleinern(t *testing.T) {
	plan := &render.Plan{
		Breite: 8, Hoehe: 8,
		Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Rahmen:  render.Rechteck{Breite: 8, Hoehe: 8},
			Bild:    render.Blatt("streifen"),
			Deckung: 1,
		}},
	}

	// Von 64 auf 8 Zeilen: jede Zielzeile deckt acht Quellzeilen ab, ein bis
	// zwei davon schwarz.
	bild, befunde, err := render.NeuerRasterer(streifenquelle{}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil || len(befunde) > 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	hellste, dunkelste := 0, 255

	for y := range bild.Bounds().Dy() {
		grau := int(bild.NRGBAAt(4, y).R)
		if grau > hellste {
			hellste = grau
		}

		if grau < dunkelste {
			dunkelste = grau
		}
	}

	// **Ein Viertel schwarz heißt: überall dasselbe Grau.** Ein Spielraum von
	// 32 Stufen lässt Rundung zu und weist Flimmern ab — bei Punktabtastung
	// liegen die Zeilen 255 auseinander.
	if hellste-dunkelste > 32 {
		t.Errorf("die Zeilen streuen um %d Stufen (%d…%d) — feine Linien flimmern",
			hellste-dunkelste, dunkelste, hellste)
	}
}

// streifenquelle gibt das Streifenbild in 64 Zeilen.
type streifenquelle struct{}

func (streifenquelle) Bild(string) (image.Image, error) { return streifen(64), nil }
