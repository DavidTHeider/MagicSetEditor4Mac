package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Bildreihe nimmt gerechnete Angaben und den Winkel des Felds** —
// Befund 4 des Zweitblicks über E/F.
//
// MSE rechnet `direction` und `alignment` wie jede Stilangabe aus
// (`MultipleChoiceStyle::update`, `src/data/field/multiple_choice.cpp`,
// Zeilen 42–47) und zeichnet in einem Kontext, der um den Ursprung des
// Feldkastens gedreht ist (`Rotation::tr`, `ValueViewer::getRotation`,
// `src/render/value/viewer.cpp`, Zeilen 44–46; `DrawPreRotatedImage` in
// `src/render/value/multiple_choice.cpp`, Zeile 64; Fassung f9d9356 vom
// 2021-07-06). Die Reihe rückt also im gedrehten Rahmen weiter.
func TestDieBildreiheNimmtGerechneteAngabenUndDenWinkel(t *testing.T) {
	quelle := massquelle{
		"/probe.mse-style/flying.png": {X: 20, Y: 16},
		"/probe.mse-style/haste.png":  {X: 18, Y: 18},
	}

	for _, fall := range []struct {
		name, ausrichtung, richtung, winkel string
		erwartet                            []render.Rechteck
		gedreht                             float64
	}{
		{
			name: "eine gerechnete Richtung", richtung: `{ "top to bottom" }`,
			erwartet: []render.Rechteck{{Links: 42, Oben: 321, Breite: 20, Hoehe: 16}, {Links: 42, Oben: 347, Breite: 18, Hoehe: 18}},
		},
		{
			name: "eine gerechnete Ausrichtung", ausrichtung: `{ "middle left" }`,
			erwartet: []render.Rechteck{{Links: 42, Oben: 333, Breite: 20, Hoehe: 16}, {Links: 72, Oben: 333, Breite: 18, Hoehe: 18}},
		},
		{
			// Das zweite Bild steht 30 Punkte weiter im Feld; um 90 Grad
			// gegen den Uhrzeigersinn gedreht heißt das: 30 Punkte nach oben.
			name: "um 90 Grad gedreht", winkel: "90",
			erwartet: []render.Rechteck{{Links: 42, Oben: 321, Breite: 20, Hoehe: 16}, {Links: 42, Oben: 291, Breite: 18, Hoehe: 18}},
			gedreht:  90,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
				Felder: []msepaket.Feldstil{{
					Feld: "keyword icons", Links: "42", Oben: "321", Breite: "297", Hoehe: "24",
					Darstellung: "image list", Ausrichtung: fall.ausrichtung, Richtung: fall.richtung,
					Abstand: "10", Winkel: fall.winkel,
					Wahlbilder: []msepaket.Wahlbild{
						{Wahl: "flying", Bild: "flying.png"},
						{Wahl: "haste", Bild: "haste.png"},
					},
				}},
			}

			plan, befunde, err := service.NeuerPlanbauer().MitBildern(quelle).Stilplan(t.Context(),
				&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{{Field: "keyword icons", Text: "flying, haste"}}},
				service.Vorlage{Stil: stil, Kartenfelder: []domain.Field{{
					Name: "keyword icons", Kind: domain.KindMultipleChoice, Choices: []string{"flying", "haste"},
				}}})
			if err != nil || len(befunde) > 0 {
				t.Fatalf("Plan: %v, Befunde %+v", err, befunde)
			}

			var gezeichnet []render.Ebene

			for _, ebene := range plan.Ebenen {
				if ebene.Art == render.ArtBild && ebene.Feld == "keyword icons" {
					gezeichnet = append(gezeichnet, ebene)
				}
			}

			if len(gezeichnet) != len(fall.erwartet) {
				t.Fatalf("%d Bilder, erwartet %d", len(gezeichnet), len(fall.erwartet))
			}

			for i := range gezeichnet {
				if gezeichnet[i].Rahmen != fall.erwartet[i] || gezeichnet[i].Winkel != fall.gedreht {
					t.Errorf("Bild %d steht bei %+v mit %v Grad, erwartet %+v mit %v Grad",
						i, gezeichnet[i].Rahmen, gezeichnet[i].Winkel, fall.erwartet[i], fall.gedreht)
				}
			}
		})
	}
}
