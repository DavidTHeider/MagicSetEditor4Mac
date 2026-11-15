package mcp_test

// **Das Metrikwerkzeug und der Griff ins Protokoll** — Werkstück O9 von
// Phase 5.2 (#1433).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// metrikprobe antwortet, wie ihr gesagt wird.
//
// **Eine beliebige Gestalt** — seit OPS-05 (#1722) reicht das Werkzeug die
// Antwort der Anwendung durch, wie der Dienst es tut, statt eine zweite
// Gestalt daneben zu pflegen, die hinter ihr zurückbliebe.
type metrikprobe struct {
	antwort any
	fehler  error
}

func (m metrikprobe) Metriken(context.Context) (any, error) {
	return m.antwort, m.fehler
}

// **Ohne Naht gibt es das Werkzeug nicht.**
func TestOhneNahtGibtEsKeinMetrikwerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.metriken" {
			t.Fatal("das Werkzeug stand ohne Naht da")
		}
	}
}

func TestDasMetrikwerkzeugNenntReihenUndGrenze(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Metriken: metrikprobe{antwort: map[string]any{
			"grenze": 40,
			"reihen": []map[string]any{{
				"flaeche": "mcp", "wirkung": "externSenden",
				"ausgang": "abgewiesen", "anzahl": 2,
			}},
		}},
	}, "lesen.metriken")

	antwort, err := werkzeug.Ruf(t.Context(), nil)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	switch {
	case !strings.Contains(antwort, `"grenze": 40`):
		t.Errorf("die Grenze fehlt: %q", antwort)
	case !strings.Contains(antwort, `"externSenden"`):
		t.Errorf("die Reihe fehlt: %q", antwort)
	}
}

func TestDasMetrikwerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Metriken: metrikprobe{fehler: errors.New("nichts zu zählen")},
	}, "lesen.metriken")

	if _, err := werkzeug.Ruf(t.Context(), nil); err == nil {
		t.Error("der Fehler ging verloren")
	}
}
