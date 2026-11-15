package main

import (
	"strings"
	"testing"
)

// **Der Verlauf an der Zeile** — Werkstück U6 von Phase 5.1 (#1417):
// `mse verlauf <set>` nennt je Schritt Vorgang, Karte, Feld, Bisher und
// Danach — und `--json` gibt, was ein Skript an `zuruecknehmen` reicht.
func TestVerlaufNenntDieSchritte(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	if code, aus, fehler := fahreAn(t, ort, "feld", setID, karteID, "artist", "Erster"); code != CodeGut {
		t.Fatalf("feld: %d (%s %s)", code, aus, fehler)
	}

	code, aus, fehler := fahreAn(t, ort, "verlauf", setID)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "artist") || !strings.Contains(aus, "Erster") {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

func TestVerlaufAlsJSON(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	if code, _, _ := fahreAn(t, ort, "felder-setzen", setID, "artist", "Masse", karteID); code != CodeGut {
		t.Fatal("felder-setzen")
	}

	code, aus, fehler := fahreAn(t, ort, "verlauf", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, `"schritte"`) || !strings.Contains(aus, `"vorgang"`) ||
		!strings.Contains(aus, `"umfang": 1`) {
		t.Errorf("die Ausgabe lautet %q", aus)
	}
}

func TestVerlaufBrauchtEineSetkennung(t *testing.T) {
	ort, _, _ := mitKarte(t)

	for name, argumente := range map[string][]string{
		"ohne Set":        {"verlauf"},
		"unbrauchbar":     {"verlauf", "keine-zahl"},
		"zu viele Wörter": {"verlauf", "1", "2"},
	} {
		t.Run(name, func(t *testing.T) {
			if code, aus, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d (%s)", code, aus)
			}
		})
	}
}
