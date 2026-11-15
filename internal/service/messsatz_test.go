package service_test

// **Die Messstufe setzt wie die Zeichenstufe**
// (#451).
//
// `content_*` und `layout` entstehen bei MSE aus der **tatsächlichen**
// Textbereitung: `TextViewer::prepareLines` setzt mit dem echten Stil, und
// `extractLayoutInfo` liest die Blöcke aus denselben Zeilen ab, die gezeichnet
// werden (`src/render/text/viewer.cpp`, Zeilen 348–389, Fassung f9d9356 vom
// 2021-07-06). Einen zweiten, ärmeren Satz gibt es dort nicht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieMessstufeKenntDieZeilenfaktoren(t *testing.T) {
	// **Zwei Absätze, harter Umbruch dazwischen.** Bei Faktor 1 stünde der
	// zweite auf 12, bei `line height hard: 2` auf 24.
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "text", Links: "10", Oben: "100", Breite: "300", Hoehe: "150",
				Schrift:      msepaket.Schrift{Name: "Probe", Groesse: "10"},
				Zeilenhoehen: msepaket.Zeilenhoehenangabe{Hart: "2"},
			},
			{
				// Liest die Mitte des zweiten Absatzes — so rechnet
				// `abil_center` bei den Planeswalkern.
				Feld: "marke", Links: "10", Breite: "40", Hoehe: "20",
				Oben:    "{ card_style.text.layout.paragraphs[1].middle }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: "eins\nzwei"},
			{Field: "marke", Text: "+1"},
		}}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %+v", befunde)
	}

	// Zeilenhöhe 12; der zweite Absatz beginnt bei 24 und ist 12 hoch, seine
	// Mitte liegt auf 30. Ohne die Faktoren wäre sie 18.
	if oben := ebeneMitText(t, plan, "+1").Rahmen.Oben; oben != 30 {
		t.Errorf("die Marke steht bei %.0f, erwartet 30", oben)
	}
}
