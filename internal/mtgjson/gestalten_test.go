package mtgjson_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// pfad baut einen Pfad unter `testdaten/mtgjson/`.
func pfad(teile ...string) string {
	return filepath.Join(append([]string{"..", "..", "testdaten", "mtgjson"}, teile...)...)
}

// setcodes liest eine Quelle und gibt ihre Setcodes in ihrer Reihenfolge.
func setcodes(t *testing.T, ort string) []string {
	t.Helper()

	quelle, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies(%q): %v", ort, err)
	}

	codes := make([]string, 0, len(quelle.Sets))
	for _, set := range quelle.Sets {
		codes = append(codes, set.Code)
	}

	return codes
}

// Drei Gestalten führen zu demselben Ergebnis.
//
// MTG JSON kommt in mehr als einer Form: als *AllSets*-Datei mit allen Sets, als
// *Individual Set*-Datei mit einem, und als Verzeichnis mit einer Datei je Set.
// Wer nur eine davon liest, kann genau die Datei nicht öffnen, die jemand hat.
//
// Die Reihenfolge ist **alphabetisch und nicht die der Datei**: eine
// JSON-Abbildung hat keine Reihenfolge, und das Dateisystem gibt keine zu.
func TestJedeGestaltWirdGelesen(t *testing.T) {
	faelle := map[string]struct {
		ort      string
		erwartet []string
	}{
		"eine AllSets-Datei":                {ort: pfad("alle.json"), erwartet: []string{"PRB", "ZWE"}},
		"eine Individual-Set-Datei":         {ort: pfad("probe.json"), erwartet: []string{"PRB"}},
		"ein Verzeichnis einzelner Dateien": {ort: pfad("einzeln"), erwartet: []string{"PRB", "ZWE"}},

		// Vor MTGJSON 5 gab es den Umschlag aus `meta` und `data` nicht. Eine
		// Datei von 2019 ist kein Sonderfall, sondern der Normalfall fremder
		// Daten (§8).
		"eine Datei ohne den Umschlag": {ort: pfad("ohneumschlag.json"), erwartet: []string{"ALT"}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			gefunden := setcodes(t, fall.ort)

			if len(gefunden) != len(fall.erwartet) {
				t.Fatalf("Setcodes: %v, erwartet %v", gefunden, fall.erwartet)
			}

			for i, code := range fall.erwartet {
				if gefunden[i] != code {
					t.Errorf("Setcode %d: %q, erwartet %q", i, gefunden[i], code)
				}
			}
		})
	}
}

// Ein Verzeichnis übergeht, was keine JSON-Datei ist.
//
// Ein bezogenes Verzeichnis enthält mehr als Sets — eine LIESMICH, eine
// Prüfsummendatei, das Übliche. Daran zu scheitern hieße, dass niemand ein
// bezogenes Verzeichnis öffnen kann, ohne es vorher aufzuräumen.
func TestVerzeichnisUebergehtFremdes(t *testing.T) {
	if codes := setcodes(t, pfad("einzeln")); len(codes) != 2 {
		t.Errorf("Setcodes: %v, erwartet genau die beiden JSON-Dateien", codes)
	}
}
