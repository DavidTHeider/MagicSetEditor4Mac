package msepaket_test

// **Was ein Bild ausgepackt kostet** — FMT-06b
// (#1974).
//
// Die Grenze zählte Bildpunkte, und ein Bildpunkt kostet je nach Kodierer
// zwischen einem und vierundzwanzig Byte. Hier steht die Messung, auf der
// die Schätzung steht: je Format und Farbmodell ein Bild, ausgepackt mit
// den registrierten Kodierern, und was sie zuteilen, bleibt unter der
// Schätzung.

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// sockel ist, was ein Kodierer unabhängig von der Größe anlegt — Tabellen,
// Puffer, Leser. Derselbe Sockel wie in `FuzzDekodiereBild`.
const sockel = 1 << 20

// messbilder gibt je Format und Farbmodell ein Bild — selbst geschrieben,
// wo die Kodierer der Standardbibliothek es können, sonst aus `testdaten`.
func messbilder(t *testing.T) map[string][]byte {
	t.Helper()

	bilder := map[string][]byte{}
	flaeche := image.Rect(0, 0, 512, 512)

	for name, bild := range map[string]image.Image{
		"png grau8":   image.NewGray(flaeche),
		"png grau16":  image.NewGray16(flaeche),
		"png rgba8":   image.NewNRGBA(flaeche),
		"png rgba16":  image.NewNRGBA64(flaeche),
		"png palette": image.NewPaletted(flaeche, color.Palette{color.Black, color.White}),
	} {
		var puffer bytes.Buffer
		if fehler := png.Encode(&puffer, bild); fehler != nil {
			t.Fatal(fehler)
		}

		bilder[name] = puffer.Bytes()
	}

	for name, bild := range map[string]image.Image{
		"jpeg grau":  image.NewGray(flaeche),
		"jpeg ycbcr": image.NewYCbCr(flaeche, image.YCbCrSubsampleRatio420),
	} {
		var puffer bytes.Buffer
		if fehler := jpeg.Encode(&puffer, bild, nil); fehler != nil {
			t.Fatal(fehler)
		}

		bilder[name] = puffer.Bytes()
	}

	var puffer bytes.Buffer
	if fehler := gif.Encode(&puffer, image.NewPaletted(flaeche, color.Palette{color.Black, color.White}), nil); fehler != nil {
		t.Fatal(fehler)
	}

	bilder["gif"] = puffer.Bytes()

	// **Was die Standardbibliothek nicht schreibt**: progressive JPEG und
	// PNG mit Zeilensprung (Adam7) — die teuersten Fälle.
	for _, datei := range []string{
		"grau-progressiv.jpg", "ycbcr444-progressiv.jpg", "cmyk.jpg", "cmyk-progressiv.jpg", "rgba16-adam7.png",
	} {
		roh, fehler := os.ReadFile(filepath.Join(testdaten, "auspackkosten", datei)) //nolint:gosec // ein fester Ort im Baum
		if fehler != nil {
			t.Fatal(fehler)
		}

		bilder[datei] = roh
	}

	return bilder
}

// **Die Schätzung deckt, was die Kodierer zuteilen** — gemessen und nicht
// aus dem Quelltext erschlossen.
//
// **Ohne `t.Parallel`**: `TotalAlloc` zählt für den ganzen Prozess, und ein
// Lauf daneben zählte mit.
func TestDieSchaetzungDecktDieMessung(t *testing.T) {
	for name, roh := range messbilder(t) {
		mass, art, fehler := image.DecodeConfig(bytes.NewReader(roh))
		if fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}

		runtime.GC()

		var vorher, nachher runtime.MemStats

		runtime.ReadMemStats(&vorher)

		if _, _, fehler := msepaket.DekodiereBild(roh); fehler != nil {
			t.Fatalf("%s: %v", name, fehler)
		}

		runtime.ReadMemStats(&nachher)

		gemessen := int64(nachher.TotalAlloc - vorher.TotalAlloc) //nolint:gosec // ein paar Megabyte
		if geschaetzt := msepaket.Auspackkosten(mass, art); gemessen > geschaetzt+sockel {
			t.Errorf("%s: %d Byte zugeteilt, geschätzt %d", name, gemessen, geschaetzt)
		}
	}
}

