package render_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Bild wird auf die Größe seines Feldes zugeschnitten**
// (#246).
//
// # Wofür
//
// MSE legt beim Einsetzen **das zugeschnittene** Bild ab, nicht das rohe:
//
//	ImageSliceWindow s(…, image, style().getSize(), mask);
//	if (s.ShowModal() == wxID_OK) {
//	  LocalFileName new_image_file = getLocalPackage().newFileName(field().name, "");
//	  img.SaveFile(getLocalPackage().nameOut(new_image_file), wxBITMAP_TYPE_PNG);
//	}
//
// `ImageValueEditor::sliceImage`, `src/gui/value/image.cpp`, Zeilen 42–51,
// Fassung f9d9356 vom 2021-07-06. Wer roh übernähme, legte für ein
// 40-Kilobyte-Kunstfenster eine 12-Megapixel-Aufnahme ins Set
// ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)).
//
// # Die Vorgabewahl
//
//	ImageSlice::ImageSlice(const Image& source, const wxSize& target_size)
//	  : selection(0, 0, source.GetWidth(), source.GetHeight())
//	  , allow_outside(false), aspect_fixed(true)
//
//	// fix aspect ratio
//	int diff = selection.width * target_size.GetHeight()
//	         - selection.height * target_size.GetWidth();
//	if (diff > 0 …) selection.width  -= int(diff / target_size.GetHeight());
//	else            selection.height -= int(-diff / target_size.GetWidth());
//
// `src/gui/image_slice_window.cpp`, Zeilen 20–52, gerufen im Konstruktor des
// Fensters (Zeile 84). Also: die **ganze** Quelle, dann die zu lange Seite
// gekürzt, bis das Seitenverhältnis stimmt — **oben links verankert**, denn
// `selection.x` und `selection.y` bleiben null.
func TestDerZuschnittFolgtMSEsVorgabewahl(t *testing.T) {
	faelle := []struct {
		name                    string
		quellbreite, quellhoehe int
		zielbreite, zielhoehe   int
		erwartetX, erwartetY    int
		erwartetB, erwartetH    int
	}{
		{
			name:        "gleiches Seitenverhältnis bleibt ganz",
			quellbreite: 200, quellhoehe: 100, zielbreite: 20, zielhoehe: 10,
			erwartetX: 0, erwartetY: 0, erwartetB: 200, erwartetH: 100,
		},
		{
			// diff = 400*10 - 100*20 = 2000; width -= 2000/10 = 200 → 200
			name:        "eine zu breite Quelle wird schmaler",
			quellbreite: 400, quellhoehe: 100, zielbreite: 20, zielhoehe: 10,
			erwartetX: 0, erwartetY: 0, erwartetB: 200, erwartetH: 100,
		},
		{
			// diff = 200*10 - 400*20 = -6000; height -= 6000/20 = 300 → 100
			name:        "eine zu hohe Quelle wird kürzer",
			quellbreite: 200, quellhoehe: 400, zielbreite: 20, zielhoehe: 10,
			erwartetX: 0, erwartetY: 0, erwartetB: 200, erwartetH: 100,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gab := render.Zuschnittflaeche(
				image.Rect(0, 0, fall.quellbreite, fall.quellhoehe),
				fall.zielbreite, fall.zielhoehe)

			soll := image.Rect(fall.erwartetX, fall.erwartetY,
				fall.erwartetX+fall.erwartetB, fall.erwartetY+fall.erwartetH)

			if gab != soll {
				t.Errorf("%v statt %v", gab, soll)
			}
		})
	}
}

// **Und das Ergebnis hat die Größe des Feldes.**
//
//	Image target(target_size.GetWidth(), target_size.GetHeight(), false);
//	resample_and_clip(source, target, selection);
//
// `ImageSlice::getSlice`, Zeilen 54–66.
func TestDerZuschnittHatDieGroesseDesFeldes(t *testing.T) {
	quelle := image.NewNRGBA(image.Rect(0, 0, 400, 100))
	for x := range 400 {
		for y := range 100 {
			// Links rot, rechts blau — damit sich zeigt, **welcher** Teil bleibt.
			if x < 200 {
				quelle.Set(x, y, color.NRGBA{R: 255, A: 255})
			} else {
				quelle.Set(x, y, color.NRGBA{B: 255, A: 255})
			}
		}
	}

	zugeschnitten := render.Zuschnitt(quelle, 20, 10)

	if zugeschnitten.Bounds() != image.Rect(0, 0, 20, 10) {
		t.Fatalf("%v", zugeschnitten.Bounds())
	}

	// **Oben links verankert**: von 400 Punkten Breite bleiben die ersten 200,
	// und die sind rot.
	ecke := zugeschnitten.NRGBAAt(1, 5)
	if ecke.R < 200 || ecke.B > 60 {
		t.Errorf("die linke Hälfte kam als %+v an — erwartet war Rot", ecke)
	}

	rechts := zugeschnitten.NRGBAAt(18, 5)
	if rechts.R < 200 || rechts.B > 60 {
		t.Errorf("rechts steht %+v — der Zuschnitt ist nicht oben links verankert", rechts)
	}
}

// **Ein leeres Ziel gibt ein leeres Bild.**
//
// Ein Feldstil darf Breite oder Höhe null tragen — `magic-m15-saga` schreibt
// `width: {if is_saga() then 0 else 153}`, und dann ist das Feld unsichtbar
// (`ValueViewer::isVisible`). Ein Zuschnitt auf null Punkte ist kein Fehler,
// sondern nichts.
func TestEinLeeresZielGibtEinLeeresBild(t *testing.T) {
	quelle := image.NewNRGBA(image.Rect(0, 0, 40, 40))

	for _, fall := range []struct{ breite, hoehe int }{{0, 10}, {10, 0}, {-5, 10}} {
		if bild := render.Zuschnitt(quelle, fall.breite, fall.hoehe); !bild.Bounds().Empty() {
			t.Errorf("%d×%d ergab %v", fall.breite, fall.hoehe, bild.Bounds())
		}
	}
}

// **Und eine leere Quelle auch.** Eine Datei, die sich zwar öffnen ließ, aber
// keine Bildpunkte trägt, ist kein Grund abzubrechen.
func TestEineLeereQuelleGibtEinLeeresBild(t *testing.T) {
	if bild := render.Zuschnitt(image.NewNRGBA(image.Rect(0, 0, 0, 0)), 20, 10); !bild.Bounds().Empty() {
		t.Errorf("%v", bild.Bounds())
	}
}
