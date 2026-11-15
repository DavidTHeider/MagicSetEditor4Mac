// **Die Exportvorlagen an jeder Fläche** —
// #1291.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// vorlagenset legt ein Set mit zwei Karten an, dessen Spiel eine Vorlage hat.
func vorlagenset(t *testing.T, umgebung []string) string {
	t.Helper()

	var angelegt struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal([]byte(mussKommandozeile(t, umgebung,
		"set-neu", "Vorlagenset", "probe", "probe-schlicht", "--json")), &angelegt); err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	for range 2 {
		mussKommandozeile(t, umgebung, "karte", angelegt.ID, "anlegen", "--json")
	}

	return angelegt.ID
}

// **Wer über die Kommandozeile, MCP oder REST eine Vorlage ausführt, bekommt
// dieselbe Datei.**
//
// # Der Befund
//
// `docs/FAEHIGKEITEN.md` führte die Zeile so:
//
//	| Über eine Exportvorlage ausgeben | `Exportvorlagen`, `VorlageAusfuehren`, `ExportoptionSetzen` | — | — | — | ✋ |
//
// Die erhobene Hand meint einen Speicherdialog, und `App.VorlageAusfuehren`
// sagt selbst: *„Der Pfad kommt vom Aufrufer, nicht aus einem Dialog."*
//
// **Es ist die Fähigkeit, für die MSE2 seine Vorlagen mitbringt**: das
// `M15-Magic-Pack` führt zwölf `.mse-export-template`.
func TestEineExportvorlageLaeuftAnJederFlaeche(t *testing.T) {
	ziel := t.TempDir()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		pfad := filepath.Join(ziel, "cli.txt")

		aus, fehler, code := kommandozeile(t, umgebung,
			"vorlagen", setID, "ausfuehren", "probe-liste.mse-export-template", pfad)
		if code != 0 {
			t.Fatalf("ausführen: %d (%s)", code, fehler)
		}

		hatInhalt(t, "die Kommandozeile", pfad, aus)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		pfad := filepath.Join(ziel, "mcp.txt")

		text, istFehler := mcpRuf(t, umgebung, "schreiben.vorlage.ausfuehren",
			map[string]any{
				"setID": zahl(t, setID),
				"paket": "probe-liste.mse-export-template", "ziel": pfad,
			})
		if istFehler {
			t.Fatalf("ausführen: %s", text)
		}

		hatInhalt(t, "MCP", pfad, text)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		an := starteDienst(t, umgebung)
		pfad := filepath.Join(ziel, "rest.txt")

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/vorlagen",
			`{"paket":"probe-liste.mse-export-template","ziel":`+alsJSONText(t, pfad)+`}`)
		if status != http.StatusCreated {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		hatInhalt(t, "REST", pfad, rumpf)
	})
}

// **Und die Liste der Vorlagen kommt an jeder Fläche an.**
func TestDieVorlagenlisteStehtAnJederFlaeche(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)

		if aus := mussKommandozeile(t, umgebung,
			"vorlagen", setID, "--json"); !strings.Contains(aus, "probe-liste") {
			t.Errorf("die Kommandozeile nennt die Vorlage nicht: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)

		text, istFehler := mcpRuf(t, umgebung, "lesen.vorlagen",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("lesen: %s", text)
		}

		if !strings.Contains(text, "probe-liste") {
			t.Errorf("MCP nennt die Vorlage nicht: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/vorlagen", "")
		if status != http.StatusOK {
			t.Fatalf("GET: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, "probe-liste") {
			t.Errorf("REST nennt die Vorlage nicht: %s", rumpf)
		}
	})
}

// **Und dieselbe fachliche Abweisung: eine Vorlage, die es nicht gibt** (§16).
func TestEineUnbekannteVorlageWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	ziel := t.TempDir()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)

		if _, _, code := kommandozeile(t, umgebung, "vorlagen", setID, "ausfuehren",
			"gibtsnicht.mse-export-template", filepath.Join(ziel, "a.txt")); code == 0 {
			t.Error("die Kommandozeile hat es hingenommen")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.vorlage.ausfuehren",
			map[string]any{
				"setID": zahl(t, setID), "paket": "gibtsnicht.mse-export-template",
				"ziel": filepath.Join(ziel, "b.txt"),
			}); !istFehler {
			t.Errorf("MCP hat es hingenommen: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/vorlagen",
			`{"paket":"gibtsnicht.mse-export-template","ziel":`+
				alsJSONText(t, filepath.Join(ziel, "c.txt"))+`}`)
		if status < 400 {
			t.Errorf("REST hat es hingenommen: %d (%s)", status, rumpf)
		}
	})
}

// hatInhalt sieht nach, dass am Pfad eine nicht leere Datei liegt.
//
// **Nicht leer und nicht bloß vorhanden**: eine Vorlage, die nichts schreibt,
// hinterließe auch eine Datei.
func hatInhalt(t *testing.T, wer, pfad, antwort string) {
	t.Helper()

	roh, err := os.ReadFile(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("%s hat keine Datei hinterlassen: %v", wer, err)
	}

	if strings.TrimSpace(string(roh)) == "" {
		t.Errorf("%s hat eine leere Datei geschrieben", wer)
	}

	if !strings.Contains(antwort, pfad) {
		t.Errorf("%s nennt den Pfad nicht: %s", wer, antwort)
	}
}
