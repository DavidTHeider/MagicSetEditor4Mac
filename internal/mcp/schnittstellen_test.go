package mcp_test

// **Die Schnittstellen als Werkzeug** — OPS-09 (#1907): rein lesend, und ohne
// Naht gibt es das Werkzeug nicht.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type schnittstellenprobe struct {
	stand  []mcp.Schnittstelle
	fehler error
}

func (s schnittstellenprobe) Schnittstellen(context.Context) ([]mcp.Schnittstelle, error) {
	return s.stand, s.fehler
}

func schnittstellenwerkzeug(t *testing.T, naht mcp.Schnittstellenauskunft) (mcp.Werkzeug, bool) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Schnittstellen: naht}) {
		if werkzeug.Name == "lesen.schnittstellen" {
			return werkzeug, true
		}
	}

	return mcp.Werkzeug{}, false
}

func TestDieSchnittstellenSindEinLesendesWerkzeug(t *testing.T) {
	t.Parallel()

	werkzeug, da := schnittstellenwerkzeug(t, schnittstellenprobe{stand: []mcp.Schnittstelle{
		{Name: "codex", Stand: "fremd", Datei: "/x/.codex/config.toml", Zeilen: []string{"# >>> MagicSetEditor4Mac"}},
	}})
	if !da || werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Fatalf("das Werkzeug: %+v", werkzeug)
	}

	text, fehler := werkzeug.Ruf(t.Context(), nil)
	if fehler != nil || !strings.Contains(text, `"stand": "fremd"`) {
		t.Errorf("Ruf: %q, %v", text, fehler)
	}

	kaputt := errors.New("kein Zuhause")
	werkzeug, _ = schnittstellenwerkzeug(t, schnittstellenprobe{fehler: kaputt})

	if _, fehler := werkzeug.Ruf(t.Context(), nil); !errors.Is(fehler, kaputt) {
		t.Errorf("der Fehler kam nicht durch: %v", fehler)
	}

	if _, da := schnittstellenwerkzeug(t, nil); da {
		t.Error("ohne Naht gibt es das Werkzeug nicht")
	}
}
