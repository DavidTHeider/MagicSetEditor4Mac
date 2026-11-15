// **Vier Bildformate, wie MSE2 sie anbietet** (#826).
package service_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// durchsichtigeEcke gibt ein Bild, dessen linke obere Ecke durchsichtig ist.
//
// **Die Ecke ist der Prüfstein.** Ob ein Format Durchsichtigkeit behält, sagt
// sich leicht; gemessen wird es an einem Bildpunkt, der sie trägt.
func durchsichtigeEcke() image.Image {
	bild := image.NewRGBA(image.Rect(0, 0, 4, 4))

	for x := range 4 {
		for y := range 4 {
			bild.Set(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}

	bild.Set(0, 0, color.RGBA{})

	return bild
}

func TestJedesFormatTraegtSeineEndung(t *testing.T) {
	t.Parallel()

	erwartet := map[service.Bildformat]string{
		service.FormatPNG:  ".png",
		service.FormatJPG:  ".jpg",
		service.FormatBMP:  ".bmp",
		service.FormatTIFF: ".tif",
	}

	for format, endung := range erwartet {
		if format.Endung() != endung {
			t.Errorf("%s gibt %q, erwartet war %q", format, format.Endung(), endung)
		}
	}

	// **Und ein erfundenes Format gibt keine.** Ohne diesen Fall sagte der
	// Test nichts darüber, was bei einer unbekannten Angabe geschieht.
	if service.Bildformat("gibtesnicht").Endung() != "" {
		t.Error("ein unbekanntes Format gab eine Endung")
	}
}

// **Was jedes Format aus einer durchsichtigen Ecke macht** — gemessen, indem
// das Geschriebene zurückgelesen wird.
//
// PNG und TIFF können Alpha und behalten es; JPG und BMP können es nicht, und
// die Ecke wird weiß. Das ist keine Vermutung über die Formate, sondern eine
// über **diese** Kodierer, und deshalb steht sie hier als Messung.
func TestWasEinFormatMitDurchsichtigkeitTut(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		format    service.Bildformat
		behaeltEs bool
		lies      func([]byte) (image.Image, error)
	}{
		{service.FormatPNG, true, func(b []byte) (image.Image, error) {
			return png.Decode(bytes.NewReader(b))
		}},
		{service.FormatTIFF, true, func(b []byte) (image.Image, error) {
			return tiff.Decode(bytes.NewReader(b))
		}},
		{service.FormatBMP, false, func(b []byte) (image.Image, error) {
			return bmp.Decode(bytes.NewReader(b))
		}},
	}

	for _, fall := range faelle {
		t.Run(string(fall.format), func(t *testing.T) {
			t.Parallel()

			bytesRaus, fehler := service.PackeBild(durchsichtigeEcke(), fall.format)
			if fehler != nil {
				t.Fatalf("packen: %v", fehler)
			}

			zurueck, fehler := fall.lies(bytesRaus)
			if fehler != nil {
				t.Fatalf("zurücklesen: %v", fehler)
			}

			_, _, _, alpha := zurueck.At(0, 0).RGBA()

			if fall.behaeltEs && alpha != 0 {
				t.Errorf("Alpha %d in der Ecke, erwartet war 0 — das Format "+
					"soll Durchsichtigkeit behalten", alpha)
			}

			if !fall.behaeltEs {
				rot, gruen, blau, _ := zurueck.At(0, 0).RGBA()
				if rot != 0xffff || gruen != 0xffff || blau != 0xffff {
					t.Errorf("die Ecke ist (%d,%d,%d), erwartet war Weiß",
						rot, gruen, blau)
				}
			}
		})
	}
}

// **Ein Format, das keines ist, wird auch hier abgewiesen** (#826).
//
// Die Aufrufer prüfen es, bevor sie packen — der Zweig ist von dort
// unerreichbar. Über den ausgeführten Einstieg ist er es nicht, und ein
// stiller Rückfall auf PNG wäre die falsche Antwort: unter welcher Endung wäre
// geschrieben worden?
func TestPackeBildWeistEinUnbekanntesFormatAb(t *testing.T) {
	t.Parallel()

	_, fehler := service.PackeBild(durchsichtigeEcke(), service.Bildformat("webp"))
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Und ein Kodiererfehler kommt heraus, statt verschluckt zu werden.**
//
// Nachgemessen am 2026-08-22: von den vier Kodierern scheitern an einem Bild
// ohne Fläche **PNG und TIFF**, JPEG und BMP nicht. Einer davon genügt, weil
// alle vier denselben Rückweg nehmen — vier eigene Rückwege bräuchten vier
// Fälle, und zwei davon gäbe es nicht.
func TestEinKodiererfehlerKommtHeraus(t *testing.T) {
	t.Parallel()

	ohneFlaeche := image.NewRGBA(image.Rect(0, 0, 0, 0))

	if _, fehler := service.PackeBild(ohneFlaeche, service.FormatPNG); fehler == nil {
		t.Error("kein Fehler für ein Bild ohne Fläche")
	}
}

// halbdurchsichtig gibt ein Bild, dessen linke obere Ecke **halb** deckt.
//
// **`NRGBA` und nicht `RGBA`.** In `image.RGBA` sind die Farbwerte mit dem
// Alpha vormultipliziert; „R 200 bei Alpha 128" ließe sich dort gar nicht
// hinschreiben, ohne zuerst zu rechnen — und dann prüfte der Test die eigene
// Rechnung mit.
func halbdurchsichtig() image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, 4, 4))

	for x := range 4 {
		for y := range 4 {
			bild.Set(x, y, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}

	bild.Set(0, 0, color.NRGBA{R: 200, G: 100, B: 50, A: 128})

	return bild
}

// halbdurchsichtigeFlaeche gibt ein Bild, das **ganz** halb deckt.
//
// **Für JPEG, und nur dafür.** [halbdurchsichtig] setzt einen einzelnen Punkt
// anders als seine Nachbarn; ein 4×4-Bild ist ein einziger 8×8-Block, und was
// dort herauskommt, misst die Blockkante statt der Weißauflage. Eine gleich
// gefärbte Fläche gibt JPEG fast unverändert zurück — dann sagt die Messung
// wieder etwas über das, was geprüft werden soll.
func halbdurchsichtigeFlaeche() image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, 16, 16))

	for x := range 16 {
		for y := range 16 {
			bild.Set(x, y, color.NRGBA{R: 200, G: 100, B: 50, A: 128})
		}
	}

	return bild
}

