package render_test

// **Die beiden Blenden: `masked_blend` über eine Maske, `linear_blend` entlang
// einer Geraden.**
//
// Getrennt von `bildausfuehrung_test.go`, weil es eine andere Frage ist: dort
// steht, **ob** ein Baum ausgeführt wird — hier, **wie** die einzelnen
// Verknüpfungen rechnen.

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`masked_blend` mischt zwei Bilder über eine Maske.**
//
// # Die Regel
//
// Wo die Maske **hell** ist, gilt `light`; wo sie **dunkel** ist, `dark`. Die
// Namen sagen es, und `magic-blends.mse-include/new-blends` bestätigt es: dort
// schachtelt sich `masked_blend` in sein eigenes `light:`, um mehr als zwei
// Farben zu mischen.
//
// **Mit 2.734 Vorkommen die häufigste Verknüpfung überhaupt.**
func TestMaskedBlendMischtUeberDieMaske(t *testing.T) {
	rot := color.RGBA{R: 255, A: 255}
	blau := color.RGBA{B: 255, A: 255}

	faelle := map[string]struct {
		maske    color.RGBA
		erwartet [3]uint8
	}{
		"weiße Maske nimmt light":   {color.RGBA{R: 255, G: 255, B: 255, A: 255}, [3]uint8{255, 0, 0}},
		"schwarze Maske nimmt dark": {color.RGBA{A: 255}, [3]uint8{0, 0, 255}},
		"graue Maske mischt":        {color.RGBA{R: 128, G: 128, B: 128, A: 255}, [3]uint8{128, 0, 127}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
				"/hell.png": rot, "/dunkel.png": blau, "/maske.png": fall.maske,
			}})

			plan := &render.Plan{Breite: 2, Hoehe: 2, Ebenen: []render.Ebene{{
				Art:    render.ArtBild,
				Rahmen: render.Rechteck{Breite: 2, Hoehe: 2},
				Bild: render.Bildbaum{Knotenart: "masked_blend", Kinder: []render.Bildkind{
					{Name: "light", Baum: render.Blatt("/hell.png")},
					{Name: "dark", Baum: render.Blatt("/dunkel.png")},
					{Name: "mask", Baum: render.Blatt("/maske.png")},
				}},
				Deckung: 1,
			}}}

			bild, befunde, err := rasterer.Zeichne(t.Context(), plan, 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			r, g, b, _ := bild.At(0, 0).RGBA()
			got := [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}

			// **Eine Stufe Spielraum**: das Mischen rundet auf ganze Bytes.
			for i := range got {
				if abstandVon(got[i], fall.erwartet[i]) > 1 {
					t.Errorf("der Bildpunkt ist %v, erwartet %v", got, fall.erwartet)

					break
				}
			}
		})
	}
}

// **Ein fehlendes Bild wird benannt, nicht geraten.**
func TestMaskedBlendOhneMaskeWirdBenannt(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/hell.png": {R: 255, A: 255}, "/dunkel.png": {B: 255, A: 255},
	}})

	if _, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "masked_blend", Kinder: []render.Bildkind{
			{Name: "light", Baum: render.Blatt("/hell.png")},
			{Name: "dark", Baum: render.Blatt("/dunkel.png")},
		},
	}); !strings.Contains(befund, "mask") {
		t.Errorf("der Befund lautet %q und nennt die fehlende Maske nicht", befund)
	}
}

// **Ein unlesbares Kind bricht die Mischung ab.**
func TestMaskedBlendMitUnlesbaremKindWirdBenannt(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{})

	if _, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "masked_blend", Kinder: []render.Bildkind{
			{Name: "light", Baum: render.Blatt("/fehlt.png")},
			{Name: "dark", Baum: render.Blatt("/auch-fehlt.png")},
			{Name: "mask", Baum: render.Blatt("/maske.png")},
		},
	}); !strings.Contains(befund, "/fehlt.png") {
		t.Errorf("der Befund lautet %q und nennt das unlesbare Bild nicht", befund)
	}
}

func abstandVon(a, b uint8) int {
	if a > b {
		return int(a) - int(b)
	}

	return int(b) - int(a)
}

// **Bilder verschiedener Größe sind ein Befund, kein Klemmen.**
//
// # Was hier stand, und warum es falsch war
//
// Der erste Anlauf sagte zu, außerhalb einer zu kleinen Maske gelte `dark` —
// mit der Begründung, eine Streckung sei eine Annahme, die keine Vorlage
// belegt. Die Begründung war erfunden.
//
// **MSE wirft dort einen Fehler:** `blend_image.cpp:66` sagt
// `Images used for blending must have the same size`. Hier ist es ein Befund
// statt eines Abbruchs (§8) — eine Karte mit einer unpassenden Maske soll ihre
// anderen sechzig Ebenen behalten —, aber es ist kein stilles Weitermachen.
//
// Aufgefallen ist das erst, als der Quelltext von MSE2 auf diesem Rechner lag.
func TestBilderVerschiedenerGroesseSindEinBefund(t *testing.T) {
	rasterer := render.NeuerRasterer(unterschiedlichGross{})

	_, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "masked_blend", Kinder: []render.Bildkind{
			{Name: "light", Baum: render.Blatt("/gross.png")},
			{Name: "dark", Baum: render.Blatt("/gross-dunkel.png")},
			{Name: "mask", Baum: render.Blatt("/klein.png")},
		},
	})

	if !strings.Contains(befund, "verschieden groß") {
		t.Errorf("der Befund lautet %q und nennt die verschiedenen Größen nicht", befund)
	}
}

