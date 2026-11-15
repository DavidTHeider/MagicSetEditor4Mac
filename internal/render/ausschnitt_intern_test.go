package render

import (
	"image"
	"image/color"
	"testing"
)

// **Ein Ausschnitt ist ein Bild wie jedes andere.**
//
// Er reicht `ColorModel` durch, weil er es muss: `image.Image` verlangt es, und
// wer die Frage stellt, will wissen, in welchem Modell `At` antwortet. Eine
// Antwort aus dem Nichts — `color.RGBAModel`, weil es sich anbietet — hieße,
// dass ein Aufrufer die Farben eines NRGBA-Bildes als RGBA läse und die
// halbdurchsichtigen Stellen dunkler bekäme.
//
// **[abtasten] fragt heute nicht danach.** Genau deshalb steht dieser Test
// hier: eine Zusage, die niemand einlöst, ist eine, die beim ersten Aufrufer
// bricht (§10).
func TestEinAusschnittReichtSeinFarbmodellDurch(t *testing.T) {
	t.Parallel()

	quelle := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	quelle.SetNRGBA(2, 2, color.NRGBA{R: 10, G: 20, B: 30, A: 128})

	teil := ausschnitt{quelle, image.Rect(1, 1, 3, 3)}

	if teil.ColorModel() != quelle.ColorModel() {
		t.Error("das Farbmodell ist nicht das der Quelle")
	}

	// **Und die Grenzen sind die des Ausschnitts, die Werte die der Quelle** —
	// die Koordinaten verschieben sich nicht.
	if teil.Bounds() != image.Rect(1, 1, 3, 3) {
		t.Errorf("Grenzen %v", teil.Bounds())
	}

	farbe, ist := teil.At(2, 2).(color.NRGBA)
	if !ist || farbe.A != 128 || farbe.G != 20 {
		t.Errorf("an (2,2) steht %+v", teil.At(2, 2))
	}
}
