package mcp_test

// **Die Parameterhilfen über MCP** — Werkstück D von Phase 4.11
// (#1342).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// hilfenprobe ist ein Doppel für `mcp.Parameterhilfen`.
type hilfenprobe struct {
	antwort any
	fehler  error

	setID, wortID int64
}

func (h *hilfenprobe) Parameterhilfen(_ context.Context, setID, wortID int64) (
	any, error,
) {
	h.setID, h.wortID = setID, wortID

	return h.antwort, h.fehler
}

// rufeHilfenwerkzeug sucht das Werkzeug und ruft es.
func rufeHilfenwerkzeug(t *testing.T, probe *hilfenprobe, argumente string) (
	string, error,
) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Parameterhilfen: probe}) {
		if werkzeug.Name == "lesen.parameterhilfen" {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatal("das Werkzeug lesen.parameterhilfen gibt es nicht")

	return "", nil
}

// **Die Hilfen kommen durch — mit beiden Kennungen.**
func TestDieParameterhilfenKommenUeberMCP(t *testing.T) {
	t.Parallel()

	probe := &hilfenprobe{antwort: map[string]any{"arten": []string{"zahl"}}}

	roh, fehler := rufeHilfenwerkzeug(t, probe, `{"setID":7,"wortID":3}`)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if probe.setID != 7 || probe.wortID != 3 || !strings.Contains(roh, "zahl") {
		t.Errorf("gefragt wurde nach %d/%d, zurück kam %s",
			probe.setID, probe.wortID, roh)
	}
}

// **Beide Kennungen werden geprüft.**
func TestDasParameterwerkzeugPrueftBeideKennungen(t *testing.T) {
	t.Parallel()

	probe := &hilfenprobe{}

	if _, fehler := rufeHilfenwerkzeug(
		t, probe, `{"setID":0,"wortID":3}`); fehler == nil {
		t.Error("die Setkennung ging durch")
	}

	if _, fehler := rufeHilfenwerkzeug(
		t, probe, `{"setID":7,"wortID":0}`); fehler == nil {
		t.Error("die Wortkennung ging durch")
	}
}

// **Kaputte Argumente sind ein Fehler.**
func TestKaputteArgumenteAmParameterwerkzeug(t *testing.T) {
	t.Parallel()

	if _, fehler := rufeHilfenwerkzeug(t, &hilfenprobe{}, `{`); fehler == nil {
		t.Error("die kaputten Argumente gingen durch")
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDerParameternahtKommtUeberMCP(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein solches Wort")

	if _, fehler := rufeHilfenwerkzeug(t, &hilfenprobe{fehler: eigener},
		`{"setID":7,"wortID":3}`); !errors.Is(fehler, eigener) {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ohne Naht gibt es das Werkzeug nicht.**
func TestOhneNahtGibtEsKeinParameterwerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.parameterhilfen" {
			t.Error("ohne Naht steht das Werkzeug im Menü")
		}
	}
}
