// `mse beispielsets`, `mse setangaben` und `mse vorbehalte`
// (#1107).
package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// **`setangaben` nennt Felder und Werte.**
func TestSetangabenNenntFelderUndWerte(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "setangaben", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		Felder []struct {
			Name string `json:"name"`
		} `json:"felder"`
		Werte []struct {
			Feld string `json:"feld"`
			Wert string `json:"wert"`
		} `json:"werte"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	// **Beide Listen sind da, auch wenn sie leer sind** — `null` zwänge jedes
	// Skript zu einer zweiten Prüfung.
	if strings.Contains(aus, "null") {
		t.Errorf("eine leere Liste kam als `null`: %s", aus)
	}
}

// **`setangaben setzen` schreibt und nennt den vorigen Wert.**
func TestSetangabenSetzenNenntDenVorigenWert(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "setangaben", setID, "setzen", "title", "Probe", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var erste struct {
		Vorher string `json:"vorher"`
		Hatte  bool   `json:"hatte"`
	}

	if err := json.Unmarshal([]byte(aus), &erste); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if erste.Hatte {
		t.Errorf("das Feld gab es angeblich schon: %+v", erste)
	}

	// **Der zweite Schreibvorgang nennt den ersten Wert** — der Beleg, dass
	// geschrieben wurde und nicht nur geantwortet.
	_, zweiter, _ := fahreAn(t, ort, "setangaben", setID, "setzen", "title", "Zweite", "--json")

	var nachher struct {
		Vorher string `json:"vorher"`
		Hatte  bool   `json:"hatte"`
	}

	if err := json.Unmarshal([]byte(zweiter), &nachher); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, zweiter)
	}

	if !nachher.Hatte || nachher.Vorher != "Probe" {
		t.Errorf("der vorige Wert lautet %+v statt »Probe«", nachher)
	}
}

// **`vorbehalte` gibt eine Liste, auch wenn sie leer ist.**
func TestVorbehalteGibtEineListe(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "vorbehalte", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "vorbehalte") || strings.Contains(aus, "null") {
		t.Errorf("die Antwort trägt keine Liste: %s", aus)
	}
}

// **`beispielsets` braucht keinen Bestand.**
func TestBeispielsetsNenntPfade(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "beispielsets", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "sets") || strings.Contains(aus, "null") {
		t.Errorf("die Antwort trägt keine Liste: %s", aus)
	}
}
