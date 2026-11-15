package msepaket_test

// **Die Bildquelle zählt** — OPS-08.AK-01
// (#1725).

import (
	"image/color"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestDieBildquelleZaehltUndLeert(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	legePNG(t, filepath.Join(wurzel, "probe-zahl.mse-style", "rahmen.png"),
		color.NRGBA{R: 1, G: 2, B: 3, A: 255})

	quelle := msepaket.NeueBildquelle(wurzel)
	for range 3 {
		if _, fehler := quelle.Bild("/probe-zahl.mse-style/rahmen.png"); fehler != nil {
			t.Fatal(fehler)
		}
	}

	zahlen := quelle.Zahlen()
	if zahlen.Treffer != 2 || zahlen.Fehlgriffe != 1 || zahlen.Eintraege != 1 ||
		zahlen.GrenzeEintraege != msepaket.GrenzeBildspeicher {
		t.Errorf("%+v", zahlen)
	}

	quelle.Leeren()

	if nachher := quelle.Zahlen(); nachher.Eintraege != 0 {
		t.Errorf("nach dem Leeren: %+v", nachher)
	}
}
