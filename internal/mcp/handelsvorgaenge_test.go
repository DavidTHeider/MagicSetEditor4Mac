package mcp_test

// **Die Handel mit einem Gegenüber, über MCP** — UX-08.AK-04
// (#1680).
//
// **Lesend wie `lesen.handel`**: welche Handel es mit jemandem gibt, ist
// Auskunft über den eigenen Bestand, und sie verlässt den Rechner nicht.

import (
	"errors"
	"strings"
	"testing"
)

func TestDieHandelsvorgaengeKommenUeberMCP(t *testing.T) {
	t.Parallel()

	probe := &handelsprobe{}

	antwort, fehler := rufeHandelswerkzeug(
		t, probe, "lesen.handelsvorgaenge", `{"gegenueber":"bdu"}`)
	if fehler != nil {
		t.Fatalf("lesen.handelsvorgaenge: %v", fehler)
	}

	if probe.gegenueber != "bdu" {
		t.Errorf("durchgereicht wurde %q", probe.gegenueber)
	}

	for _, wort := range []string{`"vorgaenge"`, "h-7", "angeboten"} {
		if !strings.Contains(antwort, wort) {
			t.Errorf("in der Antwort fehlt %q: %s", wort, antwort)
		}
	}
}

// **Ein krummes Argument und eine klagende Naht enden vor der Antwort.**
func TestDieHandelsvorgaengeReichenIhreAbsagenDurch(t *testing.T) {
	t.Parallel()

	probe := &handelsprobe{}

	if _, fehler := rufeHandelswerkzeug(
		t, probe, "lesen.handelsvorgaenge", `{`); fehler == nil {
		t.Error("ein krummes Argument kam durch")
	}

	probe.fehler = errors.New("die Platte klemmt")

	if _, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsvorgaenge",
		`{"gegenueber":"bdu"}`); !errors.Is(fehler, probe.fehler) {
		t.Errorf("die Klage kam als %v", fehler)
	}
}
