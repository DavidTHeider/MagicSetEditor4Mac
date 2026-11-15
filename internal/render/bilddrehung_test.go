package render_test

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Die Drehung legt jeden Punkt an eine gerechnete Stelle.**
//
//	inline static int offset(UInt x, UInt y, UInt w, UInt h) {
//	  int mx = y;
//	  int my = w - x - 1;
//	  return h * my + mx; // note: h, since that is the width of the target image
//	}
//
// `src/gfx/rotate_image.cpp`, Zeilen 49–82, Fassung f9d9356 vom 2021-07-06.
// Drei Strukturen, ein Muster: 90° und 270° tauschen die Seitenlängen, 180°
// lässt sie stehen.
//
// **Geprüft wird an einem ungleichseitigen Bild.** Ein Quadrat verschwiege
// genau den Fehler, den man hier macht — Breite und Höhe zu vertauschen.
func TestDieDrehungLegtJedenPunktAnSeineStelle(t *testing.T) {
	// 4 breit, 2 hoch; jeder Punkt trägt seine Spalte im Rot- und seine Zeile
	// im Grünkanal. Damit ist jede Verschiebung an der Farbe ablesbar.
	quelle := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for y := range 2 {
		for x := range 4 {
			quelle.SetNRGBA(x, y, color.NRGBA{R: uint8(x), G: uint8(y), A: 255})
		}
	}

	for _, fall := range []struct {
		name           string
		winkel         float64
		breite, hoehe  int
		zielX, zielY   int
		quellX, quellY uint8
	}{
		// Aus Rotate90deg: mx = y, my = w − x − 1. Quellpunkt (0,0) landet
		// damit auf (0, 3).
		{"90 Grad", 90, 2, 4, 0, 3, 0, 0},
		{"180 Grad", 180, 4, 2, 3, 1, 0, 0},
		// Aus Rotate270deg: mx = h − y − 1, my = x.
		{"270 Grad", 270, 2, 4, 1, 0, 0, 0},
		{"0 Grad", 0, 4, 2, 0, 0, 0, 0},
		{"360 Grad", 360, 4, 2, 0, 0, 0, 0},
		{"minus 90 Grad", -90, 2, 4, 1, 0, 0, 0},
	} {
		t.Run(fall.name, func(t *testing.T) {
			bild, befund := render.NeuerRasterer(gemaltequelle{bild: quelle}).
				AusBaumFuerTest(render.Bildbaum{
					Knotenart: "rotate",
					Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("/a.png")}},
					Zahlen:    map[string]float64{"angle": fall.winkel},
				})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			if rahmen := bild.Bounds(); rahmen.Dx() != fall.breite || rahmen.Dy() != fall.hoehe {
				t.Fatalf("das Bild ist %dx%d, erwartet %dx%d",
					rahmen.Dx(), rahmen.Dy(), fall.breite, fall.hoehe)
			}

			// **Der Vergleich läuft über die Schnittstelle**, nicht über
			// eine Typzusicherung: `Convert` gibt eine `color.NRGBA`, und
			// eine Zusicherung darauf wäre ein ungeprüfter Rückgabewert.
			hier := color.NRGBAModel.Convert(bild.At(fall.zielX, fall.zielY))
			erwartet := color.NRGBA{R: fall.quellX, G: fall.quellY, A: 255}

			if hier != color.Color(erwartet) {
				t.Errorf("auf (%d,%d) steht %+v, erwartet der Quellpunkt (%d,%d)",
					fall.zielX, fall.zielY, hier, fall.quellX, fall.quellY)
			}
		})
	}
}

// **Ein schräger Winkel wird benannt, nicht geraten.**
//
// MSE fällt dafür auf `wxImage::Rotate` zurück — eine Neuabtastung mit
// eigenem Randverhalten. Hier steht dieselbe Antwort wie bei einer schräg
// gestellten Ebene ([drehung.go]): das Ausgangsbild und ein Befund. Im vollen
// Pack steht `rotate(` **null**-mal; eine Neuabtastung, die niemand aufruft,
// wäre ungeprüfte Fläche.
func TestDieDrehungNenntWasSieNichtKann(t *testing.T) {
	quelle := zweifarbig()

	for _, fall := range []struct {
		name   string
		zahlen map[string]float64
		nennt  string
	}{
		{"ohne angle", nil, "rotate braucht `angle:`"},
		{"schräg", map[string]float64{"angle": 30}, "ist nicht gerade"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			bild, befund := render.NeuerRasterer(gemaltequelle{bild: quelle}).
				AusBaumFuerTest(render.Bildbaum{
					Knotenart: "rotate",
					Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("/a.png")}},
					Zahlen:    fall.zahlen,
				})
			if bild == nil {
				t.Fatal("ohne Bild")
			}

			if !strings.Contains(befund, fall.nennt) {
				t.Errorf("der Befund lautet %q, erwartet wurde %q darin", befund, fall.nennt)
			}
		})
	}
}

// **Ohne lesbares Ausgangsbild dreht sich nichts** — und der Befund von unten
// bleibt stehen, statt von „braucht `angle:`" verdeckt zu werden (§8).
func TestDieDrehungOhneAusgangsbild(t *testing.T) {
	bild, befund := render.NeuerRasterer(bildquelle{}).
		AusBaumFuerTest(render.Bildbaum{
			Knotenart: "rotate",
			Kinder:    []render.Bildkind{{Name: "input", Baum: render.Blatt("/fehlt.png")}},
			Zahlen:    map[string]float64{"angle": 90},
		})

	if bild != nil || !strings.Contains(befund, "/fehlt.png") {
		t.Errorf("Bild %v, Befund %q", bild != nil, befund)
	}
}
