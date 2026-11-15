package render

import (
	"image"
	"image/color"
	"testing"
)

// **`punktfarbe` hält sich am Rand fest** — an allen vier Kanten.
//
// # Warum das ein eigener Test ist
//
// Bis zum 2026-08-19 wurde die **untere** Klemmung nebenbei erreicht: die
// bilineare Abtastung las beim Vergrößern links und oben über den Rand hinaus.
// Seit #532
// geht Vergrößerung durch das Flächenmittel, und dessen [spanne] greift nie
// darüber hinaus.
//
// **Die Zusage bleibt trotzdem** — `punktfarbe` ist ein allgemeiner Helfer und
// bekommt Indizes von mehreren Aufrufern. Sie hier zu prüfen ist ehrlicher,
// als sie über einen Aufrufer zu erreichen, der sie zufällig auslöst.
func TestPunktfarbeHaeltSichAmRandFest(t *testing.T) {
	t.Parallel()

	// Ein Teilbild ab (10,10), damit auch der versetzte Ursprung zählt.
	gross := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := range 20 {
		for x := range 20 {
			gross.SetNRGBA(x, y, color.NRGBA{R: uint8(x), G: uint8(y), B: 7, A: 255})
		}
	}

	innen := image.Rect(10, 10, 14, 14)

	faelle := []struct {
		name       string
		x, y       int
		erwarteteR uint8
		erwarteteG uint8
	}{
		{"links daneben", 5, 12, 10, 12},
		{"oben daneben", 12, 5, 12, 10},
		{"rechts daneben", 99, 12, 13, 12},
		{"unten daneben", 12, 99, 12, 13},
		{"weit links oben", -100, -100, 10, 10},
		{"innerhalb", 11, 12, 11, 12},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			farbe := punktfarbe(gross.SubImage(innen), innen, fall.x, fall.y)
			if farbe.R != fall.erwarteteR || farbe.G != fall.erwarteteG {
				t.Errorf("(%d,%d) gibt R=%d G=%d, erwartet R=%d G=%d",
					fall.x, fall.y, farbe.R, farbe.G,
					fall.erwarteteR, fall.erwarteteG)
			}
		})
	}
}