// **Halbe Deckung, nicht nur ganz oder gar nicht** —
// #851.
//
// # Warum der Fall dazugehört
//
// `TestWasEinFormatMitDurchsichtigkeitTut` misst Alpha **0** und Alpha
// **255**. Beide Endpunkte sehen auch dann richtig aus, wenn dazwischen falsch
// gemischt wird — eine Ebene mit `Deckung: 0.5` ist auf einer Karte der
// Normalfall, nicht die Ausnahme. Der Zweitblick über Phase P hat die Lücke
// angestrichen.
//
// # Woher die erwarteten Werte kommen
//
// **Aus der Formel, nicht aus `aufWeiss`.** Wer die zu prüfende Seite auch als
// Maßstab nähme, verglich sie mit sich selbst (§10):
//
//	aus = c · a/255 + 255 · (1 − a/255)
//
// Mit a = 128 gibt das für (200, 100, 50) die Werte (227, 177, 152).
func TestWasEinFormatMitHalberDeckungTut(t *testing.T) {
	t.Parallel()

	// **Ein Schritt Spielraum je Kanal.** `draw.Over` rechnet in 16 Bit und
	// rundet beim Zurückschreiben; mehr als ein Schritt wäre kein Rundungsrest
	// mehr, sondern eine andere Formel.
	const spielraum = 1

	t.Run("bmp legt auf Weiß", func(t *testing.T) {
		t.Parallel()

		roh, fehler := service.PackeBild(halbdurchsichtig(), service.FormatBMP)
		if fehler != nil {
			t.Fatalf("packen: %v", fehler)
		}

		zurueck, fehler := bmp.Decode(bytes.NewReader(roh))
		if fehler != nil {
			t.Fatalf("zurücklesen: %v", fehler)
		}

		rot, gruen, blau, alpha := zurueck.At(0, 0).RGBA()

		for _, kanal := range []struct {
			name              string
			gegeben, erwartet uint32
		}{
			{"Rot", rot >> 8, 227},
			{"Grün", gruen >> 8, 177},
			{"Blau", blau >> 8, 152},
		} {
			if kanalabstand(kanal.gegeben, kanal.erwartet) > spielraum {
				t.Errorf("%s ist %d, gemischt auf Weiß wären es %d",
					kanal.name, kanal.gegeben, kanal.erwartet)
			}
		}

		if alpha>>8 != 255 {
			t.Errorf("Alpha %d — BMP kennt keine Durchsichtigkeit und soll voll decken",
				alpha>>8)
		}
	})

	// **JPG geht denselben `aufWeiss`-Weg — durch einen anderen Kodierer.**
	//
	// Der Nachlauf zum Zweitblick über Phase P hat angestrichen, dass die
	// Zusage JPG ausließ: produktiv teilt es mit BMP die Weißauflage, aber
	// danach kommt ein **verlustbehafteter** Kodierer. Der große Goldwächter
	// misst nur aggregierte Abstände und ersetzt diese Zusage nicht
	// (#860).
	t.Run("jpg legt ebenfalls auf Weiß", func(t *testing.T) {
		t.Parallel()

		roh, fehler := service.PackeBild(halbdurchsichtigeFlaeche(), service.FormatJPG)
		if fehler != nil {
			t.Fatalf("packen: %v", fehler)
		}

		zurueck, fehler := jpeg.Decode(bytes.NewReader(roh))
		if fehler != nil {
			t.Fatalf("zurücklesen: %v", fehler)
		}

		rot, gruen, blau, alpha := zurueck.At(0, 0).RGBA()

		// **Eine eigene Toleranz, und sie ist begründet.**
		//
		// JPEG rechnet in Blöcken von 8×8 und tastet die Farbkanäle
		// unter. Für BMP gilt weiter **ein** Schritt — es ist verlustfrei.
		//
		// **Und ein eigenes Bild**: an [halbdurchsichtig] steht **ein**
		// Bildpunkt anders als seine drei Nachbarn, und ein 4×4-Bild ist ein
		// einziger Block. Gemessen am 2026-08-22 kam Rot dort als 181 statt
		// 227 heraus — das misst die Blockkante und nicht die Weißauflage.
		// Eine Toleranz von 50 hätte den Fall durchgelassen, in dem gar nicht
		// auf Weiß gelegt wird (Rot bliebe 200); der Test hätte dann nichts
		// mehr gehalten (§10).
		const jpgSpielraum = 3

		for _, kanal := range []struct {
			name              string
			gegeben, erwartet uint32
		}{
			{"Rot", rot >> 8, 227},
			{"Grün", gruen >> 8, 177},
			{"Blau", blau >> 8, 152},
		} {
			if kanalabstand(kanal.gegeben, kanal.erwartet) > jpgSpielraum {
				t.Errorf("%s ist %d, gemischt auf Weiß wären es %d",
					kanal.name, kanal.gegeben, kanal.erwartet)
			}
		}

		if alpha>>8 != 255 {
			t.Errorf("Alpha %d — JPG kennt keine Durchsichtigkeit", alpha>>8)
		}
	})

	t.Run("tiff behält das Alpha", func(t *testing.T) {
		t.Parallel()

		roh, fehler := service.PackeBild(halbdurchsichtig(), service.FormatTIFF)
		if fehler != nil {
			t.Fatalf("packen: %v", fehler)
		}

		zurueck, fehler := tiff.Decode(bytes.NewReader(roh))
		if fehler != nil {
			t.Fatalf("zurücklesen: %v", fehler)
		}

		_, _, _, alpha := zurueck.At(0, 0).RGBA()

		if kanalabstand(alpha>>8, 128) > spielraum {
			t.Errorf("Alpha %d, erwartet waren 128 — TIFF soll die halbe "+
				"Deckung tragen", alpha>>8)
		}
	})
}

// kanalabstand gibt den Betrag der Differenz zweier Kanalwerte.
func kanalabstand(a, b uint32) uint32 {
	if a > b {
		return a - b
	}

	return b - a
}
