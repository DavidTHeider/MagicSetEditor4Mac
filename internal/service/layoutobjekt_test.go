package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`card_style.<feld>.layout` ist das gesetzte Layout, nicht eine Zahl.**
//
// MSE gibt der Vorlage das ganze Objekt: `lines`, `paragraphs`, `blocks` und
// `separators`, je Eintrag mit `width`, `top`, `height`, `bottom` und `middle`
// (`src/data/field/text.cpp`, Zeilen 114–133, und `src/data/field/text.hpp`,
// Zeilen 50–67, Fassung f9d9356 vom 2021-07-06).
//
// **Es ist die größte verbliebene Befundklasse.** Über die 190 Stile von
// `M15-Magic-Pack` schlagen rund 250 Befunde auf `"blocks" gibt es nur an einer
// Abbildung` und Verwandte; die Planeswalker-Stile setzen damit jede
// Fähigkeitszeile
// (#338).
func TestDasLayoutObjektStehtDerVorlageZurVerfuegung(t *testing.T) {
	faelle := []struct {
		name     string
		angabe   string
		erwartet float64
	}{
		// Drei Zeilen à 12 Punkte (Größe 10 × Zeilenhöhe 1,2).
		{"die Zahl der Zeilen", `{ length(card_style.text.layout.lines) }`, 3},
		{"die Zahl der Absätze", `{ length(card_style.text.layout.paragraphs) }`, 2},
		{"ein Block, solange `<line>` nicht schneidet", `{ length(card_style.text.layout.blocks) }`, 1},
		{"die Unterkante des ersten Blocks", `{ card_style.text.layout.blocks[0].bottom }`, 36},
		{"die Mitte des ersten Blocks", `{ card_style.text.layout.blocks[0].middle }`, 18},
		// Der zweite Absatz beginnt mit der **zweiten** Zeile und trägt
		// deren beide — Oberkante 12, Höhe 24.
		{"die Oberkante des zweiten Absatzes", `{ card_style.text.layout.paragraphs[1].top }`, 12},
		{"die Höhe des zweiten Absatzes", `{ card_style.text.layout.paragraphs[1].height }`, 24},
		{"die Breite des Ganzen", `{ card_style.text.layout.width }`, 95},

		// **Ohne Blockgrenze gibt es keinen Trenner** — er liegt zwischen
		// zweien.
		{"keine Trenner", `{ length(card_style.text.layout.separators) }`, 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{
					{
						Feld: "text", Links: "0", Oben: "0", Breite: "100", Hoehe: "154",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
					{
						Feld: "melder", Links: fall.angabe,
						Oben: "0", Breite: "50", Hoehe: "12",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
				},
			}

			// Ein Zeichen ist 5 breit, der Kasten 100: zwanzig passen in eine
			// Zeile. Der zweite Absatz hat 24 Zeichen und bricht einmal um —
			// drei Zeilen, zwei Absätze.
			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "text", Text: "aaaaaaaa\nbbbb cccc dddd eeee ffff"},
				{Field: "melder", Text: "x"},
			}}

			plan, befunde, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, befund := range befunde {
				t.Logf("Befund %s: %s", befund.Ort, befund.Meldung)
			}

			for stelle := range plan.Ebenen {
				if plan.Ebenen[stelle].Art == render.ArtText &&
					plan.Ebenen[stelle].Rahmen.Hoehe == 12 {
					if links := plan.Ebenen[stelle].Rahmen.Links; links != fall.erwartet {
						t.Errorf("%s ergibt %g, erwartet %g", fall.angabe, links, fall.erwartet)
					}

					return
				}
			}

			t.Fatal("der Plan führt das Feld `melder` nicht")
		})
	}
}
