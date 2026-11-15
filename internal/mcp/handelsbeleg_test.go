package mcp_test

// **Der Beleg über MCP** — wer unterschrieben hat, ohne einen Schritt
// (Hüllenprobe `nat04handel`,
// #1672).

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func (h *handelsprobe) Handelsbeleg(_ context.Context, vorgang string) (any, error) {
	h.vorgang = vorgang

	if h.fehler != nil {
		return nil, h.fehler
	}

	return map[string]any{
		"zustand": "beidseitigVerbucht", "hatBeleg": true,
		"beleg": map[string]any{"beidseitig": true, "gegenzeichner": "peer-bert"},
	}, nil
}

func TestDasBelegwerkzeugNenntDenGegenzeichner(t *testing.T) {
	t.Parallel()

	probe := &handelsprobe{}

	antwort, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsbeleg", `{"vorgang":"v-9"}`)
	if fehler != nil || probe.vorgang != "v-9" ||
		!strings.Contains(antwort, "peer-bert") || !strings.Contains(antwort, "beidseitigVerbucht") {
		t.Fatalf("Antwort %s, %v", antwort, fehler)
	}

	probe.fehler = errors.New("Bestand zu")
	if _, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsbeleg", `{"vorgang":"v-9"}`); fehler == nil {
		t.Error("der Fehler ging verloren")
	}

	if _, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsbeleg", `{"vorgang":7}`); fehler == nil {
		t.Error("eine Zahl als Vorgang ging durch")
	}
}
