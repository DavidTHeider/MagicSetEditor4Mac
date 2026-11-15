package main

// **Format und Vorlage eines Turniers an der Kommandozeile** — DOM-04.AK-01
// und AK-03 (#1699).
//
// **Die fachliche Zusage prüft `internal/flaechen`**; hier stehen die
// Bedienfehler und die Gestalt der Eingabe.

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEinTurnierMitFormatUndNachVorlage(t *testing.T) {
	ort, _ := eingefahren(t)

	formatID := strings.TrimSpace(mussLaufen(t, ort,
		"format", "neu", "konstruiert", "Pauper", "magic"))

	var angelegt struct {
		Turnier struct {
			ID       string `json:"id"`
			FormatID string `json:"formatID"`
		} `json:"turnier"`
	}

	// **`--format` darf an jeder Stelle stehen** — die Wörter davor und
	// danach behalten ihre Bedeutung.
	roh := mussLaufen(t, ort, "turnier", "neu", "Freitag", "magic", "--format", formatID,
		"Hinterzimmer", "einzelspiel", "--json")
	if err := json.Unmarshal([]byte(roh), &angelegt); err != nil {
		t.Fatalf("die Antwort: %v (%s)", err, roh)
	}

	if angelegt.Turnier.FormatID != formatID {
		t.Errorf("das Turnier trägt Format %q, erwartet %q", angelegt.Turnier.FormatID, formatID)
	}

	mussLaufen(t, ort, "turnier", "anmelden", angelegt.Turnier.ID, "Anna")

	kopie := strings.TrimSpace(mussLaufen(t, ort, "turnier", "kopieren",
		angelegt.Turnier.ID, "Freitag II", "--mit-teilnehmenden"))
	if kopie == "" || kopie == angelegt.Turnier.ID {
		t.Fatalf("die Kopie hat die Kennung %q", kopie)
	}

	if aus := mussLaufen(t, ort, "turnier", kopie); !strings.Contains(aus, "Anna") {
		t.Errorf("die Kopie nennt Anna nicht: %q", aus)
	}

	// **Ohne den Schalter nur die Einstellungen.**
	ohne := strings.TrimSpace(mussLaufen(t, ort, "turnier", "kopieren",
		angelegt.Turnier.ID, "Freitag III"))
	if aus := mussLaufen(t, ort, "turnier", ohne); strings.Contains(aus, "Anna") {
		t.Errorf("die Kopie ohne Schalter nennt Anna: %q", aus)
	}
}

// **Bedienfehler enden mit dem Bedienungscode** — und vor dem Bestand.
func TestDieVorlagenbefehleWeisenBedienfehlerAb(t *testing.T) {
	ort, _ := eingefahren(t)

	for name, argumente := range map[string][]string{
		"--format ohne Wert":    {"turnier", "neu", "Freitag", "magic", "--format"},
		"kopieren ohne Namen":   {"turnier", "kopieren", "1"},
		"kopieren ohne Kennung": {"turnier", "kopieren"},
	} {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, argumente...); code != CodeBedienung {
				t.Errorf("%v gab %d, erwartet %d", argumente, code, CodeBedienung)
			}
		})
	}
}
