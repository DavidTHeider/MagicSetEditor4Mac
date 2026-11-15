package render

// **Der Schatten einer Schrift** — F02
// (#2244).
//
// MSE zeichnet jedes Textstück zweimal:
//
//	void RotatedDC::DrawTextWithShadow(const String& text, const Font& font, const RealPoint& pos, double scale, double stretch) {
//	  DrawText(text, pos + font.shadow_displacement * scale, font.shadow_color, font.shadow_blur * scale, 1, stretch);
//	  DrawText(text, pos, font.color, 0, 1, stretch);
//	}
//
// (`src/util/rotation.cpp`, Zeilen 221–224, Fassung f9d9356 vom 2021-07-06).
// **Ohne Abfrage, ob es einen Schatten gibt**: ein durchsichtiger fällt in
// `DrawText` heraus (`if (color.Alpha() == 0) return;`, Zeile 191), und die
// Vorgabe ist durchsichtig (`Color() : r(0), g(0), b(0), a(0)`,
// `src/gfx/color.hpp`, Zeile 32).
//
// Die Proben hier kommen ohne Goldbild aus, weil der Schatten **derselbe**
// Text ist: ein Schatten ohne Versatz unter durchsichtigem Text ist der Text,
// einer mit ganzzahligem Versatz ist er, verschoben.

import (
	"image"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func TestEinSchattenOhneVersatzIstDerText(t *testing.T) {
	text := zeileMitSchatten(t, Farbe{A: 255}, Textschatten{})
	schatten := zeileMitSchatten(t, Farbe{}, Textschatten{Farbe: Farbe{A: 255}})

	if !gleicheDeckung(text, schatten) {
		t.Error("ein Schatten ohne Versatz unter durchsichtigem Text sieht anders aus als der Text")
	}
}

func TestEinSchattenSitztUmSeinenVersatz(t *testing.T) {
	text := zeileMitSchatten(t, Farbe{A: 255}, Textschatten{})
	schatten := zeileMitSchatten(t, Farbe{}, Textschatten{Farbe: Farbe{A: 255}, X: 3, Y: -2})

	verschoben := image.NewNRGBA(text.Bounds())
	for y := text.Bounds().Min.Y; y < text.Bounds().Max.Y; y++ {
		for x := text.Bounds().Min.X; x < text.Bounds().Max.X; x++ {
			if quelle := (image.Point{X: x - 3, Y: y + 2}); quelle.In(text.Bounds()) {
				verschoben.SetNRGBA(x, y, text.NRGBAAt(quelle.X, quelle.Y))
			}
		}
	}

	if !gleicheDeckung(verschoben, schatten) {
		t.Error("der Schatten steht nicht um (3, −2) neben dem Text")
	}
}

// **Ein durchsichtiger Schatten zeichnet nichts** — die Vorgabe, und die
// meisten Angaben im Pack (`rgba(0,0,0,0)` für jeden gewöhnlichen Rahmen).
func TestEinDurchsichtigerSchattenZeichnetNichts(t *testing.T) {
	text := zeileMitSchatten(t, Farbe{A: 255}, Textschatten{})
	mit := zeileMitSchatten(t, Farbe{A: 255}, Textschatten{X: 3, Y: 3, Weichheit: 2})

	if !gleicheDeckung(text, mit) {
		t.Error("ein durchsichtiger Schatten hat etwas gezeichnet")
	}
}

// **Der Schatten liegt unter dem Text**, nicht darüber: ein roter ohne
// Versatz unter schwarzem Text lässt dort, wo der Text deckt, Schwarz stehen.
func TestDerSchattenLiegtUnterDemText(t *testing.T) {
	bild := zeileMitSchatten(t, Farbe{A: 255}, Textschatten{Farbe: Farbe{R: 255, A: 255}})

	schwarz := 0
	for i := 0; i < len(bild.Pix); i += 4 {
		if bild.Pix[i+3] == 255 && bild.Pix[i] == 0 {
			schwarz++
		}
	}

	if schwarz == 0 {
		t.Error("kein voll gedeckter Punkt ist schwarz — der Schatten liegt über dem Text")
	}
}

// **Ein halbdurchsichtiger Schatten deckt halb** — `set_alpha` nimmt die
// Deckung der Farbe ganzzahlig hinein, bevor geweicht wird.
func TestEinHalbdurchsichtigerSchattenDecktHalb(t *testing.T) {
	bild := zeileMitSchatten(t, Farbe{}, Textschatten{Farbe: Farbe{A: 128}})

	hoechste := uint8(0)
	for i := 3; i < len(bild.Pix); i += 4 {
		hoechste = max(hoechste, bild.Pix[i])
	}

	if hoechste != 128 {
		t.Errorf("die stärkste Deckung ist %d, erwartet 128", hoechste)
	}
}

// **Weich wird in einem Zug, an Ort und Stelle** — `blur_image_alpha`:
//
//	(2·Mitte + links + oben + rechts + unten) / 6
//
// mit dem Rand als Mitte, und **links und oben sind schon weich**, weil der
// Zug in denselben Speicher schreibt (`src/gfx/resample_text.cpp`, Zeilen
// 144–162, Fassung f9d9356 vom 2021-07-06). Ein Zug mit getrennter Ausgabe
// gäbe [10, 40, 10].
func TestWeichzeichnenRechnetWieMSE(t *testing.T) {
	faelle := []struct {
		name          string
		breite, hoehe int
		ein, aus      []uint8
	}{
		{"eine Zeile", 3, 1, []uint8{0, 60, 0}, []uint8{10, 41, 6}},
		{"eine Spalte", 1, 3, []uint8{0, 60, 0}, []uint8{10, 41, 6}},
		{"ein Feld", 2, 2, []uint8{120, 0, 0, 0}, []uint8{80, 13, 13, 4}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			deckung := append([]uint8(nil), fall.ein...)
			weichzeichne(deckung, fall.breite, fall.hoehe)

			for i := range deckung {
				if deckung[i] != fall.aus[i] {
					t.Fatalf("%v, erwartet %v", deckung, fall.aus)
				}
			}
		})
	}
}

