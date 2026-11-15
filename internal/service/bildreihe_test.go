package service_test

// **Eine Mehrfachwahl als Reihe ihrer Wahlbilder** — #2297.
//
//	} else if (style().render_style & RENDER_LIST) {
//	  // render only selected choices
//	  FOR_EACH(choice, selected) {
//	    drawChoice(dc, pos, choice);
//	  }
//
// `MultipleChoiceValueViewer::draw`, `src/render/value/multiple_choice.cpp`,
// Zeilen 23–49, Fassung f9d9356 vom 2021-07-06. Je Wahl das Bild in **eigener
// Größe** an `pos` (`GeneratedImage::Options(0,0,…)`, Zeilen 59–70), dann
// `pos = move_in_direction(direction, pos, size, spacing)` (Zeile 82) — auch
// für eine Wahl ohne Bild, dann um den Abstand allein. `pos` beginnt bei
// `align_in_rect(alignment, RealSize(0,0), rect)` (Zeile 27): bei `middle`
// also auf halber Höhe, und das Bild **beginnt** dort.

import (
	"errors"
	"image"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// massquelle kennt nur die Größe jedes Bilds.
type massquelle map[string]image.Point

func (q massquelle) Bild(pfad string) (image.Image, error) {
	groesse, da := q[pfad]
	if !da {
		return nil, errors.New("das Bild gibt es nicht")
	}

	return image.NewNRGBA(image.Rect(0, 0, groesse.X, groesse.Y)), nil
}

func TestEineMehrfachwahlAlsBildreihe(t *testing.T) {
	quelle := massquelle{
		"/probe.mse-style/flying.png": {X: 20, Y: 16},
		"/probe.mse-style/haste.png":  {X: 18, Y: 18},
	}

	faelle := []struct {
		name, wert, ausrichtung, richtung string
		erwartet                          []render.Rechteck
		befunde                           int
	}{
		{
			"nebeneinander, auf halber Höhe", "haste, flying", "middle left", "",
			[]render.Rechteck{{Links: 42, Oben: 333, Breite: 20, Hoehe: 16}, {Links: 72, Oben: 333, Breite: 18, Hoehe: 18}},
			0,
		},
		{
			"ohne Ausrichtung oben links", "flying", "", "",
			[]render.Rechteck{{Links: 42, Oben: 321, Breite: 20, Hoehe: 16}},
			0,
		},
		{
			"untereinander", "flying, haste", "", "top to bottom",
			[]render.Rechteck{{Links: 42, Oben: 321, Breite: 20, Hoehe: 16}, {Links: 42, Oben: 347, Breite: 18, Hoehe: 18}},
			0,
		},
		{
			"eine Wahl ohne Bild rückt um den Abstand", "deathtouch, flying", "", "",
			[]render.Rechteck{{Links: 52, Oben: 321, Breite: 20, Hoehe: 16}},
			0,
		},
		{
			"ein Bild, das fehlt, wird benannt", "lifelink, flying", "", "", // flying kommt in der Erklärung zuerst
			[]render.Rechteck{{Links: 42, Oben: 321, Breite: 20, Hoehe: 16}},
			1,
		},
		{"leer", "", "", "", nil, 0},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
				Felder: []msepaket.Feldstil{{
					Feld: "keyword icons", Links: "42", Oben: "321", Breite: "297", Hoehe: "24",
					Darstellung: "image list", Ausrichtung: fall.ausrichtung, Richtung: fall.richtung, Abstand: "10",
					Wahlbilder: []msepaket.Wahlbild{
						{Wahl: "flying", Bild: "flying.png"},
						{Wahl: "haste", Bild: "haste.png"},
						{Wahl: "lifelink", Bild: "lifelink.png"},
					},
				}},
			}

			karte := &domain.CardRow{ID: 1, Number: 1}
			if fall.wert != "" {
				karte.Values = []domain.Value{{Field: "keyword icons", Text: fall.wert}}
			}

			plan, befunde, err := service.NeuerPlanbauer().MitBildern(quelle).Stilplan(t.Context(), karte, service.Vorlage{
				Stil: stil,
				Kartenfelder: []domain.Field{{
					Name: "keyword icons", Kind: domain.KindMultipleChoice,
					Choices: []string{"deathtouch", "flying", "haste", "lifelink"},
				}},
			})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(befunde) != fall.befunde {
				t.Errorf("Befunde: %+v, erwartet %d", befunde, fall.befunde)
			}

			var gezeichnet []render.Rechteck

			for _, ebene := range plan.Ebenen {
				if ebene.Art == render.ArtBild && ebene.Feld == "keyword icons" {
					gezeichnet = append(gezeichnet, ebene.Rahmen)
				}
			}

			if len(gezeichnet) != len(fall.erwartet) {
				t.Fatalf("%d Bilder %+v, erwartet %d %+v", len(gezeichnet), gezeichnet, len(fall.erwartet), fall.erwartet)
			}

			for i := range gezeichnet {
				if gezeichnet[i] != fall.erwartet[i] {
					t.Errorf("Bild %d steht bei %+v, erwartet %+v", i, gezeichnet[i], fall.erwartet[i])
				}
			}
		})
	}
}
