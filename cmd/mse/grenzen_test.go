package main

import (
	"strings"
	"testing"
)

// **Die Grenzen dieses Prozesses** — Werkstück O5 von Phase 5.2 (#1429).

func TestDieGrenzenNennenKostenPlaetzeUndGuthaben(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "grenzen")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	for _, erwartet := range []string{
		"rendern\t2", "einfuhr\t1", "lesen\t1", "masse\t10", "principal\t120", "Reserve 120", "antwort\t",
		"kontingente\tan",
		// **Auftrags- und Verbindungsgrenze und die Kostentafel** (#2003).
		"verbindungen\t16", "auftraege\t4 laufend, 16 behalten, 3600 s",
		"\tKartenbilder schreiben\trendern\t5\tmse zeichnen · schreiben.karte.zeichnen · POST /karten/{}/bild",
		"\tEinen laufenden Indexaufbau abbrechen\tabbruch\t0\tDELETE /index/aufbau",
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("die Ausgabe nennt %q nicht:\n%s", erwartet, aus)
		}
	}

	code, aus, _ = fahreAn(t, ort, "grenzen", "--json")
	if code != CodeGut || !strings.Contains(aus, `"principal"`) || !strings.Contains(aus, `"kosten"`) ||
		!strings.Contains(aus, `"verbindungen": 16`) || !strings.Contains(aus, `"kostentafel"`) {
		t.Errorf("als JSON: %d %s", code, aus)
	}

	if code, _, _ := fahreAn(t, ort, "grenzen", "mehr"); code != CodeBedienung {
		t.Errorf("mit Argument: %d", code)
	}
}

// **Der Schalter aus dem Bestand** — QA-08.AK-03 (#1739): die Kommandozeile
// nennt, was ein jetzt gestarteter Dienst über diesem Bestand anwendete.
func TestDieGrenzenNennenAbgeschalteteKontingente(t *testing.T) {
	ort, _, _ := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "einstellungen", "setzen", "grenzen.kontingente", "aus"); code != CodeGut {
		t.Fatalf("setzen: %d (%s)", code, fehler)
	}

	code, aus, fehler := fahreAn(t, ort, "grenzen")
	if code != CodeGut || !strings.Contains(aus, "kontingente\taus") || !strings.Contains(aus, "principal\t120") {
		t.Errorf("Text: %d %s %s", code, aus, fehler)
	}

	code, aus, _ = fahreAn(t, ort, "grenzen", "--json")
	if code != CodeGut || !strings.Contains(aus, `"kontingente": "aus"`) {
		t.Errorf("als JSON: %d %s", code, aus)
	}
}
