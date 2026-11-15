package mcp_test

// **Das Werkzeug über die eigenen Partien** —
// #1600.
//
// **Eine eigene Datei**, weil `partiewerkzeug_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge — und weil der Schnitt der Frage folgt: dort geht
// es um **eine** Partie, hier darum, welche es gibt.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// **`lesen.partien` nennt die Partien und keine Karte** —
// #1600.
//
// **Die zweite Zusicherung ist die wichtigere**: eine Liste ist keine Sicht.
// Was in einer Partie liegt, sagt `lesen.partie` für **einen** Sitz — wer es
// hier mitgäbe, umginge die Projektion, statt sie zu benutzen.
func TestLesenPartienNenntDiePartienUndKeineKarte(t *testing.T) {
	t.Parallel()

	roh, fehler := rufePartiewerkzeug(t, &partienprobe{}, "lesen.partien", `{}`)
	if fehler != nil {
		t.Fatalf("lesen.partien: %v", fehler)
	}

	var zeilen []struct {
		ID          string `json:"id"`
		Stand       string `json:"stand"`
		Objekte     int    `json:"objekte"`
		EigenerSitz string `json:"eigenerSitz"`
		Dran        bool   `json:"dran"`

		Sitze []struct {
			Kennung string `json:"kennung"`
			Name    string `json:"name"`
		} `json:"sitze"`
	}

	if err := json.Unmarshal([]byte(roh), &zeilen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	switch {
	case len(zeilen) != 1:
		t.Fatalf("es kamen %d Zeilen", len(zeilen))
	case zeilen[0].ID != "p1":
		t.Errorf("die Kennung ist %q", zeilen[0].ID)
	case zeilen[0].Objekte != 12:
		t.Errorf("gezählt sind %d Objekte", zeilen[0].Objekte)
	case len(zeilen[0].Sitze) != 1 || zeilen[0].Sitze[0].Name != "Anna":
		t.Errorf("die Sitze sind %+v", zeilen[0].Sitze)
	case zeilen[0].EigenerSitz != "anna" || !zeilen[0].Dran:
		t.Errorf("eigener Sitz %q, dran=%v", zeilen[0].EigenerSitz, zeilen[0].Dran)
	}

	// **Und kein Feld nennt eine Karte** — geprüft am rohen JSON, nicht an
	// einer Gestalt, die nur die Felder kennt, die ich hineingeschrieben habe.
	for _, wort := range []string{"karte", "objekte\":[", "zone"} {
		if strings.Contains(roh, wort) {
			t.Errorf("die Antwort nennt %q: %s", wort, roh)
		}
	}
}

// **Und ein Fehler der Naht kommt durch.**
func TestLesenPartienReichtDenFehlerDurch(t *testing.T) {
	t.Parallel()

	probe := &partienprobe{klagt: errors.New("der Bestand ist zu")}

	if _, fehler := rufePartiewerkzeug(t, probe, "lesen.partien",
		`{}`); fehler == nil {
		t.Error("der Fehler wurde verschluckt")
	}
}
