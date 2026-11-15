package mcp_test

// **Die Liste der Handelsvorgänge über MCP** — TRADE-01
// (#1715):
// lesend wie `lesen.handel`, und der Filter kommt unverändert an.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestDieUebersichtKommtUeberMCP(t *testing.T) {
	t.Parallel()

	probe := &handelsprobe{}

	antwort, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsuebersicht",
		`{"lagen":["dran"],"suche":"bolt","ordnung":"lage","archiv":true,"seit":"2026-09-01"}`)
	if fehler != nil {
		t.Fatalf("lesen.handelsuebersicht: %v", fehler)
	}

	if !slices.Equal(probe.filter.Lagen, []string{"dran"}) || probe.filter.Suche != "bolt" ||
		probe.filter.Ordnung != "lage" || !probe.filter.Archiv || probe.filter.Seit != "2026-09-01" {
		t.Errorf("durchgereicht wurde %+v", probe.filter)
	}

	for _, wort := range []string{`"zeilen"`, "h-8", "Turfer", `"lage": "dran"`} {
		if !strings.Contains(antwort, wort) {
			t.Errorf("in der Antwort fehlt %q: %s", wort, antwort)
		}
	}
}

func TestDieUebersichtReichtIhreAbsagenDurch(t *testing.T) {
	t.Parallel()

	if _, fehler := rufeHandelswerkzeug(t, &handelsprobe{fehler: errors.New("klemmt")},
		"lesen.handelsuebersicht", `{}`); fehler == nil {
		t.Error("die Absage ging verloren")
	}

	if _, fehler := rufeHandelswerkzeug(t, &handelsprobe{},
		"lesen.handelsuebersicht", `{"lagen":"dran"}`); fehler == nil {
		t.Error("ein krummer Filter ging durch")
	}
}
