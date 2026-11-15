package mcp_test

// **Der Preisanhalt über MCP** — TRADE-02b
// (#1951).

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func (h *handelsprobe) Handelsanhalt(_ context.Context, vorgang string) (any, error) {
	h.vorgang = vorgang

	if h.fehler != nil {
		return nil, h.fehler
	}

	return map[string]any{"hauptwaehrung": "EUR", "differenzen": []any{}}, nil
}

func TestDasAnhaltswerkzeugNenntDieHauptwaehrung(t *testing.T) {
	t.Parallel()

	probe := &handelsprobe{}

	antwort, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsanhalt", `{"vorgang":"v-9"}`)
	if fehler != nil || probe.vorgang != "v-9" || !strings.Contains(antwort, "EUR") {
		t.Fatalf("Antwort %s, %v", antwort, fehler)
	}

	probe.fehler = errors.New("Bestand zu")
	if _, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsanhalt", `{"vorgang":"v-9"}`); fehler == nil {
		t.Error("der Fehler ging verloren")
	}

	if _, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsanhalt", `{"vorgang":7}`); fehler == nil {
		t.Error("eine Zahl als Vorgang ging durch")
	}
}
