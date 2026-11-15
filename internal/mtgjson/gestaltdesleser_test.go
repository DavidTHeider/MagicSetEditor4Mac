package mtgjson_test

// **Prüfung und Leser sehen dieselbe Gestalt** — Fund der Fuzzkampagne vom
// 2026-10-10 (C02, #2227).
//
// Der Leser dekodiert in eine Struktur, und dort gilt ein Schlüssel auch in
// anderer Schreibung: „Code" ist `code`, die Datei ein Set. Die Prüfung las
// eine Abbildung und unterschied die Schreibung — sie sah eine Abbildung von
// Sets und meldete, „Code" sei kein Set.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

func TestDiePruefungSiehtDieGestaltDesLesers(t *testing.T) {
	t.Parallel()

	zuordnung, da := mtgjson.ZuordnungFuer("magic")
	if !da {
		t.Fatal("die Zuordnung für `magic` fehlt")
	}

	roh := []byte(`{"Code":"X","Cards":[]}`)

	quelle, fehler := mtgjson.LiesInhalt(roh, "probe.json", zuordnung)
	if fehler != nil {
		t.Fatalf("LiesInhalt: %v", fehler)
	}

	if len(quelle.Sets) != 1 || quelle.Sets[0].Code != "X" {
		t.Fatalf("der Leser las %+v — erwartet ein Set X", quelle.Sets)
	}

	for _, befund := range quelle.Befunde {
		if strings.Contains(befund.Meldung, "ist kein Set") {
			t.Errorf("die Prüfung sah eine Abbildung von Sets, wo der Leser ein Set las: %q — %q", befund.Ort, befund.Meldung)
		}
	}
}
