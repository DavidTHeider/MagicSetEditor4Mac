package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Wahlbild wird eingepasst — außer die Vorlage sagt etwas anderes.**
//
//	opts.width  = (int) rot.trX(style.width);
//	opts.height = (int) rot.trY(style.height);
//	opts.preserve_aspect = (style.alignment & ALIGN_STRETCH) ? ASPECT_STRETCH : ASPECT_FIT;
//
// `get_options`, `src/render/value/choice.cpp`, Zeilen 111–124, Fassung f9d9356
// vom 2021-07-06. Die Maße werden **nur** in diesem Zweig gesetzt: zeichnet das
// Feld auch Text (`render style: both`), gilt die Originalgröße.
//
// **Daran hängt das Seltenheitssymbol.** Sein Kasten ist breiter als hoch;
// gestreckt wurde daraus ein Balken quer über die Typzeile
// (#270).
func TestEinWahlbildWirdEingepasst(t *testing.T) {
	faelle := []struct {
		name         string
		darstellung  string
		ausrichtung  string
		art          domain.FieldKind
		eingepasst   bool
		ausgerichtet render.Ausrichtung
	}{
		{
			name: "ein Auswahlfeld als Bild", darstellung: "image",
			ausrichtung: "middle right", art: domain.KindChoice,
			eingepasst: true, ausgerichtet: render.RechtsMitte,
		},
		{
			// **`both` zeichnet Bild und Text**, und dann setzt MSE die Maße
			// gar nicht — die Einpassung in den Kasten entfällt.
			name: "Bild und Text", darstellung: "both",
			ausrichtung: "middle right", art: domain.KindChoice,
		},
		{
			name: "ausdrücklich gestreckt", darstellung: "image",
			ausrichtung: "middle right stretch", art: domain.KindChoice,
		},
		{
			// **Ein Bildfeld wird gestreckt** — MSEs `image.Rescale(w, h)`
			// (`src/render/value/image.cpp`, Zeile 38). Das Kunstfenster füllt
			// seinen Kasten.
			//
			// Fassung f9d9356 vom 2021-07-06.
			name: "ein Bildfeld", darstellung: "image",
			ausrichtung: "middle right", art: domain.KindImage,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
				Felder: []msepaket.Feldstil{{
					Feld: "rarity", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
					Darstellung: fall.darstellung, Ausrichtung: fall.ausrichtung,
					Wahlbilder: []msepaket.Wahlbild{
						{Wahl: "rare", Bild: "/probe.mse-style/selten.png"},
					},
				}},
			}

			karte := &domain.CardRow{ID: 1, Values: []domain.Value{
				{Field: "rarity", Text: "rare"},
			}}

			vorlage := service.Vorlage{
				Stil:         stil,
				Kartenfelder: []domain.Field{{Name: "rarity", Kind: fall.art}},
			}

			plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			ebene := bildebeneAus(t, plan)

			if (ebene.Passung != nil) != fall.eingepasst {
				t.Fatalf("die Passung ist %+v, eingepasst erwartet: %v",
					ebene.Passung, fall.eingepasst)
			}

			if fall.eingepasst && ebene.Passung.Ausrichtung != fall.ausgerichtet {
				t.Errorf("ausgerichtet wird %q", ebene.Passung.Ausrichtung)
			}
		})
	}
}

func bildebeneAus(t *testing.T, plan *render.Plan) render.Ebene {
	t.Helper()

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art == render.ArtBild {
			return plan.Ebenen[i]
		}
	}

	t.Fatal("der Plan führt keine Bildebene")

	return render.Ebene{}
}

// **Ohne Setsymbol zeichnet `symbol_variation` MSEs Standardsymbol.**
//
//	if (filename.empty()) {
//	  the_symbol = default_symbol();
//	}
//
// `SymbolToImage::generate`, `src/gfx/generated_image.cpp`, Zeilen 457–459,
// Fassung f9d9356 vom 2021-07-06. Ein Set ohne eigenes Symbol ist der
// Normalfall — die vier Referenzkarten sind solche —, und MSE malt an der
// Typzeile ein Quadrat, wo dieses Programm „die Symbolzeichnung liegt nicht
// vor" meldete
// (#270).
func TestOhneSetsymbolGiltDasStandardsymbol(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Symbolfassungen: []msepaket.Symbolfassung{{Feld: "symbol", Name: "rare", Fuellart: "solid"}},
		Felder: []msepaket.Feldstil{{
			Feld: "rarity", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
			Darstellung: "image",
			Bild:        `{ symbol_variation(symbol: set.symbol, variation: "rare") }`,
		}},
	}

	vorlage := service.Vorlage{
		Stil:      stil,
		Setfelder: []domain.Field{{Name: "symbol", Kind: domain.KindImage}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	ebene := bildebeneAus(t, plan)
	if len(ebene.Bild.Symbol) != 1 || len(ebene.Bild.Symbol[0].Punkte) != 4 {
		t.Errorf("die Symbolgestalt ist %+v", ebene.Bild.Symbol)
	}
}
