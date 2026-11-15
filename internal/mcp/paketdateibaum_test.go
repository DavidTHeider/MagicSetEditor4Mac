// **Der Dateibaum am MCP-Server** — FMT-02.AK-01
// (#1691).
package mcp_test

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

// gelesen macht aus der Antwort eines Werkzeugs eine Abbildung.
func gelesen(t *testing.T, antwort string) map[string]any {
	t.Helper()

	var abbildung map[string]any
	if err := json.Unmarshal([]byte(antwort), &abbildung); err != nil {
		t.Fatalf("keine JSON-Antwort: %v\n%s", err, antwort)
	}

	return abbildung
}

func TestDerDateibaumUeberMCP(t *testing.T) {
	probe := &werkstattprobe{stand: "neu", beiwerk: []string{"rahmen.png"}}

	liste, fehler := rufeWerkstatt(t, probe, "lesen.paketdateien", `{"paket": "p"}`)
	if fehler != nil {
		t.Fatal(fehler)
	}

	felder := gelesen(t, liste)
	if beiwerk, istListe := felder["Beiwerk"].([]any); !istListe || !slices.Equal(beiwerk, []any{"rahmen.png"}) {
		t.Errorf("Beiwerk %v", felder["Beiwerk"])
	}

	// **Leer heißt `[]` und nicht `null`.**
	if dateien, liste := felder["Dateien"].([]any); !liste || len(dateien) != 0 {
		t.Errorf("Dateien %v", felder["Dateien"])
	}

	angelegt, fehler := rufeWerkstatt(t, probe, "schreiben.paketdatei.anlegen",
		`{"paket": "p", "datei": "skripte/neu", "inhalt": "# leer"}`)
	if fehler != nil || gelesen(t, angelegt)["Stand"] != "neu" || probe.datei != "skripte/neu" || probe.wert != "# leer" {
		t.Errorf("Anlegen %s (%v), %+v", angelegt, fehler, probe)
	}

	umbenannt, fehler := rufeWerkstatt(t, probe, "schreiben.paketdatei.umbenennen",
		`{"paket": "p", "datei": "a", "neu": "b", "stand": "s"}`)
	if fehler != nil || gelesen(t, umbenannt)["Datei"] != "b" || probe.datei != "a" || probe.neu != "b" || probe.erwartet != "s" {
		t.Errorf("Umbenennen %s (%v), %+v", umbenannt, fehler, probe)
	}

	entfernt, fehler := rufeWerkstatt(t, probe, "schreiben.paketdatei.entfernen", `{"paket": "p", "datei": "c"}`)
	if fehler != nil || gelesen(t, entfernt)["Entfernt"] != "c" || probe.datei != "c" || probe.erwartet != "" {
		t.Errorf("Entfernen %s (%v), %+v", entfernt, fehler, probe)
	}
}

func TestDerDateibaumMeldetSeineFehler(t *testing.T) {
	probe := &werkstattprobe{fehler: errors.New("klemmt")}

	for name, argumente := range map[string]string{
		"lesen.paketdateien":              `{"paket": "p"}`,
		"schreiben.paketdatei.anlegen":    `{"paket": "p", "datei": "a", "inhalt": ""}`,
		"schreiben.paketdatei.umbenennen": `{"paket": "p", "datei": "a", "neu": "b"}`,
		"schreiben.paketdatei.entfernen":  `{"paket": "p", "datei": "a"}`,
	} {
		if _, fehler := rufeWerkstatt(t, probe, name, argumente); fehler == nil {
			t.Errorf("%s: kein Fehler", name)
		}

		if _, fehler := rufeWerkstatt(t, probe, name, `{"paket": 1}`); fehler == nil {
			t.Errorf("%s: falsche Argumente kamen durch", name)
		}
	}
}
