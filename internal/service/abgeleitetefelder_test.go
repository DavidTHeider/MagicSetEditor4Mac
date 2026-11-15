package service_test

// **`type` und `card color` werden abgeleitet**
// (#298).
//
// Nachgelesen in `magic.mse-game/card_fields`:
//
//	name: type
//	save value: false
//	script:
//		combined_editor(
//			field1: card.super_type,
//			separator: language().type_separator,
//			field2: card.sub_type,
//			…
//		)
//
// Zeilen 963–991. **`save value: false`** ist der stärkste Beleg: ein Feld,
// das nie gespeichert wird, muss gerechnet werden.
//
//	name: card color
//	default: stylesheet; card_color(casting_cost: card.casting_cost, rules_text: card.rule_text, type: …)
//	script: card_color_filter(value)
//
// Damit ist die zweite Deutung des Issues bestätigt und die erste widerlegt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerTypWirdAusOberUndUntertypGerechnet(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "type", Links: "10", Oben: "300", Breite: "300", Hoehe: "20",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	vorlage := service.Vorlage{
		Stil: stil,
		Kartenfelder: []domain.Field{
			{Name: "super type", Kind: domain.KindText},
			{Name: "sub type", Kind: domain.KindText},
			{
				Name: "type", Kind: domain.KindText,
				// Dieselbe Zusammensetzung wie im Pack, auf das Nötige gekürzt.
				Skript: `combined_editor(field1: card.super_type, separator: " — ", field2: card.sub_type)`,
			},
		},
	}

	// **Die Karte führt `type` nicht** — genau der Fall aus dem Issue.
	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "super type", Text: "Creature"},
			{Field: "sub type", Text: "Goblin Mage"},
		}}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %+v", befunde)
	}

	if breite := zugbreite(t, plan, "Creature — Goblin Mage"); breite == 0 {
		t.Error("der Plan führt keine Typzeile")
	}
}

// **Und `card color` über sein `default:`.**
//
// Der Unterschied zu `type` ist die Stufe, nicht die Sache: ein `script:`
// rechnet **immer**, ein `default:` nur, wenn die Karte nichts führt. MSE
// schreibt `card color` als `default:`, weil der Anwender die Farbe
// überschreiben können soll — `type` dagegen als `script:` mit
// `save value: false`.
func TestDieKartenfarbeKommtAusIhremVorgabeskript(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "card color", Links: "10", Oben: "300", Breite: "300", Hoehe: "20",
			Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	vorlage := service.Vorlage{
		Stil: stil,
		Kartenfelder: []domain.Field{
			{Name: "casting cost", Kind: domain.KindText},
			{
				Name: "card color", Kind: domain.KindText,
				Default: `{ if contains(card.casting_cost, match: "R") then "red" else "colorless" }`,
			},
		},
	}

	for _, fall := range []struct{ kosten, erwartet string }{
		{kosten: "1R", erwartet: "red"},
		{kosten: "2", erwartet: "colorless"},
	} {
		t.Run(fall.kosten, func(t *testing.T) {
			plan, befunde, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "casting cost", Text: fall.kosten},
				}}, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(befunde) > 0 {
				t.Errorf("Befunde: %+v", befunde)
			}

			if breite := zugbreite(t, plan, fall.erwartet); breite == 0 {
				t.Errorf("der Plan führt %q nicht", fall.erwartet)
			}
		})
	}
}
