package service_test

// **Das Layout beschreibt den Satz, der gezeichnet wird — samt Ausrichtung**
// (#584,
// #494).
//
//	// store information about the content/layout, allow this to change alignment
//	if (style.alignment.isScripted()) {
//	  style.layout = extractLayoutInfo();
//	  style.alignment.update(ctx);
//	}
//	// align
//	alignLines(dc, chars, style);
//	…
//	// make layout available to scripts
//	style.layout = extractLayoutInfo();
//
// `TextViewer::prepareLines`, `src/render/text/viewer.cpp`, Zeilen 405–425,
// Fassung f9d9356 vom 2021-07-06. **Das Layout, das die Skripte sehen, wird
// nach dem Ausrichten abgelesen** — die zweite Zuweisung, nicht die erste. Die
// erste gibt es nur, damit ein Ausrichtungsskript überhaupt etwas zu lesen hat.
//
// # Was das gekostet hat
//
// `magic-m15-saga.mse-style` setzt die Laschen einer Class-Karte über
//
//	lev2 := { … card_style.text.top + card_style.text.layout.separator[1] - 13 … }
//
// Unser Messsatz richtete **gar nicht** aus — er stand immer oben. Bei
// `middle`-Ausrichtung und einem Kasten, der höher ist als der Text, fehlt
// damit genau die halbe Differenz, und die Laschen landeten mitten im
// Regeltext.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDasLayoutTraegtDieAusrichtung(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		ausrichtung string
		erwartet    float64
	}{
		// Drei Zeilen à 12 in einem Kasten von 100: der Text ist 36 hoch, es
		// bleiben 64. Oben ausgerichtet beginnt die erste Zeile bei 0 …
		"oben": {"top left", 0},

		// … mittig bei 32 …
		"mittig": {"middle left", 32},

		// … und unten bei 64.
		"unten": {"bottom left", 64},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{
					{
						Feld: "text", Links: "0", Oben: "0", Breite: "100", Hoehe: "100",
						Ausrichtung: fall.ausrichtung,
						Schrift:     msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
					{
						Feld: "melder", Links: `{ card_style.text.layout.lines[0].top }`,
						Oben: "200", Breite: "50", Hoehe: "12",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
				},
			}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "text", Text: "aaaaaaaa\nbbbb cccc dddd eeee ffff"},
				{Field: "melder", Text: "x"},
			}}

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for stelle := range plan.Ebenen {
				if plan.Ebenen[stelle].Art == render.ArtText &&
					plan.Ebenen[stelle].Rahmen.Oben == 200 {
					if links := plan.Ebenen[stelle].Rahmen.Links; links != fall.erwartet {
						t.Errorf("die erste Zeile steht bei %g, erwartet %g",
							links, fall.erwartet)
					}

					return
				}
			}

			t.Fatal("der Plan führt das Feld `melder` nicht")
		})
	}
}
