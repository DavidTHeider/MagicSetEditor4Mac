package main

// **Die Folgen einer Berichtigung an der Kommandozeile** — DOM-04.AK-02
// (#1699).

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDieFolgenEinerBerichtigungAnDerKommandozeile(t *testing.T) {
	ort, _ := eingefahren(t)

	turnierID := strings.TrimSpace(mussLaufen(t, ort, "turnier", "neu", "Freitag", "magic"))
	for _, name := range []string{"Anna", "Bert"} {
		mussLaufen(t, ort, "turnier", "anmelden", turnierID, name)
	}

	mussLaufen(t, ort, "turnier", "runde", turnierID)

	var gelesen struct {
		Turnier struct {
			Runden []struct {
				Paarungen []struct {
					ID string `json:"id"`
				} `json:"paarungen"`
			} `json:"runden"`
		} `json:"turnier"`
	}

	roh := mussLaufen(t, ort, "turnier", turnierID, "--json")
	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil || len(gelesen.Turnier.Runden) != 1 {
		t.Fatalf("das Turnier: %v (%s)", err, roh)
	}

	paarung := gelesen.Turnier.Runden[0].Paarungen[0].ID

	mussLaufen(t, ort, "turnier", "ergebnis", turnierID, paarung, "2", "0")

	// **Die Frage schreibt nicht** — und nennt, wer seinen Platz wechselt.
	folgen := mussLaufen(t, ort, "turnier", "folgen", turnierID, paarung, "0", "2")
	if !strings.Contains(folgen, "Anna") || !strings.Contains(folgen, "Bert") {
		t.Errorf("die Folgen nennen nicht beide: %q", folgen)
	}

	if rangliste := mussLaufen(t, ort, "turnier", "rangliste", turnierID); !strings.HasPrefix(
		rangliste, "1.\tAnna") {
		t.Errorf("die Frage hat geschrieben — die Rangliste: %q", rangliste)
	}

	// **»weg« fragt nach dem Zurücknehmen**, `--json` gibt die Gestalt.
	if roh := mussLaufen(t, ort, "turnier", "folgen", turnierID, paarung, "weg", "--json"); !strings.Contains(roh, `"verschoben"`) {
		t.Errorf("die JSON-Gestalt: %q", roh)
	}

	if code, _, _ := fahreAn(t, ort, "turnier", "folgen", turnierID); code != CodeBedienung {
		t.Errorf("ohne Paarung und Stand: %d, erwartet %d", code, CodeBedienung)
	}

	if code, _, _ := fahreAn(t, ort, "turnier", "folgen", turnierID, paarung, "zwei"); code != CodeBedienung {
		t.Errorf("mit einem Wort statt einer Zahl: %d, erwartet %d", code, CodeBedienung)
	}
}
