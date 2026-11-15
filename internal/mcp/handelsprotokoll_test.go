package mcp_test

// **Das Protokoll über MCP** — TRADE-03.AK-03
// (#1717).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func (h *handelsprobe) Handelsprotokoll(_ context.Context, vorgang string) (
	[]mcp.Handelsprotokollzeile, error,
) {
	h.vorgang = vorgang

	if h.fehler != nil {
		return nil, h.fehler
	}

	return []mcp.Handelsprotokollzeile{
		{Revision: 1, Nach: "entwurf", Zeit: "2026-09-30T12:00:00Z"},
		{Revision: 1, Von: "entwurf", Nach: "angeboten", Zeit: "2026-09-30T12:01:00Z"},
	}, nil
}

func TestDasProtokollwerkzeugNenntDieWechsel(t *testing.T) {
	t.Parallel()

	probe := &handelsprobe{}

	antwort, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsprotokoll", `{"vorgang":"v-9"}`)
	if fehler != nil || probe.vorgang != "v-9" ||
		!strings.Contains(antwort, "angeboten") || !strings.Contains(antwort, "2026-09-30T12:01:00Z") {
		t.Fatalf("Antwort %s, %v", antwort, fehler)
	}

	probe.fehler = errors.New("Bestand zu")
	if _, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsprotokoll", `{"vorgang":"v-9"}`); fehler == nil {
		t.Error("der Fehler ging verloren")
	}

	if _, fehler := rufeHandelswerkzeug(t, probe, "lesen.handelsprotokoll", `{"vorgang":7}`); fehler == nil {
		t.Error("eine Zahl als Vorgang ging durch")
	}
}
