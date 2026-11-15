package mcp_test

// **Bereiche und Funktionen abschalten, am Werkzeugserver** — OPS-10
// (#1908).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Ein abgeschaltetes Werkzeug läuft nicht** — es antwortet mit der
// Abweisung samt Schalter, und ein anderes läuft weiter.
func TestEinAbgeschaltetesWerkzeugLaeuftNicht(t *testing.T) {
	t.Parallel()

	gerufen := map[string]int{}
	werkzeug := func(name string) mcp.Werkzeug {
		return mcp.Werkzeug{
			Name: name, Art: mcp.Lesend, Erklaerung: "zählt",
			Schema: map[string]any{"type": "object"},
			Ruf: func(context.Context, json.RawMessage) (string, error) {
				gerufen[name]++

				return "gut", nil
			},
		}
	}

	var gefragt []string

	server := mcp.NeuerServer([]mcp.Werkzeug{werkzeug("lesen.partien"), werkzeug("lesen.sets")}).
		MitSchaltern(func(_ context.Context, name string) error {
			gefragt = append(gefragt, name)
			if name == "lesen.partien" {
				return &domain.Abschaltung{Schalter: "spielen"}
			}

			return nil
		})

	for _, name := range []string{"lesen.partien", "lesen.sets"} {
		antwort := antworten(t, server,
			`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"`+name+`","arguments":{}}}`)

		abgewiesen := len(antwort) == 1 && strings.Contains(antwort[0], "abgeschaltet") &&
			strings.Contains(antwort[0], "spielen") && strings.Contains(antwort[0], `"isError":true`)
		if abgewiesen != (name == "lesen.partien") {
			t.Errorf("%s: %v", name, antwort)
		}
	}

	if gerufen["lesen.partien"] != 0 || gerufen["lesen.sets"] != 1 || len(gefragt) != 2 {
		t.Errorf("gerufen %v, gefragt %v", gerufen, gefragt)
	}
}

type schalterprobe struct {
	stand  []mcp.Schalter
	fehler error
}

func (s schalterprobe) Schalter(context.Context) ([]mcp.Schalter, error) {
	return s.stand, s.fehler
}

// **Die Schalter sind ein lesendes Werkzeug**, und ohne Naht gibt es keines.
func TestDieSchalterSindEinLesendesWerkzeug(t *testing.T) {
	t.Parallel()

	finde := func(naht mcp.Schalterauskunft) (mcp.Werkzeug, bool) {
		for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Schalter: naht}) {
			if werkzeug.Name == "lesen.schalter" {
				return werkzeug, true
			}
		}

		return mcp.Werkzeug{}, false
	}

	werkzeug, da := finde(schalterprobe{stand: []mcp.Schalter{{Kennung: "ki", Art: "funktion", An: false}}})
	if !da || werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Fatalf("%+v", werkzeug)
	}

	if text, fehler := werkzeug.Ruf(t.Context(), nil); fehler != nil || !strings.Contains(text, `"kennung": "ki"`) {
		t.Errorf("Ruf: %q, %v", text, fehler)
	}

	kaputt := errors.New("kein Bestand")
	werkzeug, _ = finde(schalterprobe{fehler: kaputt})

	if _, fehler := werkzeug.Ruf(t.Context(), nil); !errors.Is(fehler, kaputt) {
		t.Errorf("der Fehler kam nicht durch: %v", fehler)
	}

	if _, da := finde(nil); da {
		t.Error("ohne Naht gibt es das Werkzeug nicht")
	}
}
