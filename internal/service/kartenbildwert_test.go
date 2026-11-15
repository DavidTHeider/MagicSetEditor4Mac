package service_test

// **`card.<bildfeld>` ist ein Bild aus dem Set-Paket**
// (#440).
//
//	void ImageValue::reflect(GetDefaultMember& handler) {
//	  // convert to ScriptImageP for scripting
//	  handler.handle( (ScriptValueP)make_intrusive<ImageValueToImage>(filename, last_update) );
//	}
//
// `src/data/field/image.cpp`, Zeilen 49–53, Fassung f9d9356 vom 2021-07-06.
// `ImageValueToImage::generate` öffnet die Datei aus `opt.local_package`
// (`src/gfx/generated_image.cpp`, Zeilen 489–496) — dem Set.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinKartenbildKommtAusDemSetpaket(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "spiegel", Links: "0", Oben: "0", Breite: "375", Hoehe: "523",
			Bild: "{ card.image }",
		}},
	}

	vorlage := service.Vorlage{
		Stil:     stil,
		Setpaket: "/wo/auch/immer/Probe.mse-set",
		Kartenfelder: []domain.Field{
			{Name: "image", Kind: domain.KindImage},
			{Name: "spiegel", Kind: domain.KindText},
		},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "image", Text: "image16"},
		}}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %+v", befunde)
	}

	for i := range plan.Ebenen {
		blaetter := plan.Ebenen[i].Bild.Blaetter()
		if len(blaetter) == 0 {
			continue
		}

		ort, datei, ist := msepaket.ZerlegeSetpaketverweis(blaetter[0])
		if !ist {
			t.Errorf("das Blatt %q ist kein Setpaketverweis", blaetter[0])

			return
		}

		if ort != "/wo/auch/immer/Probe.mse-set" || datei != "image16" {
			t.Errorf("Verweis auf %q in %q", datei, ort)
		}

		return
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Ohne Bildfeld bleibt alles Text** — und ohne Ort des Set-Pakets auch.
//
// Zwei Gegenproben zu [TestEinKartenbildKommtAusDemSetpaket]: die Wandlung
// greift nur, wo ein Feld der Art `image` steht **und** der Ort bekannt ist.
func TestOhneBildfeldOderOrtBleibtEsText(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "spiegel", Links: "0", Oben: "0", Breite: "375", Hoehe: "523",
			Bild: "{ card.image }",
		}},
	}

	for _, fall := range []struct {
		name     string
		vorlage  service.Vorlage
		erwartet string
	}{
		{
			name: "kein Feld der Art image",
			vorlage: service.Vorlage{
				Stil: stil, Setpaket: "/wo/Probe.mse-set",
				Kartenfelder: []domain.Field{{Name: "image", Kind: domain.KindText}},
			},
			erwartet: "/probe.mse-style/image16",
		},
		{
			name: "kein Ort des Set-Pakets",
			vorlage: service.Vorlage{
				Stil:         stil,
				Kartenfelder: []domain.Field{{Name: "image", Kind: domain.KindImage}},
			},
			erwartet: "/probe.mse-style/image16",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			plan, _, err := service.NeuerPlanbauer().
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "image", Text: "image16"},
				}}, fall.vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for i := range plan.Ebenen {
				if blaetter := plan.Ebenen[i].Bild.Blaetter(); len(blaetter) > 0 {
					if blaetter[0] != fall.erwartet {
						t.Errorf("Blatt %q, erwartet %q", blaetter[0], fall.erwartet)
					}

					return
				}
			}

			t.Fatal("der Plan führt keine Bildebene")
		})
	}
}
