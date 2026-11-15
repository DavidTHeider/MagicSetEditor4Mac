package render_test

// **`linear_blend`: der Verlauf entlang einer Geraden.**
//
// Getrennt von `blenden_test.go`, das `masked_blend` hält — die beiden sind
// zwei Rechnungen und nicht eine, und ihre Zusagen stehen dort, wo sie gelten.

import (
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`linear_blend` blendet zwei Bilder entlang einer Geraden.**
//
// # Die Regel, nachgelesen im Pack
//
// `magic-blends.mse-include/new-blends:70` baut den Hybridrahmen damit:
//
//	card_hybrid_2 := {linear_blend(
//		image1: template(colors[0]),
//		image2: template(colors[1]),
//		x1: 0.4, y1: 0
//		x2: 0.6, y2: 0)}
//
// **Die Koordinaten sind Anteile, keine Bildpunkte** — 0,4 und 0,6 sind vierzig
// und sechzig Prozent der Breite. Bei (x1,y1) gilt `image1`, bei (x2,y2)
// `image2`, dazwischen wird linear geblendet und außerhalb geklemmt.
//
// **Dass die Achsen tauschbar sind, steht daneben:** Zeile 91 bindet
// `linear_blend` um und setzt `y1: x1, y2: x2` — eine senkrechte Version
// derselben Mischung. Die Funktion muss also beide Achsen tragen und darf
// keine davon annehmen.
//
// # Warum das die schlechteste Zahl im Bestand betrifft
//
// `Old-Money` ist eine Hybridkarte und liegt bei 92,1 Abstand — 99,9 Prozent
// abweichende Bildpunkte, weil sie **gar keinen Rahmen** bekommt
// (#306).
func TestLinearBlendBlendetEntlangEinerGeraden(t *testing.T) {
	rot := color.RGBA{R: 255, A: 255}
	blau := color.RGBA{B: 255, A: 255}

	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": rot, "/zwei.png": blau,
	}})

	// Vier Punkte breit: die Anteile 0 und 1 spannen über die ganze Breite.
	baum := render.Bildbaum{
		Knotenart: "linear_blend",
		Kinder: []render.Bildkind{
			{Name: "image1", Baum: render.Blatt("/eins.png")},
			{Name: "image2", Baum: render.Blatt("/zwei.png")},
		},
		Zahlen: map[string]float64{"x1": 0, "y1": 0, "x2": 1, "y2": 0},
	}

	bild, befund := rasterer.AusBaumFuerTest(baum)
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	// **Links überwiegt `image1`, rechts `image2`** — und nicht mehr.
	//
	// Die Endwerte werden nicht geprüft: das Prüfbild ist zwei Punkte breit,
	// und ein Verlauf über Punktmitten erreicht seine Enden dort nie. Eine
	// Erwartung von 255 prüfte die Breite des Prüfbilds und nicht die Regel.
	linksRot, _, linksBlau, _ := bild.At(0, 0).RGBA()
	rechtsRot, _, rechtsBlau, _ := bild.At(bild.Bounds().Dx()-1, 0).RGBA()

	if linksRot <= linksBlau {
		t.Errorf("links steht (%d,_,%d), erwartet mehr Rot als Blau", linksRot>>8, linksBlau>>8)
	}

	if rechtsBlau <= rechtsRot {
		t.Errorf("rechts steht (%d,_,%d), erwartet mehr Blau als Rot", rechtsRot>>8, rechtsBlau>>8)
	}

	// **Und der Verlauf geht in die richtige Richtung**, nicht nur irgendwohin.
	if rechtsBlau <= linksBlau {
		t.Errorf("Blau nimmt von %d auf %d ab, erwartet zu", linksBlau>>8, rechtsBlau>>8)
	}
}

// **Ohne Gerade gibt es nichts zu blenden.**
//
// Fallen (x1,y1) und (x2,y2) zusammen, ist die Richtung nicht bestimmt. Das
// ist ein Befund und kein Grund, eine Richtung zu erfinden.
func TestLinearBlendOhneRichtungWirdBenannt(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 255, A: 255}, "/zwei.png": {B: 255, A: 255},
	}})

	_, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "linear_blend",
		Kinder: []render.Bildkind{
			{Name: "image1", Baum: render.Blatt("/eins.png")},
			{Name: "image2", Baum: render.Blatt("/zwei.png")},
		},
		Zahlen: map[string]float64{"x1": 0.5, "y1": 0.5, "x2": 0.5, "y2": 0.5},
	})

	if !strings.Contains(befund, "Richtung") {
		t.Errorf("der Befund lautet %q und nennt die fehlende Richtung nicht", befund)
	}
}

// **Ein fehlendes `image2` wird benannt.**
//
// Die beiden Bilder sind gleichrangig: `linear_blend` ohne das zweite ist kein
// halber Verlauf, sondern ein unvollständiger Aufruf.
func TestLinearBlendOhneZweitesBildWirdBenannt(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 255, A: 255},
	}})

	if _, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "linear_blend",
		Kinder:    []render.Bildkind{{Name: "image1", Baum: render.Blatt("/eins.png")}},
		Zahlen:    map[string]float64{"x2": 1},
	}); !strings.Contains(befund, "image2") {
		t.Errorf("der Befund lautet %q und nennt das fehlende image2 nicht", befund)
	}
}

