// Der Arbeitsablauf, einmal über jede Fläche.
package flaechen_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// kommandozeile ruft `mse` und gibt Ausgabe, Fehlerstrom und Rückgabewert.
func kommandozeile(
	t *testing.T, umgebung []string, argumente ...string,
) (aus, fehler string, code int) {
	t.Helper()

	return mitEingabeUndCode(t, umgebung, "", argumente...)
}

// mitEingabeUndCode ruft `mse` mit einem Text auf der Standardeingabe.
//
// **Weil die Zwischenablage ohne Fenster stdin und stdout ist** — Werkstück 1
// von Phase 3.7
// (#1168).
// Ein Lauf ohne Eingabe prüfte `ablage-einfuegen` gar nicht.
func mitEingabeUndCode(
	t *testing.T, umgebung []string, eingabe string, argumente ...string,
) (aus, fehler string, code int) {
	t.Helper()

	// **Maschinenlesbar, wenn `abweisungGleich` es verlangt** (F03, #2342).
	if maschinenlesbar(t) {
		argumente = append(argumente, "--json")
	}

	// #nosec G204 -- gebautes Programm aus TestMain, feste Argumente
	lauf := exec.CommandContext(t.Context(), gebaut.mse, argumente...)
	lauf.Env = umgebung
	lauf.Stdin = strings.NewReader(eingabe)

	var raus, rfehler bytes.Buffer
	lauf.Stdout, lauf.Stderr = &raus, &rfehler

	if err := lauf.Run(); err != nil {
		var beendet *exec.ExitError
		if !errorsAs(err, &beendet) {
			t.Fatalf("mse %v: %v", argumente, err)
		}

		code = beendet.ExitCode()
	}

	return raus.String(), rfehler.String(), code
}

// mitEingabe ist mitEingabeUndCode und bricht ab, wenn es nicht glückt.
func mitEingabe(
	t *testing.T, umgebung []string, eingabe string, argumente ...string,
) string {
	t.Helper()

	aus, fehler, code := mitEingabeUndCode(t, umgebung, eingabe, argumente...)
	if code != 0 {
		t.Fatalf("mse %v: %d %s%s", argumente, code, aus, fehler)
	}

	return aus
}

// mcpRuf ruft ein MCP-Werkzeug über stdio und gibt seinen Text.
//
// **Über den echten Weg**: Anmeldung, dann `tools/call` — dieselbe Folge, die
// ein Sprachmodell geht.
func mcpRuf(
	t *testing.T, umgebung []string, name string, argumente map[string]any,
) (text string, istFehler bool) {
	t.Helper()

	teile, strukturiert, istFehler := mcpErgebnis(t, umgebung, name, argumente)

	// **Strukturiert, wenn `abweisungGleich` es verlangt** (F03, #2342).
	if maschinenlesbar(t) {
		return teile[0] + "\n" + string(strukturiert), istFehler
	}

	return teile[0], istFehler
}

// mcpTeile ist [mcpRuf] samt **jedem** Teil des Ergebnisses — für die Wege,
// die einem Ergebnis einen zweiten Teil anhängen (#1970).
func mcpTeile(
	t *testing.T, umgebung []string, name string, argumente map[string]any,
) (teile []string, istFehler bool) {
	t.Helper()

	teile, _, istFehler = mcpErgebnis(t, umgebung, name, argumente)

	return teile, istFehler
}

// mcpErgebnis ist [mcpTeile] samt `structuredContent`.
func mcpErgebnis(
	t *testing.T, umgebung []string, name string, argumente map[string]any,
) (teile []string, strukturiert json.RawMessage, istFehler bool) {
	t.Helper()

	const anmeldung = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":` +
		`{"protocolVersion":"2025-06-18","capabilities":{},` +
		`"clientInfo":{"name":"abnahme","version":"1"}}}` + "\n"

	ruf := mussJSON(t, map[string]any{
		"jsonrpc": "2.0", "id": 2, "method": "tools/call",
		"params": map[string]any{"name": name, "arguments": argumente},
	})

	eingabe := anmeldung + string(ruf) + "\n"

	// #nosec G204 -- gebautes Programm aus TestMain
	lauf := exec.CommandContext(t.Context(), gebaut.mcp)
	lauf.Env = umgebung
	lauf.Stdin = strings.NewReader(eingabe)

	var aus bytes.Buffer
	lauf.Stdout = &aus

	if err := lauf.Run(); err != nil {
		t.Fatalf("mse-mcp %s: %v", name, err)
	}

	zeilen := strings.Split(strings.TrimSpace(aus.String()), "\n")

	var antwort struct {
		Ergebnis struct {
			Inhalt []struct {
				Text string `json:"text"`
			} `json:"content"`
			IstFehler    bool            `json:"isError"`
			Strukturiert json.RawMessage `json:"structuredContent"`
		} `json:"result"`
	}

	if err := json.Unmarshal([]byte(zeilen[len(zeilen)-1]), &antwort); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v\n%s", err, aus.String())
	}

	if len(antwort.Ergebnis.Inhalt) == 0 {
		t.Fatalf("die Antwort hat keinen Inhalt: %s", aus.String())
	}

	for _, teil := range antwort.Ergebnis.Inhalt {
		teile = append(teile, teil.Text)
	}

	return teile, antwort.Ergebnis.Strukturiert, antwort.Ergebnis.IstFehler
}

