// **Include und Exportvorlage aus einem Gerüst — an jeder Fläche** —
// FMT-02.AK-01 (#1691).
//
// Und das Entscheidende an der Exportvorlage: **sie läuft**. Ein Gerüst, das
// sich anlegen, aber nicht ausführen lässt, wäre ein Name in der Paketliste
// und keine Vorlage.
package flaechen_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIncludeUndExportvorlageAnJederFlaeche(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		for _, aufruf := range [][]string{
			{"paket-anlegen", "mse-include", "bausteine", "Bausteine"},
			{"paket-anlegen", "mse-export-template", "namen", "probe", "Namen"},
		} {
			if aus, fehler, code := kommandozeile(t, umgebung, aufruf...); code != 0 {
				t.Fatalf("%v: %d (%s) %s", aufruf, code, fehler, aus)
			}
		}

		imKatalog(t, umgebung, "die Kommandozeile", "bausteine")
		imKatalog(t, umgebung, "die Kommandozeile", "probe-namen")

		// **Die Vorlage läuft** — über dem Set, das die Saat für Exportvorlagen
		// anlegt, mit zwei Karten.
		setID := vorlagenset(t, umgebung)
		ziel := filepath.Join(t.TempDir(), "namen.txt")

		if _, fehler, code := kommandozeile(t, umgebung,
			"vorlagen", setID, "ausfuehren", "probe-namen.mse-export-template", ziel); code != 0 {
			t.Fatalf("ausführen: %d (%s)", code, fehler)
		}

		inhalt, err := os.ReadFile(ziel) //nolint:gosec // Pfad aus t.TempDir()
		if err != nil || strings.Count(string(inhalt), "\n") != 2 {
			t.Errorf("die Ausgabe hat nicht zwei Zeilen: %q (%v)", inhalt, err)
		}

		if _, fehler, code := kommandozeile(t, umgebung,
			"paket-anlegen", "mse-export-template", "waise"); code == 0 || !strings.Contains(fehler, "Spiel") {
			t.Errorf("eine Exportvorlage ohne Spiel: %d (%s)", code, fehler)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)

		for _, argumente := range []map[string]any{
			{"art": "mse-include", "name": "bausteine"},
			{"art": "mse-export-template", "name": "namen", "spiel": "probe"},
		} {
			if text, istFehler := mcpRuf(t, umgebung, "schreiben.paket.anlegen", argumente); istFehler {
				t.Fatalf("%v: %s", argumente, text)
			}
		}

		imKatalog(t, umgebung, "MCP", "bausteine")
		imKatalog(t, umgebung, "MCP", "probe-namen")
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		an := starteDienst(t, umgebung)

		for _, rumpf := range []string{
			`{"art":"mse-include","name":"bausteine"}`,
			`{"art":"mse-export-template","name":"namen","spiel":"probe"}`,
		} {
			if status, antwort := an.ruf(t, http.MethodPost, "/pakete/neu", rumpf); status != http.StatusCreated {
				t.Fatalf("%s: %d (%s)", rumpf, status, antwort)
			}
		}

		imKatalog(t, umgebung, "REST", "bausteine")
		imKatalog(t, umgebung, "REST", "probe-namen")
	})
}
