package mcp_test

// **Eine Datei ansehen über den Werkzeugserver** — UX-13 (#1685).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type ansichtsprobe struct{ gefragt string }

func (a *ansichtsprobe) DateiAnsehen(_ context.Context, pfad string) (any, error) {
	a.gefragt = pfad

	if pfad == "/fehlt" {
		return nil, domain.ErrNotFound
	}

	return map[string]any{"name": "liste.txt", "text": "4 Blitzschlag"}, nil
}

func rufeAnsicht(t *testing.T, naht mcp.Dateiansicht, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Dateien: naht}) {
		if werkzeug.Name == "lesen.datei.ansicht" {
			if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
				t.Errorf("%+v", werkzeug)
			}

			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatal("das Werkzeug steht nicht in der Liste")

	return "", nil
}

func TestDerWerkzeugserverSiehtEineDateiAn(t *testing.T) {
	t.Parallel()

	probe := &ansichtsprobe{}

	aus, fehler := rufeAnsicht(t, probe, `{"pfad":"/tmp/liste.txt"}`)
	if fehler != nil || probe.gefragt != "/tmp/liste.txt" || !strings.Contains(aus, "4 Blitzschlag") {
		t.Errorf("%v %q", fehler, aus)
	}

	if _, fehler := rufeAnsicht(t, probe, `{"pfad":"/fehlt"}`); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("fehlt: %v", fehler)
	}

	if _, fehler := rufeAnsicht(t, probe, `{}`); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Pfad: %v", fehler)
	}

	if _, fehler := rufeAnsicht(t, probe, `kein json`); fehler == nil {
		t.Error("kaputte Argumente ohne Fehler")
	}

	// **Ohne Naht gibt es das Werkzeug nicht.**
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.datei.ansicht" {
			t.Error("ohne Naht steht es trotzdem da")
		}
	}
}
