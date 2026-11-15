package mcp_test

// **Das anonyme Konto über MCP** — Werkstück 6 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"
)

// **Das Werkzeug legt einen Nutzer ohne jede Angabe an.**
func TestDasAnonymeWerkzeugLegtEinenNutzerAn(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{}

	antwort := rufeRoh(t, mitNutzern(t, probe, "schreiben.konto.anonym"), `{}`)

	if !probe.anonym {
		t.Error("die Naht wurde nicht gerufen")
	}

	if !strings.Contains(antwort, "anonym-abc") {
		t.Errorf("die Antwort ist %q", antwort)
	}
}

// **Und ein Fehler der Anwendung kommt als Fehler an** (§8).
//
// **Nicht als Text in einer geglückten Antwort**: ein Sprachmodell läse ihn
// dann als Ergebnis und arbeitete damit weiter.
func TestDasAnonymeWerkzeugReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	probe := &nutzerprobe{fehler: errors.New("kaputt")}

	if !rufeMitFehler(t, mitNutzern(t, probe, "schreiben.konto.anonym"), `{}`) {
		t.Error("der Fehler ging verloren")
	}
}
