package mcp_test

// **Die Verträglichkeit über MCP** — Werkstück O6 von Phase 5.2
// (#1430).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type vertraeglichkeitsprobe struct {
	antwort mcp.Vertraeglichkeit
	fehler  error
}

func (v vertraeglichkeitsprobe) Vertraeglichkeit(
	context.Context,
) (mcp.Vertraeglichkeit, error) {
	return v.antwort, v.fehler
}

func TestOhneNahtGibtEsKeinVertraeglichkeitswerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.vertraeglichkeit" {
			t.Fatal("das Werkzeug stand ohne Naht da")
		}
	}
}

func TestDasVertraeglichkeitswerkzeugNenntUrteilUndHindernisse(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Vertraeglichkeit: vertraeglichkeitsprobe{antwort: mcp.Vertraeglichkeit{
			Fassung: "0.6.0", Urteil: "unvertraeglich",
			Hindernisse: []mcp.Hindernis{
				{Art: "wanderungFehlt", Stelle: "0052_deckzuordnungen.sql"},
			},
		}},
	}, "lesen.vertraeglichkeit")

	antwort, err := werkzeug.Ruf(t.Context(), nil)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	for _, teil := range []string{`"urteil": "unvertraeglich"`, `"0052_deckzuordnungen.sql"`} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("Antwort %q nennt %q nicht", antwort, teil)
		}
	}
}

func TestDasVertraeglichkeitswerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Vertraeglichkeit: vertraeglichkeitsprobe{fehler: errors.New("kein Netz")},
	}, "lesen.vertraeglichkeit")

	if _, err := werkzeug.Ruf(t.Context(), nil); err == nil {
		t.Error("der Fehler ging verloren")
	}
}
