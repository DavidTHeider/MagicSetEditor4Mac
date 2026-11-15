package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Anfang und Schritt einer Bildreihe** — `align_in_rect` mit der Größe null
// und `move_in_direction`, `src/util/alignment.cpp`, Zeilen 15–33 und
// 99–124, Fassung f9d9356 vom 2021-07-06.
func TestReihenanfangUndSchritt(t *testing.T) {
	rahmen := render.Rechteck{Links: 10, Oben: 20, Breite: 100, Hoehe: 40}

	anfaenge := map[string][2]float64{
		"":             {10, 20},
		"top center":   {60, 20},
		"bottom right": {110, 60},
		"middle left":  {10, 40},
		"stretch":      {10, 20},
	}

	for ausrichtung, erwartet := range anfaenge {
		if x, y := reihenanfang(rahmen, ausrichtung); x != erwartet[0] || y != erwartet[1] {
			t.Errorf("%q beginnt bei (%v, %v), erwartet %v", ausrichtung, x, y, erwartet)
		}
	}

	schritte := map[string][2]float64{
		"":                         {28, 0},
		"left to right":            {28, 0},
		"horizontal":               {28, 0},
		"right to left":            {-28, 0},
		"top to bottom":            {0, 13},
		"vertical":                 {0, 13},
		"bottom to top":            {0, -13},
		"top-right to bottom-left": {-28, 13},
		"bottom-left to top-right": {28, -13},
		"top-left to bottom-right": {28, 13},
		"bottom-right to top-left": {-28, -13},
	}

	for richtung, erwartet := range schritte {
		if x, y := inRichtung(richtung, 0, 0, 20, 5, 8); x != erwartet[0] || y != erwartet[1] {
			t.Errorf("%q rückt nach (%v, %v), erwartet %v", richtung, x, y, erwartet)
		}
	}
}

// **Ein Wahlbild, dessen Größe hier niemand kennt, wird benannt.**
func TestEinWahlbildOhneBekannteGroesseIstEinBefund(t *testing.T) {
	gerechnet := render.Bildbaum{Knotenart: "masked_blend"}
	if _, befund := wahlbildebene("f", "w", gerechnet, 0, 0, nil); befund == nil {
		t.Error("ein gerechnetes Wahlbild ging ohne Befund durch")
	}

	datei := render.Bildbaum{Knotenart: "datei", Pfad: "/p.mse-style/a.png"}
	if _, befund := wahlbildebene("f", "w", datei, 0, 0, nil); befund == nil {
		t.Error("ohne Bildquelle ging das Wahlbild ohne Befund durch")
	}
}
