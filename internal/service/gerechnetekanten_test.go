package service_test

// **`card_style.<feld>.top` darf gerechnet sein**
// (#449).
//
// Bei MSE sind die sechs Kanten `Scriptable<double>` und werden in
// `Style::update` ausgewertet (`src/data/field.cpp`, Zeilen 141–151, Fassung
// f9d9356 vom 2021-07-06) — vor allem anderen.
//
// Über `Full-Magic-Pack` sind **4.566 von 28.420** Kantenangaben Skripte;
// gezählt an der Zeile in den Dateien `style` (§10).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineGerechneteKanteStehtInCardStyle(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "text", Links: "48", Breite: "300", Hoehe: "183",
				// **Das `top:` als Skript** — so schreibt es
				// `magic-m15-mainframe-planeswalker`.
				Oben:    "{ 200 + 92 }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
			{
				// Ein zweites Feld liest die Kante des ersten — genau wie
				// `abil_center` es tut.
				Feld: "marke", Links: "10", Breite: "40", Hoehe: "20",
				Oben:    "{ card_style.text.top }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: "Regeltext"},
			{Field: "marke", Text: "+1"},
		}}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %+v", befunde)
	}

	if oben := ebeneMitText(t, plan, "+1").Rahmen.Oben; oben != 292 {
		t.Errorf("die Marke steht bei %.0f, erwartet 292", oben)
	}
}

// ebeneMitText sucht die Textebene mit diesem Klartext.
func ebeneMitText(t *testing.T, plan *render.Plan, text string) *render.Ebene {
	t.Helper()

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Klartext() == text {
			return &plan.Ebenen[i]
		}
	}

	t.Fatalf("der Plan führt keine Ebene %q", text)

	return nil
}
