package render_test

// **Ein fehlendes Symbolbild in einem Textzug meldet ein fehlendes Bild**
// (#422).
//
// Bis dahin fiel es in dieselbe Meldung wie ein fehlender Glyph: „in der
// Schrift MPlantin fehlen Zeichen: /magic-mana-small.mse-symbol-font/…". Wer
// das las, suchte eine Schrift — und der Mangel lag im Bild.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestEinFehlendesSymbolbildImTextMeldetEinBild(t *testing.T) {
	plan := &render.Plan{Breite: 60, Hoehe: 20, Ebenen: []render.Ebene{{
		Art:    render.ArtText,
		Rahmen: render.Rechteck{Breite: 60, Hoehe: 20},
		Abschnitte: []render.Abschnitt{
			{Text: "A"},
			{
				Text:        "R",
				Symbole:     []render.Symbol{{Bild: render.Blatt("/fehlt.png")}},
				Symbolmasse: render.Symbolmasse{Groesse: 10, Bildgroesse: 10},
			},
		},
		Satz:    render.Satz{Schrift: "Probe", Groesse: 10, Ausrichtung: render.LinksOben},
		Deckung: 1,
	}}}

	_, befunde, err := render.NeuerRasterer(bildmasse{}).
		MitSchriften(schriften{"Probe": einheitsschrift{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %+v", len(befunde), befunde)
	}

	if strings.Contains(befunde[0].Meldung, "fehlen Zeichen") {
		t.Errorf("die Meldung spricht von Zeichen: %s", befunde[0].Meldung)
	}

	if !strings.Contains(befunde[0].Meldung, "Symbolbilder") ||
		!strings.Contains(befunde[0].Meldung, "/fehlt.png") {
		t.Errorf("die Meldung nennt das fehlende Bild nicht: %s", befunde[0].Meldung)
	}
}

// **Ein Symbol ohne Bild behält sein Quadrat.**
//
// Der Baum ist dann leer — ein `symbol:`-Block ohne `image:`, oder eine
// Angabe, die sich nicht ausrechnen ließ. Eine Breite von null setzte das
// Symbol auf keine Breite, und der Rest der Zeile rutschte nach links; der
// Befund zeigte dann auf eine Stelle, an der nichts mehr steht.
func TestEinSymbolOhneBildBehaeltSeinQuadrat(t *testing.T) {
	breite, hoehe := render.Symbolzugmass(
		[]render.Symbol{{}},
		render.Symbolmasse{Groesse: 10, Bildgroesse: 10, AbstandWaagerecht: 2},
		bildmasse{})

	if breite != 12 || hoehe != 10 {
		t.Errorf("%v × %v, erwartet 12 × 10", breite, hoehe)
	}
}
