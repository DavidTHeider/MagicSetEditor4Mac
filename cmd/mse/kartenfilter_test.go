// **Der Schnellfilter an der Zeile** — Werkstück U5 von Phase 5.1
// (#1416).
package main

import (
	"strings"
	"testing"
)

func TestKartenfilterGibtDieKennungenDerTreffer(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	if code, aus, fehler := fahreAn(t, ort, "feld", setID, karteID, "name", "Grizzly Bears"); code != CodeGut {
		t.Fatalf("feld: %d %s %s", code, aus, fehler)
	}

	code, aus, fehler := fahreAn(t, ort, "kartenfilter", setID, "name:bear")
	if code != CodeGut {
		t.Fatalf("kartenfilter: %d %s %s", code, aus, fehler)
	}

	if strings.TrimSpace(aus) != karteID {
		t.Errorf("die Zeile nennt %q, erwartet %q", strings.TrimSpace(aus), karteID)
	}

	// **Zwei Wörter sind eine Anfrage**, und `--json` gibt die Kennungen als Liste.
	code, aus, _ = fahreAn(t, ort, "kartenfilter", setID, "-goblin", "bears", "--json")
	if code != CodeGut || !strings.Contains(aus, `"kartenIDs"`) || !strings.Contains(aus, `"`+karteID+`"`) {
		t.Errorf("mit --json: %d %s", code, aus)
	}

	code, aus, _ = fahreAn(t, ort, "kartenfilter", setID, "dragon")
	if code != CodeGut || strings.TrimSpace(aus) != "" {
		t.Errorf("ohne Treffer: %d %q, erwartet eine leere Ausgabe", code, aus)
	}
}

func TestKartenfilterWeistFalscheAufrufeAb(t *testing.T) {
	ort, setID, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "kartenfilter", setID); code != CodeBedienung ||
		!strings.Contains(fehler, "Anfrage") {
		t.Errorf("ohne Anfrage: %d %s", code, fehler)
	}

	if code, _, _ := fahreAn(t, ort, "kartenfilter", "x", "bear"); code != CodeBedienung {
		t.Errorf("mit unlesbarer Kennung: %d", code)
	}

	if code, _, _ := fahreAn(t, ort, "kartenfilter", "4711", "bear"); code != CodeFachlich {
		t.Errorf("mit unbekanntem Set: %d", code)
	}
}
