package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Jede Ebene weiß, aus welchem Feld sie kommt**
// (#246).
//
// Daran hängen zwei Dinge: ein Klick in die Vorschau soll das Feld darunter
// wählen, und ein Befund soll sein Feld nennen statt nur seiner Nummer.
func TestJedeEbeneNenntIhrFeld(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "text", Links: "10", Oben: "20", Breite: "300", Hoehe: "100",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{Feld: "card color", Links: "0", Oben: "0", Breite: "375", Hoehe: "523"},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "Fliegend."},
		{Field: "card color", Text: "rot"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{
			Stil:         stil,
			Kartenfelder: []domain.Field{{Name: "card color", Kind: domain.KindColor}},
		})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	gefunden := map[string]bool{}

	for _, ebene := range plan.Ebenen {
		if ebene.Feld != "" {
			gefunden[ebene.Feld] = true
		}
	}

	for _, erwartet := range []string{"text", "card color"} {
		if !gefunden[erwartet] {
			t.Errorf("keine Ebene nennt das Feld %q", erwartet)
		}
	}

	// **Der weiße Grund gehört zu keinem Feld.** Er ist unsere Zutat, damit
	// eine Karte ohne Rahmenbild nicht durchsichtig ist — und ein Klick darauf
	// soll kein Feld aufschlagen.
	if plan.Ebenen[0].Art != render.ArtFlaeche || plan.Ebenen[0].Feld != "" {
		t.Errorf("die erste Ebene ist %+v", plan.Ebenen[0])
	}
}