// **Je Format und Farbmodell ein Faktor** — und wer keinen kennt, bekommt
// den teuersten.
func TestJeFormatUndFarbmodellEinFaktor(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		art    string
		modell color.Model
		je     int64
	}{
		{"png", color.GrayModel, 3},
		{"png", color.Gray16Model, 5},
		{"png", color.RGBAModel, 9},
		{"png", color.NRGBAModel, 9},
		{"png", color.RGBA64Model, 17},
		{"png", color.NRGBA64Model, 17},
		{"png", color.Palette{color.Black}, 3},
		{"gif", color.Palette{color.Black}, 3},
		{"jpeg", color.GrayModel, 6},
		{"jpeg", color.YCbCrModel, 16},
		{"jpeg", color.CMYKModel, 25},
		{"webp", color.RGBAModel, 25},
	} {
		mass := image.Config{ColorModel: fall.modell, Width: 16, Height: 16}
		if geschaetzt := msepaket.Auspackkosten(mass, fall.art); geschaetzt != 256*fall.je {
			t.Errorf("%s %T: %d, erwartet %d", fall.art, fall.modell, geschaetzt, 256*fall.je)
		}
	}

	// **Ein JPEG legt Blöcke zu sechzehn Bildpunkten an** — 17 × 17 kostet
	// wie 32 × 32.
	if geschaetzt := msepaket.Auspackkosten(image.Config{ColorModel: color.GrayModel, Width: 17, Height: 17}, "jpeg"); geschaetzt != 32*32*6 {
		t.Errorf("Blockfläche: %d", geschaetzt)
	}
}

// riesenJPEG ist ein kleines JPEG, dessen Kopf andere Maße behauptet.
func riesenJPEG(t *testing.T, breite, hoehe uint16) []byte {
	t.Helper()

	var puffer bytes.Buffer
	if fehler := jpeg.Encode(&puffer, image.NewYCbCr(image.Rect(0, 0, 8, 8), image.YCbCrSubsampleRatio444), nil); fehler != nil {
		t.Fatal(fehler)
	}

	roh := puffer.Bytes()
	stelle := bytes.Index(roh, []byte{0xFF, 0xC0})

	// SOF0: Marke, Länge (2), Genauigkeit (1), dann Höhe und Breite.
	roh[stelle+5], roh[stelle+6] = byte(hoehe>>8), byte(hoehe)
	roh[stelle+7], roh[stelle+8] = byte(breite>>8), byte(breite)

	return roh
}

// **Ein Bild, das unter der Bildpunktgrenze liegt und ausgepackt mehr als
// die Speichergrenze bräuchte, wird abgewiesen** — mit eigenem Schlüssel,
// und ohne dass es ausgepackt würde.
func TestEinBildUeberDerSpeichergrenzeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	// 9.000 × 9.000 sind 81 Millionen Bildpunkte, unter den hundert der
	// Bildpunktgrenze — und als JPEG in YCbCr gut 1,3 GB.
	for name, roh := range map[string][]byte{
		"Speicher":   riesenJPEG(t, 9000, 9000),
		"Bildpunkte": riesenJPEG(t, 32000, 32000),
	} {
		_, _, fehler := msepaket.DekodiereBild(roh)

		var befund *domain.FieldError
		if !errors.Is(fehler, msepaket.ErrBildZuGross) || !errors.As(fehler, &befund) ||
			befund.Key != domain.SchluesselBildZuGross {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}

// **Wer ein Bild liest, reicht die Abweisung weiter, wie sie ist** — mit
// ihrem Schlüssel; jeder andere Fehler wird, was der Rufer daraus macht.
func TestZuGrossOderReichtNurDieAbweisungDurch(t *testing.T) {
	t.Parallel()

	_, _, zuGross := msepaket.DekodiereBild(riesenJPEG(t, 9000, 9000))
	sonst := errors.New("kein lesbares Bild")

	if fehler := msepaket.ZuGrossOder(zuGross, sonst); !errors.Is(fehler, msepaket.ErrBildZuGross) {
		t.Errorf("die Abweisung ging verloren: %v", fehler)
	}

	if fehler := msepaket.ZuGrossOder(errors.New("kaputt"), sonst); !errors.Is(fehler, sonst) {
		t.Errorf("ein anderer Fehler blieb: %v", fehler)
	}
}
