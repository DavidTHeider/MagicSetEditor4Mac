package render_test

// **Die Maske eines Textfelds bestimmt die Breite seiner Zeilen**
// (#2299).
//
// MSE fragt beim Umbruch für jede Zeile ihre Oberkante nach dem linken und
// rechten Rand der Maske:
//
//	double TextViewer::lineLeft(RotatedDC& dc, const TextStyle& style, double y) const {
//	  return style.mask.getFromCache().rowLeft(y, dc.getInternalSize()) + style.padding_left;
//	}
//
// `src/render/text/viewer.cpp`, Zeilen 736–741; `rowLeft` und `rowRight`
// nehmen je Bildzeile den ersten und letzten Punkt mit Rotkanal ≥ 128 und
// rechnen ihn auf die Feldgröße um (`src/gfx/mask_image.cpp`, Zeilen
// 145–178). Beides Fassung f9d9356 vom 2021-07-06.

import (
	"cmp"
	"errors"
	"image"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// zeilenmaskenquelle gibt unter `maske.png` ein Bild in der Größe des Felds,
// 100 mal 60 Punkte, in sechs Bändern zu zehn Zeilen; sie sagt je Band, von
// welcher bis zu welcher Spalte (in Zehnern) es weiß ist — `{-1, -1}` heißt:
// gar nicht. So bleibt das Umrechnen aufs Feld außen vor; das prüft
// [TestDieMaskeWirdAufDasFeldGebracht].
//
// **Der Rest ist schwarz und durchsichtig, das Weiße durchsichtig auch**:
// gelesen wird der rohe Rotkanal, wie wxWidgets ihn aus der Datei gibt —
// nicht die Deckung.
type zeilenmaskenquelle [6][2]int

func (z zeilenmaskenquelle) Bild(pfad string) (image.Image, error) {
	if pfad != "maske.png" {
		return nil, errors.New("kein Bild unter " + pfad)
	}

	bild := image.NewNRGBA(image.Rect(0, 0, 100, 60))
	for y := range 60 {
		spanne := z[y/10]
		for x := range 100 {
			farbe := color.NRGBA{}
			if x/10 >= spanne[0] && x/10 <= spanne[1] {
				farbe = color.NRGBA{R: 255, G: 255, B: 255}
			}

			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild, nil
}

var (
	ganz   = [2]int{0, 9}
	links  = [2]int{0, 4}
	rechts = [2]int{5, 9}
	keine  = [2]int{-1, -1}
)

// worteJeZeile setzt Wörter zu je 20 Punkten in einen Kasten von 100 mal 60
// und zählt, wie viele auf jeder Zeile stehen.
func worteJeZeile(t *testing.T, satz render.Satz, quelle render.Bildquelle, worte int) ([]int, []render.Zeile) {
	t.Helper()

	satz.Schrift, satz.Groesse = "Probe", 10
	satz.Zeilenfaktoren = render.Zeilenfaktoren{Weich: 1, Hart: 1, Linie: 1}

	gesetzt := render.SetzeAbschnitte(satz,
		render.Rechteck{Breite: 100, Hoehe: 60},
		[]render.Abschnitt{{Text: strings.TrimSpace(strings.Repeat("aaaa ", worte))}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, quelle)

	zahlen := make([]int, 0, len(gesetzt.Zeilen))
	for _, zeile := range gesetzt.Zeilen {
		zahlen = append(zahlen, len(strings.Fields(zeile.Text)))
	}

	return zahlen, gesetzt.Zeilen
}

func TestDieMaskeBestimmtDieBreiteDerZeilen(t *testing.T) {
	// Ein Wort misst 20, mit Lücke 25: vier passen in 99 (95), zwei in 49 (45).
	// Eine weiße Zeile endet beim letzten weißen Punkt — Spalte 99, nicht 100.
	faelle := []struct {
		name     string
		satz     render.Satz
		maske    zeilenmaskenquelle
		erwartet []int
	}{
		{
			// **Ohne Maske die ganze Breite** — die Gegenprobe.
			name: "ohne Maske", satz: render.Satz{},
			maske:    zeilenmaskenquelle{ganz, ganz, ganz, links, links, links},
			erwartet: []int{4, 4, 4, 4, 3},
		},
		{
			// **Unten schmal**: ab 30 endet die Zeile bei 49. Die Zeilen
			// beginnen bei 0, 12, 24, 36 und 48 — und die bei 60 liegt unter
			// dem Feld, wo die Maske nicht mehr gilt (`y >= resize.height`,
			// `mask_image.cpp`, Zeile 164).
			name: "unten schmal", satz: render.Satz{Zeilenmaske: "maske.png"},
			maske:    zeilenmaskenquelle{ganz, ganz, ganz, links, links, links},
			erwartet: []int{4, 4, 4, 2, 2, 3},
		},
		{
			// **Das Polster oben zählt mit**: gefragt wird mit `line.top`, und
			// die erste Zeile beginnt bei `padding_top` — 8, 20, 32, 44, 56,
			// 68. Schon die dritte liegt im schmalen Teil.
			name: "Polster oben", satz: render.Satz{
				Zeilenmaske: "maske.png", Polster: render.Polster{Oben: 8},
			},
			maske:    zeilenmaskenquelle{ganz, ganz, ganz, links, links, links},
			erwartet: []int{4, 4, 2, 2, 2, 4, 1},
		},
		{
			// **Eine Maske, die nicht zu lesen ist, ändert nichts.**
			name: "unlesbar", satz: render.Satz{Zeilenmaske: "fehlt.png"},
			maske:    zeilenmaskenquelle{ganz, ganz, ganz, links, links, links},
			erwartet: []int{4, 4, 4, 4, 3},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			zahlen, _ := worteJeZeile(t, fall.satz, fall.maske, 19)

			if !gleicheZahlen(zahlen, fall.erwartet) {
				t.Errorf("Wörter je Zeile %v, erwartet %v", zahlen, fall.erwartet)
			}
		})
	}
}

// **Der linke Rand der Maske verschiebt die Zeile** — die Stellen beginnen bei
// MSE an `margin_left + lineLeft(top)` (`fitLineWidth`, Zeile 548).
func TestDerLinkeRandDerMaskeVerschiebtDieZeile(t *testing.T) {
	zahlen, zeilen := worteJeZeile(t, render.Satz{Zeilenmaske: "maske.png"},
		zeilenmaskenquelle{ganz, ganz, ganz, rechts, rechts, rechts}, 19)

	if !gleicheZahlen(zahlen, []int{4, 4, 4, 2, 2, 3}) {
		t.Fatalf("Wörter je Zeile %v, erwartet [4 4 4 2 2 3]", zahlen)
	}

	if zeilen[0].Links != 0 || zeilen[3].Links != 50 {
		t.Errorf("Zeilen beginnen bei %.1f und %.1f, erwartet 0 und 50",
			zeilen[0].Links, zeilen[3].Links)
	}
}

// **Passt an einer Stelle nichts, rückt die Zeile nach unten** — je einen
// Punkt, bis etwas passt:
//
//	while (line.top < dc.getHeight() && line_size.width + 1 >= dc.getWidth() - style.padding_right - line.margin_right) {
//	  line.top += 1;
//	  …
//	}
//
// `fitLineWidth`, `src/render/text/viewer.cpp`, Zeilen 546–554, Fassung
// f9d9356 vom 2021-07-06. Die zweite Zeile begänne bei 12, in einer schwarzen
// Bildzeile (10 bis 20); sie rückt auf 20, und die dritte folgt ihr bei 32.
func TestEineLeereMaskenzeileSchiebtDieZeileHinunter(t *testing.T) {
	_, zeilen := worteJeZeile(t, render.Satz{Zeilenmaske: "maske.png"},
		zeilenmaskenquelle{ganz, keine, ganz, ganz, ganz, ganz}, 11)

	if len(zeilen) < 3 {
		t.Fatalf("%d Zeilen: %v", len(zeilen), zeilen)
	}

	for nummer, erwartet := range []float64{0, 20, 32} {
		if zeilen[nummer].Oben != erwartet {
			t.Errorf("Zeile %d beginnt bei %.1f, erwartet %.1f",
				nummer+1, zeilen[nummer].Oben, erwartet)
		}
	}
}

func gleicheZahlen(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// kleinemaske gibt unter `klein.png` ein Bild von zehn mal sechs Punkten,
// weiß in den Spalten bis zur genannten.
type kleinemaske int

func (k kleinemaske) Bild(string) (image.Image, error) {
	bild := image.NewNRGBA(image.Rect(0, 0, 10, 6))
	for y := range 6 {
		for x := range 10 {
			if x <= int(k) {
				bild.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}

	return bild, nil
}

// **Die Maske wird auf das Feld gebracht, bevor sie misst.**
//
//	return conform_image(image, options);
//
// `ScriptableImage::generate`, `src/script/image.cpp`, Zeile 36, mit der
// Größe des Felds (`ValueViewer::getMask`, `src/render/value/viewer.cpp`,
// Zeilen 94–100), umgerechnet mit MSEs Kastenfilter (`resample`,
// `src/gfx/resample_image.cpp`). Beides Fassung f9d9356 vom 2021-07-06. Eine
// Maske von zehn Spalten wird auf hundert gedehnt: weiß bis Spalte 4 heißt
// dann weiß bis 49 — nicht 4 · 100/10 = 40. Und ganz weiß heißt 99.
func TestDieMaskeWirdAufDasFeldGebracht(t *testing.T) {
	faelle := []struct {
		name   string
		spalte kleinemaske
		hoehe  float64
		text   string
		zeilen int
	}{
		// Zwei Wörter, 45 breit: sie passen in 49, nicht in 40.
		{name: "weiß bis Spalte 4", spalte: 4, text: "aaaa aaaa", zeilen: 1},
		// Ist das Feld so hoch wie die Maske, wird nur waagerecht umgerechnet.
		{name: "nur in der Breite", spalte: 4, hoehe: 6, text: "aaaa aaaa", zeilen: 1},
		// 95 breit: passt in 99, nicht in 9 · 100/10 = 90.
		{name: "ganz weiß, 95 breit", spalte: 9, text: "aaaaaaaaa aaaaaaaaa", zeilen: 1},
		// 100 breit: passt ohne Maske, in 99 nicht.
		{name: "ganz weiß, 100 breit", spalte: 9, text: "aaaaaaaaa aaaaaaaaaa", zeilen: 2},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			gesetzt := render.SetzeAbschnitte(
				render.Satz{Schrift: "Probe", Groesse: 10, Zeilenmaske: "klein.png"},
				render.Rechteck{Breite: 100, Hoehe: cmp.Or(fall.hoehe, 60)},
				[]render.Abschnitt{{Text: fall.text}},
				render.Schriftsatz{Aufrecht: einheitsschrift{}}, fall.spalte)

			if len(gesetzt.Zeilen) != fall.zeilen {
				t.Errorf("%d Zeilen, erwartet %d: %v", len(gesetzt.Zeilen), fall.zeilen, gesetzt.Zeilen)
			}
		})
	}
}

// **Ein Feld ohne Breite hat keine Maske.** Die Messstufe setzt ein Feld,
// dessen Breite sich nicht ausrechnen ließ, mit `math.MaxFloat64` — so bricht
// es nie um (`service.umbruchbreite`). Eine Maske darauf umzurechnen hieße, ein
// Bild dieser Breite anzulegen; der Kanonlauf vom 2026-10-11 brach daran ab.
func TestEinFeldOhneBreiteHatKeineMaske(t *testing.T) {
	gesetzt := render.SetzeAbschnitte(
		render.Satz{Schrift: "Probe", Groesse: 10, Zeilenmaske: "klein.png"},
		render.Rechteck{Breite: math.MaxFloat64, Hoehe: 60},
		[]render.Abschnitt{{Text: "aaaa aaaa"}},
		render.Schriftsatz{Aufrecht: einheitsschrift{}}, kleinemaske(4))

	if len(gesetzt.Zeilen) != 1 {
		t.Errorf("%d Zeilen, erwartet eine: %v", len(gesetzt.Zeilen), gesetzt.Zeilen)
	}
}
