package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Feldstil nennt auch seine Textangaben — Polster und Zeilenhöhen.**
//
//	REFLECT(padding_left);  REFLECT(padding_right);
//	REFLECT(padding_top);   REFLECT(padding_bottom);
//	REFLECT(line_height_soft); REFLECT(line_height_hard);
//	REFLECT(line_height_line);
//	REFLECT(line_height_soft_max); … REFLECT(paragraph_height);
//
// `IMPLEMENT_REFLECTION(TextStyle)`, `src/data/field/text.cpp`, Zeilen 143–167,
// Fassung f9d9356 vom 2021-07-06.
//
// **Eine Vorlage rechnet mit den Angaben einer anderen Ebene**:
// `magic-genevensis` liest `card_style.text.line_height_hard`, um seine Stufen
// gleich hoch zu setzen. 14 Befunde über die 190 Stile von `M15-Magic-Pack`
// (#270).
func TestEinFeldstilNenntSeineTextangaben(t *testing.T) {
	gelesen := func(ausdruck string) string {
		t.Helper()

		vorlage := service.Vorlage{
			Stil: msepaket.Stil{
				Spiel: "probe", Breite: 100, Hoehe: 100,
				Felder: []msepaket.Feldstil{
					{
						Feld: "text", Links: "0", Oben: "0", Breite: "50", Hoehe: "20",
						Polster: msepaket.Polsterangabe{
							Links: "3", Rechts: "4", Oben: "5", Unten: "6",
						},
						Zeilenhoehen: msepaket.Zeilenhoehenangabe{
							Weich: "1.1", Hart: "2.2", Linie: "3.3",
							WeichMax: "4.4", HartMax: "5.5", LinieMax: "6.6",
							Absatz: "7.7",
						},
						Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
					},
					{
						Feld: "name", Links: "0", Oben: "50", Breite: "50", Hoehe: "20",
						Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
					},
				},
			},
			Kartenfelder: []domain.Field{
				{Name: "text", Kind: domain.KindText},
				{Name: "name", Kind: domain.KindText, Skript: "to_string(" + ausdruck + ")"},
			},
		}

		karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "text", Text: "x"}}}

		plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		for _, befund := range befunde {
			if strings.Contains(befund.Ort, "name") {
				t.Fatalf("Befund: %s", befund.Meldung)
			}
		}

		for _, ebene := range plan.Ebenen {
			if ebene.Klartext() != "" && ebene.Klartext() != "x" {
				return ebene.Klartext()
			}
		}

		return ""
	}

	faelle := map[string]string{
		"card_style.text.padding_left":         "3",
		"card_style.text.padding_right":        "4",
		"card_style.text.padding_top":          "5",
		"card_style.text.padding_bottom":       "6",
		"card_style.text.line_height_soft":     "1.1",
		"card_style.text.line_height_hard":     "2.2",
		"card_style.text.line_height_line":     "3.3",
		"card_style.text.line_height_soft_max": "4.4",
		"card_style.text.line_height_hard_max": "5.5",
		"card_style.text.line_height_line_max": "6.6",
		"card_style.text.paragraph_height":     "7.7",
	}

	for ausdruck, erwartet := range faelle {
		t.Run(ausdruck, func(t *testing.T) {
			if haben := gelesen(ausdruck); haben != erwartet {
				t.Errorf("ergibt %q, erwartet %q", haben, erwartet)
			}
		})
	}
}
