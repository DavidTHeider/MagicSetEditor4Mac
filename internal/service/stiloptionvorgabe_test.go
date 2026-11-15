package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Eine Stiloption ohne `initial:` hat ihre erste Wahl.**
//
//	ChoiceValue::ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice)
//	  : value( !field->initial.empty() ? field->initial
//	         : initial_first_choice    ? field->choices->choiceName(0)
//	         :                           _("") , true)
//
// `src/data/field/choice.cpp`, Zeilen 279–284, und `choice.hpp:192` — der
// Vorgabewert des Schalters ist **`true`**. Fassung f9d9356 vom 2021-07-06.
//
// # Was das gekostet hat
//
// Für **Kartenfelder** ist die Regel seit
// #341
// portiert, für **Stiloptionen** nicht. `magic-archenemy` schreibt
//
//	styling field:
//		type: choice
//		name: artist font color
//		choice: white
//		choice: black
//
// ohne `initial:` und setzt damit die Schriftfarbe des Fußes
// (`color: {styling.artist_font_color}`). Bei uns war der Wert **leer**, die
// Farbe fiel auf Schwarz zurück — auf MSE2s eigenem Bild steht dort **weiß**
// (#409).
//
// Über die 376 Stile des vollen Packs: **194** `styling field:` der Art
// `choice` nennen kein `initial:`, 855 nennen eines. Gezählt am Text der
// Stildateien (§10).
func TestEineStiloptionOhneInitialHatDieErsteWahl(t *testing.T) {
	feld := flaeche("name")
	feld.Schrift = msepaket.Schrift{
		Name: "Helvetica", Groesse: "12", Farbe: "{ styling.artist_font_color }",
	}

	stil := bildstil(feld)
	stil.Stiloptionen = []domain.StyleField{{
		Name: "artist font color", Kind: domain.KindChoice, Choices: []string{"white", "black"},
	}}

	plan, befunde := gedeutetMit(t, stilkarte("rot"), stil)

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if gesetzt := textebene(t, plan).Farbe; gesetzt != weiss {
		t.Errorf("Farbe %+v, erwartet Weiß — die erste Wahl der Option", gesetzt)
	}
}

// **Mit `initial:` gewinnt das `initial:`** — die andere Hälfte derselben
// Regel: `!field->initial.empty() ? field->initial : …`.
func TestEineStiloptionMitInitialBehaeltEs(t *testing.T) {
	feld := flaeche("name")
	feld.Schrift = msepaket.Schrift{
		Name: "Helvetica", Groesse: "12", Farbe: "{ styling.artist_font_color }",
	}

	stil := bildstil(feld)
	stil.Stiloptionen = []domain.StyleField{{
		Name: "artist font color", Kind: domain.KindChoice,
		Initial: "black", Choices: []string{"white", "black"},
	}}

	plan, _ := gedeutetMit(t, stilkarte("rot"), stil)

	if gesetzt := textebene(t, plan).Farbe; gesetzt != (render.Farbe{A: 255}) {
		t.Errorf("Farbe %+v, erwartet Schwarz", gesetzt)
	}
}
