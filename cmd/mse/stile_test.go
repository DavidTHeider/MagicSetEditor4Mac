// `mse stile` und `mse stiloption`
// (#1107).
package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// **`stile` nennt die wählbaren Stile und markiert den gewählten.**
func TestStileNenntDenGewaehlten(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "stile", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		Stile []struct {
			Paket    string `json:"paket"`
			Gewaehlt bool   `json:"gewaehlt"`
		} `json:"stile"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if len(antwort.Stile) == 0 {
		t.Fatalf("kein Stil — die Saat müsste welche haben:\n%s", aus)
	}

	gewaehlt := false

	for _, stil := range antwort.Stile {
		if stil.Gewaehlt {
			gewaehlt = true
		}
	}

	if !gewaehlt {
		t.Errorf("kein Stil ist markiert: %+v", antwort.Stile)
	}
}

// **`stile … folgen` sagt, was ein Wechsel kostet — ohne ihn zu tun.**
func TestStilfolgenAendernNichts(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	vorher := gewaehlterStil(t, ort, setID)

	code, _, fehler := fahreAn(t, ort, "stile", setID, "folgen", "probe-einzeiler")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if nachher := gewaehlterStil(t, ort, setID); nachher != vorher {
		t.Errorf("»folgen« hat gewechselt: %s → %s", vorher, nachher)
	}
}

// **`stile … wechseln` wechselt wirklich.**
func TestStilWechselnWechselt(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	vorher := gewaehlterStil(t, ort, setID)

	code, _, fehler := fahreAn(t, ort, "stile", setID, "wechseln", "probe-einzeiler")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	nachher := gewaehlterStil(t, ort, setID)
	if nachher == vorher {
		t.Fatalf("der Stil steht unverändert bei %s", nachher)
	}

	if nachher != "probe-einzeiler" {
		t.Errorf("gewechselt zu %q statt zu »probe-einzeiler«", nachher)
	}
}

// gewaehlterStil liest den markierten Stil über `mse stile`.
func gewaehlterStil(t *testing.T, ort, setID string) string {
	t.Helper()

	_, aus, _ := fahreAn(t, ort, "stile", setID, "--json")

	var antwort struct {
		Stile []struct {
			Paket    string `json:"paket"`
			Gewaehlt bool   `json:"gewaehlt"`
		} `json:"stile"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	for _, stil := range antwort.Stile {
		if stil.Gewaehlt {
			return stil.Paket
		}
	}

	return ""
}

// **`stiloption setzen` schreibt, `zuruecksetzen` nimmt zurück.**
func TestStiloptionSetzenUndZuruecksetzen(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	_, aus, _ := fahreAn(t, ort, "stiloption", setID, "--json")

	var antwort struct {
		Optionen []struct {
			Name    string `json:"name"`
			Wert    string `json:"wert"`
			Gesetzt bool   `json:"gesetzt"`
		} `json:"optionen"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if len(antwort.Optionen) == 0 {
		t.Skipf("der Saatstil hat keine Optionen:\n%s", aus)
	}

	name := antwort.Optionen[0].Name

	if code, _, fehler := fahreAn(t, ort,
		"stiloption", setID, "setzen", name, "probe"); code != CodeGut {
		t.Fatalf("setzen: Rückgabewert %d (%s)", code, fehler)
	}

	if wert, gesetzt := option(t, ort, setID, name); wert != "probe" || !gesetzt {
		t.Fatalf("nach dem Setzen: %q (gesetzt: %v)", wert, gesetzt)
	}

	if code, _, fehler := fahreAn(t, ort,
		"stiloption", setID, "zuruecksetzen", name); code != CodeGut {
		t.Fatalf("zurücksetzen: Rückgabewert %d (%s)", code, fehler)
	}

	if _, gesetzt := option(t, ort, setID, name); gesetzt {
		t.Error("die Option gilt nach dem Zurücksetzen noch als gesetzt")
	}
}

// option liest eine Stiloption über `mse stiloption`.
func option(t *testing.T, ort, setID, name string) (wert string, gesetzt bool) {
	t.Helper()

	_, aus, _ := fahreAn(t, ort, "stiloption", setID, "--json")

	var antwort struct {
		Optionen []struct {
			Name    string `json:"name"`
			Wert    string `json:"wert"`
			Gesetzt bool   `json:"gesetzt"`
		} `json:"optionen"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	for _, option := range antwort.Optionen {
		if option.Name == name {
			return option.Wert, option.Gesetzt
		}
	}

	t.Fatalf("die Option %q steht nicht in der Liste:\n%s", name, aus)

	return "", false
}

// **Ein falscher Unterbefehl ist ein Bedienfehler.**
func TestStileWeistFalscheAufrufeAb(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	faelle := [][]string{
		{"stile"},
		{"stile", "keinezahl"},
		{"stile", setID, "folgen"},
		{"stile", setID, "gibtsnicht", "probe-schlicht"},
		{"stiloption"},
		{"stiloption", setID, "setzen"},
		{"stiloption", setID, "zuruecksetzen", "a", "b"},
	}

	for _, fall := range faelle {
		code, _, fehler := fahreAn(t, ort, fall...)
		if code != CodeBedienung {
			t.Errorf("%v endete mit %d statt %d (%s)",
				fall, code, CodeBedienung, strings.SplitN(fehler, "\n", 2)[0])
		}
	}
}
