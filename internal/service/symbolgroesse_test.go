package service_test

// **Die Symbolgröße darf gerechnet sein**
// (#444).
//
// `SymbolFontRef::size` ist bei MSE ein `Scriptable<double>` und wird in
// `TextStyle::update` ausgewertet (`src/data/field/text.cpp`, Zeile 86,
// Fassung f9d9356 vom 2021-07-06) — genau wie `font: size:`.
//
// `magic-m15-altered.mse-style` schreibt für den Regeltext
// `symbol font: size: {body_font_size()}`. Roh gelesen fiel die Größe auf die
// **Höhe des Feldes** zurück — 154 Punkte —, und die Manasymbole in
// `{4}{W}: …` deckten die halbe Karte zu.
//
// Über `Full-Magic-Pack` sind **272 von 937** `symbol font: size:`-Angaben
// Skripte; gezählt am Block in `style` und `game` (§10).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEineGerechneteSymbolgroesseGilt(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "29", Oben: "327", Breite: "314", Hoehe: "154",
			Schrift:       msepaket.Schrift{Name: "Probe", Groesse: "10"},
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "{ 6 + 6 }"},
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
		"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{
			{Code: "W", Bild: "w.png"},
		}},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "text", Text: "<sym>W</sym> und Text"},
		}}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for i := range plan.Ebenen {
		for _, abschnitt := range plan.Ebenen[i].Abschnitte {
			if len(abschnitt.Symbole) == 0 {
				continue
			}

			if abschnitt.Symbolmasse.Groesse != 12 {
				t.Errorf("Symbolgröße %.0f, erwartet 12 (nicht die Feldhöhe 154)",
					abschnitt.Symbolmasse.Groesse)
			}

			return
		}
	}

	t.Fatal("der Plan führt keinen Symbolabschnitt")
}

// **Ohne Angabe bleibt es die Höhe des Feldes** — die Gegenprobe.
//
// So misst MSE ein einzeiliges Symbolfeld, und die Manakosten von `magic-m15`
// sind genau das.
func TestOhneSymbolgroesseGiltDieFeldhoehe(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "casting cost", Links: "300", Oben: "29", Breite: "40", Hoehe: "23",
			Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole"},
			ImmerSymbole:  true,
		}},
	}

	plan, _, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "casting cost", Text: "W"},
		}}, service.Vorlage{Stil: stil, Symbolschriften: map[string]msepaket.Symbolschrift{
			"probe-symbole": {Name: "probe-symbole", Symbole: []msepaket.Symbol{
				{Code: "W", Bild: "w.png"},
			}},
		}})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if masse := symbolmasseAus(t, plan); masse.Groesse != 23 {
		t.Errorf("Symbolgröße %.0f, erwartet 23", masse.Groesse)
	}
}

// symbolmasseAus holt die Maße des einen Symbolzugs eines Plans.
func symbolmasseAus(t *testing.T, plan *render.Plan) render.Symbolmasse {
	t.Helper()

	for i := range plan.Ebenen {
		for _, abschnitt := range plan.Ebenen[i].Abschnitte {
			if len(abschnitt.Symbole) > 0 {
				return abschnitt.Symbolmasse
			}
		}
	}

	t.Fatal("der Plan führt keinen Symbolzug")

	return render.Symbolmasse{}
}
