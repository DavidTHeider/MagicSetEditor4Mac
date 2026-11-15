package msepaket

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"runtime"
	"testing"
	"testing/iotest"
)

// **Ein Bild wird vermessen, bevor es ausgepackt wird** — FMT-06.AK-01
// (#1695).
//
// Ein PNG von ein paar hundert Bytes darf 40.000 × 40.000 Bildpunkte
// behaupten; `image.Decode` teilt dafür 6,4 GB zu, bevor es merkt, dass die
// Daten fehlen. Sieben Stellen dekodierten ohne Vorprüfung.
//
// **Nicht parallel**, aus demselben Grund wie [TestEinZipKopfWirdNichtGeglaubt].
func TestEinBildWirdVermessenBevorEsAusgepacktWird(t *testing.T) {
	roh := pngMitKopf(40_000, 40_000)

	var vorher, nachher runtime.MemStats

	runtime.ReadMemStats(&vorher)

	_, _, err := DekodiereBild(roh)

	runtime.ReadMemStats(&nachher)

	if zugeteilt := nachher.TotalAlloc - vorher.TotalAlloc; zugeteilt > 64<<20 {
		t.Errorf("%d MiB zugeteilt — ausgepackt, bevor vermessen", zugeteilt>>20)
	}

	if !errors.Is(err, ErrBildZuGross) {
		t.Errorf("Fehler %v, erwartet ErrBildZuGross", err)
	}
}

// **Was unter der Grenze liegt, kommt durch**, auch an ihrem Rand — der Rand
// an der Maßprüfung allein, denn darunter wird wirklich ausgepackt.
func TestEinBildUnterDerGrenzeWirdGelesen(t *testing.T) {
	t.Parallel()

	var puffer bytes.Buffer

	bild := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	bild.SetNRGBA(2, 1, color.NRGBA{R: 200, A: 255})

	if err := png.Encode(&puffer, bild); err != nil {
		t.Fatal(err)
	}

	gelesen, art, err := DekodiereBild(puffer.Bytes())
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if art != "png" || gelesen.Bounds().Dx() != 3 || gelesen.Bounds().Dy() != 2 {
		t.Errorf("%s, %v — erwartet png, 3 × 2", art, gelesen.Bounds())
	}

	// **Aus einer Quelle derselbe Weg** — `DekodiereBildAus` liest höchstens
	// die Grenze und reicht an `DekodiereBild` weiter.
	ausQuelle, artAusQuelle, err := DekodiereBildAus(bytes.NewReader(puffer.Bytes()))
	if err != nil {
		t.Fatalf("aus der Quelle lesen: %v", err)
	}

	if artAusQuelle != "png" || ausQuelle.Bounds() != gelesen.Bounds() {
		t.Errorf("aus der Quelle %s, %v — erwartet png, %v", artAusQuelle, ausQuelle.Bounds(), gelesen.Bounds())
	}

	for _, fall := range []struct {
		breite, hoehe int
		zuGross       bool
	}{
		{GrenzeBildseite, 1, false},
		{GrenzeBildseite + 1, 1, true},
		{10_000, 10_000, false},
		{10_000, 10_001, true},
	} {
		err := bildmassPruefen(image.Config{Width: fall.breite, Height: fall.hoehe})
		if errors.Is(err, ErrBildZuGross) != fall.zuGross {
			t.Errorf("%d × %d: Fehler %v, zu groß erwartet: %v", fall.breite, fall.hoehe, err, fall.zuGross)
		}
	}
}

// pngMitKopf baut ein PNG, das diese Maße behauptet und keine Bilddaten trägt:
// Signatur, `IHDR`, ein leeres `IDAT`, `IEND`.
func pngMitKopf(breite, hoehe int) []byte {
	var aus bytes.Buffer

	aus.WriteString("\x89PNG\r\n\x1a\n")

	kopf := make([]byte, 13)
	binary.BigEndian.PutUint32(kopf[0:], uint32(breite)) //nolint:gosec // Testdaten
	binary.BigEndian.PutUint32(kopf[4:], uint32(hoehe))  //nolint:gosec // Testdaten
	kopf[8] = 8                                          // Bittiefe
	kopf[9] = 6                                          // RGBA

	stueck(&aus, "IHDR", kopf)
	stueck(&aus, "IDAT", []byte{0x78, 0x9c, 0x03, 0x00, 0x00, 0x00, 0x00, 0x01})
	stueck(&aus, "IEND", nil)

	return aus.Bytes()
}

// stueck hängt ein PNG-Stück samt Länge und Prüfsumme an.
func stueck(aus *bytes.Buffer, art string, daten []byte) {
	laenge := make([]byte, 4)
	binary.BigEndian.PutUint32(laenge, uint32(len(daten))) //nolint:gosec // Testdaten
	aus.Write(laenge)

	summe := crc32.NewIEEE()
	_, _ = summe.Write([]byte(art)) //nolint:errcheck // eine Prüfsumme scheitert nicht
	_, _ = summe.Write(daten)       //nolint:errcheck // dito

	aus.WriteString(art)
	aus.Write(daten)

	pruef := make([]byte, 4)
	binary.BigEndian.PutUint32(pruef, summe.Sum32())
	aus.Write(pruef)
}

// **Aus einer Quelle wird nicht mehr gelesen als die Grenze**, und ein
// Lesefehler kommt als solcher an — beides an `dekodiereBildBisZu`, damit der
// Fall nicht 256 MiB braucht, um sich zu zeigen.
func TestEinBildAusEinerQuelleHaeltDieGrenze(t *testing.T) {
	t.Parallel()

	if _, _, err := dekodiereBildBisZu(bytes.NewReader(make([]byte, 9)), 8); !errors.Is(err, ErrZuGross) {
		t.Errorf("neun Byte über einer Grenze von acht: %v, erwartet ErrZuGross", err)
	}

	kaputt := errors.New("der Datenträger antwortet nicht")
	if _, _, err := dekodiereBildBisZu(iotest.ErrReader(kaputt), 8); !errors.Is(err, kaputt) {
		t.Errorf("ein Lesefehler kam als %v an", err)
	}
}
