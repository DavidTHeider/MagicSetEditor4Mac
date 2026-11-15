package render_test

// **`recolor_image`: die häufigste Verknüpfung, die es noch nicht gab.**
//
// 965 der 3.192 Befunde über die 190 Stile von `M15-Magic-Pack` lauteten
// darauf (#270).

import (
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein Bildpunkt wird als Mischung aus Rot, Grün, Blau und Weiß gelesen.**
//
//	int lo = min(x.r,min(x.g,x.b));
//	int nr = x.r - lo; int ng = x.g - lo; int nb = x.b - lo; int nw = lo;
//	int total = max(255, nr+ng+nb+nw);
//	return RGB((nr*cr.r + ng*cg.r + nb*cb.r + nw*cw.r) / total, …);
//
// `recolor`, `src/gfx/image_effects.cpp`, Zeilen 76–93, Fassung f9d9356 vom
// 2021-07-06.
//
// **Der Anteil einer Farbe ist ihr Abstand zum dunkelsten Kanal**, der Rest ist
// Weiß. Ein reines Rot (255,0,0) ist damit ganz Rot, ein Grau ganz Weiß.
//
// **Geteilt wird durch die Summe, nicht durch 255** — MSEs Kommentar sagt
// warum: „just to be sure", falls die Anteile zusammen mehr als 255 ergeben.
func TestRecolorImageLiestEinenPunktAlsMischung(t *testing.T) {
	faelle := []struct {
		name     string
		quelle   color.RGBA
		erwartet color.RGBA
	}{
		{
			// Reines Rot wird ganz zur Rotfarbe.
			name:   "reines Rot",
			quelle: color.RGBA{R: 255, A: 255},

			erwartet: color.RGBA{R: 10, G: 20, B: 30, A: 255},
		},
		{
			// Reines Grün wird ganz zur Grünfarbe.
			name:     "reines Grün",
			quelle:   color.RGBA{G: 255, A: 255},
			erwartet: color.RGBA{R: 40, G: 50, B: 60, A: 255},
		},
		{
			// **Weiß ist reines Weiß** — `lo` ist 255, und nichts bleibt übrig.
			name:     "Weiß",
			quelle:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
			erwartet: color.RGBA{R: 200, G: 210, B: 220, A: 255},
		},
		{
			// **Schwarz bleibt schwarz**: alle vier Anteile sind null.
			name:     "Schwarz",
			quelle:   color.RGBA{A: 255},
			erwartet: color.RGBA{A: 255},
		},
		{
			// **Halb Rot, halb Weiß**: 128 Rot und 127 Weiß, durch 255 geteilt
			// — und **abgeschnitten**, nicht gerundet. Grün rechnet sich zu
			// (128·20 + 127·210) / 255 = 114,6, und daraus wird 114.
			name:     "helles Rot",
			quelle:   color.RGBA{R: 255, G: 127, B: 127, A: 255},
			erwartet: color.RGBA{R: 104, G: 114, B: 124, A: 255},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
				"/eins.png": fall.quelle,
			}})

			baum := render.Bildbaum{
				Knotenart: "recolor_image",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
				Farben: map[string]render.Farbe{
					"red":   {R: 10, G: 20, B: 30, A: 255},
					"green": {R: 40, G: 50, B: 60, A: 255},
					"blue":  {R: 70, G: 80, B: 90, A: 255},
					"white": {R: 200, G: 210, B: 220, A: 255},
				},
			}

			bild, befund := rasterer.AusBaumFuerTest(baum)
			if befund != "" {
				t.Fatalf("Befund: %s", befund)
			}

			rot, gruen, blau, alpha := bild.At(0, 0).RGBA()

			haben := color.RGBA{
				R: uint8(rot >> 8), G: uint8(gruen >> 8),
				B: uint8(blau >> 8), A: uint8(alpha >> 8),
			}

			if haben != fall.erwartet {
				t.Errorf("%+v wird %+v, erwartet %+v", fall.quelle, haben, fall.erwartet)
			}
		})
	}
}

