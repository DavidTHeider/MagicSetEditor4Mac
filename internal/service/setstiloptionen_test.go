package service_test

// **Ein Set bringt setweite Stiloptionen mit** —
// #557.
//
//	IndexMap<FieldP, ValueP>& Set::stylingDataFor(const StyleSheet& stylesheet) {
//	  return styling_data.get(stylesheet.name(), stylesheet.styling_fields);
//	}
//	IndexMap<FieldP, ValueP>& Set::stylingDataFor(const CardP& card) {
//	  if (card && card->has_styling) return card->styling_data;
//	  else                           return stylingDataFor(stylesheetFor(card));
//	}
//
// `src/data/set.cpp`, Zeilen 99–105, Fassung f9d9356 vom 2021-07-06. **Zwei
// Zweige**: der erste stand seit
// #547, der
// zweite fehlte.
//
// # Warum das zählt, obwohl der Kanon ihn nie erreicht
//
// Alle 25.148 kanonischen Karten tragen `has_styling: true`; über sie ist der
// Zweig unerreichbar. **Die dritte Quelle schreibt aber genau ihn** (§11):
// `magic-set-generator` legt die Stiloptionen setweit ab, geschlüsselt nach
// dem **präfixierten** Paketnamen (`src/mse.rs`, Zeilen 151–174).
//
// **Und jedes kanonische Set führt den Block ebenfalls** — nachgesehen am
// 2026-08-20: alle zehn tragen `styling:` mit einer Ebene je Stilpaket.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieSetweitenStiloptionenGelten(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		karte    []domain.Value
		erwartet float64
	}{
		// **Ohne eigene Wahl gilt die des Sets** — der fehlende Zweig.
		"eine Karte ohne eigene Stiloptionen": {
			karte:    []domain.Value{{Field: "melder", Text: "x"}},
			erwartet: 10,
		},

		// **Mit `has styling` gilt die der Karte** — der Zweig von #547, und
		// die Reihenfolge ist damit geprüft und nicht behauptet.
		"eine Karte mit eigenen": {
			karte: []domain.Value{
				{Field: "melder", Text: "x"},
				{Field: "has styling", Text: "true"},
				{Field: "styling data", Text: "probe: karte"},
			},
			erwartet: 20,
		},

		// **Und ohne beides die Vorgabe des Stils.**
		"weder noch": {
			karte:    []domain.Value{{Field: "melder", Text: "x"}},
			erwartet: 30,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			setangaben := []domain.Value{{
				Field: "styling",
				// **Zeilen, die keine Option sind, werden übergangen** — eine
				// ohne Doppelpunkt und eine ohne Namen. Sie entstehen nicht aus
				// dem Import, aber der Bestand ist keine geschlossene Welt: der
				// MCP-Server schreibt auf dieselbe Datei
				// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
				Text: "magic-probe/probe: setweit\n" +
					"magic-anders/probe: falsch\n" +
					"magic-probe/ohne Doppelpunkt\n" +
					"magic-probe/: ohne Namen\n" +
					"ohne Schrägstrich: egal\n",
			}}

			if fall.erwartet == 30 {
				setangaben = nil
			}

			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "magic-probe.mse-style",
				Stiloptionen: []domain.StyleField{
					{Name: "probe", Kind: domain.KindText, Initial: "vorgabe"},
				},
				// **Die Wahl steht in der Lage**, nicht im Text: `styling.probe`
				// ist eine Zeichenkette, und der Text einer Ebene kommt von der
				// Karte. Derselbe Weg wie in `layoutobjekt_test.go`.
				Felder: []msepaket.Feldstil{{
					Feld: "melder", Oben: "100", Breite: "100", Hoehe: "20",
					Links: `{ if styling.probe == "setweit" then 10 ` +
						`else if styling.probe == "karte" then 20 else 30 }`,
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				}},
			}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: fall.karte}

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{
					Stil: stil, Setangaben: setangaben,
				})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if gab := melderlage(plan); gab != fall.erwartet {
				t.Errorf("`styling.probe` ergibt %g, erwartet %g", gab, fall.erwartet)
			}
		})
	}
}

// melderlage gibt die Linkskante der Ebene, die auf `styling.probe` steht.
func melderlage(plan *render.Plan) float64 {
	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art == render.ArtText && plan.Ebenen[i].Rahmen.Oben == 100 {
			return plan.Ebenen[i].Rahmen.Links
		}
	}

	return -1
}
