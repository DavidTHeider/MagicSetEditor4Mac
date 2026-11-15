package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Exportoption ohne Vorgabe nimmt die erste Wahl.**
//
// # Die Regel steht am Feld, nicht am Export
//
//	ChoiceValue::ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice)
//	  : Value(field)
//	  , value( !field->initial.empty() ? field->initial
//	         : initial_first_choice    ? field->choices->choiceName(0)
//	         :                           _("")
//	         , true)
//	{}
//
//	ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice = true);
//
// `src/data/field/choice.cpp`, Zeilen 279–285, und `choice.hpp`, Zeile 192.
// Fassung f9d9356 vom 2021-07-06.
//
// **Und eine Exportvorlage führt genau solche Felder**: `option_fields` ist
// ein `vector<FieldP>` (`src/data/export_template.hpp`, Zeile 32) — dieselbe
// Klasse wie ein Kartenfeld. Was für `spiel.go` und `stil.go` schon gilt
// (#456,
// #531), gilt
// hier ebenso.
//
// # Was das Pack dazu sagt
//
// Über seine Exportvorlagen: **7** boolesche Optionen ohne `default:`/`initial:`
// und **21** `choice`-Optionen ohne. Sie standen alle leer da, wo MSE `yes`
// beziehungsweise die erste Wahl liest.
//
// # Und eine Ausnahme, die hier nicht auftritt
//
// `MultipleChoiceValue` reicht **`false`** durch (`multiple_choice.hpp`,
// Zeile 52): eine Mehrfachwahl ohne `initial:` beginnt **leer**. Eine
// Exportoption dieser Art gibt es im Pack nicht; wer sie hinzunimmt, prüft das
// nach.
func TestEineExportoptionOhneVorgabeNimmtDieErsteWahl(t *testing.T) {
	faelle := []struct {
		name     string
		option   msepaket.Exportoption
		skript   string
		erwartet string
	}{
		{
			name:     "boolean ohne Vorgabe ist yes",
			option:   msepaket.Exportoption{Art: "boolean", Name: "mit bild"},
			skript:   `if options.mit_bild then "ja" else "nein"`,
			erwartet: "ja",
		},
		{
			name: "eine Auswahl ohne Vorgabe nimmt die erste",
			option: msepaket.Exportoption{
				Art: "choice", Name: "format", Auswahl: []string{"kurz", "lang"},
			},
			skript:   `options.format`,
			erwartet: "kurz",
		},
		{
			name: "eine gesetzte Vorgabe schlägt die erste Wahl",
			option: msepaket.Exportoption{
				Art: "choice", Name: "format", Vorgabe: "lang",
				Auswahl: []string{"kurz", "lang"},
			},
			skript:   `options.format`,
			erwartet: "lang",
		},
		{
			name:     "boolean mit no bleibt no",
			option:   msepaket.Exportoption{Art: "boolean", Name: "mit bild", Vorgabe: "no"},
			skript:   `if options.mit_bild then "ja" else "nein"`,
			erwartet: "nein",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{}).Fuehre(
				t.Context(),
				msepaket.Exportvorlage{
					Skript:   fall.skript,
					Optionen: []msepaket.Exportoption{fall.option},
				}, nil, nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v — %v", err, befunde)
			}

			if ausgabe != fall.erwartet {
				t.Errorf("%q, erwartet %q", ausgabe, fall.erwartet)
			}
		})
	}
}