// **Mit einer einzigen Farbe entscheidet ihre Helligkeit über Grün und Blau.**
//
//	Byte to_grayscale(RGB x) { return (6969*x.r + 23434*x.g + 2365*x.b) / 32768; }
//	void recolor(Image& img, RGB cr) {
//	  bool dark = to_grayscale(cr) < 100;
//	  recolor(img, cr, dark ? black : white, dark ? white : black, white);
//	}
//
// Zeilen 103–112. **Grün wird die ähnliche, Blau die entgegengesetzte Farbe** —
// eine dunkle Farbe bekommt Schwarz für Grün und Weiß für Blau, eine helle
// umgekehrt. So bleibt ein Symbol lesbar, gleich in welcher Farbe es steht.
//
// Die Gewichte stammen aus libpng, und sie sind **nicht** die des
// Menschenauges nach Rec. 601 — wer hier 0,299/0,587/0,114 einsetzte, bekäme
// bei Grenzfällen die andere Entscheidung.
func TestRecolorImageMitEinerFarbeEntscheidetDieHelligkeit(t *testing.T) {
	faelle := []struct {
		name        string
		farbe       render.Farbe
		gruen, blau color.RGBA
	}{
		{
			// Dunkelblau: Graustufe 2.365·255/32768 ≈ 18, also dunkel.
			name:  "eine dunkle Farbe",
			farbe: render.Farbe{B: 255, A: 255},
			gruen: color.RGBA{A: 255},
			blau:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			// Gelb: 6.969 + 23.434 von 32.768 ≈ 236, also hell.
			name:  "eine helle Farbe",
			farbe: render.Farbe{R: 255, G: 255, A: 255},
			gruen: color.RGBA{R: 255, G: 255, B: 255, A: 255},
			blau:  color.RGBA{A: 255},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
				"/gruen.png": {G: 255, A: 255},
				"/blau.png":  {B: 255, A: 255},
			}})

			for pfad, erwartet := range map[string]color.RGBA{
				"/gruen.png": fall.gruen, "/blau.png": fall.blau,
			} {
				baum := render.Bildbaum{
					Knotenart: "recolor_image",
					Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt(pfad)}},
					Farben:    map[string]render.Farbe{"color": fall.farbe},
				}

				bild, befund := rasterer.AusBaumFuerTest(baum)
				if befund != "" {
					t.Fatalf("Befund: %s", befund)
				}

				rot, gruen, blau, alpha := bild.At(0, 0).RGBA()

				haben := color.RGBA{
					R: uint8(rot >> 8), G: uint8(gruen >> 8),
					B: uint8(blau >> 8), A: uint8(alpha >> 8),
				}

				if haben != erwartet {
					t.Errorf("%s wird %+v, erwartet %+v", pfad, haben, erwartet)
				}
			}
		})
	}
}

// **Ohne Farbe gibt es einen Befund und das Ausgangsbild.**
//
// `recolor_image` liest `color` als Pflichtargument, sobald `red` fehlt
// (`src/script/functions/image.cpp`, Zeilen 88–99). Eine Vorlage ohne beides
// ist falsch geschrieben — und ein stillschweigend durchgereichtes Bild
// verdeckte das (§8).
//
// Fassung f9d9356 vom 2021-07-06.
func TestRecolorImageOhneFarbeIstEinBefund(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 255, A: 255},
	}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "recolor_image",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
	})

	if !strings.Contains(befund, "color") {
		t.Errorf("Befund %q nennt das fehlende Argument nicht", befund)
	}

	if bild == nil {
		t.Error("ohne Farbe wurde gar nichts gezeichnet — erwartet das Ausgangsbild")
	}
}

// **Das Alpha bleibt, wie es ist.**
//
// MSE rechnet über `GetData()`, und das sind **drei** Kanäle; die
// Durchsichtigkeit eines Rahmens gehört zu seiner Form und nicht zu seiner
// Farbe.
func TestRecolorImageLaesstDasAlphaStehen(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/halb.png": {R: 255, A: 128},
	}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "recolor_image",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/halb.png")}},
		Farben:    map[string]render.Farbe{"color": {R: 10, G: 200, B: 10, A: 255}},
	})

	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	if _, _, _, alpha := bild.At(0, 0).RGBA(); alpha>>8 != 128 {
		t.Errorf("das Alpha ist %d, erwartet 128", alpha>>8)
	}
}

// **Ohne `green:` und `blue:` ist auch `red:` nicht genug.**
//
//	SCRIPT_OPTIONAL_PARAM(Color, red) {
//	  SCRIPT_PARAM(Color, green);
//	  SCRIPT_PARAM(Color, blue);
//	  …
//	}
//
// `src/script/functions/image.cpp`, Zeilen 90–94: hinter `red:` sind die
// beiden anderen **Pflicht**, und `white:` hat eine Vorgabe.
//
// Fassung f9d9356 vom 2021-07-06.
func TestRecolorImageMitRotBrauchtGruenUndBlau(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 255, A: 255},
	}})

	_, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "recolor_image",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/eins.png")}},
		Farben:    map[string]render.Farbe{"red": {R: 255, A: 255}},
	})

	if !strings.Contains(befund, "green") {
		t.Errorf("Befund %q nennt das fehlende Argument nicht", befund)
	}
}

// **Ein unlesbares Ausgangsbild bricht die Umfärbung ab.**
//
// Kein halber Stand: wo nichts zu färben ist, wird nichts gezeichnet, und der
// Befund nennt die Datei (§8).
func TestRecolorImageOhneAusgangsbild(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "recolor_image",
		Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/fehlt.png")}},
		Farben:    map[string]render.Farbe{"color": {R: 255, A: 255}},
	})

	if bild != nil || !strings.Contains(befund, "/fehlt.png") {
		t.Errorf("Bild %v, Befund %q", bild != nil, befund)
	}
}
