// **Der Dateibaum eines Pakets — an jeder Fläche** — FMT-02.AK-01
// (#1691).
//
// Derselbe Ablauf über Kommandozeile, MCP und REST: in einer eigenen Kopie
// eine Vorlage anlegen, umbenennen und entfernen — und dieselbe Abweisung,
// wenn eine Vorlage ohne ihren Stand umbenannt werden soll (§16).
package flaechen_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ohneStandSatz ist der Satz, mit dem der Dienst eine Vorlage ohne Stand
// abweist — an jeder Fläche derselbe.
const ohneStandSatz = "ist eine Vorlage — ohne ihren Stand"

// kopiepaket ist die eigene Kopie der Saat, an der gearbeitet wird.
const kopiepaket = "probe-schlicht-kopie.mse-style"

func standVon(inhalt string) string {
	roh := sha256.Sum256([]byte(inhalt))

	return hex.EncodeToString(roh[:])
}

// mitKopie legt die eigene Kopie an — die Voraussetzung, nicht der Gegenstand.
func mitKopie(t *testing.T) []string {
	t.Helper()

	umgebung := mitEigenerSaat(t)
	if _, fehler, code := kommandozeile(t, umgebung, "paket-kopieren", "mse-style", "probe-schlicht"); code != 0 {
		t.Fatalf("die Kopie anlegen: %d (%s)", code, fehler)
	}

	return umgebung
}

func TestDerDateibaumAnJederFlaeche(t *testing.T) {
	t.Parallel()

	stand := standVon("# neu\n")

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitKopie(t)
		quelle := filepath.Join(t.TempDir(), "neu.txt")
		if err := os.WriteFile(quelle, []byte("# neu\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		for _, schritt := range [][]string{
			{"paketdatei-anlegen", kopiepaket, "skripte/neu", quelle},
			{"paketdatei-umbenennen", kopiepaket, "skripte/neu", "skripte/alt", stand},
			{"paketdatei-entfernen", kopiepaket, "skripte/alt", stand},
		} {
			if aus, fehler, code := kommandozeile(t, umgebung, schritt...); code != 0 {
				t.Fatalf("%s: %d (%s) %s", schritt[0], code, fehler, aus)
			}
		}

		if _, fehler, code := kommandozeile(t, umgebung,
			"paketdatei-umbenennen", kopiepaket, "style", "stil"); code == 0 || !strings.Contains(fehler, ohneStandSatz) {
			t.Errorf("ohne Stand: %d (%s)", code, fehler)
		}

		if aus, _, _ := kommandozeile(t, umgebung, "paketdateien", kopiepaket); strings.Contains(aus, "skripte/") {
			t.Errorf("die Datei steht noch da: %s", aus)
		}

		// **Einfügen gibt es nur hier und am Fenster** — REST und MCP lesen
		// keine Datei des Rechners (`docs/FAEHIGKEITEN.md`).
		bild := filepath.Join(t.TempDir(), "maske.png")
		if err := os.WriteFile(bild, []byte("\x89PNG\r\n\x1a\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		if aus, fehler, code := kommandozeile(t, umgebung,
			"paketdatei-einfuegen", kopiepaket, bild, "bilder/maske.png"); code != 0 {
			t.Fatalf("einfügen: %d (%s) %s", code, fehler, aus)
		}

		// **Das Beiwerk steht im JSON**, nicht in der Textausgabe — dort wäre
		// es eine brechende Änderung (0033, Regel 6).
		if aus, _, _ := kommandozeile(t, umgebung, "--json", "paketdateien", kopiepaket); !strings.Contains(aus, "bilder/maske.png") {
			t.Errorf("das Beiwerk fehlt im JSON: %s", aus)
		}

		if aus, _, _ := kommandozeile(t, umgebung, "paketdateien", kopiepaket); strings.Contains(aus, "maske.png") {
			t.Errorf("das Beiwerk steht in der Textausgabe: %s", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitKopie(t)

		for name, argumente := range map[string]map[string]any{
			"schreiben.paketdatei.anlegen": {"paket": kopiepaket, "datei": "skripte/neu", "inhalt": "# neu\n"},
		} {
			if text, istFehler := mcpRuf(t, umgebung, name, argumente); istFehler || !strings.Contains(text, stand) {
				t.Fatalf("%s: %s", name, text)
			}
		}

		for _, schritt := range []struct {
			name      string
			argumente map[string]any
		}{
			{"schreiben.paketdatei.umbenennen", map[string]any{"paket": kopiepaket, "datei": "skripte/neu", "neu": "skripte/alt", "stand": stand}},
			{"schreiben.paketdatei.entfernen", map[string]any{"paket": kopiepaket, "datei": "skripte/alt", "stand": stand}},
		} {
			if text, istFehler := mcpRuf(t, umgebung, schritt.name, schritt.argumente); istFehler {
				t.Fatalf("%s: %s", schritt.name, text)
			}
		}

		if text, istFehler := mcpRuf(t, umgebung, "schreiben.paketdatei.umbenennen",
			map[string]any{"paket": kopiepaket, "datei": "style", "neu": "stil"}); !istFehler || !strings.Contains(text, ohneStandSatz) {
			t.Errorf("ohne Stand: %s", text)
		}

		if text, _ := mcpRuf(t, umgebung, "lesen.paketdateien", map[string]any{"paket": kopiepaket}); strings.Contains(text, "skripte/") {
			t.Errorf("die Datei steht noch da: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitKopie(t)
		an := starteDienst(t, umgebung)
		weg := "/pakete/" + kopiepaket + "/dateien"

		if status, rumpf := an.ruf(t, http.MethodPost, weg,
			`{"name":"skripte/neu","inhalt":"# neu\n"}`); status != http.StatusCreated || !strings.Contains(rumpf, stand) {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		if status, rumpf := an.ruf(t, http.MethodPatch, weg+"/skripte/neu",
			`{"neu":"skripte/alt","stand":"`+stand+`"}`); status != http.StatusOK {
			t.Fatalf("PATCH: %d (%s)", status, rumpf)
		}

		if status, rumpf := an.ruf(t, http.MethodDelete, weg+"/skripte/alt?stand="+stand, ""); status != http.StatusOK {
			t.Fatalf("DELETE: %d (%s)", status, rumpf)
		}

		if status, rumpf := an.ruf(t, http.MethodPatch, weg+"/style",
			`{"neu":"stil"}`); status != http.StatusBadRequest || !strings.Contains(rumpf, ohneStandSatz) {
			t.Errorf("ohne Stand: %d (%s)", status, rumpf)
		}

		if _, rumpf := an.ruf(t, http.MethodGet, weg, ""); strings.Contains(rumpf, "skripte/") {
			t.Errorf("die Datei steht noch da: %s", rumpf)
		}
	})
}
