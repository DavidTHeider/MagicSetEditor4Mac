package mcp_test

// **Auszählen nach einer erklärten Dimension, über MCP** — Werkstück C von
// Phase 4.11 (#1338).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// zaehlprobe ist ein Doppel für `mcp.Auszaehlung`.
type zaehlprobe struct {
	angebot  any
	gezaehlt any
	fehler   error

	gefragt  int64
	gewaehlt string
}

func (z *zaehlprobe) Angebot(_ context.Context, setID int64) (any, error) {
	z.gefragt = setID

	return z.angebot, z.fehler
}

func (z *zaehlprobe) Auszaehlen(_ context.Context, setID int64, name string) (any, error) {
	z.gefragt = setID
	z.gewaehlt = name

	return z.gezaehlt, z.fehler
}

// rufeZaehlwerkzeug sucht ein Werkzeug und ruft es.
func rufeZaehlwerkzeug(
	t *testing.T, probe *zaehlprobe, name, argumente string,
) (string, error) {
	t.Helper()

	werkzeuge := mcp.Werkzeuge(mcp.Naehte{Auszaehlung: probe})

	for _, werkzeug := range werkzeuge {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return "", nil
}

// **Das Angebot kommt durch.**
func TestDasAuszaehlangebotKommtUeberMCP(t *testing.T) {
	t.Parallel()

	probe := &zaehlprobe{angebot: []map[string]any{{"name": "farbe"}}}

	roh, fehler := rufeZaehlwerkzeug(t, probe,
		"lesen.auszaehldimensionen", `{"setID":7}`)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if probe.gefragt != 7 || !strings.Contains(roh, "farbe") {
		t.Errorf("gefragt wurde nach %d, zurück kam %s", probe.gefragt, roh)
	}
}

// **Und die Auszählung — mit ihrem Namen als Zeichenkette.**
//
// **Der Punkt ist die Zeichenkette.** `kennung` deutet alle Argumente als
// Zahlen; `name` ginge daran zugrunde, und der Fehler hieße „invalid number
// literal". Dieselbe Falle wie bei `schreiben.kartenskript`.
func TestDieAuszaehlungKommtUeberMCP(t *testing.T) {
	t.Parallel()

	probe := &zaehlprobe{gezaehlt: map[string]any{"dimension": "farbe"}}

	roh, fehler := rufeZaehlwerkzeug(t, probe,
		"lesen.auszaehlung", `{"setID":7,"name":"farbe"}`)
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if probe.gewaehlt != "farbe" || !strings.Contains(roh, "farbe") {
		t.Errorf("gewählt war %q, zurück kam %s", probe.gewaehlt, roh)
	}
}

// **Eine Kennung, die keine ist, wird abgewiesen — an beiden Werkzeugen.**
func TestDieAuszaehlwerkzeugePruefenIhreKennung(t *testing.T) {
	t.Parallel()

	probe := &zaehlprobe{}

	if _, fehler := rufeZaehlwerkzeug(t, probe,
		"lesen.auszaehldimensionen", `{"setID":0}`); fehler == nil {
		t.Error("die Null ging als Kennung durch")
	}

	if _, fehler := rufeZaehlwerkzeug(t, probe,
		"lesen.auszaehlung", `{"setID":0,"name":"farbe"}`); fehler == nil {
		t.Error("dieselbe beim Auszählen")
	}
}

// **Kaputte Argumente sind ein Fehler und kein leeres Ergebnis.**
func TestKaputteArgumenteAnDenAuszaehlwerkzeugen(t *testing.T) {
	t.Parallel()

	probe := &zaehlprobe{}

	if _, fehler := rufeZaehlwerkzeug(t, probe,
		"lesen.auszaehldimensionen", `{`); fehler == nil {
		t.Error("die kaputten Argumente gingen durch")
	}

	if _, fehler := rufeZaehlwerkzeug(t, probe,
		"lesen.auszaehlung", `{`); fehler == nil {
		t.Error("dieselben beim Auszählen")
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDerAuszaehlnahtKommtUeberMCP(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Spielpaket")
	probe := &zaehlprobe{fehler: eigener}

	if _, fehler := rufeZaehlwerkzeug(t, probe,
		"lesen.auszaehldimensionen", `{"setID":7}`); !errors.Is(fehler, eigener) {
		t.Errorf("der Fehler lautet %v", fehler)
	}

	if _, fehler := rufeZaehlwerkzeug(t, probe,
		"lesen.auszaehlung", `{"setID":7,"name":"farbe"}`); !errors.Is(fehler, eigener) {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
//
// §16: eine Fähigkeit, die nicht eingerichtet ist, steht auch nicht im Menü —
// ein Werkzeug, das immer scheitert, wäre schlimmer als keines.
func TestOhneNahtGibtEsKeineAuszaehlwerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.HasPrefix(werkzeug.Name, "lesen.auszaehl") {
			t.Errorf("ohne Naht steht %q im Menü", werkzeug.Name)
		}
	}
}
