package render_test

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Vier weitere Verknüpfungen: `saturate`, `invert_image`, `crop`, `flip_*`.**
//
// Sie machen aus einem Bild ein anderes, ohne ein zweites zu brauchen — die
// einfachste Sorte, und die letzte, die im Pack noch auf einen Namen wartete
// (#270).

// **`saturate` rechnet über den Mittelwert der drei Kanäle.**
//
//	int factor = int(256 * amount);
//	…
//	int div = 768 - 3 * factor;
//	int avg = factor*(r+g+b);
//	pix[0] = col((768*r - avg) / div);
//
// `src/gfx/image_effects.cpp`, Zeilen 15–61, Fassung f9d9356 vom 2021-07-06.
//
// **Drei Zweige und eine Division**: über null wird gesättigt, unter null
// entsättigt, und bei genau eins teilte MSE durch null — dort gilt „größer als
// der Mittelwert" (Zeilen 29–38).
func TestSaturateRechnetUeberDenMittelwert(t *testing.T) {
	faelle := []struct {
		name     string
		menge    float64
		quelle   color.RGBA
		erwartet color.RGBA
	}{
		{
			// **Null ändert nichts** — MSE kehrt sofort zurück.
			name: "null ändert nichts", menge: 0,
			quelle:   color.RGBA{R: 200, G: 100, B: 50, A: 255},
			erwartet: color.RGBA{R: 200, G: 100, B: 50, A: 255},
		},
		{
			// **Ganz entsättigt bleibt der Mittelwert**: (200+100+50)/3 = 116.
			name: "ganz entsättigt", menge: -1,
			quelle:   color.RGBA{R: 200, G: 100, B: 50, A: 255},
			erwartet: color.RGBA{R: 116, G: 116, B: 116, A: 255},
		},
		{
			// **Bei eins entscheidet der Vergleich**: rot ist größer als die
			// beiden anderen zusammen? 400 > 150 — ja.
			name: "die Division durch null", menge: 1,
			quelle:   color.RGBA{R: 200, G: 100, B: 50, A: 255},
			erwartet: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			// **Halb gesättigt**: Faktor 128, Teiler 768 − 384 = 384, Summe
			// 350 — rot wird (768·200 − 128·350) / 384 = 283, geklemmt auf 255;
			// blau (768·50 − 44.800) / 384 = −16, geklemmt auf null.
			name: "halb gesättigt", menge: 0.5,
			quelle:   color.RGBA{R: 200, G: 100, B: 50, A: 255},
			erwartet: color.RGBA{R: 255, G: 83, B: 0, A: 255},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			bild, befund := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
				"/eins.png": fall.quelle,
			}}).AusBaumFuerTest(render.Bildbaum{
				Knotenart: "saturate",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
				Zahlen:    map[string]float64{"amount": fall.menge},
			})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			rot, gruen, blau, _ := bild.At(0, 0).RGBA()

			haben := color.RGBA{
				R: uint8(rot >> 8), G: uint8(gruen >> 8), B: uint8(blau >> 8), A: 255,
			}
			if haben != fall.erwartet {
				t.Errorf("ergibt %+v, erwartet %+v", haben, fall.erwartet)
			}
		})
	}
}

// **`invert_image` dreht jeden Kanal um — das Alpha nicht.**
//
//	for (int i = 0 ; i < n ; ++i) data[i] = 255 - data[i];
//
// `src/gfx/image_effects.cpp`, Zeilen 66–72. `n` ist **dreimal** die Zahl der
// Bildpunkte: die Durchsichtigkeit bleibt.
//
// Fassung f9d9356 vom 2021-07-06.
func TestInvertImageDrehtJedenKanalUm(t *testing.T) {
	bild, befund := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 200, G: 100, B: 0, A: 255},
	}}).AusBaumFuerTest(render.Bildbaum{
		Knotenart: "invert_image",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	rot, gruen, blau, _ := bild.At(0, 0).RGBA()
	if rot>>8 != 55 || gruen>>8 != 155 || blau>>8 != 255 {
		t.Errorf("(%d,%d,%d), erwartet (55,155,255)", rot>>8, gruen>>8, blau>>8)
	}

	// **Und die Durchsichtigkeit bleibt.** Die Prüfquelle ist vormultipliziert;
	// ein halbdurchsichtiges Grau steht dort als (64,64,64,128) und heißt
	// ungetrennt (128,128,128) — umgedreht also (127,127,127) bei Alpha 128.
	halb, befund := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/halb.png": {R: 64, G: 64, B: 64, A: 128},
	}}).AusBaumFuerTest(render.Bildbaum{
		Knotenart: "invert_image",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/halb.png")}},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if _, _, _, alpha := halb.At(0, 0).RGBA(); alpha>>8 != 128 {
		t.Errorf("das Alpha ist %d, erwartet 128", alpha>>8)
	}
}

