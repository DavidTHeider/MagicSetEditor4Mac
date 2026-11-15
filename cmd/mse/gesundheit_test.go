package main

// **Der Betriebszustand auf der Kommandozeile** — Werkstück O9 von Phase 5.2
// (#1433).
//
// **Der Rückgabewert ist die Antwort**: 0 bereit, 1 nicht bereit — dieselbe
// Aussage, die der Dienst als 200 oder 503 gibt.

import (
	"strings"
	"testing"
)

func TestDerBetriebszustandSagtLebtUndBereit(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "gesundheit")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, erwartet := range []string{"lebt\ttrue", "bereit\ttrue"} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("die Ausgabe nennt %q nicht:\n%s", erwartet, aus)
		}
	}

	// **Kein Grund, wenn es keinen gibt** — eine leere Zeile wäre eine ohne
	// Aussage.
	if strings.Contains(aus, "grund") {
		t.Errorf("die Ausgabe nennt einen Grund:\n%s", aus)
	}
}

func TestDerBetriebszustandAlsJSONUndMitArgument(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, _ := fahreAn(t, ort, "gesundheit", "--json")

	switch {
	case code != CodeGut:
		t.Errorf("als JSON: %d %s", code, aus)
	case !strings.Contains(aus, `"bereit": true`):
		t.Errorf("als JSON: %s", aus)
	case !strings.Contains(aus, `"fremde": []`):
		t.Errorf("die Fremdenliste war nicht leer, sondern: %s", aus)
	}

	if code, _, _ := fahreAn(t, ort, "gesundheit", "mehr"); code != CodeBedienung {
		t.Errorf("mit Argument: %d", code)
	}
}
