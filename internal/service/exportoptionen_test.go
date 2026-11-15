package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Option mit `type: boolean` kommt als Wahrheitswert an.**
//
// # Die Prämisse des Issues war falsch
//
// #202 fragte,
// was MSE aus einer **Zeichenkette** als Wahrheitswert macht — `"yes"`? nicht
// leer? — und nannte es „nicht erschlossen".
//
// **Ausgezählt stellt sich die Frage nicht.** Die neunzehn Exportvorlagen von
// `M15-Magic-Pack` führen 36 Optionen mit `type: boolean`; `if options.text_costs
// then …` liest also keine Zeichenkette, sondern eine Option, die ihren Typ
// selbst nennt. Zu deuten ist nur ihre **Vorgabe**, und die ist in allen 29
// Fällen mit `initial:` entweder `yes` (11) oder `no` (18).
func TestEineBoolescheOptionKommtAlsWahrheitswertAn(t *testing.T) {
	vorlage := msepaket.Exportvorlage{
		Skript: `if options.mit_bild then "ja" else "nein"`,
		Optionen: []msepaket.Exportoption{
			{Art: "boolean", Name: "mit bild", Vorgabe: "yes"},
		},
	}

	ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, nil, nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v — %v", err, befunde)
	}

	if ausgabe != "ja" {
		t.Errorf("%q, erwartet ja", ausgabe)
	}
}

// **`no` ist falsch.**
func TestEineBoolescheOptionMitNo(t *testing.T) {
	vorlage := msepaket.Exportvorlage{
		Skript: `if options.mit_bild then "ja" else "nein"`,
		Optionen: []msepaket.Exportoption{
			{Art: "boolean", Name: "mit bild", Vorgabe: "no"},
		},
	}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, nil, nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if ausgabe != "nein" {
		t.Errorf("%q, erwartet nein", ausgabe)
	}
}

// **Die Wahl schlägt die Vorgabe, und auch sie wird gedeutet.**
func TestDieWahlEinerBoolschenOptionWirdGedeutet(t *testing.T) {
	vorlage := msepaket.Exportvorlage{
		Skript: `if options.mit_bild then "ja" else "nein"`,
		Optionen: []msepaket.Exportoption{
			{Art: "boolean", Name: "mit bild", Vorgabe: "no"},
		},
	}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, nil, nil, map[string]string{"mit bild": "yes"})
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if ausgabe != "ja" {
		t.Errorf("%q, erwartet ja", ausgabe)
	}
}

// **Eine Option ohne `type: boolean` bleibt Text.**
//
// `options.set_code` wird verkettet, nicht abgefragt; daraus einen
// Wahrheitswert zu machen ergäbe an jeder Stelle Unsinn.
func TestEineTextoptionBleibtText(t *testing.T) {
	vorlage := msepaket.Exportvorlage{
		Skript:   `options.kuerzel + "!"`,
		Optionen: []msepaket.Exportoption{{Art: "text", Name: "kuerzel", Vorgabe: "PRB"}},
	}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{Verzeichnis: "ausfuhr-files"}).
		Fuehre(t.Context(), vorlage, nil, nil, nil)
	if err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	if ausgabe != "PRB!" {
		t.Errorf("%q", ausgabe)
	}
}