// unterschiedlichGross gibt ein 4×4-Bild und eine 1×1-Maske.
type unterschiedlichGross struct{}

func (unterschiedlichGross) Bild(pfad string) (image.Image, error) {
	seite, farbe := 4, color.RGBA{R: 255, A: 255}

	switch pfad {
	case "/klein.png":
		seite, farbe = 1, color.RGBA{R: 255, G: 255, B: 255, A: 255}

	case "/gross-dunkel.png":
		farbe = color.RGBA{B: 255, A: 255}
	}

	bild := image.NewRGBA(image.Rect(0, 0, seite, seite))
	for y := range seite {
		for x := range seite {
			bild.SetRGBA(x, y, farbe)
		}
	}

	return bild, nil
}

// **Abgeschnitten, nicht gerundet** — wie die Ganzzahldivision in C++.
//
// `blend_image.cpp:73` rechnet `(a*m + b*(255-m)) / 255` **ganzzahlig**. Bei
// `light=0, dark=1, maske=127` gibt MSE **0**; eine Rundung gäbe 1.
//
// Der erste Anlauf benutzte `alsAnteil`, das rundet — und der Wächter darüber
// ließ eine Stufe Abweichung zu und konnte den Unterschied deshalb nicht sehen.
// Angestrichen vom Nachlauf des Zweitblicks.
func TestDieMischungSchneidetAbUndRundetNicht(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/null.png": {A: 255},
		"/eins.png": {R: 1, G: 1, B: 1, A: 255},
		"/halb.png": {R: 127, G: 127, B: 127, A: 255},
	}})

	bild, befund := rasterer.AusBaumFuerTest(render.Bildbaum{
		Knotenart: "masked_blend", Kinder: []render.Bildkind{
			{Name: "light", Baum: render.Blatt("/null.png")},
			{Name: "dark", Baum: render.Blatt("/eins.png")},
			{Name: "mask", Baum: render.Blatt("/halb.png")},
		},
	})
	if befund != "" {
		t.Fatalf("Befund: %s", befund)
	}

	// (0*127 + 1*128) / 255 = 0 ganzzahlig, 0,502 gerundet wäre 1.
	if r, _, _, _ := bild.At(0, 0).RGBA(); r>>8 != 0 {
		t.Errorf("der Kanal ist %d, erwartet 0 — es wurde gerundet statt abgeschnitten", r>>8)
	}
}

// **Eine Maske mit Warnung zeichnet und meldet.**
//
// Für die Kinder von `masked_blend` war das behoben, **für die Maske einer
// Ebene nicht** — der gleichartige Nebenkanal blieb stehen. Angestrichen vom
// Nachlauf des Zweitblicks.
func TestEineMaskeMitWarnungZeichnetUndMeldet(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/bild.png":  {R: 255, A: 255},
		"/maske.png": {R: 255, G: 255, B: 255, A: 255},
	}})

	_, befunde, err := rasterer.Zeichne(t.Context(), &render.Plan{
		Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
			Art:    render.ArtBild,
			Rahmen: render.Rechteck{Breite: 8, Hoehe: 8},
			Bild:   render.Blatt("/bild.png"),
			Maske: render.Bildbaum{
				Knotenart: "crop",
				Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt("/maske.png")}},
			},
			Deckung: 1,
		}},
	}, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].String(), "crop") {
		t.Errorf("die Befunde lauten %v, erwartet einer über die nicht gebaute Maske", befunde)
	}
}

// **Beide Meldungen, wenn beide etwas zu sagen haben.**
//
// Eine Ebene kann an der Maske *und* am Bild etwas melden. Wer nur eine nähme,
// verlöre die andere — genau die Klasse, die der Nachlauf des Zweitblicks
// zweimal gefunden hat. **Als zwei Befunde** seit QA-03 (#1656): in einer
// Meldung bekämen beide die Art des ersten Treffers.
func TestMaskeUndBildMeldenBeide(t *testing.T) {
	rasterer := render.NeuerRasterer(bildquelle{da: map[string]color.RGBA{
		"/bild.png":  {R: 255, A: 255},
		"/maske.png": {R: 255, G: 255, B: 255, A: 255},
	}})

	nichtGebaut := func(pfad string) render.Bildbaum {
		return render.Bildbaum{
			// Eine erfundene Art — der Grund steht bei [nochNichtGebaut].
			Knotenart: nochNichtGebaut,
			Kinder:    []render.Bildkind{{Name: "image", Baum: render.Blatt(pfad)}},
		}
	}

	_, befunde, err := rasterer.Zeichne(t.Context(), &render.Plan{
		Breite: 8, Hoehe: 8, Ebenen: []render.Ebene{{
			Art:     render.ArtBild,
			Rahmen:  render.Rechteck{Breite: 8, Hoehe: 8},
			Bild:    nichtGebaut("/bild.png"),
			Maske:   nichtGebaut("/maske.png"),
			Deckung: 1,
		}},
	}, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 2 {
		t.Fatalf("%d Befunde, erwartet zwei: %v", len(befunde), befunde)
	}

	if maske, bild := befunde[0].String(), befunde[1].String(); !strings.Contains(maske, "Maske") ||
		!strings.Contains(bild, "/bild.png") || strings.Contains(bild, "Maske") {
		t.Errorf("die Befunde lauten %q und %q — erwartet die Maske, dann das Bild", maske, bild)
	}
}
