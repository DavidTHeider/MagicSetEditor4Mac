package service

// **Ein Eintrag in einem JSON-Objekt, ohne die Ordnung daneben zu
// verlieren** — OPS-09b
// (#1961).

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestEinEintragWirdGesetztUndDieOrdnungBleibt(t *testing.T) {
	t.Parallel()

	vorher := `{"zuerst": 1, "mcpServers": {"anderer": {"command": "x"}}, "zuletzt": [true]}`

	nachher, fehler := jsonEintragSetzen([]byte(vorher), "magicseteditor4mac",
		json.RawMessage(`{"command":"/a/mse-mcp"}`))
	if fehler != nil {
		t.Fatal(fehler)
	}

	text := string(nachher)

	geordnet := strings.Index(text, `"zuerst"`) < strings.Index(text, `"mcpServers"`) &&
		strings.Index(text, `"mcpServers"`) < strings.Index(text, `"zuletzt"`) &&
		strings.Index(text, `"anderer"`) < strings.Index(text, `"magicseteditor4mac"`)
	if !geordnet {
		t.Errorf("die Ordnung ging verloren:\n%s", text)
	}

	wert, da, fehler := jsonEintrag(nachher, "magicseteditor4mac")
	if fehler != nil || !da || !strings.Contains(string(wert), "/a/mse-mcp") {
		t.Errorf("gelesen: %s %v %v", wert, da, fehler)
	}

	// **Ein zweites Setzen ersetzt** und hängt nicht an.
	zweimal, fehler := jsonEintragSetzen(nachher, "magicseteditor4mac",
		json.RawMessage(`{"command":"/b/mse-mcp"}`))
	if fehler != nil || strings.Count(string(zweimal), "magicseteditor4mac") != 1 || !strings.Contains(string(zweimal), "/b/mse-mcp") {
		t.Errorf("zweimal:\n%s", zweimal)
	}

	ohne, fehler := jsonEintragEntfernen(zweimal, "magicseteditor4mac")
	if fehler != nil || strings.Contains(string(ohne), "magicseteditor4mac") || !strings.Contains(string(ohne), `"anderer"`) {
		t.Errorf("entfernt:\n%s %v", ohne, fehler)
	}
}

func TestEinLeeresDokumentBekommtSeinObjekt(t *testing.T) {
	t.Parallel()

	for _, vorher := range []string{"", "  \n", "{}"} {
		nachher, fehler := jsonEintragSetzen([]byte(vorher), "m", json.RawMessage(`{"command":"c"}`))
		if fehler != nil {
			t.Fatalf("%q: %v", vorher, fehler)
		}

		wert, da, fehler := jsonEintrag(nachher, "m")

		var kompakt bytes.Buffer
		if fehler != nil || !da || json.Compact(&kompakt, wert) != nil || kompakt.String() != `{"command":"c"}` {
			t.Errorf("%q → %s", vorher, nachher)
		}
	}

	if _, da, fehler := jsonEintrag(nil, "m"); da || fehler != nil {
		t.Errorf("leer: %v %v", da, fehler)
	}
}

func TestEinKaputtesDokumentWirdNichtAngefasst(t *testing.T) {
	t.Parallel()

	for _, vorher := range []string{
		`[1, 2]`, `{"a": 1`, `{"a": 1]`, `{"a": 1} {}`, `{"mcpServers": 5}`, `{"mcpServers": {"m": }}`, `{1: 2}`,
		`{"mcpServers": {"x": 1}, "mcpServers": {}}`,
		// **Ein kaputter Nachlauf ist kein Ende** (Zweitblick über Phase
		// 6.12, Z8): `Token` gibt dort einen Syntaxfehler und nicht
		// `io.EOF`, und beides galt als Ende — beim Umschreiben verschwand
		// der Nachlauf.
		`{"a": 1} @`, `{"a": 1} "`, `{"a": 1}]`,
	} {
		if _, fehler := jsonEintragSetzen([]byte(vorher), "m", json.RawMessage(`{}`)); !errors.Is(fehler, errKeinObjekt) {
			t.Errorf("setzen %q: %v", vorher, fehler)
		}

		if _, _, fehler := jsonEintrag([]byte(vorher), "m"); !errors.Is(fehler, errKeinObjekt) {
			t.Errorf("lesen %q: %v", vorher, fehler)
		}

		if _, fehler := jsonEintragEntfernen([]byte(vorher), "m"); !errors.Is(fehler, errKeinObjekt) {
			t.Errorf("entfernen %q: %v", vorher, fehler)
		}
	}
}

// **Ein Objekt, das beim Entfernen leer wird, bleibt als leeres stehen** —
// es gehörte dem Menschen, bevor dieses Programm etwas eintrug.
func TestEntfernenOhneEintragAendertNichts(t *testing.T) {
	t.Parallel()

	vorher := []byte(`{"mcpServers": {}}`)

	nachher, fehler := jsonEintragEntfernen(vorher, "m")
	if fehler != nil || !strings.Contains(string(nachher), `"mcpServers": {}`) {
		t.Errorf("%s %v", nachher, fehler)
	}

	ohneAussen, fehler := jsonEintragEntfernen([]byte(`{"a": 1}`), "m")
	if fehler != nil || strings.Contains(string(ohneAussen), "mcpServers") {
		t.Errorf("%s %v", ohneAussen, fehler)
	}
}
