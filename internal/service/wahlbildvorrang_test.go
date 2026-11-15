package service_test

// **`choice images:` gelten nur ohne `image:`**
// (#454).
//
//	void ChoiceStyle::initImage() {
//	  if (image.isSet() || choice_images.empty()) return;
//
// `src/data/field/choice.cpp`, Zeile 172, Fassung f9d9356 vom 2021-07-06.
// Steht ein `image:` da, werden die Wahlbilder **gar nicht erst** in das
// Skript übersetzt — sie sind dann tot. Ergibt das `image:` nichts, wird
// nichts gezeichnet.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinLeeresBildSchlaegtDieWahlbilder(t *testing.T) {
	for _, fall := range []struct {
		name     string
		bild     string
		erwartet string
	}{
		{
			// **Der Fall aus dem Pack**: `card stamp` führt beides, und sein
			// Skript gibt für „none" die leere Zeichenkette.
			name: "mit image:, das nichts ergibt", bild: `{ "" }`,
			erwartet: "",
		},
		{
			name: "mit image:, das etwas ergibt", bild: `{ "gestempelt.png" }`,
			erwartet: "/probe.mse-style/gestempelt.png",
		},
		{
			// **Die Gegenprobe**: ohne `image:` gilt die Wahlliste weiter.
			name: "ohne image:", bild: "",
			erwartet: "/probe.mse-style/keiner.png",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
				Felder: []msepaket.Feldstil{{
					Feld: "stempel", Links: "166", Oben: "473", Breite: "43", Hoehe: "43",
					Bild:       fall.bild,
					Wahlbilder: []msepaket.Wahlbild{{Wahl: "keiner", Bild: "keiner.png"}},
				}},
			}

			plan, _, err := service.NeuerPlanbauer().
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "stempel", Text: "keiner"},
				}}, service.Vorlage{
					Stil:         stil,
					Kartenfelder: []domain.Field{{Name: "stempel", Kind: domain.KindChoice}},
				})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			var gefunden string

			for i := range plan.Ebenen {
				if plan.Ebenen[i].Art == render.ArtBild {
					gefunden = plan.Ebenen[i].Bild.Beschreibung()
				}
			}

			if gefunden != fall.erwartet {
				t.Errorf("Bild %q, erwartet %q", gefunden, fall.erwartet)
			}
		})
	}
}
