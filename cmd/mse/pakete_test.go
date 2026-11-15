// `mse pakete` und `mse paket-beziehen`
// (#1107).
package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// **`pakete` frischt auf und zeigt, was auf der Platte liegt.**
//
// **Auffrischen ist der Punkt**: der Katalog steht im Bestand und ist auf
// einem frischen Rechner leer. Ohne diesen Zug gäbe die Kommandozeile eine
// andere Antwort als das Fenster (§16).
func TestPaketeZeigtDieSaat(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "pakete", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var antwort struct {
		Pakete []struct {
			Art  string `json:"art"`
			Name string `json:"name"`
		} `json:"pakete"`
		Befunde []string `json:"befunde"`
	}

	if err := json.Unmarshal([]byte(aus), &antwort); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if len(antwort.Pakete) == 0 {
		t.Fatalf("kein Paket — die Saat liegt im Zuhause:\n%s", aus)
	}

	gefunden := false

	for _, paket := range antwort.Pakete {
		if paket.Name == "probe" {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("»probe« fehlt: %+v", antwort.Pakete)
	}

	if antwort.Befunde == nil {
		t.Error("die Befunde kamen als null statt als leere Liste")
	}
}

// **Ein Bezug ohne Netz scheitert fachlich und nicht als Bedienfehler.**
//
// **Die Adresse zeigt ins Leere und nicht ins Netz** (§3): ein Test, der
// wirklich lädt, prüft die Leitung und nicht den Code — und er lädt bei jedem
// Lauf.
func TestPaketBeziehenScheitertFachlich(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, _, fehler := fahreAn(t, ort, "paket-beziehen",
		"http://127.0.0.1:1/gibtsnicht", "abc1234", "data", "probe")
	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d statt %d (%s)", code, CodeFachlich, fehler)
	}

	if fehler == "" {
		t.Error("es kam keine Meldung")
	}
}

// **Falsche Aufrufe sind Bedienfehler.**
func TestPaketbefehleWeisenFalscheAufrufeAb(t *testing.T) {
	ort, _, _ := mitKarte(t)

	faelle := [][]string{
		{"pakete", "zuviel"},
		{"paket-beziehen"},
		{"paket-beziehen", "a", "b", "c"},
	}

	for _, fall := range faelle {
		code, _, fehler := fahreAn(t, ort, fall...)
		if code != CodeBedienung {
			t.Errorf("%v endete mit %d statt %d (%s)",
				fall, code, CodeBedienung, strings.SplitN(fehler, "\n", 2)[0])
		}
	}
}
