package service

// **Der Anschnitt verlängert den Rand der Karte** — I04
// (#2262).

import (
	"image"
	"image/color"
	"testing"
)

func TestDerAnschnittVerlaengertDenRandUndLaesstDieKarteStehen(t *testing.T) {
	t.Parallel()

	// Drei mal zwei Bildpunkte, jeder anders gefärbt.
	karte := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for y := range 2 {
		for x := range 3 {
			karte.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 80), G: uint8(y * 120), B: 7, A: 255})
		}
	}

	mit := mitAnschnitt(karte, 7, 6)

	if mit.Bounds() != image.Rect(0, 0, 7, 6) {
		t.Fatalf("die Fläche ist %v, erwartet 7 × 6", mit.Bounds())
	}

	for y := range 6 {
		for x := range 7 {
			// **Die Karte steht unverändert innen**, und außen steht der
			// nächste Randpunkt — nicht Weiß und nicht eine gestreckte Karte.
			quelle := karte.NRGBAAt(min(max(x-2, 0), 2), min(max(y-2, 0), 1))

			if ist := mit.NRGBAAt(x, y); ist != quelle {
				t.Errorf("(%d, %d) ist %v, erwartet %v", x, y, ist, quelle)
			}
		}
	}

	// **Ohne Platz bleibt das Bild, wie es ist** — auch wenn es größer ist als
	// verlangt: kleiner wird nichts.
	if ohne := mitAnschnitt(karte, 2, 1); ohne != karte {
		t.Error("ohne Anschnitt entstand ein neues Bild")
	}

	// **Ein ungerader Rest geht nach rechts und unten.**
	ungerade := mitAnschnitt(karte, 6, 5)
	if ungerade.Bounds() != image.Rect(0, 0, 6, 5) ||
		ungerade.NRGBAAt(1, 1) != karte.NRGBAAt(0, 0) || ungerade.NRGBAAt(0, 0) != karte.NRGBAAt(0, 0) ||
		ungerade.NRGBAAt(5, 4) != karte.NRGBAAt(2, 1) || ungerade.NRGBAAt(3, 2) != karte.NRGBAAt(2, 1) {
		t.Errorf("der ungerade Rest liegt falsch: %v", ungerade.Bounds())
	}
}
