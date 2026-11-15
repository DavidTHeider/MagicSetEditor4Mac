package render

import (
	"image"
	"math"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// **Quartz' Schriftglättung, nachgebildet** —
// #1806.
//
// Die Zahlen stehen mit ihrer Messung in `docs/SCHRIFTUNTERSUCHUNG.md`; gegen
// Core Text selbst hält `TestDieGlaettungTrifftQuartz` hinter `fremdpakete`.

func TestQuartzVerdicktNachDerGroesse(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		ppem, je float64
	}{
		{6, 6.0 / 64},
		{16, 0.25},
		{19.2, 0.3},
		{56, 0.3},
		{400, 0.3},
		{0, 0},
	} {
		if got := quartzVerdickung(fall.ppem); math.Abs(got-fall.je) > 1e-12 {
			t.Errorf("bei %v Bildpunkten je Geviert %v, erwartet %v", fall.ppem, got, fall.je)
		}
	}
}

func TestEinUmrissWirdJeSeiteVerdickt(t *testing.T) {
	t.Parallel()

	// Ein Quadrat von 10 Bildpunkten, einmal in jeder Umlaufrichtung — welche
	// außen ist, sagt die Fläche des ganzen Umrisses.
	for name, ecken := range map[string][]fixed.Point26_6{
		"im Uhrzeigersinn":        {fixed.P(0, 0), fixed.P(10, 0), fixed.P(10, 10), fixed.P(0, 10)},
		"gegen den Uhrzeigersinn": {fixed.P(0, 0), fixed.P(0, 10), fixed.P(10, 10), fixed.P(10, 0)},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dick := verdickt(vieleck(ecken), 0.3)

			// **Waagerecht nach beiden Seiten, senkrecht nur nach oben** — die
			// Unterkante bleibt, die Oberkante rückt um das Doppelte.
			grenzen := dick.Bounds()
			if got, want := grenzen, (fixed.Rectangle26_6{
				Min: fixed.Point26_6{X: -19, Y: -38}, Max: fixed.Point26_6{X: 640 + 19, Y: 640},
			}); got != want {
				t.Errorf("Grenzen %v, erwartet %v", got, want)
			}

			if len(dick) != len(vieleck(ecken)) {
				t.Errorf("%d Stücke statt %d", len(dick), len(vieleck(ecken)))
			}
		})
	}
}

// **Auch ein kubischer Umriss** — CFF-Schriften (`.otf`) beschreiben ihre
// Glyphen in Kubiken, und deren Stützpunkte wandern mit. Ein Quadrat aus vier
// geraden Kubiken wird wie eines aus Geraden verdickt.
func TestEinKubischerUmrissWirdVerdickt(t *testing.T) {
	t.Parallel()

	ecken := []fixed.Point26_6{fixed.P(0, 0), fixed.P(10, 0), fixed.P(10, 10), fixed.P(0, 10)}

	umriss := make(sfnt.Segments, 0, len(ecken)+1)
	umriss = append(umriss, sfnt.Segment{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{ecken[0]}})

	for i := range ecken {
		von, nach := ecken[i], ecken[(i+1)%len(ecken)]
		schritt := nach.Sub(von)

		umriss = append(umriss, sfnt.Segment{Op: sfnt.SegmentOpCubeTo, Args: [3]fixed.Point26_6{
			von.Add(schritt.Mul(fixed.I(1) / 3)), von.Add(schritt.Mul(fixed.I(2) / 3)), nach,
		}})
	}

	if got, want := verdickt(umriss, 0.3).Bounds(), verdickt(vieleck(ecken), 0.3).Bounds(); got != want {
		t.Errorf("kubisch %v, aus Geraden %v", got, want)
	}
}

// **Ein Loch wird enger** — eine innere Kontur läuft andersherum, und dieselbe
// Regel, die die äußere weitet, zieht sie zusammen. Sonst würde ein `o` zu
// einem Ring mit dünnerem Strich statt einem fetteren.
func TestEinLochWirdEnger(t *testing.T) {
	t.Parallel()

	aussen := vieleck([]fixed.Point26_6{fixed.P(0, 0), fixed.P(20, 0), fixed.P(20, 20), fixed.P(0, 20)})
	innen := vieleck([]fixed.Point26_6{fixed.P(5, 5), fixed.P(5, 15), fixed.P(15, 15), fixed.P(15, 5)})

	dick := verdickt(append(aussen, innen...), 0.3)
	loch := dick[len(aussen):].Bounds()

	if want := (fixed.Rectangle26_6{
		Min: fixed.Point26_6{X: 320 + 19, Y: 320}, Max: fixed.Point26_6{X: 960 - 19, Y: 960 - 38},
	}); loch != want {
		t.Errorf("das Loch hat %v, erwartet %v", loch, want)
	}
}

