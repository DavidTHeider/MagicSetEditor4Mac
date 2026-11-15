package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`visible:` entscheidet, ob eine Ebene überhaupt gezeichnet wird.**
//
//	bool ValueViewer::isVisible() const {
//	  return getStyle()->visible
//	    && bounding_box.width > 0
//	    && bounding_box.height > 0
//	    && fabs(bounding_box.x) < 100000
//	    && fabs(bounding_box.y) < 100000;
//	}
//
// `src/render/value/viewer.cpp`, Zeilen 36–42, Fassung f9d9356 vom
// 2021-07-06; gerufen in `CardViewer::draw`, Zeile 62.
//
// # Was das gekostet hat
//
// `magic-m15` schreibt am Kasten für Stärke und Widerstandskraft
// `visible: { card.pt != "" }`. Ohne diese Angabe steht auf **jeder
// Verzauberung** ein leerer weißer Kasten unten rechts — auf *Old-Money* zu
// sehen (#270).
//
// **Der Nullwert ist sichtbar**, so wie bei MSE (`visible(true)`,
// `src/data/field.cpp`, Zeile 106).
func TestVisibleEntscheidetObEineEbeneGezeichnetWird(t *testing.T) {
	faelle := []struct {
		name     string
		angabe   string
		erwartet int
	}{
		{name: "ohne Angabe wird gezeichnet", angabe: "", erwartet: 2},
		{name: "wahr wird gezeichnet", angabe: "{ true }", erwartet: 2},
		{name: "falsch wird weggelassen", angabe: "{ false }", erwartet: 1},

		// **Die Angabe rechnet mit der Karte** — das ist ihr Sinn.
		{name: "eine Bedingung an der Karte", angabe: `{ card.name != "" }`, erwartet: 2},
		{name: "und die andere Seite davon", angabe: `{ card.name == "" }`, erwartet: 1},

		// **Was sich nicht ausrechnen lässt, wird gezeichnet.** Eine Ebene
		// wegzulassen, weil ihre Bedingung scheiterte, versteckte den Mangel
		// — sichtbar falsch schlägt unsichtbar fehlend (§8).
		{name: "eine kaputte Bedingung zeichnet", angabe: "{ gibtsnicht() }", erwartet: 2},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			vorlage := service.Vorlage{
				Stil: msepaket.Stil{
					Spiel: "probe", Breite: 100, Hoehe: 100,
					Felder: []msepaket.Feldstil{{
						Feld: "name", Links: "0", Oben: "0", Breite: "50", Hoehe: "10",
						Sichtbar: fall.angabe,
						Schrift:  msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
					}},
				},
				Kartenfelder: []domain.Field{{Name: "name", Kind: domain.KindText}},
			}

			karte := &domain.CardRow{
				ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
			}

			plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(plan.Ebenen) != fall.erwartet {
				t.Errorf("%d Ebenen, erwartet %d", len(plan.Ebenen), fall.erwartet)
			}
		})
	}
}
