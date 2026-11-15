package render_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// gleichbreit ist eine Schrift, in der jedes Zeichen gleich breit ist.
type gleichbreit struct{ breite float64 }

func (g gleichbreit) Breite(rune) float64 { return g.breite }
func (gleichbreit) Zeilenhoehe() float64  { return 1.2 }

// **Die Breite eines Textzugs, ohne ihn zu setzen.**
//
// # Wofür das gebraucht wird
//
// `card_style.<feld>.content_width` fragt, wie breit der **gesetzte Inhalt**
// eines Feldes ist — und Feldangaben rechnen damit, bevor der Rahmen feststeht.
// [render.Setze] kann das nicht beantworten: es braucht einen Rahmen, um
// umzubrechen, und der Rahmen ist hier gerade das Gesuchte.
//
// Gemessen wird deshalb der Zug, **wie er ohne Umbruch dastünde**. Das ist die
// Frage, die `content_width` stellt: nicht „wie hoch baut sich der Text in
// diesem Kasten", sondern „wie breit ist er".
//
// # Warum die längste Zeile und nicht die Summe
//
// Ein Text mit einem harten Umbruch besteht aus mehreren Zeilen. Seine Breite
// ist die der **längsten** — die Summe wäre die Breite, die er hätte, wenn man
// ihn wieder zu einer Zeile zusammenzöge, und das tut niemand.
func TestDieBreiteEinesTextzugsOhneUmbruch(t *testing.T) {
	schrift := gleichbreit{breite: 0.5}

	faelle := []struct {
		name     string
		text     string
		groesse  float64
		erwartet float64
	}{
		{name: "eine Zeile", text: "abcd", groesse: 10, erwartet: 20},
		{name: "die Größe geht ein", text: "abcd", groesse: 20, erwartet: 40},
		{name: "die längste Zeile gewinnt", text: "ab\nabcdef\nabc", groesse: 10, erwartet: 30},
		{name: "leerer Text ist null breit", text: "", groesse: 10, erwartet: 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if breite := render.Textbreite(fall.text, fall.groesse, schrift); breite != fall.erwartet {
				t.Errorf("%.1f breit, erwartet %.1f", breite, fall.erwartet)
			}
		})
	}
}

// **Ohne Schrift keine Messung, und keine erfundene Zahl.**
//
// Eine Breite von null ist hier kein Messergebnis, sondern die Auskunft „nicht
// gemessen". Der Aufrufer unterscheidet das an der Null; eine geschätzte Zahl
// sähe aus wie eine Messung und wäre keine.
func TestOhneSchriftIstDieTextbreiteNull(t *testing.T) {
	if breite := render.Textbreite("abcd", 10, nil); breite != 0 {
		t.Errorf("%.1f breit, erwartet 0", breite)
	}
}