// **Eine Spitze wandert nicht zur Seite** — dreht der Umriss um mehr als
// rund 160 Grad, ginge die Gehrung ins Unendliche. FreeType lässt den Punkt
// stehen, und dieser Nachbau auch; angehoben wird er mit dem ganzen Umriss.
func TestEineSpitzeWandertNicht(t *testing.T) {
	t.Parallel()

	spitz := vieleck([]fixed.Point26_6{fixed.P(0, 0), fixed.P(100, 1), fixed.P(0, 2)})
	dick := verdickt(spitz, 0.3)

	if got, want := dick[1].Args[0], (fixed.Point26_6{X: 6400, Y: 64 - 19}); got != want {
		t.Errorf("die Spitze wanderte nach %v, erwartet %v", got, want)
	}
}

// **Kurven behalten ihre Gestalt**: Art und Zahl der Stücke bleiben, nur die
// Punkte wandern — auch die Stützpunkte.
func TestKurvenBehaltenIhreStuecke(t *testing.T) {
	t.Parallel()

	schrift, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	kennung, _ := schrift.kennung('o')

	umriss, err := schrift.schrift.LoadGlyph(&schrift.puffer, kennung, fixed.I(56), nil)
	if err != nil {
		t.Fatalf("Umriss: %v", err)
	}

	umriss = append(sfnt.Segments(nil), umriss...)
	dick := verdickt(umriss, 0.3)

	if len(dick) != len(umriss) {
		t.Fatalf("%d Stücke statt %d", len(dick), len(umriss))
	}

	bewegt := 0

	for i := range umriss {
		if dick[i].Op != umriss[i].Op {
			t.Fatalf("Stück %d ist %v statt %v", i, dick[i].Op, umriss[i].Op)
		}

		if dick[i].Args != umriss[i].Args {
			bewegt++
		}
	}

	if bewegt < len(umriss)/2 {
		t.Errorf("nur %d von %d Stücken bewegt", bewegt, len(umriss))
	}
}

// **Ohne Verdickung bleibt alles, wo es ist** — und ein Umriss ohne Fläche
// wird nur angehoben, denn er hat keine Seite, die außen wäre.
func TestNullVerdicktNichts(t *testing.T) {
	t.Parallel()

	quadrat := vieleck([]fixed.Point26_6{fixed.P(0, 0), fixed.P(10, 0), fixed.P(10, 10), fixed.P(0, 10)})
	punkt := sfnt.Segments{{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{fixed.P(3, 3)}}}

	for name, fall := range map[string]struct {
		umriss, erwartet sfnt.Segments
		um               float64
	}{
		"null":   {quadrat, quadrat, 0},
		"nichts": {nil, nil, 0.3},
		"ein einzelner Punkt": {
			punkt,
			sfnt.Segments{{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{{X: 3 * 64, Y: 3*64 - 19}}}},
			0.3,
		},
	} {
		dick := verdickt(fall.umriss, fall.um)

		if len(dick) != len(fall.erwartet) {
			t.Fatalf("%s: %d Stücke statt %d", name, len(dick), len(fall.erwartet))
		}

		for i := range dick {
			if dick[i] != fall.erwartet[i] {
				t.Errorf("%s: Stück %d steht bei %v, erwartet %v", name, i, dick[i].Args, fall.erwartet[i].Args)
			}
		}
	}
}

// **Ein Stamm wird um die gemessene Breite fetter** — die Probe aus der
// Messung, an Go Regular: die Summe der Deckung quer durch den Stamm von `l`
// wächst um zweimal die Verdickung.
func TestEinStammWirdUmDieGemesseneBreiteFetter(t *testing.T) {
	t.Parallel()

	schrift, err := SchriftAusDatei(goregular.TTF)
	if err != nil {
		t.Fatalf("Go Regular: %v", err)
	}

	kennung, _ := schrift.kennung('l')

	schrift.schloss.Lock()
	duenn, duennEcke, _ := schrift.maskeDerKennung(kennung, 56, 0)
	dick, dickEcke, _ := schrift.maskeDerKennung(kennung, 56, quartzVerdickung(56))
	schrift.schloss.Unlock()

	// **Dieselbe Zeile im Glyphenraum**, zwanzig Bildpunkte über der
	// Grundlinie: die Maske des dicken beginnt weiter oben.
	quer := func(maske image.Image, ecke image.Point) float64 {
		summe := 0.0
		for x := range maske.Bounds().Dx() {
			summe += float64(alphaAn(maske, x, -20-ecke.Y).A) / 255
		}

		return summe
	}

	if zuwachs := quer(dick, dickEcke) - quer(duenn, duennEcke); math.Abs(zuwachs-0.6) > 0.02 {
		t.Errorf("der Stamm wuchs um %.3f Bildpunkte, gemessen sind 0,6", zuwachs)
	}
}

// vieleck baut eine geschlossene Kontur aus Geraden.
func vieleck(ecken []fixed.Point26_6) sfnt.Segments {
	umriss := make(sfnt.Segments, 0, len(ecken)+1)
	umriss = append(umriss, sfnt.Segment{Op: sfnt.SegmentOpMoveTo, Args: [3]fixed.Point26_6{ecken[0]}})

	for _, ecke := range append(ecken[1:], ecken[0]) {
		umriss = append(umriss, sfnt.Segment{Op: sfnt.SegmentOpLineTo, Args: [3]fixed.Point26_6{ecke}})
	}

	return umriss
}
