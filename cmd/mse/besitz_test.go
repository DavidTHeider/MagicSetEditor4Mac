package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// **Ein Nutzer schreibt die Karte eines anderen nicht um** —
// #1815.
//
// Nachgestellt am 2026-09-23: im eigenen Profil schrieb `mse feld 1 1 name
// Übernommen` den Kartennamen des ersten Nutzers um — die Ablage glaubte die
// Kennung, ohne zu fragen, wem ihr Set gehört. Hier an der Fläche, lesend und
// schreibend, dazu die Gegenprobe: beim Eigner steht der Name unverändert.
func TestEinNutzerSchreibtKeineFremdeKarteUm(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "nutzer-neu", "Bob"); code != CodeGut {
		t.Fatalf("Bob anlegen: %d (%s)", code, fehler)
	}

	if code, _, fehler := fahreAn(t, ort, "nutzer-wechseln", "2"); code != CodeGut {
		t.Fatalf("zu Bob wechseln: %d (%s)", code, fehler)
	}

	for name, argumente := range map[string][]string{
		"schreiben": {"feld", setID, karteID, "name", "Übernommen"},
		"lesen":     {"karten", setID},
	} {
		code, aus, fehler := fahreAn(t, ort, argumente...)
		if code != CodeFachlich || !strings.Contains(fehler, "nicht gefunden") {
			t.Errorf("Bob beim %s: %d %q (%s), erwartet „nicht gefunden“", name, code, aus, fehler)
		}
	}

	if code, _, fehler := fahreAn(t, ort, "nutzer-wechseln", "1"); code != CodeGut {
		t.Fatalf("zurückwechseln: %d (%s)", code, fehler)
	}

	_, aus, _ := fahreAn(t, ort, "karten", setID, "--json")

	var gelesen kartenAntwort
	if fehler := json.Unmarshal([]byte(aus), &gelesen); fehler != nil {
		t.Fatalf("kein JSON: %v (%q)", fehler, aus)
	}

	if len(gelesen.Karten) != 1 || gelesen.Karten[0].Name != "Feuerball" {
		t.Errorf("beim Eigner steht danach %+v", gelesen.Karten)
	}
}