// **Und ein weicher Schatten reicht weiter** als ein harter.
func TestEinWeicherSchattenReichtWeiter(t *testing.T) {
	hart := zeileMitSchatten(t, Farbe{}, Textschatten{Farbe: Farbe{A: 255}})
	weich := zeileMitSchatten(t, Farbe{}, Textschatten{Farbe: Farbe{A: 255}, Weichheit: 3})

	if gedeckt(weich) <= gedeckt(hart) {
		t.Errorf("weich deckt %d Punkte, hart %d — erwartet mehr", gedeckt(weich), gedeckt(hart))
	}
}

func gedeckt(bild *image.NRGBA) int {
	anzahl := 0
	for i := 3; i < len(bild.Pix); i += 4 {
		if bild.Pix[i] > 0 {
			anzahl++
		}
	}

	return anzahl
}

// zeileMitSchatten zeichnet „Schatten" in Go Regular bei 20 Punkt.
func zeileMitSchatten(t *testing.T, farbe Farbe, schatten Textschatten) *image.NRGBA {
	t.Helper()

	schrift, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}

	plan := &Plan{Breite: 160, Hoehe: 50, Ebenen: []Ebene{{
		Art:        ArtText,
		Rahmen:     Rechteck{Links: 12, Oben: 10, Breite: 136, Hoehe: 32},
		Abschnitte: Aufrecht("Schatten"),
		Farbe:      farbe,
		Satz:       Satz{Schrift: "Go", Groesse: 20, Ausrichtung: LinksOben, Schatten: schatten},
		Deckung:    1,
	}}}

	bild, befunde, err := NeuerRasterer(nil).MitSchriften(eineSchriftquelle{schrift}).
		Zeichne(t.Context(), plan, 1, GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("zeichnen: %v %v", err, befunde)
	}

	return bild
}
