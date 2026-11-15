package werkzeuge_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// **Die Grafik der Saat ist gezeichnet, nicht gefunden.**
//
// # Warum sie überhaupt ein Wächter ist
//
// [0005](../../docs/entscheidungen/0005-keine-fremden-vorlagen.md) verbietet
// fremde Vorlagen im Repository, und `LICENSE` sagt, dass Inhalte Dritter nicht
// unter MIT stehen. Eine Grafik, die einfach dalag, ließe sich davon nicht
// unterscheiden: ein PNG sieht einem PNG gleich, und „selbst gemacht" ist eine
// Behauptung, die niemand nachprüfen kann.
//
// Hier steht sie als **Rechenvorschrift**. Wer wissen will, woher die Datei
// kommt, liest die dreißig Zeilen darunter — und wer sie austauscht, wird rot.
//
// # Warum nicht ein Werkzeug, das man einmal fährt
//
// Ein Erzeuger neben dem Repository wäre nach dem ersten Lauf wieder eine
// Behauptung: er sagt, wie die Datei entstanden **wäre**. Als Wächter sagt er,
// wie sie ist.
func TestDieSaatgrafikIstSelbstGezeichnet(t *testing.T) {
	saat := filepath.Join("..", "..", "saat")

	for _, fall := range []struct {
		paket, name string
		gezeichnet  image.Image
	}{
		{"probe-bild.mse-style", "rahmen.png", saatrahmen()},
		{"probe-bild.mse-style", "rand_maske.png", saatmaske()},

		// **Die Wahlbilder der Stiloption `hintergrund`**
		// (#1225).
		{"probe-schlicht.mse-style", "hell.png", saatwahl(wahlHell)},
		{"probe-schlicht.mse-style", "dunkel.png", saatwahl(wahlDunkel)},
	} {
		t.Run(fall.paket+"/"+fall.name, func(t *testing.T) {
			haltePNG(t, filepath.Join(saat, fall.paket, fall.name), fall.gezeichnet)
		})
	}
}

// **Die zwei Farben der Wahlbilder** — dieselben, die `grund` rechnet.
//
// Ein Wahlbild zeigt, was die Wahl bewirkt; wer hier eine dritte Farbe
// hinschriebe, zeigte etwas anderes, als die Karte danach tut.
var (
	wahlHell   = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	wahlDunkel = color.NRGBA{R: 38, G: 44, B: 52, A: 255}
)

// saatwahl zeichnet ein Wahlbild: eine Fläche mit dunklem Rand.
//
// **Quadratisch und klein.** MSE rastert seine Wahlbilder auf 18 × 18
// (`thumbnail_size`, `src/gui/value/choice.cpp`, Zeile 17, Fassung f9d9356 vom
// 2021-07-06); hier stehen 24 Punkte, damit die Datei auch auf einem
// Bildschirm mit doppelter Auflösung nicht geraten werden muss.
func saatwahl(fuellung color.NRGBA) image.Image {
	const kante = 24

	rand := color.NRGBA{R: 30, G: 41, B: 59, A: 255}

	bild := image.NewNRGBA(image.Rect(0, 0, kante, kante))

	for y := range kante {
		for x := range kante {
			farbe := fuellung
			if x < 2 || y < 2 || x >= kante-2 || y >= kante-2 {
				farbe = rand
			}

			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild
}

// haltePNG hält eine Datei gegen ein gezeichnetes Bild.
//
// **Fehlt sie, wird sie angelegt — und der Lauf ist trotzdem rot.** Niemand
// soll eine Grafik erben, ohne sie gesehen zu haben; das ist dieselbe Regel wie
// beim Goldbild.
func haltePNG(t *testing.T, pfad string, gezeichnet image.Image) {
	t.Helper()

	var frisch bytes.Buffer
	if err := png.Encode(&frisch, gezeichnet); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	daliegend, err := os.ReadFile(pfad) // #nosec G304 -- fester Pfad im Repository
	if err != nil {
		if fehler := os.MkdirAll(filepath.Dir(pfad), 0o750); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		if fehler := os.WriteFile(pfad, frisch.Bytes(), 0o600); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}

		t.Fatalf("die Grafik fehlte und wurde gezeichnet: %s — bitte ansehen", pfad)
	}

	if !bytes.Equal(daliegend, frisch.Bytes()) {
		t.Errorf("%s stimmt nicht mit der Vorschrift überein — %d statt %d Bytes",
			pfad, len(daliegend), frisch.Len())
	}
}

// Die Maße der Saatkarte, in Punkten der Vorlage.
const (
	saatBreite = 375
	saatHoehe  = 523
	saatRand   = 24
	saatRadius = 18
)

// saatrahmen zeichnet den Kartenrahmen: eine dunkle Fläche mit heller Mitte.
//
// **Zwei Rechtecke und sonst nichts.** Die Saat zeigt, dass der Weg vom Stil
// zum Bild trägt, nicht wie eine schöne Karte aussieht — und jede Linie mehr
// wäre eine Linie, die der Goldbildvergleich mitschleppt.
func saatrahmen() image.Image {
	dunkel := color.NRGBA{R: 30, G: 41, B: 59, A: 255}
	hell := color.NRGBA{R: 248, G: 250, B: 252, A: 255}

	bild := image.NewNRGBA(image.Rect(0, 0, saatBreite, saatHoehe))

	for y := range saatHoehe {
		for x := range saatBreite {
			farbe := dunkel

			if x >= saatRand && x < saatBreite-saatRand &&
				y >= saatRand && y < saatHoehe-saatRand {
				farbe = hell
			}

			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild
}

// saatmaske zeichnet die Maske: weiß innerhalb der abgerundeten Karte.
//
// **Grau an der Kante, nicht schwarz oder weiß.** Eine harte Kante ergäbe eine
// Treppe an jeder Ecke, und der Rasterer könnte sie nicht glätten — er sieht
// nur die Maske, die er bekommt. Gemittelt wird über 4 × 4 Proben je Bildpunkt.
func saatmaske() image.Image {
	bild := image.NewGray(image.Rect(0, 0, saatBreite, saatHoehe))

	for y := range saatHoehe {
		for x := range saatBreite {
			anteil := deckung(float64(x), float64(y))
			bild.SetGray(x, y, color.Gray{Y: uint8(math.Round(anteil * 255))})
		}
	}

	return bild
}

// deckung sagt, wie viel eines Bildpunkts innerhalb der abgerundeten Ecke liegt.
func deckung(x, y float64) float64 {
	const proben = 4

	var innen float64

	for senkrecht := range proben {
		for waagerecht := range proben {
			pruefX := x + (float64(waagerecht)+0.5)/proben
			pruefY := y + (float64(senkrecht)+0.5)/proben

			if drin(pruefX, pruefY) {
				innen++
			}
		}
	}

	return innen / (proben * proben)
}

// drin sagt, ob ein Punkt in der abgerundeten Kartenfläche liegt.
func drin(x, y float64) bool {
	// Die Mitte des nächstgelegenen Eckkreises — waagerecht wie senkrecht.
	mitteX := math.Min(math.Max(x, saatRadius), saatBreite-saatRadius)
	mitteY := math.Min(math.Max(y, saatRadius), saatHoehe-saatRadius)

	return math.Hypot(x-mitteX, y-mitteY) <= saatRadius
}