// **Ein unlesbares erstes Bild bricht den Verlauf ab.**
func TestLinearBlendMitUnlesbaremBildWirdBenannt(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{})

	if _, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "linear_blend",
		Kinder: []render.Bildkind{
			{Name: "image1", Baum: render.Blatt("/fehlt.png")},
			{Name: "image2", Baum: render.Blatt("/auch.png")},
		},
		Zahlen: map[string]float64{"x2": 1},
	}); !strings.Contains(befund, "/fehlt.png") {
		t.Errorf("der Befund lautet %q und nennt das unlesbare Bild nicht", befund)
	}
}

// **Ein senkrechter Verlauf geht senkrecht.**
//
// `magic-blends.mse-include/new-blends:91` bindet `linear_blend` um und setzt
// `y1: x1, y2: x2` — beide Achsen müssen tragen.
func TestLinearBlendKannAuchSenkrecht(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 255, A: 255}, "/zwei.png": {B: 255, A: 255},
	}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "linear_blend",
		Kinder: []render.Bildkind{
			{Name: "image1", Baum: render.Blatt("/eins.png")},
			{Name: "image2", Baum: render.Blatt("/zwei.png")},
		},
		Zahlen: map[string]float64{"x1": 0, "y1": 0, "x2": 0, "y2": 1},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	_, _, obenBlau, _ := bild.At(0, 0).RGBA()
	_, _, untenBlau, _ := bild.At(0, bild.Bounds().Dy()-1).RGBA()

	if untenBlau <= obenBlau {
		t.Errorf("Blau nimmt von oben %d nach unten %d ab, erwartet zu",
			obenBlau>>8, untenBlau>>8)
	}
}

// **Der Befund eines Kindes reist mit.**
//
// # Der Befund des Zweitblicks
//
// [render.Rasterer] gibt für eine nicht gebaute Verknüpfung **Bild und Befund**
// zurück. Der erste Anlauf der Blenden übernahm den Befund nur, wenn das Bild
// `nil` war — ein `recolor_image` **innerhalb** eines `masked_blend` zeichnete
// damit sein Ersatzbild, und die Warnung verschwand.
//
// **Halbe Unterstützung sah dann aus wie vollständige, und der Befundzähler
// sank.** Das ist die schlimmste Art, eine Zahl zu verbessern.
func TestDerBefundEinesKindesReistMit(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/hell.png": {R: 255, A: 255}, "/dunkel.png": {B: 255, A: 255},
		"/maske.png": {R: 255, G: 255, B: 255, A: 255},
	}})

	// Das helle Bild kommt aus einer Verknüpfung, die es noch nicht gibt.
	_, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "masked_blend", Kinder: []render.Bildkind{
			{Name: "light", Baum: render.Bildbaum{
				Knotenart: "recolor_image",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/hell.png")}},
			}},
			{Name: "dark", Baum: render.Blatt("/dunkel.png")},
			{Name: "mask", Baum: render.Blatt("/maske.png")},
		},
	})

	if !strings.Contains(befund, "recolor_image") {
		t.Errorf("der Befund lautet %q und nennt die nicht gebaute Verknüpfung des Kindes nicht",
			befund)
	}

	if !strings.Contains(befund, "light") {
		t.Errorf("der Befund lautet %q und sagt nicht, welches Kind es war", befund)
	}
}

// **Fehlt eine Koordinate, gilt sie als null.**
//
// `magic-blends.mse-include/new-blends:91` schreibt nur `x1`, `x2`, `y1`, `y2`
// — und Zeile 70 lässt `y1` und `y2` auf null stehen. Ein fehlender Wert ist
// dort kein Mangel, sondern die Achse, um die es nicht geht.
func TestEineFehlendeKoordinateGiltAlsNull(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/eins.png": {R: 255, A: 255}, "/zwei.png": {B: 255, A: 255},
	}})

	// Nur `x2` steht da; x1, y1, y2 sind null.
	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "linear_blend",
		Kinder: []render.Bildkind{
			{Name: "image1", Baum: render.Blatt("/eins.png")},
			{Name: "image2", Baum: render.Blatt("/zwei.png")},
		},
		Zahlen: map[string]float64{"x2": 1},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	_, _, linksBlau, _ := bild.At(0, 0).RGBA()
	_, _, rechtsBlau, _ := bild.At(bild.Bounds().Dx()-1, 0).RGBA()

	if rechtsBlau <= linksBlau {
		t.Errorf("Blau nimmt von %d auf %d ab, erwartet zu", linksBlau>>8, rechtsBlau>>8)
	}
}

// **Ungleich große Bilder sind auch beim Verlauf ein Befund.**
//
// MSE wirft `images used for blending must have the same size`
// (`blend_image.cpp:21`). Hier ist es ein Befund statt eines Abbruchs (§8).
func TestLinearBlendMitUngleichenGroessenWirdBenannt(t *testing.T) {
	rasterer := render.NeuerRasterer(unterschiedlichGross{})

	if _, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "linear_blend",
		Kinder: []render.Bildkind{
			{Name: "image1", Baum: render.Blatt("/gross.png")},
			{Name: "image2", Baum: render.Blatt("/klein.png")},
		},
		Zahlen: map[string]float64{"x2": 1},
	}); !strings.Contains(befund, "verschieden groß") {
		t.Errorf("der Befund lautet %q und nennt die verschiedenen Größen nicht", befund)
	}
}
