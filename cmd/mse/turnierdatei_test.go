package main

// **Turnierdatei und Bericht an der Kommandozeile** — DOM-04.AK-03
// (#1699).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEinTurnierGehtAnDerKommandozeileHinUndZurueck(t *testing.T) {
	ort, _ := eingefahren(t)

	turnierID := strings.TrimSpace(mussLaufen(t, ort, "turnier", "neu", "Freitag", "magic"))

	// **Mit einer Liste, wie die Leitung sie abzieht** — ein echter Druck
	// steht dort ohne Bezeichnung; der Rundlauf darf daran nicht scheitern
	// (Befund 1 des Zweitblicks über Phase 6.6).
	deckID := strings.TrimSpace(mussLaufen(t, ort, "deck", "neu", "Aggro", "magic"))
	mussLaufen(t, ort, "deck", "legen", deckID, "a1b2c3d4-0000-4000-8000-00000000000a", "4")
	mussLaufen(t, ort, "turnier", "anmelden", turnierID, "Anna", deckID)

	// **Ohne Ziel auf die Ausgabe** — zum Weiterreichen in eine Leitung.
	if aus := mussLaufen(t, ort, "turnier", "ausgeben", turnierID); !strings.Contains(aus, `"name": "Freitag"`) {
		t.Errorf("die Ausgabe: %q", aus)
	}

	datei := filepath.Join(t.TempDir(), "freitag.mseturnier")
	mussLaufen(t, ort, "turnier", "ausgeben", turnierID, datei)

	// **Derselbe Name ein zweites Mal ist ein Konflikt** — fachlich.
	if code, _, _ := fahreAn(t, ort, "turnier", "einlesen", datei); code != CodeFachlich {
		t.Errorf("zweimal eingelesen: %d, erwartet %d", code, CodeFachlich)
	}

	roh, err := os.ReadFile(datei) //nolint:gosec // ein Pfad dieses Laufs
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	umbenannt := strings.Replace(string(roh), `"Freitag"`, `"Samstag"`, 1)
	umbenannt = strings.Replace(umbenannt, `"fassung": 1,`, `"fassung": 1, "zukunft": 1,`, 1)

	if err := os.WriteFile(datei, []byte(umbenannt), 0o600); err != nil { //nolint:gosec // ein Pfad dieses Laufs
		t.Fatalf("schreiben: %v", err)
	}

	// **Ein fremdes Feld ist ein Befund** und wird genannt (§8).
	if aus := mussLaufen(t, ort, "turnier", "einlesen", datei); !strings.Contains(aus, "eingelesen") ||
		!strings.Contains(aus, "Befund: Das Feld »zukunft«") {
		t.Errorf("das Einlesen: %q", aus)
	}

	if samstag := mussLaufen(t, ort, "turniere"); !strings.Contains(samstag, "Samstag") {
		t.Errorf("das eingelesene Turnier fehlt: %q", samstag)
	}

	if !strings.Contains(string(roh), `"printing": "a1b2c3d4-0000-4000-8000-00000000000a"`) {
		t.Errorf("die Liste reist ohne ihren Druck: %s", roh)
	}

	if bericht := mussLaufen(t, ort, "turnier", "bericht", turnierID); !strings.HasPrefix(bericht, "# Freitag\n") ||
		!strings.Contains(bericht, "## Rangliste") {
		t.Errorf("der Bericht: %q", bericht)
	}

	berichtsdatei := filepath.Join(t.TempDir(), "freitag.md")
	mussLaufen(t, ort, "turnier", "bericht", turnierID, berichtsdatei)

	if inhalt, err := os.ReadFile(berichtsdatei); err != nil || !strings.HasPrefix(string(inhalt), "# Freitag") { //nolint:gosec // dito
		t.Errorf("die Berichtsdatei: %q (%v)", inhalt, err)
	}

	if roh := mussLaufen(t, ort, "turnier", "bericht", turnierID, "--json"); !strings.Contains(roh, `"inhalt"`) {
		t.Errorf("die JSON-Gestalt: %q", roh)
	}

	for _, befehl := range []string{"ausgeben", "einlesen", "bericht"} {
		if code, _, _ := fahreAn(t, ort, "turnier", befehl); code != CodeBedienung {
			t.Errorf("%s ohne Angabe: %d, erwartet %d", befehl, code, CodeBedienung)
		}
	}
}
