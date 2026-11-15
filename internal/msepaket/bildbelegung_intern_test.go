package msepaket

import (
	"image"
	"image/color"
	"testing"
)

// fremdesBild ist eine Art, die `image` nicht kennt.
type fremdesBild struct{ image.Image }

// **Was ein entschlüsseltes Bild im Arbeitsspeicher belegt** — PERF03
// (#2353),
// Befund A-19: die Puffer der Bildpunkte, je Art so, wie `image` sie anlegt.
func TestBelegteBytesZaehltDiePuffer(t *testing.T) {
	t.Parallel()

	feld := image.Rect(0, 0, 10, 4)
	palette := color.Palette{color.Black, color.White}

	faelle := []struct {
		name  string
		bild  image.Image
		bytes int64
	}{
		{name: "NRGBA", bild: image.NewNRGBA(feld), bytes: 160},
		{name: "RGBA", bild: image.NewRGBA(feld), bytes: 160},
		{name: "NRGBA64", bild: image.NewNRGBA64(feld), bytes: 320},
		{name: "RGBA64", bild: image.NewRGBA64(feld), bytes: 320},
		{name: "Gray", bild: image.NewGray(feld), bytes: 40},
		{name: "Gray16", bild: image.NewGray16(feld), bytes: 80},
		{name: "Alpha", bild: image.NewAlpha(feld), bytes: 40},
		{name: "Alpha16", bild: image.NewAlpha16(feld), bytes: 80},
		{name: "CMYK", bild: image.NewCMYK(feld), bytes: 160},
		{name: "Paletted", bild: image.NewPaletted(feld, palette), bytes: 40 + 2*16},
		{name: "YCbCr 4:2:0", bild: image.NewYCbCr(feld, image.YCbCrSubsampleRatio420), bytes: 40 + 2*10},
		{name: "NYCbCrA 4:4:4", bild: image.NewNYCbCrA(feld, image.YCbCrSubsampleRatio444), bytes: 4 * 40},
		{name: "eine fremde Art, vier Byte je Punkt", bild: fremdesBild{image.NewGray(feld)}, bytes: 160},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if belegt := belegteBytes(fall.bild); belegt != fall.bytes {
				t.Errorf("%d Bytes, erwartet %d", belegt, fall.bytes)
			}
		})
	}
}

// **An der Bytegrenze räumt die Quelle ganz, und ein Bild über ihr wird gar
// nicht gemerkt** — dieselbe Wahl wie an der Zahl der Einträge
// ([GrenzeBildspeicher]): wer die Grenze erreicht, hat keinen Arbeitssatz,
// der hineinpasst.
func TestDieBildquelleRaeumtAnDerBytegrenze(t *testing.T) {
	t.Parallel()

	quelle := NeueBildquelle()
	quelle.grenzeBytes = 1000

	vierhundert := image.NewGray(image.Rect(0, 0, 20, 20))

	quelle.merken("/a.mse-style/eins.png", vierhundert)
	quelle.merken("/a.mse-style/zwei.png", vierhundert)

	// **Dasselbe Bild zweimal** — zwei Läufe, die es zugleich nicht fanden —
	// zählt einmal.
	quelle.merken("/a.mse-style/zwei.png", vierhundert)

	if zahlen := quelle.Zahlen(); zahlen.Eintraege != 2 || zahlen.Bytes != 800 || zahlen.GrenzeBytes != 1000 {
		t.Fatalf("zwei Bilder: %+v", zahlen)
	}

	quelle.merken("/a.mse-style/drei.png", vierhundert)

	if zahlen := quelle.Zahlen(); zahlen.Eintraege != 1 || zahlen.Bytes != 400 {
		t.Errorf("über der Grenze nicht ganz geräumt: %+v", zahlen)
	}

	quelle.merken("/a.mse-style/riesig.png", image.NewGray(image.Rect(0, 0, 40, 40)))

	if _, da := quelle.nachschlagen("/a.mse-style/riesig.png"); da {
		t.Error("ein Bild über der ganzen Grenze wurde gemerkt")
	}

	if zahlen := quelle.Zahlen(); zahlen.Bytes > zahlen.GrenzeBytes || zahlen.Eintraege != 1 {
		t.Errorf("nach dem riesigen Bild: %+v", zahlen)
	}

	quelle.Leeren()

	if zahlen := quelle.Zahlen(); zahlen.Eintraege != 0 || zahlen.Bytes != 0 {
		t.Errorf("nach dem Leeren: %+v", zahlen)
	}
}

// **Die Grenze im Betrieb** — 128 MiB, PERF03.
func TestDieBytegrenzeDerBildquelle(t *testing.T) {
	t.Parallel()

	if zahlen := NeueBildquelle().Zahlen(); zahlen.GrenzeBytes != GrenzeBildbytes || GrenzeBildbytes != 128<<20 {
		t.Errorf("Grenze %d, Konstante %d", zahlen.GrenzeBytes, GrenzeBildbytes)
	}
}
