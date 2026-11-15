package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Feld der Art `multiple choice` steht in Deklarationsreihenfolge**
// (#556).
//
// # Was MSE damit tut
//
//	void MultipleChoiceValue::normalForm() {
//	  String& val = value.mutateDontChangeDefault();
//	  vector<bool> seen(field().choices->lastId());
//	  …
//	  // now put them back in the right order
//	  val.clear();
//	  for (size_t i = 0 ; i < seen.size() ; ++i) {
//	    if (seen[i]) {
//	      if (!val.empty()) val += _(", ");
//	      val += field().choices->choiceName((int)i);
//	    }
//	  }
//	}
//
// `src/data/field/multiple_choice.cpp`, Zeilen 82–110, gerufen aus
// `MultipleChoiceValue::update`, Zeile 59. Fassung f9d9356 vom 2021-07-06.
//
// # Woran es aufgefallen ist
//
// `magic.mse-game/script`, Zeile 3605, baut daraus einen Dateinamen:
//
//	shape := remove_comma(loyalty_cost_box_field(input + (instance-1)*8))
//	name  := "loyalty cost " + shape + ".png"
//
// und `remove_comma` streicht **nur** Kommata. Ohne Normalform suchte dieses
// Programm `loyalty cost shadow 0.png`, wo `loyalty cost 0 shadow.png` liegt —
// über 790 kanonische Karten **zehnmal**, und das war jeder verbliebene
// Vorlagenbefund.
func TestEineMehrfachwahlStehtInDeklarationsreihenfolge(t *testing.T) {
	faelle := []struct {
		name, wert, erwartet string
	}{
		{"der Fall aus dem Pack", "shadow, 0", "0, shadow"},
		{"schon in Ordnung bleibt", "0, shadow", "0, shadow"},
		{"eine einzelne Wahl", "shadow", "shadow"},
		{"drei, verdreht", "ult, +, shadow", "+, ult, shadow"},

		// **Ohne Leerzeichen nach dem Komma** — MSE überliest es
		// (`normalForm`, Zeile 87) und schreibt beim Zusammensetzen immer
		// `", "`.
		{"ohne Leerzeichen geschrieben", "shadow,0", "0, shadow"},

		// **Was keine Wahl ist, fällt weg.** MSE trägt nur ein, was es in der
		// Liste findet.
		{"eine unbekannte Wahl", "gibtsnicht, 0", "0"},

		// **Ein leerer Wert steht in `mehrfachwahl_einheit_test.go`** — er
		// ergibt keine Textebene, und das ist eine andere Regel als diese.
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), &domain.CardRow{
					ID: 1, Number: 1,
					Values: []domain.Value{{Field: "kostenkasten", Text: fall.wert}},
				}, mehrfachwahlvorlage())
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Art != render.ArtText {
					continue
				}

				if ebene.Klartext() != fall.erwartet {
					t.Errorf("der Wert lautet %q, erwartet %q",
						ebene.Klartext(), fall.erwartet)
				}

				return
			}

			t.Fatal("der Plan führt keine Textebene")
		})
	}
}

// mehrfachwahlvorlage baut eine Vorlage mit einem `multiple choice`-Feld.
//
// **Dieselbe Reihenfolge wie `choices_cost` im Pack** — sie ist der Gegenstand
// der Prüfung.
func mehrfachwahlvorlage() service.Vorlage {
	return service.Vorlage{
		Stil: msepaket.Stil{
			Breite: 375, Hoehe: 523,
			Felder: []msepaket.Feldstil{{
				Feld: "kostenkasten", Links: "0", Oben: "0",
				Breite: "375", Hoehe: "100",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			}},
		},
		Kartenfelder: []domain.Field{{
			Name: "kostenkasten",
			Kind: domain.KindMultipleChoice,
			Choices: []string{
				"none", "anti ult", "+", "0", "-", "ult", "shadow",
			},
		}},
	}
}
