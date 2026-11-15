// **Der Druckbogen an jeder Fläche** —
// #1291.
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Wer über die Kommandozeile, MCP oder REST einen Bogen legt, bekommt ein PDF.**
//
// # Der Befund
//
// `docs/FAEHIGKEITEN.md` führte die Zeile so:
//
//	| Einen Druckbogen legen | `BogenLegen` | — | — | — | ✋ |
//
// Drei Gedankenstriche und eine erhobene Hand. Die Hand steht für den
// Dateidialog — und der ist **kein Grund**: er ist der Weg des Fensters zu
// einem Pfad, den `App.BogenLegen` längst entgegennimmt. Sieben Zeilen im
// Katalog waren so begründet, gemessen am 2026-08-30; diese ist die erste.
//
// # Was hier gefahren wird
//
// Derselbe Ablauf dreimal: ein Set einfahren, den Bogen legen, nachsehen, dass
// eine **PDF-Datei** entstanden ist — nicht bloß eine Datei mit Inhalt.
func TestEinDruckbogenEntstehtAnJederFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	ziel := t.TempDir()

	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	t.Run("Kommandozeile", func(t *testing.T) {
		pfad := filepath.Join(ziel, "cli.pdf")

		aus, fehler, code := kommandozeile(t, umgebung, "bogen", setID, pfad)
		if code != 0 {
			t.Fatalf("bogen: %d (%s)", code, fehler)
		}

		istEinPDF(t, "die Kommandozeile", pfad, aus)
	})

	t.Run("MCP", func(t *testing.T) {
		pfad := filepath.Join(ziel, "mcp.pdf")

		text, istFehler := mcpRuf(t, umgebung, "schreiben.set.bogen",
			map[string]any{"setID": zahl(t, setID), "ziel": pfad})
		if istFehler {
			t.Fatalf("bogen: %s", text)
		}

		istEinPDF(t, "MCP", pfad, text)
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)
		pfad := filepath.Join(ziel, "rest.pdf")

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/bogen",
			`{"ziel":`+alsJSONText(t, pfad)+`}`)
		if status != http.StatusCreated {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		istEinPDF(t, "REST", pfad, rumpf)
	})
}

// **Und dieselbe fachliche Abweisung an jeder Fläche** (§16).
//
// Eine Setkennung, die auf nichts zeigt, ist an allen dreien dasselbe Nein —
// und nicht dreimal eine leere PDF-Datei. Genau das war der Ausgang: der Leger
// schrieb 0 Seiten und meldete Erfolg, weil `Karten` für ein leeres Set und
// für ein Set, das es nicht gibt, dieselbe leere Liste gibt.
func TestEinSetOhneDaseinWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	ziel := t.TempDir()

	t.Run("Kommandozeile", func(t *testing.T) {
		pfad := filepath.Join(ziel, "cli.pdf")

		if _, _, code := kommandozeile(t, umgebung, "bogen", "999", pfad); code == 0 {
			t.Error("die Kommandozeile hat es hingenommen")
		}

		nichtsEntstanden(t, "die Kommandozeile", pfad)
	})

	t.Run("MCP", func(t *testing.T) {
		pfad := filepath.Join(ziel, "mcp.pdf")

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.set.bogen",
			map[string]any{"setID": 999, "ziel": pfad}); !istFehler {
			t.Errorf("MCP hat es hingenommen: %s", text)
		}

		nichtsEntstanden(t, "MCP", pfad)
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)
		pfad := filepath.Join(ziel, "rest.pdf")

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/999/bogen",
			`{"ziel":`+alsJSONText(t, pfad)+`}`)
		if status < 400 {
			t.Errorf("REST hat es hingenommen: %d (%s)", status, rumpf)
		}

		nichtsEntstanden(t, "REST", pfad)
	})
}

// istEinPDF sieht nach, dass am Pfad ein PDF liegt und die Antwort ihn nennt.
func istEinPDF(t *testing.T, wer, pfad, antwort string) {
	t.Helper()

	roh, err := os.ReadFile(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("%s hat keine Datei hinterlassen: %v", wer, err)
	}

	// **Ein PDF und keine Datei mit Inhalt.** Der Unterschied ist der ganze
	// Punkt — eine leere Datei sähe von außen genauso aus.
	if !strings.HasPrefix(string(roh), "%PDF-") {
		t.Errorf("%s hat kein PDF geschrieben: %q", wer, roh[:min(20, len(roh))])
	}

	if !strings.Contains(antwort, pfad) {
		t.Errorf("%s nennt den Pfad nicht: %s", wer, antwort)
	}
}

// nichtsEntstanden sieht nach, dass eine Abweisung keine Datei hinterlässt.
func nichtsEntstanden(t *testing.T, wer, pfad string) {
	t.Helper()

	if _, err := os.Stat(pfad); err == nil {
		t.Errorf("%s hat trotz Abweisung eine Datei angelegt", wer)
	}
}
