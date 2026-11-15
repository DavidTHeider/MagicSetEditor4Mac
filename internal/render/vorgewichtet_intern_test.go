package render

import (
	"image"
	"image/color"
	"testing"
)

// **Der Schnellweg rechnet Zeichen für Zeichen dasselbe wie `At(…).RGBA()`**
// (#672).
//
// Nicht „ungefähr dasselbe": bitgleich. Die Gegenprobe zu dieser Zusicherung
// ist teuer bezahlt — der erste Anlauf verbreiterte das Alpha vor der
// Gewichtung und war damit um den Faktor 257 daneben.
//
// **Über den ganzen Wertebereich, nicht an drei Stellen.** Ein Fehler in der
// Rundung zeigt sich an den mittleren Werten, nicht an 0 und 255.
func TestDerSchnellwegGibtDasselbe(t *testing.T) {
	t.Parallel()

	stufen := []uint8{0, 1, 2, 17, 63, 64, 127, 128, 129, 200, 254, 255}

	t.Run("NRGBA", func(t *testing.T) {
		t.Parallel()

		for _, a := range stufen {
			for _, kanal := range stufen {
				bild := image.NewNRGBA(image.Rect(0, 0, 1, 1))
				bild.SetNRGBA(0, 0, color.NRGBA{R: kanal, G: kanal / 2, B: 255 - kanal, A: a})

				vergleiche(t, bild)
			}
		}
	})

	t.Run("RGBA", func(t *testing.T) {
		t.Parallel()

		for _, a := range stufen {
			for _, kanal := range stufen {
				bild := image.NewRGBA(image.Rect(0, 0, 1, 1))
				// **Vorgewichtet, also nie über dem Alpha** — ein RGBA mit
				// R > A ist kein gültiger Wert, und was `At` daraus macht,
				// ist keine Zusage, die dieser Schnellweg halten müsste.
				bild.SetRGBA(0, 0, color.RGBA{R: min(kanal, a), G: min(kanal/2, a), B: 0, A: a})

				vergleiche(t, bild)
			}
		}
	})

	// **Beide Arten von Tafeleintrag.** Der PNG-Leser gibt `color.NRGBA`,
	// aber eine Tafel darf jede Farbe führen — und `color.RGBA` ist die
	// zweite, die ohne Umweg gelesen wird.
	t.Run("mit Farbtafel aus NRGBA", func(t *testing.T) {
		t.Parallel()

		for _, a := range stufen {
			tafel := color.Palette{color.NRGBA{R: 10, G: 20, B: 30, A: a}}
			bild := image.NewPaletted(image.Rect(0, 0, 1, 1), tafel)

			vergleiche(t, bild)
		}
	})

	t.Run("mit Farbtafel aus RGBA", func(t *testing.T) {
		t.Parallel()

		for _, a := range stufen {
			tafel := color.Palette{
				color.RGBA{R: min(10, a), G: min(20, a), B: min(30, a), A: a},
			}
			bild := image.NewPaletted(image.Rect(0, 0, 1, 1), tafel)

			vergleiche(t, bild)
		}
	})

	// **Und eine Art, für die es keinen Schnellweg gibt**, geht den alten Weg.
	t.Run("Graustufen gehen den alten Weg", func(t *testing.T) {
		t.Parallel()

		for _, stufe := range stufen {
			bild := image.NewGray(image.Rect(0, 0, 1, 1))
			bild.SetGray(0, 0, color.Gray{Y: stufe})

			vergleiche(t, bild)
		}
	})
}

// vergleiche hält den Schnellweg gegen `At(…).RGBA()`.
func vergleiche(t *testing.T, bild image.Image) {
	t.Helper()

	sollR, sollG, sollB, sollA := bild.At(0, 0).RGBA()
	istR, istG, istB, istA := vorgewichteteWerte(bild, 0, 0)

	// **Ein durchsichtiger Punkt gibt vier Nullen**, und das ist der einzige
	// Unterschied: der Aufrufer wirft ihn ohnehin weg, und `At` gibt dort je
	// nach Art auch schon Nullen.
	if sollA == 0 && istR == 0 && istG == 0 && istB == 0 && istA == 0 {
		return
	}

	if istR != sollR || istG != sollG || istB != sollB || istA != sollA {
		t.Errorf("%T: Schnellweg (%d,%d,%d,%d), At (%d,%d,%d,%d)",
			bild, istR, istG, istB, istA, sollR, sollG, sollB, sollA)
	}
}
