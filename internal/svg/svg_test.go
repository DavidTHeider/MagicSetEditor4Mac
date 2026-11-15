// **Eine Karte als SVG aus dem Zeichenplan** (#837), mit dem **ganzen** Plan
// (#852).
package svg_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// TestJedeEbeneStehtImSVG hält die Zeichenreihenfolge.
func TestJedeEbeneStehtImSVG(t *testing.T) {
	t.Parallel()

	geschrieben, _, fehler := svg.AusPlan(t.Context(), probeplan(), mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.HasPrefix(geschrieben, "<svg") {
		t.Errorf("fängt nicht mit <svg an: %.40s", geschrieben)
	}

	// **Die Reihenfolge ist die Zeichenreihenfolge** — hinten nach vorn. Wer
	// sie vertauscht, malt den Rahmen über den Text.
	stellen := []int{
		strings.Index(geschrieben, `data-feld="grund"`),
		strings.Index(geschrieben, `data-feld="rahmen"`),
		strings.Index(geschrieben, `data-feld="name"`),
	}

	vorige := -1

	for i, stelle := range stellen {
		if stelle < 0 {
			t.Fatalf("Ebene %d fehlt im SVG", i)
		}

		if stelle < vorige {
			t.Errorf("Ebene %d steht vor Ebene %d", i, i-1)
		}

		vorige = stelle
	}
}

// **Text steht als Text, nicht als Bild.**
//
// Das ist der ganze Gewinn des Formats: markierbar, beliebig skalierbar,
// durchsuchbar. Ein `<image>` an dieser Stelle wäre der Name ohne die Sache.
func TestTextStehtAlsText(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Satz.Kursivschrift = "Helvetica Oblique"

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	// **Die Kursive steht als Schriftname da, nicht als `font-style`** — ein
	// Betrachter, der `italic` sieht, erfindet eine, und ihre Vorschübe sind
	// andere als die, mit denen hier gesetzt wurde. Gesehen bei der
	// Sichtprüfung an vier echten Karten: der Flavortext lief zusammen.
	for _, teil := range []string{
		">Feuerball<", `font-family="Helvetica Oblique"`, `font-weight="bold"`,
		`font-family="Helvetica"`,
	} {
		if !strings.Contains(geschrieben, teil) {
			t.Errorf("%q fehlt im SVG", teil)
		}
	}
}

// **Eine Fläche ist Geometrie.**
func TestEineFlaecheIstGeometrie(t *testing.T) {
	t.Parallel()

	geschrieben, _, fehler := svg.AusPlan(t.Context(), probeplan(), mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, "<rect") {
		t.Error("die Fläche steht nicht als <rect>")
	}
}

// **Ein Bild wird als Datenadresse eingebettet.**
func TestEinBildWirdEingebettet(t *testing.T) {
	t.Parallel()

	geschrieben, _, fehler := svg.AusPlan(t.Context(), probeplan(), mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, `href="data:image/png;base64,`) {
		t.Error("das Bild steht nicht als Datenadresse im SVG")
	}
}

// **Ein Kartenname mit Sonderzeichen bricht das SVG nicht.**
//
// „Ach! Hans, run!" ist eine echte Magic-Karte; Anführungszeichen und
// kaufmännisches Und kommen in Kartennamen vor, und der Text steht hier
// sowohl zwischen Marken als auch **in einem Attribut**.
func TestSonderzeichenBrechenDasSVGNicht(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	textebene := &plan.Ebenen[2]
	textebene.Feld = `na"me`
	textebene.Abschnitte = []render.Abschnitt{{Text: "Ach! <Hans> & \"run\"\x1b"}}

	geschrieben, _, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, `<Hans>`) {
		t.Error("die spitzen Klammern stehen roh im SVG")
	}

	if strings.Contains(geschrieben, `data-feld="na"me"`) {
		t.Error("das Anführungszeichen bricht das Attribut")
	}

	// **Und ein Steuerzeichen ist gar keines mehr.** XML 1.0 verbietet sie
	// (§2.2), auch als Zahlenverweis — ein `\x1b` aus einem Paket von 2007
	// machte die ganze Datei unlesbar.
	if strings.ContainsRune(geschrieben, '\x1b') {
		t.Error("ein Steuerzeichen steht im SVG")
	}
}

// **Ohne Plan gibt es einen Fehler und kein leeres SVG.**
func TestOhnePlanGibtEsEinenFehler(t *testing.T) {
	t.Parallel()

	if _, _, fehler := svg.AusPlan(t.Context(), nil, mitAllem()); !errors.Is(fehler, render.ErrPlan) {
		t.Errorf("Fehler %v, erwartet war ErrPlan", fehler)
	}
}

// **Ein abgebrochener Kontext hält den Lauf an.**
//
// Ein Stapel von 3.700 Karten in Druckauflösung ist eine Zahl; wer abbricht,
// soll nicht auf das Ende warten.
func TestEinAbgebrochenerKontextHaeltAn(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, _, fehler := svg.AusPlan(ctx, probeplan(), mitAllem()); !errors.Is(fehler, context.Canceled) {
		t.Errorf("Fehler %v, erwartet war context.Canceled", fehler)
	}
}
