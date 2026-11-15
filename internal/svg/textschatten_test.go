package svg_test

// **Der Schatten einer Schrift auch im SVG** — F02
// (#2244).
// Der zweite Verbraucher desselben Plans: was der Rasterer zeichnet, schreibt
// dieser Weg auch, sonst sähe eine ausgeführte SVG anders aus als die Vorschau.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

func TestDerSchattenStehtVorDemText(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Satz.Schatten = render.Textschatten{Farbe: render.Farbe{R: 255, A: 255}, X: -1, Y: 1}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	schatten := strings.Index(geschrieben, `fill="rgba(255,0,0,1)"`)
	text := strings.Index(geschrieben, `fill="rgba(0,0,0,1)"`)

	if schatten < 0 || text < 0 || schatten > text {
		t.Fatalf("der Schatten steht bei %d, der Text bei %d — erwartet beide, der Schatten zuerst", schatten, text)
	}

	if !strings.Contains(geschrieben, `<g transform="translate(-1,1)" aria-hidden="true">`) {
		t.Error("der Schatten ist nicht um (−1, 1) versetzt oder nicht verborgen")
	}

	if strings.Contains(geschrieben, "feGaussianBlur") {
		t.Error("ein harter Schatten bekam einen Weichzeichner")
	}
}

// **Ein weicher Schatten bekommt einen Weichzeichner.** MSE weicht `n`-mal mit
// dem Kern (2·Mitte + vier Nachbarn) / 6; seine Varianz ist je Achse ein
// Drittel je Durchgang, die Breite also √(n/3). Das ist eine Näherung —
// dort an Ort und Stelle und unsymmetrisch, hier ein Gaußfilter.
func TestEinWeicherSchattenBekommtEinenWeichzeichner(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Satz.Schatten = render.Textschatten{Farbe: render.Farbe{A: 255}, Weichheit: 3}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, `<feGaussianBlur stdDeviation="1"/>`) ||
		!strings.Contains(geschrieben, `filter="url(#`) {
		t.Errorf("kein Weichzeichner mit σ = 1: %s", geschrieben)
	}
}

// **Ein durchsichtiger Schatten schreibt nichts.**
func TestEinDurchsichtigerSchattenSchreibtNichts(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Satz.Schatten = render.Textschatten{X: 2, Y: 2, Weichheit: 2}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Count(geschrieben, ">Feuerball<") != 1 || strings.Contains(geschrieben, "translate(2,2)") {
		t.Error("ein durchsichtiger Schatten wurde geschrieben")
	}
}
