package app

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// **Ein großes Bild geht verkleinert über die Bindung** —
// #576.
//
// Eine Digitalkamera liefert 4000 × 3000; als PNG mit Base64 sind das über
// zwanzig Megabyte durch die Wails-Brücke, für ein Fenster, das sie ohnehin
// verkleinert zeigt. MSE rechnet dafür ebenso herunter
// (`ImageSliceSelector::createBitmap`, `src/gui/image_slice_window.cpp`,
// Zeilen 610–640, Fassung f9d9356 vom 2021-07-06).
//
// **Ein interner Test**, weil `verkleinert` keine Bindung ist: was über die
// Brücke geht, ist der Data-URI, und aus ihm die Kantenlänge zurückzurechnen
// prüfte die Kodierung statt der Regel.
func TestEinGrossesBildWirdFuerDieAnzeigeVerkleinert(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		breite, hoehe        int
		erwartetB, erwartetH int
	}{
		"quer":                       {4000, 3000, 1000, 750},
		"hoch":                       {3000, 4000, 750, 1000},
		"genau an der Grenze":        {1000, 1000, 1000, 1000},
		"kleiner bleibt, wie es ist": {400, 100, 400, 100},

		// **Eine Kante bleibt stehen, auch wenn sie herunterfiele.** 20.000 × 1
		// ergäbe rechnerisch null; ein Bild ohne Höhe zeigt niemand.
		"sehr flach": {20000, 1, 1000, 1},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gab := verkleinert(image.NewNRGBA(image.Rect(0, 0, fall.breite, fall.hoehe)))
			if gab.Bounds().Dx() != fall.erwartetB || gab.Bounds().Dy() != fall.erwartetH {
				t.Errorf("gab %v, erwartet %d × %d", gab.Bounds(), fall.erwartetB, fall.erwartetH)
			}
		})
	}
}

// **Und eine Kodierung, die scheitert, wird gemeldet** (§8).
//
// Mit einem `bytes.Buffer` kann `png.Encode` nicht scheitern — deshalb steht
// die Kodierung als Naht am [App] und nicht als Aufruf im Code. Ohne sie wäre
// dieser Zweig eine Behauptung darüber, was im Ernstfall geschieht.
func TestEineGescheiterteKodierungDerQuelleWirdGemeldet(t *testing.T) {
	quelle := filepath.Join(t.TempDir(), "bild.png")

	bild := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	bild.Set(0, 0, color.NRGBA{R: 255, A: 255})

	datei, err := os.Create(quelle) //nolint:gosec // Testpfad
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := png.Encode(datei, bild); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	anwendung := NeueApp("")
	anwendung.packen = func(io.Writer, image.Image) error { return errKodierung }

	if antwort := anwendung.vorbereiten(t.Context(), quelle, 10, 10); antwort.Fehler == nil {
		t.Error("die gescheiterte Kodierung kam durch")
	}

	// **Und die Naht ist wirklich die, die sonst schreibt.**
	var gepackt bytes.Buffer
	if err := NeueApp("").packen(&gepackt, bild); err != nil || gepackt.Len() == 0 {
		t.Errorf("die Vorgabenaht packt nicht: %v, %d Bytes", err, gepackt.Len())
	}
}
