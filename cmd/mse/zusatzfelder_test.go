package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// **`mse zusatzfeld` liest und setzt die Zusatzfelder einer Karte** (#2298) —
// dieselbe Fähigkeit wie im Fenster, über MCP und REST, gegen die Saat, deren
// Stil `vermerk` führt.
func TestZusatzfeldLesenUndSetzen(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort,
		"zusatzfeld", setID, karteID, "setzen", "vermerk", "Vorabdruck"); code != CodeGut {
		t.Fatalf("setzen: Rückgabewert %d (%s)", code, fehler)
	}

	_, aus, _ := fahreAn(t, ort, "zusatzfeld", setID, karteID, "--json")

	var antwort struct {
		Felder []struct {
			Form struct {
				Name string `json:"name"`
			} `json:"form"`
			Wert        string `json:"wert"`
			Gespeichert bool   `json:"gespeichert"`
		} `json:"felder"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if len(antwort.Felder) != 1 || antwort.Felder[0].Form.Name != "vermerk" ||
		antwort.Felder[0].Wert != "Vorabdruck" || !antwort.Felder[0].Gespeichert {
		t.Errorf("Felder: %s", aus)
	}

	// **Und als Text** — Name und Wert in einer Zeile.
	if _, text, _ := fahreAn(t, ort, "zusatzfeld", setID, karteID); !strings.Contains(text, "vermerk") ||
		!strings.Contains(text, "Vorabdruck") {
		t.Errorf("Text: %s", text)
	}
}

// **Was nicht geht, wird benannt** — falsche Gestalt als Bedienfehler, ein
// gerechnetes Feld als fachlicher.
func TestZusatzfeldWeistAbWasNichtGeht(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	for _, fall := range []struct {
		name      string
		argumente []string
		code      int
	}{
		{name: "ohne Karte", argumente: []string{"zusatzfeld", setID}, code: CodeBedienung},
		{name: "eine falsche Kennung", argumente: []string{"zusatzfeld", setID, "x"}, code: CodeBedienung},
		{name: "eine falsche Setkennung", argumente: []string{"zusatzfeld", "x", karteID}, code: CodeBedienung},
		{name: "setzen ohne Wert", argumente: []string{"zusatzfeld", setID, karteID, "setzen", "vermerk"}, code: CodeBedienung},
		{name: "ein gerechnetes Feld", argumente: []string{"zusatzfeld", setID, karteID, "setzen", "grund", "x"}, code: CodeFachlich},
		{name: "eine Karte, die es nicht gibt", argumente: []string{"zusatzfeld", setID, "999"}, code: CodeFachlich},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, fall.argumente...); code != fall.code {
				t.Errorf("Rückgabewert %d, erwartet %d", code, fall.code)
			}
		})
	}
}
