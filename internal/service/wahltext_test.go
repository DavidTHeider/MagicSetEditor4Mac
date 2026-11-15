package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein gezeichneter Wahlwert beginnt groß** —
// #561.
//
// # Warum das an einem Locale-Issue hängt
//
// MSE schlägt einen Wahlwert im Sprachpaket nach, **bevor** es ihn zeichnet:
//
//	String text = tr(viewer.getStylePackage(), value, capitalize_sentence);
//
// `draw_choice_viewer`, `src/render/value/choice.cpp`, Zeile 99, und
// wortgleich `MultipleChoiceValueViewer::draw`,
// `src/render/value/multiple_choice.cpp`, Zeile 74. Beides Fassung f9d9356
// vom 2021-07-06.
//
// Das dritte Argument ist der **Rückfall**: was gilt, wenn das Sprachpaket
// nichts sagt.
//
//	void capitalize_sentence_in_place(String& s) {
//	  if (!s.empty()) s[0] = toUpper(s[0]);
//	}
//
// `src/util/string.cpp`, Zeilen 115–119.
//
// # Und der Rückfall ist der Normalfall
//
// Nachgemessen am 2026-08-20 über `full-magic-pack`: **alle dreizehn**
// Sprachpakete führen unter `package:` genau zwei Blöcke — `magic.mse-game`
// und `magic-mana-*.mse-symbol-font`. Der Aufruf oben schlägt aber unter dem
// **Stilpaket** nach (`magic-m15.mse-style`), und darauf passt keiner der
// beiden — auch nicht über MSEs Platzhalterregel (`find_wildcard`,
// `src/data/locale.cpp`, Zeilen 57–66). Es gibt im ganzen Pack keine
// Übersetzung, die dieser Weg fände.
//
// **Die Reichweite der Regel selbst**: über 376 Stilpakete zeichnen von 1.359
// Auswahlfeldstilen **fünf** Text — zwei mit ausdrücklichem `render style:
// text`, drei ohne Angabe (`ChoiceStyle::render_style` steht in der Vorgabe auf
// `RENDER_TEXT`, `src/data/field/choice.cpp`, Zeile 164). Im kanonischen Stil
// sind es **null**; der Abstand ändert sich dadurch nicht.
func TestEinGezeichneterWahlwertBeginntGross(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		art      domain.FieldKind
		wert     string
		erwartet string
	}{
		"eine Auswahl":                   {domain.KindChoice, "common", "Common"},
		"eine Auswahl, schon groß":       {domain.KindChoice, "Common", "Common"},
		"eine leere Auswahl":             {domain.KindChoice, "", ""},
		"ein Textfeld bleibt, wie es is": {domain.KindText, "fliegend", "fliegend"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{{
					Feld: "rarity", Links: "10", Oben: "20", Breite: "300", Hoehe: "40",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				}},
			}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "rarity", Text: fall.wert},
			}}

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{
					Stil:         stil,
					Kartenfelder: []domain.Field{{Name: "rarity", Kind: fall.art}},
				})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if gab := ersterText(plan); gab != fall.erwartet {
				t.Errorf("gezeichnet wird %q, erwartet %q", gab, fall.erwartet)
			}
		})
	}
}

// ersterText gibt den Text der ersten Textebene eines Plans.
func ersterText(plan *render.Plan) string {
	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art != render.ArtText {
			continue
		}

		text := ""
		for _, abschnitt := range plan.Ebenen[i].Abschnitte {
			text += abschnitt.Text
		}

		return text
	}

	return ""
}
