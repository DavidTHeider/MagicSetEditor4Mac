package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Feldname in einer MSE-Datei ist kanonisch: `_` steht für ein
// Leerzeichen.**
//
//	void canonical_name_form_in_place(String& str) {
//	  for (String::iterator it = str.begin(); it != str.end(); ++it) {
//	    if (*it == ' ') *it = '_';
//	  }
//	}
//
// `src/util/string.cpp`, Zeilen 121–126, Fassung f9d9356 vom 2021-07-06. Der
// Kommentar am Kopf sagt es: „Convert a field name to canonical form".
//
// **Was das gekostet hat:** eine kanonische Karte trägt `rule_text:`, der
// Feldstil des Spiels heißt `rule text` — der Regeltext kam nie an, und weil
// ein Textfeld ohne Wert gar nicht gezeichnet wird, auch ohne Befund
// (#396).
func TestEinFeldnameWirdEntkanonisiert(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "kanonisch.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse_version: 2.1.2\ngame: probe\nstylesheet: probe-schlicht\n" +
		"set_info:\n\tset_code: PRB\ncard:\n\tname: Feuerball\n" +
		"\trule_text: Füge drei Schaden zu.\n\tcasting_cost: 2R\n"

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	speicher := &speicherDouble{}
	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), ort); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if len(speicher.set.Cards) != 1 {
		t.Fatalf("%d Karten, erwartet eine", len(speicher.set.Cards))
	}

	erwartet := map[string]string{
		"name":         "Feuerball",
		"rule text":    "Füge drei Schaden zu.",
		"casting cost": "2R",
	}

	gefunden := map[string]string{}
	for _, wert := range speicher.set.Cards[0].Values {
		gefunden[wert.Field] = wert.Text
	}

	for feld, text := range erwartet {
		if gefunden[feld] != text {
			t.Errorf("Feld %q trägt %q, erwartet %q — gefunden wurde: %v",
				feld, gefunden[feld], text, schluessel(gefunden))
		}
	}

	// **Auch die Setangaben.** `set_info:` trägt dieselben kanonischen Namen.
	var setcode string

	for _, wert := range speicher.set.Values {
		if wert.Field == "set code" {
			setcode = wert.Text
		}
	}

	if setcode != "PRB" {
		t.Errorf("das Setfeld `set code` trägt %q, erwartet PRB", setcode)
	}
}

// schluessel gibt die Feldnamen einer Abbildung, sortiert genug zum Lesen.
func schluessel(werte map[string]string) string {
	namen := make([]string, 0, len(werte))
	for name := range werte {
		namen = append(namen, name)
	}

	return strings.Join(namen, ", ")
}

// **Und geschrieben wird wieder kanonisch.**
//
// MSEs Schreiber wandelt jedes Leerzeichen in einen Unterstrich
// (`canonical_name_form`, `src/util/string.cpp`, Zeilen 121–126). Eine Datei,
// die wir schreiben, sieht damit aus wie eine von MSE — und ein Rundlauf
// verändert die Feldnamen nicht.
//
// Fassung f9d9356 vom 2021-07-06.
func TestEinFeldnameWirdKanonischGeschrieben(t *testing.T) {
	quelle := filepath.Join(t.TempDir(), "kanonisch.mse-set")
	if err := os.MkdirAll(quelle, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse_version: 2.1.2\ngame: probe\nstylesheet: probe-schlicht\n" +
		"card:\n\tname: Feuerball\n\trule_text: Füge drei Schaden zu.\n"

	if err := os.WriteFile(filepath.Join(quelle, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	speicher := &speicherDouble{}
	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), quelle); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Der Rundlauf: was gelesen wurde, muss unverändert wieder hinausgehen.
	gelesen, err := os.ReadFile(filepath.Join(quelle, "set")) // #nosec G304 -- t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.Contains(string(gelesen), "rule_text:") {
		t.Errorf("die Quelldatei führt kein `rule_text:` mehr: %q", gelesen)
	}

	// Und das Modell trägt den Namen des Spiels.
	if _, da := speicher.set.Cards[0].Value("rule text"); !da {
		t.Error("die Karte führt kein Feld `rule text`")
	}
}
