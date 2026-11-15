package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Ein unsichtbares Zeichen fehlt nicht — es hat keine Gestalt.**
//
// Ein Nullbreiten-Leerzeichen (U+200B) trennt Wörter, ohne selbst etwas zu
// zeichnen; dasselbe gilt für den weichen Trennstrich (U+00AD) und die
// Byte-Reihenfolgemarke (U+FEFF). Eine Schrift führt für sie keine Glyphe, und
// **das ist ihre Bauform**, kein Mangel.
//
// # Was das gekostet hat
//
// Der Lauf über die selbst geschriebenen Demokarten meldete an einer
// Saga-Karte zweimal „in der Schrift MPlantin fehlen Zeichen: ​" — mit einem
// Zeichen, das niemand sehen kann. Eine Meldung, die auf nichts zeigt, ist
// schlimmer als keine: sie steht in derselben Liste wie die echten
// (#270).
func TestEinUnsichtbaresZeichenFehltNicht(t *testing.T) {
	schrift := echteSchrift(t)

	plan := func(text string) *render.Plan {
		return &render.Plan{Breite: 200, Hoehe: 40, Ebenen: []render.Ebene{{
			Art:        render.ArtText,
			Rahmen:     render.Rechteck{Breite: 200, Hoehe: 40},
			Satz:       render.Satz{Schrift: "Andale", Groesse: 12},
			Abschnitte: []render.Abschnitt{{Text: text}},
			Farbe:      render.Farbe{A: 255},
			Deckung:    1,
		}}}
	}

	for _, fall := range []struct {
		name, text string
		meldet     bool
	}{
		{name: "Nullbreiten-Leerzeichen", text: "ab\u200bcd"},
		{name: "weicher Trennstrich", text: "ab\u00adcd"},
		{name: "Reihenfolgemarke", text: "ab\ufeffcd"},
		// **Ein Zeichen, das wirklich fehlt, wird weiterhin benannt.** Sonst
		// hielte diese Änderung die halbe Zusage und machte die andere
		// unsichtbar (§10).
		{name: "ein wirklich fehlendes Zeichen", text: "ab\U0010FFFDcd", meldet: true},
	} {
		t.Run(fall.name, func(t *testing.T) {
			_, befunde, err := render.NeuerRasterer(bilder{}).
				MitSchriften(schriften{"Andale": schrift}).
				Zeichne(t.Context(), plan(fall.text), 1, render.GueteMittel)
			if err != nil {
				t.Fatalf("zeichnen: %v", err)
			}

			if (len(befunde) > 0) != fall.meldet {
				t.Errorf("Befunde: %v, gemeldet erwartet: %v", befunde, fall.meldet)
			}
		})
	}
}
