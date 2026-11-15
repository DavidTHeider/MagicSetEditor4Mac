package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`content_lines` ist die Zahl der Zeilen **nach** dem Umbruch.**
//
// # Der Befund
//
// MSE gibt sie aus dem gesetzten Text zurück:
//
//	REFLECT_N("content_lines",  layout->lines.size());
//
// `src/data/field/text.cpp`, Zeile 120, Fassung f9d9356 vom 2021-07-06.
//
// **221-mal steht sie in den 190 Stilen von `M15-Magic-Pack`**, 174-mal für das
// Feld `text`. `magic-m15` entscheidet damit die Ausrichtung des Regeltextes:
//
//	if (… and card_style.text.content_lines <= 2) or styling.center_text == "always"
//	then "middle center" else "middle left"
//
// Ohne die Zahl rechnet diese Bedingung mit nichts — und der Regeltext einer
// kurzen Karte steht links statt mittig
// (#291).
func TestDieZeilenzahlKommtAusDemUmbruch(t *testing.T) {
	faelle := []struct {
		name     string
		text     string
		erwartet float64
	}{
		// Ein Zeichen ist 5 breit, der Kasten 100: zwanzig passen in eine Zeile.
		{"eine Zeile", "aaaa bbbb", 1},
		{"zwei Zeilen", "aaaaaaaaaa bbbbbbbbbb ccccc", 2},
		{"ein harter Umbruch zählt mit", "aa\nbb\ncc", 3},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := zeilenplan(t, fall.text)

			// Das Feld `probe` legt seine linke Kante auf die Zeilenzahl —
			// so wird sie sichtbar, ohne dass der Test in die Umgebung greift.
			for _, ebene := range plan.Ebenen {
				if ebene.Art == render.ArtText && ebene.Rahmen.Hoehe == 12 {
					if ebene.Rahmen.Links != fall.erwartet {
						t.Errorf("content_lines ist %g, erwartet %g",
							ebene.Rahmen.Links, fall.erwartet)
					}

					return
				}
			}

			t.Fatal("der Plan führt das Feld `probe` nicht")
		})
	}
}

// zeilenplan baut einen Plan mit einem Regeltextfeld und einem Melder daneben.
func zeilenplan(t *testing.T, text string) *render.Plan {
	t.Helper()

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "text", Links: "0", Oben: "0", Breite: "100", Hoehe: "154",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{
				Feld: "probe", Links: "{ card_style.text.content_lines }",
				Oben: "0", Breite: "50", Hoehe: "12",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: text},
		{Field: "probe", Text: "x"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return plan
}
