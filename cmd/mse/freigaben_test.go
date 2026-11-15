package main

// **Die Freigaben an der Kommandozeile** — Werkstück 11 von Phase 4.12
// (#1351).

import (
	"encoding/json"
	"strings"
	"testing"
)

// zweiterNutzerAnDerZeile legt einen zweiten Nutzer an und gibt seine Kennung.
func zweiterNutzerAnDerZeile(t *testing.T, ort string) string {
	t.Helper()

	code, aus, fehler := fahreAn(t, ort, "nutzer-neu", "Zweiter", "--json")
	if code != CodeGut {
		t.Fatalf("den zweiten Nutzer anlegen: %d (%s)", code, fehler)
	}

	var antwort struct {
		Nutzer struct {
			ID string `json:"id"`
		} `json:"nutzer"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("die Antwort lesen: %v — %s", err, aus)
	}

	return antwort.Nutzer.ID
}

// **Der ganze Weg: vergeben, nennen, zurückziehen.**
func TestDerGanzeFreigabenwegAnDerZeile(t *testing.T) {
	ort, _, _ := mitKarte(t)

	nutzerID := zweiterNutzerAnDerZeile(t, ort)

	// **Der arbeitende Mensch ist Nutzer 1** — der Zeiger steht auf seinem
	// Profil, und damit verwaltet er es.
	if code, _, fehler := fahreAn(
		t, ort, "freigabe-setzen", "1", nutzerID, "mitarbeiter"); code != CodeGut {
		t.Fatalf("setzen: %d (%s)", code, fehler)
	}

	code, aus, fehler := fahreAn(t, ort, "freigaben", "1")
	if code != CodeGut {
		t.Fatalf("freigaben: %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "mitarbeiter") || !strings.Contains(aus, "Zweiter") {
		t.Errorf("die Freigabe fehlt in %q", aus)
	}

	if code, _, fehler := fahreAn(
		t, ort, "freigabe-loeschen", "1", nutzerID); code != CodeGut {
		t.Fatalf("löschen: %d (%s)", code, fehler)
	}

	if _, nachher, _ := fahreAn(t, ort, "freigaben", "1"); strings.Contains(
		nachher, "Zweiter") {
		t.Errorf("die zurückgezogene Freigabe steht noch da: %q", nachher)
	}
}

// **Die eigene Rolle steht in der JSON-Antwort.**
func TestDieEigeneRolleStehtInDerAntwort(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "freigaben", "1", "--json")
	if code != CodeGut {
		t.Fatalf("freigaben: %d (%s)", code, fehler)
	}

	// **Der Eigner ist Verwalter — ohne Zeile im Bestand.**
	if !strings.Contains(aus, `"eigeneRolle": "verwalter"`) {
		t.Errorf("die eigene Rolle fehlt in %q", aus)
	}

	if !strings.Contains(aus, `"sichtbarkeit": "privat"`) {
		t.Errorf("die Sichtbarkeit fehlt in %q", aus)
	}
}

// **Was nicht geht, wird abgewiesen.**
func TestEineFreigabeAufNichts(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string][]string{
		"ein Profil, das es nicht gibt":    {"freigaben", "4711"},
		"eine Kennung, die keine Zahl ist": {"freigaben", "keine-zahl"},
		"eine Rolle, die es nicht gibt": {
			"freigabe-setzen", "1", "1", "gott",
		},
		"eine Freigabe an den eigenen Eigner": {
			"freigabe-setzen", "1", "1", "leser",
		},
		"eine Freigabe, die es nicht gibt": {"freigabe-loeschen", "1", "4711"},
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, argumente...); code != CodeFachlich {
				t.Errorf("Rückgabewert %d, erwartet war CodeFachlich", code)
			}
		})
	}
}

// **Und jeder der drei sagt, wie er heißen will.**
func TestDieFreigabenbefehleWeisenFalscheAufrufeAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := map[string][]string{
		"freigaben ohne Profil":         {"freigaben"},
		"freigaben mit zweien":          {"freigaben", "1", "2"},
		"freigabe-setzen ohne Rolle":    {"freigabe-setzen", "1", "2"},
		"freigabe-loeschen ohne Nutzer": {"freigabe-loeschen", "1"},
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet war CodeBedienung", code)
			}
		})
	}
}
