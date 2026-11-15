package mcp_test

// **Ein Set vorab ansehen über den Werkzeugserver** — FMT-03.AK-01 (#1692).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type vorschauprobe struct{ gefragt string }

func (v *vorschauprobe) SetVorschau(_ context.Context, pfad string) (any, error) {
	v.gefragt = pfad

	if pfad == "/kein.mse-game" {
		return nil, domain.ErrValidation
	}

	return map[string]any{"name": "Mein Probeset", "karten": 3}, nil
}

func rufeVorschau(t *testing.T, naht mcp.Setvorschau, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Setvorschau: naht}) {
		if werkzeug.Name == "lesen.set.vorschau" {
			if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
				t.Errorf("%+v", werkzeug)
			}

			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatal("das Werkzeug steht nicht in der Liste")

	return "", nil
}

func TestDerWerkzeugserverZeigtEineSetvorschau(t *testing.T) {
	t.Parallel()

	probe := &vorschauprobe{}

	aus, fehler := rufeVorschau(t, probe, `{"pfad":"/tmp/probe.mse-set"}`)
	if fehler != nil || probe.gefragt != "/tmp/probe.mse-set" || !strings.Contains(aus, "Mein Probeset") {
		t.Errorf("%v %q", fehler, aus)
	}

	if _, fehler := rufeVorschau(t, probe, `{"pfad":"/kein.mse-game"}`); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("kein Set: %v", fehler)
	}

	for _, argumente := range []string{`{}`, `kaputt`} {
		if _, fehler := rufeVorschau(t, probe, argumente); fehler == nil {
			t.Errorf("%s: kein Fehler", argumente)
		}
	}
}

func TestOhneNahtGibtEsKeineSetvorschau(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.set.vorschau" {
			t.Error("das Werkzeug steht ohne Naht in der Liste")
		}
	}
}
