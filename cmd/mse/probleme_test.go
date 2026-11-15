package main

// **`probleme` prüft ein Set und nennt den Stand** — PERF02
// (#2352).

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestProblemePrueftUndNenntDenStand(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "probleme", setID, "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var stand struct {
		Urteil   string            `json:"urteil"`
		Sicher   bool              `json:"sicher"`
		Geprueft int               `json:"geprueft"`
		Gesamt   int               `json:"gesamt"`
		Karten   []json.RawMessage `json:"karten"`
	}

	if err := json.Unmarshal([]byte(aus), &stand); err != nil {
		t.Fatalf("JSON lesen: %v (%s)", err, aus)
	}

	if stand.Urteil != "aktuell" || !stand.Sicher || stand.Gesamt == 0 || stand.Geprueft != stand.Gesamt || stand.Karten == nil {
		t.Errorf("Stand %+v", stand)
	}

	// **Ohne `--json` ein Satz**, und der Fortschritt auf dem Fehlerkanal.
	code, aus, fehler = fahreAn(t, ort, "probleme", setID)
	if code != CodeGut || !strings.Contains(aus, "geprüft") || !strings.Contains(fehler, "100 %") {
		t.Errorf("ohne --json: %d, %q, %q", code, aus, fehler)
	}
}

func TestProblemeBrauchtGenauEinSet(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, _ := fahreAn(t, ort, "probleme"); code != CodeBedienung {
		t.Errorf("ohne Set: %d", code)
	}

	if code, _, _ := fahreAn(t, ort, "probleme", "keins"); code != CodeBedienung {
		t.Errorf("unlesbare Kennung: %d", code)
	}

	if code, _, _ := fahreAn(t, ort, "probleme", "999"); code != CodeFachlich {
		t.Errorf("unbekanntes Set: %d", code)
	}
}

// **Ohne `--json` sagt der Satz, wie weit die Zahl gilt** — Befund 16 des
// Zweitblicks über Phase 3: ein Set, das sich während der Prüfung änderte,
// stand im Text als „3 mit Problemen" da, uneingeschränkt.
func TestDerSatzNenntDasUrteil(t *testing.T) {
	t.Parallel()

	for urteil, erwartet := range map[string]string{
		"aktuell":        "3 mit Problemen.",
		"veraltet":       "veraltet",
		"unvollstaendig": "unvollständig",
		"ungeprueft":     "ungeprüft",
	} {
		var raus, rfehler strings.Builder

		an := &schreiber{ziel: &raus, fehler: &rfehler}
		stand := app.ProblemstandDTO{Urteil: urteil, Zahl: 3, Sicher: urteil == "aktuell", Geprueft: 4, Gesamt: 4}

		if code := problemeZeigen(an, stand); code != CodeGut || !strings.Contains(raus.String(), erwartet) {
			t.Errorf("%s: %d, %q — erwartet %q", urteil, code, raus.String(), erwartet)
		}

		if urteil != "aktuell" && strings.Contains(raus.String(), "3 mit Problemen.") {
			t.Errorf("%s: %q nennt die Zahl ohne Vorbehalt", urteil, raus.String())
		}
	}
}
