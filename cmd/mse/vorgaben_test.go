package main

// **Die Profilvorgaben an der Kommandozeile** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.

import (
	"strings"
	"testing"
)

// **Alle fünf Arten kommen — auch die leeren.**
func TestDieVorgabenNennenAlleFuenf(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "vorgaben", "1")
	if code != CodeGut {
		t.Fatalf("vorgaben: %d (%s)", code, fehler)
	}

	for _, art := range []string{"format", "spiel", "vorlage", "deck", "lagerort"} {
		if !strings.Contains(aus, art) {
			t.Errorf("die Art %q fehlt in %q", art, aus)
		}
	}
}

// **Eine Vorgabe wird gesetzt und kommt zurück.**
//
// **Über das Spiel**, weil `mitKarte` ein Spielpaket einrichtet: es ist die
// einzige der fünf Arten, für die dieser Bestand ein Angebot hat.
func TestEineVorgabeWirdGesetztUndKommtZurueck(t *testing.T) {
	ort, _, _ := mitKarte(t)

	mussEingabe(t, ort, `["probe"]`, "vorgabe-setzen", "1", "spiel")

	code, aus, fehler := fahreAn(t, ort, "vorgaben", "1")
	if code != CodeGut {
		t.Fatalf("vorgaben: %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "probe") {
		t.Errorf("die gesetzte Vorgabe fehlt in %q", aus)
	}

	t.Run("und ein leeres Feld räumt sie ab", func(t *testing.T) {
		mussEingabe(t, ort, `[]`, "vorgabe-setzen", "1", "spiel")

		_, nachher, _ := fahreAn(t, ort, "vorgaben", "1")
		if strings.Contains(nachher, "probe") {
			t.Errorf("die abgeräumte Vorgabe steht noch da: %q", nachher)
		}
	})
}

// **Was es nicht gibt, wird abgewiesen.**
func TestEineVorgabeAufNichts(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string]struct {
		eingabe   string
		argumente []string
	}{
		"ein Spiel, das es nicht gibt": {
			`["kartenrücken"]`, []string{"vorgabe-setzen", "1", "spiel"},
		},
		"eine Art, die es nicht gibt": {
			`["1"]`, []string{"vorgabe-setzen", "1", "kartenrücken"},
		},
		"ein Profil, das es nicht gibt": {
			"", []string{"vorgaben", "4711"},
		},
		"eine Kennung, die keine Zahl ist": {
			"", []string{"vorgaben", "keine-zahl"},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			code, _, _ := fahreAnMitEingabe(t, ort, fall.eingabe, fall.argumente...)
			if code != CodeFachlich {
				t.Errorf("Rückgabewert %d, erwartet war CodeFachlich", code)
			}
		})
	}
}

// **Beide Befehle sagen, wie sie heißen wollen.**
func TestDieVorgabenbefehleWeisenFalscheAufrufeAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string][]string{
		"vorgaben ohne Profil":      {"vorgaben"},
		"vorgaben mit zweien":       {"vorgaben", "1", "2"},
		"vorgabe-setzen ohne Art":   {"vorgabe-setzen", "1"},
		"vorgabe-setzen mit dreien": {"vorgabe-setzen", "1", "spiel", "probe"},
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet war CodeBedienung", code)
			}
		})
	}

	t.Run("und ein krummes Feld ist keiner", func(t *testing.T) {
		if code, _, _ := fahreAnMitEingabe(
			t, ort, "[", "vorgabe-setzen", "1", "spiel"); code == CodeGut {
			t.Error("das krumme Feld kam durch")
		}
	})
}

// **Und beide geben auf Wunsch JSON.**
func TestDieVorgabenbefehleGebenJSON(t *testing.T) {
	ort, _, _ := mitKarte(t)

	_, aus, _ := fahreAn(t, ort, "vorgaben", "1", "--json")
	if !strings.Contains(aus, `"moeglich"`) {
		t.Errorf("das Angebot fehlt in %q", aus)
	}

	gesetzt := mussEingabe(t, ort, `["probe"]`, "vorgabe-setzen", "1", "spiel",
		"--json")
	if !strings.Contains(gesetzt, `"vorgaben"`) {
		t.Errorf("der Stand fehlt in %q", gesetzt)
	}
}
