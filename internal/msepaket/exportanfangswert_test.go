package msepaket_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Was für eine Exportoption gilt, wenn niemand wählt.**
//
//	ChoiceValue::ChoiceValue(const ChoiceFieldP& field, bool initial_first_choice)
//	  : Value(field)
//	  , value( !field->initial.empty() ? field->initial
//	         : initial_first_choice    ? field->choices->choiceName(0)
//	         :                           _("")
//	         , true)
//
// `src/data/field/choice.cpp`, Zeilen 279–285, mit `initial_first_choice`
// vorbelegt auf `true` (`choice.hpp`, Zeile 192). Eine Exportvorlage führt in
// `option_fields` dieselbe Feldklasse (`export_template.hpp`, Zeile 32).
// Fassung f9d9356 vom 2021-07-06.
func TestDerAnfangswertEinerExportoption(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		option   msepaket.Exportoption
		erwartet string
	}{
		{
			name:     "eine gesetzte Vorgabe gilt",
			option:   msepaket.Exportoption{Art: "choice", Vorgabe: "lang", Auswahl: []string{"kurz", "lang"}},
			erwartet: "lang",
		},
		{
			name:     "ohne Vorgabe die erste Wahl",
			option:   msepaket.Exportoption{Art: "choice", Auswahl: []string{"kurz", "lang"}},
			erwartet: "kurz",
		},
		{
			// **Auch ohne Wahlliste.** [msepaket.Exportoption] wird von
			// `optionAus` mit `yes`/`no` gebaut, aber die Zusage gehört
			// dorthin, wo sie eingelöst wird (§10).
			name:     "boolean ohne alles ist yes",
			option:   msepaket.Exportoption{Art: "boolean"},
			erwartet: "yes",
		},
		{
			name:     "ein Text ohne Vorgabe bleibt leer",
			option:   msepaket.Exportoption{Art: "text"},
			erwartet: "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if ist := fall.option.Anfangswert(); ist != fall.erwartet {
				t.Errorf("Anfangswert %q, erwartet %q", ist, fall.erwartet)
			}
		})
	}
}

// **Und der Leser legt einer booleschen Option ihre zwei Wahlen an.**
//
//	BooleanField::BooleanField() {
//	  choices->choices.push_back(make_intrusive<Choice>(_("yes")));
//	  choices->choices.push_back(make_intrusive<Choice>(_("no")));
//
// `src/data/field/boolean.cpp`, Zeilen 14–18, Fassung f9d9356 vom 2021-07-06.
// Über die Exportvorlagen des Packs: **7** boolesche Optionen ohne
// `default:`/`initial:` und **21** `choice`-Optionen ohne.
func TestEineBoolescheExportoptionHatZweiWahlen(t *testing.T) {
	t.Parallel()

	pfad := probeexportpaket(t, `mse version: 2.0.0
option field:
	type: boolean
	name: mit bild
option field:
	type: choice
	name: format
	choice: kurz
	choice:
		name: lang
script: ""
`)

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	vorlage := paket.Exportmodell()
	if len(vorlage.Optionen) != 2 {
		t.Fatalf("%d Optionen, erwartet zwei", len(vorlage.Optionen))
	}

	if ist := vorlage.Optionen[0].Anfangswert(); ist != "yes" {
		t.Errorf("die boolesche Option beginnt mit %q, erwartet yes", ist)
	}

	// **Auch die Blockform trägt ihren Namen** — dieselbe Regel wie an einem
	// Kartenfeld ([msepaket.wahlname]).
	if ist := vorlage.Optionen[1].Auswahl; len(ist) != 2 || ist[1] != "lang" {
		t.Errorf("Wahlen %q, erwartet [kurz lang]", ist)
	}
}

// probeexportpaket legt eine `.mse-export-template` mit dem gegebenen Inhalt an.
func probeexportpaket(t *testing.T, inhalt string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.mse-export-template")
	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	datei := filepath.Join(pfad, "export-template")
	if err := os.WriteFile(datei, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return pfad
}