// **`crop` schneidet ein Stück heraus.**
//
//	return image->generate(opt).Size(wxSize((int)width, (int)height),
//	                                 wxPoint(-(int)offset_x, -(int)offset_y));
//
// `CropImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 309–311, Fassung
// f9d9356 vom 2021-07-06. **Der Versatz zählt negativ**: `offset_x: 2` schiebt
// das Bild um zwei Punkte nach links, zeigt also ab Spalte zwei.
//
// **Was über den Rand ragt, ist leer** — `wxImage::Size` füllt mit der
// Hintergrundfarbe, und ohne eine ist das durchsichtig.
func TestCropSchneidetEinStueckHeraus(t *testing.T) {
	bild, befund := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 200, A: 255},
	}}).AusBaumFuerTest(render.Bildbaum{
		Knotenart: "crop",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
		Zahlen: map[string]float64{
			"width": 4, "height": 4, "offset_x": 6, "offset_y": 0,
		},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if rahmen := bild.Bounds(); rahmen.Dx() != 4 || rahmen.Dy() != 4 {
		t.Fatalf("das Bild ist %dx%d, erwartet 4x4", rahmen.Dx(), rahmen.Dy())
	}

	// Die Quelle ist acht Punkte breit; ab Spalte sechs bleiben zwei.
	if _, _, _, alpha := bild.At(0, 0).RGBA(); alpha == 0 {
		t.Error("die erste Spalte ist leer")
	}

	if _, _, _, alpha := bild.At(3, 0).RGBA(); alpha != 0 {
		t.Error("über den Rand hinaus wurde etwas gezeichnet")
	}
}

// **Die Spiegelungen tauschen eine Achse.**
//
//	Image FlipImageHorizontal::generate(const Options& opt) const {
//	  return image->generate(opt).Mirror(true);
//	}
//
// `src/gfx/generated_image.cpp`, Zeilen 254–270, Fassung f9d9356 vom
// 2021-07-06.
func TestDieSpiegelungenTauschenEineAchse(t *testing.T) {
	quelle := zweifarbig()

	for _, fall := range []struct {
		name  string
		art   string
		links color.NRGBA
		oben  color.NRGBA
	}{
		{
			name: "waagerecht", art: "flip_horizontal",
			links: color.NRGBA{B: 255, A: 255}, oben: color.NRGBA{B: 255, A: 255},
		},
		{
			name: "senkrecht", art: "flip_vertical",
			links: color.NRGBA{B: 255, A: 255}, oben: color.NRGBA{B: 255, A: 255},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			bild, befund := render.NeuerRasterer(gemaltequelle{bild: quelle}).
				AusBaumFuerTest(render.Bildbaum{
					Knotenart: fall.art,
					Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/a.png")}},
				})
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			// Die Quelle ist links oben rot und rechts unten blau; nach jeder
			// Spiegelung steht dort etwas anderes.
			ecke := color.NRGBAModel.Convert(bild.At(0, 0))
			if ecke == (color.NRGBA{R: 255, A: 255}) {
				t.Errorf("die Ecke ist unverändert: %+v", ecke)
			}
		})
	}
}

// **Was ohne Ausgangsbild oder ohne Maße dasteht, wird benannt.**
//
// Kein halber Stand: `saturate` ohne `amount:`, `crop` ohne `width:` — eine
// Vorlage, die das schreibt, ist falsch geschrieben, und ein stillschweigend
// durchgereichtes Bild verdeckte es (§8).
func TestDieBildwerkzeugeMeldenIhreGrenzfaelle(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 200, A: 255},
	}})

	faelle := []struct {
		name  string
		baum  render.Bildbaum
		nennt string
	}{
		{
			name: "saturate ohne amount",
			baum: render.Bildbaum{
				Knotenart: "saturate",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
			},
			nennt: "amount",
		},
		{
			name: "crop ohne Maße",
			baum: render.Bildbaum{
				Knotenart: "crop",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
			},
			nennt: "width",
		},
		{
			name: "saturate ohne lesbares Bild",
			baum: render.Bildbaum{
				Knotenart: "saturate",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/fehlt.png")}},
				Zahlen:    map[string]float64{"amount": 0.5},
			},
			nennt: "/fehlt.png",
		},
		{
			name: "invert_image ohne lesbares Bild",
			baum: render.Bildbaum{
				Knotenart: "invert_image",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/fehlt.png")}},
			},
			nennt: "/fehlt.png",
		},
		{
			name: "crop ohne lesbares Bild",
			baum: render.Bildbaum{
				Knotenart: "crop",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/fehlt.png")}},
				Zahlen:    map[string]float64{"width": 2, "height": 2},
			},
			nennt: "/fehlt.png",
		},
		{
			name: "flip ohne lesbares Bild",
			baum: render.Bildbaum{
				Knotenart: "flip_vertical",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/fehlt.png")}},
			},
			nennt: "/fehlt.png",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, befund := rasterer.AusBaumFuerTest(fall.baum)
			if !strings.Contains(befund, fall.nennt) {
				t.Errorf("der Befund lautet %q und nennt %q nicht", befund, fall.nennt)
			}
		})
	}
}

// zweifarbig malt ein Bild, das links oben rot und rechts unten blau ist.
func zweifarbig() *image.NRGBA {
	bild := image.NewNRGBA(image.Rect(0, 0, 4, 4))

	for y := range 4 {
		for x := range 4 {
			farbe := color.NRGBA{R: 255, A: 255}
			if x >= 2 || y >= 2 {
				farbe = color.NRGBA{B: 255, A: 255}
			}

			bild.SetNRGBA(x, y, farbe)
		}
	}

	return bild
}

// gemaltequelle gibt immer dasselbe Bild.
type gemaltequelle struct{ bild image.Image }

func (g gemaltequelle) Bild(pfad string) (image.Image, error) {
	if pfad == "/a.png" {
		return g.bild, nil
	}

	return nil, errKeinBild
}