// **Der Arbeitsablauf über die Kommandozeile.**
func TestDerAblaufUeberDieKommandozeile(t *testing.T) {
	umgebung := eigenerBestand(t)
	ziel := t.TempDir()

	// 1. einfahren
	aus, fehler, code := kommandozeile(t, umgebung, "oeffnen", probeset(t), "--json")
	if code != 0 {
		t.Fatalf("einfahren: %d (%s)", code, fehler)
	}

	setID := setkennungAus(t, aus)

	// 2. eine Karte finden
	aus, _, code = kommandozeile(t, umgebung, "karten", setID, "--json")
	if code != 0 {
		t.Fatalf("karten: %d", code)
	}

	karteID := ersteKarte(t, aus)

	// 3. ändern
	if _, fehler, code = kommandozeile(t, umgebung,
		"feld", setID, karteID, "name", "Aus der Abnahme"); code != 0 {
		t.Fatalf("feld: %d (%s)", code, fehler)
	}

	// 4. zeichnen
	bild := filepath.Join(ziel, "karte.png")
	if _, fehler, code = kommandozeile(t, umgebung, "zeichnen", karteID, bild); code != 0 {
		t.Fatalf("zeichnen: %d (%s)", code, fehler)
	}

	istBild(t, bild)

	// 5. sichern
	raus := filepath.Join(ziel, "raus.mse-set")
	if _, fehler, code = kommandozeile(t, umgebung, "sichern", setID, raus); code != 0 {
		t.Fatalf("sichern: %d (%s)", code, fehler)
	}

	istWiederLesbaresSet(t, umgebung, raus, "Aus der Abnahme")

	// 6. und die Änderung steht wirklich darin
	aus, _, _ = kommandozeile(t, umgebung, "karten", setID, "--json")
	if !strings.Contains(aus, "Aus der Abnahme") {
		t.Error("die Änderung steht nicht im Bestand")
	}
}

// **Derselbe Arbeitsablauf über MCP.**
func TestDerAblaufUeberMCP(t *testing.T) {
	umgebung := eigenerBestand(t)
	ziel := t.TempDir()

	// 1. einfahren
	text, istFehler := mcpRuf(t, umgebung, "schreiben.set.einfahren",
		map[string]any{"pfad": probeset(t)})
	if istFehler {
		t.Fatalf("einfahren: %s", text)
	}

	setID := setkennungAus(t, text)

	// 2. eine Karte finden
	text, istFehler = mcpRuf(t, umgebung, "lesen.karten",
		map[string]any{"setID": zahl(t, setID)})
	if istFehler {
		t.Fatalf("karten: %s", text)
	}

	karteID := ersteKarteMCP(t, text)

	// 3. ändern
	text, istFehler = mcpRuf(t, umgebung, "schreiben.feld", map[string]any{
		"setID": zahl(t, setID), "karteID": zahl(t, karteID),
		"feld": "name", "wert": "Aus der Abnahme",
	})
	if istFehler {
		t.Fatalf("feld: %s", text)
	}

	// 4. zeichnen
	bild := filepath.Join(ziel, "karte.png")

	text, istFehler = mcpRuf(t, umgebung, "schreiben.karte.zeichnen",
		map[string]any{"karteID": zahl(t, karteID), "ziel": bild})
	if istFehler {
		t.Fatalf("zeichnen: %s", text)
	}

	istBild(t, bild)

	// 5. sichern
	raus := filepath.Join(ziel, "raus.mse-set")

	text, istFehler = mcpRuf(t, umgebung, "schreiben.set.sichern",
		map[string]any{"setID": zahl(t, setID), "ziel": raus})
	if istFehler {
		t.Fatalf("sichern: %s", text)
	}

	istWiederLesbaresSet(t, umgebung, raus, "Aus der Abnahme")

	// 6. und die Änderung steht wirklich darin
	text, _ = mcpRuf(t, umgebung, "lesen.karten", map[string]any{"setID": zahl(t, setID)})
	if !strings.Contains(text, "Aus der Abnahme") {
		t.Error("die Änderung steht nicht im Bestand")
	}
}
