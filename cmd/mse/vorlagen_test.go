package main

// **Die Exportvorlagen an der Kommandozeile** (#1291).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// vorlagenset legt ein Set an, dessen Spiel eine Exportvorlage hat.
func vorlagenset(t *testing.T) (ort, setID string) {
	t.Helper()

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatBereitstellen(t, zuhause)

	ort = frischerBestand(t)

	_, aus, fehler := fahreAn(t, ort,
		"set-neu", "Vorlagenset", "probe", "probe-schlicht", "--json")

	var angelegt setAnlegeAntwort
	if err := json.Unmarshal([]byte(aus), &angelegt); err != nil {
		t.Fatalf("Set anlegen: %v (%s)", err, fehler)
	}

	if code, _, kartenfehler := fahreAn(t, ort,
		"karte", angelegt.ID, "anlegen"); code != CodeGut {
		t.Fatalf("Karte anlegen: %d (%s)", code, kartenfehler)
	}

	return ort, angelegt.ID
}

// **`vorlagen <set>` nennt die Exportvorlagen.**
func TestVorlagenNenntDieVorlagen(t *testing.T) {
	ort, setID := vorlagenset(t)

	code, aus, fehler := fahreAn(t, ort, "vorlagen", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var gelesen vorlagenAntwort
	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, aus)
	}

	if len(gelesen.Vorlagen) == 0 {
		t.Fatalf("keine Vorlage: %s", aus)
	}

	// **Die Befunde sind nie `nil`** (§3).
	if gelesen.Befunde == nil {
		t.Errorf("die Befunde sind null: %s", aus)
	}
}

// **`vorlagen <set> ausfuehren` schreibt das Ergebnis.**
func TestVorlagenAusfuehrenSchreibtEineDatei(t *testing.T) {
	ort, setID := vorlagenset(t)

	ziel := filepath.Join(t.TempDir(), "ausfuhr.txt")

	code, aus, fehler := fahreAn(t, ort, "vorlagen", setID, "ausfuehren",
		"probe-liste.mse-export-template", ziel)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}

	if !strings.Contains(aus, ziel) {
		t.Errorf("die Ausgabe nennt den Pfad nicht: %q", aus)
	}
}

// **Eine Vorlage, die es nicht gibt, ist ein fachliches Nein.**
func TestVorlagenAusfuehrenMitEinerUnbekanntenVorlage(t *testing.T) {
	ort, setID := vorlagenset(t)

	if code, _, _ := fahreAn(t, ort, "vorlagen", setID, "ausfuehren",
		"gibtsnicht.mse-export-template",
		filepath.Join(t.TempDir(), "raus.txt")); code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeFachlich)
	}
}

// **Und die Bedienfehler.**
func TestVorlagenBedienfehler(t *testing.T) {
	ort, setID := vorlagenset(t)

	faelle := map[string][]string{
		"ohne Setkennung":      {"vorlagen"},
		"krumme Setkennung":    {"vorlagen", "x"},
		"unbekannter Zug":      {"vorlagen", setID, "irgendwas"},
		"ausfuehren ohne Ziel": {"vorlagen", setID, "ausfuehren", "probe-liste.mse-export-template"},
		"option ohne Wert":     {"vorlagen", setID, "option", "probe-liste.mse-export-template", "feld"},
		"krummes Paar": {
			"vorlagen", setID, "ausfuehren", "probe-liste.mse-export-template",
			"/tmp/a.txt", "ohneGleich",
		},
	}

	for name, aufruf := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, aufruf...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
			}
		})
	}
}

// **Ein Optionswert lässt sich behalten.**
func TestVorlagenOptionSetzen(t *testing.T) {
	ort, setID := vorlagenset(t)

	_, aus, _ := fahreAn(t, ort, "vorlagen", setID, "--json")

	var gelesen vorlagenAntwort
	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil {
		t.Fatalf("Antwort lesen: %v", err)
	}

	var paket, feld string

	for _, vorlage := range gelesen.Vorlagen {
		if len(vorlage.Optionen) > 0 {
			paket, feld = vorlage.Paket, vorlage.Optionen[0].Name

			break
		}
	}

	if paket == "" {
		t.Skip("keine Vorlage der Saat führt eine Option")
	}

	if code, _, fehler := fahreAn(t, ort,
		"vorlagen", setID, "option", paket, feld, "geändert"); code != CodeGut {
		t.Fatalf("setzen: %d (%s)", code, fehler)
	}

	_, danach, _ := fahreAn(t, ort, "vorlagen", setID, "--json")
	if !strings.Contains(danach, "geändert") {
		t.Errorf("der Wert kam nicht zurück: %s", danach)
	}
}
