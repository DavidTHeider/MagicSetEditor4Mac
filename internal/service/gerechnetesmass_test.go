package service_test

// **Die Messstufe rechnet ihre Angaben aus** — sie liest sie nicht roh
// (#420).
//
// Getrennt von `textbreite_test.go`, das die Messung selbst hält: hier steht
// nicht, *ob* gemessen wird, sondern *woraus* die Messung ihre Angaben nimmt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zugbreite gibt die Breite der Ebene, die den gesuchten Klartext trägt.
func zugbreite(t *testing.T, plan *render.Plan, text string) float64 {
	t.Helper()

	for i := range plan.Ebenen {
		if ebene := &plan.Ebenen[i]; ebene.Art == render.ArtText && ebene.Klartext() == text {
			return ebene.Rahmen.Breite
		}
	}

	t.Fatalf("der Plan führt keine Textebene %q", text)

	return 0
}

// **Ein gerechneter Schriftname und eine gerechnete Größe werden gemessen.**
//
// MSE wertet beide vor dem Satz aus: `Font::update`
// (`src/data/font.cpp`, Zeilen 27–36, Fassung f9d9356 vom 2021-07-06), gerufen
// über `TextStyle::update` und `SetScriptManager::updateStyles`. Über die 190
// Stile von `M15-Magic-Pack` sind **664 von 3.119** `font: name:` und **2.528
// von 3.119** `font: size:` Skripte — gezählt am Feldstil, nicht an
// gezeichneten Ebenen (§10).
func TestGerechneteSchriftangabenWerdenGemessen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
				Schrift: msepaket.Schrift{
					Name:    `{ if true then "Probe" else "Fehlt" }`,
					Groesse: `{ if true then 10 else 20 }`,
				},
			},
			{
				Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
				Rechts:  "{ card_style.name.content_width }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "name", Text: "abcd"},
			{Field: "type", Text: "x"},
		}}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// `abcd` à 0,5 bei Größe 10 ergibt 20; die Typzeile reicht von 10 bis 20.
	// Ohne die Auswertung bleibt `content_width` bei der deklarierten 100, und
	// die Typzeile wird 90 breit.
	if breite := zugbreite(t, plan, "x"); breite != 10 {
		t.Errorf("die Typzeile ist %.0f breit, erwartet 10", breite)
	}
}

// **Was sich nicht ausrechnen lässt, wird nicht gemessen.**
//
// Der Ersatzwert ist die deklarierte Breite — dieselbe Antwort wie bei einer
// fehlenden Schrift. Eine geratene Zahl wäre schlechter: sie stünde in
// `content_width`, und niemand sähe, dass sie erfunden ist.
func TestOhneAusrechenbareAngabeWirdNichtGemessen(t *testing.T) {
	// `gibtsnicht` ist kein Name der Umgebung; die Auswertung ergibt einen
	// Fehlerwert, und der ist weder ein Schriftname noch eine Zahl.
	for _, fall := range []struct {
		name string
		feld msepaket.Feldstil
	}{
		{
			name: "der Schriftname ergibt nichts",
			feld: msepaket.Feldstil{
				Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
				Schrift: msepaket.Schrift{Name: "{ if false then \"Probe\" }", Groesse: "10"},
			},
		},
		{
			name: "die Schriftgröße ergibt keine Zahl",
			feld: msepaket.Feldstil{
				Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "{ gibtsnicht }"},
			},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{fall.feld, {
					Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
					Rechts:  "{ card_style.name.content_width }",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				}},
			}

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "name", Text: "abcd"},
					{Field: "type", Text: "x"},
				}}, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			// Es bleibt bei der deklarierten Breite 100: 10 bis 100 sind 90.
			if breite := zugbreite(t, plan, "x"); breite != 90 {
				t.Errorf("die Typzeile ist %.0f breit, erwartet 90", breite)
			}
		})
	}
}

// **Und dasselbe für den Symbolzug.**
//
// Ergibt weder die `size:` des Symbolschriftblocks noch die Höhe des Feldes
// eine Zahl, so gibt es keine Kantenlänge — und ohne die keine Breite.
func TestOhneAusrechenbareSymbolgroesseWirdNichtGemessen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "casting cost", Links: "300", Oben: "29", Breite: "40",
			Hoehe:         "{ gibtsnicht }",
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "{ gibtsnicht }"},
			ImmerSymbole:  true,
		}, {
			Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
			Rechts:  "{ card_style.casting_cost.content_width }",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "casting cost", Text: "WU"},
			{Field: "type", Text: "x"},
		}}, service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
			"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{
				{Code: "W", Bild: "w.png"}, {Code: "U", Bild: "u.png"},
			}},
		}})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	// Es bleibt bei der deklarierten Breite 40: 10 bis 40 sind 30.
	if breite := zugbreite(t, plan, "x"); breite != 30 {
		t.Errorf("die Typzeile ist %.0f breit, erwartet 30", breite)
	}
}
