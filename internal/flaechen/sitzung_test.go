package flaechen_test

// **Die Sitzung an der Kommandozeile** — ID-02
// (#1701).
//
// **Nur hier und am Fenster**: MCP und Dienst bieten Kennwort und Sitzung
// nicht an, mit Grund in `docs/FAEHIGKEITEN.md`. Das Fenster fährt denselben
// Dienst über `internal/app/sitzung_test.go` und die Hüllenprobe `id02`.

import (
	"strings"
	"testing"
)

func TestDieSitzungAnDerKommandozeile(t *testing.T) {
	umgebung := eigenerBestand(t)

	if aus := mussKommandozeile(t, umgebung, "sitzung"); !strings.HasPrefix(aus, "angemeldet\t") {
		t.Fatalf("mse sitzung: %q", aus)
	}

	if aus := mitEingabe(t, umgebung, "\n"+probekennwort, "kennwort-wechseln"); aus != "gesetzt\n" {
		t.Fatalf("mse kennwort-wechseln: %q", aus)
	}

	// **Jetzt verlangt der Wechsel das bisherige** — und ein falsches sagt es.
	if _, fehler, code := mitEingabeUndCode(t, umgebung, "falsch\nnoch-ein-langes-kennwort", "kennwort-wechseln"); code == 0 ||
		!strings.Contains(fehler, "Kennwort stimmt nicht") {
		t.Errorf("ohne das bisherige: %d %q", code, fehler)
	}

	// **Angemeldet gibt es nichts wiederherzustellen** — ein Zettel ändert
	// an einer offenen Sitzung nichts.
	if _, fehler, code := mitEingabeUndCode(t, umgebung, "code\nnoch-ein-langes-kennwort", "kennwort-wiederherstellen", "1"); code == 0 {
		t.Errorf("mse kennwort-wiederherstellen angemeldet: %q", fehler)
	}
}
