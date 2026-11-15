package main

// **Ein fachlicher Fehler trägt im JSON seinen Code** — F03
// (#2342),
// Befunde A-03 und A-24.
//
// # Warum
//
// Mit `--json` gab ein Erfolg JSON und ein Fehler nur einen deutschen Satz auf
// der Fehlerausgabe. Ein Skript, das nach der Klasse verzweigen wollte, musste
// den Satz lesen — und ein Satz ist keine Zusage. Der Dienst nennt den Code
// längst im Rumpf; die Kommandozeile jetzt auch.
//
// # Was die Prüfungen halten
//
// Mit `--json` steht auf der Ausgabe `{"fehler": …}` mit Code und
// Vorgangskennung, **zusätzlich**: der Satz bleibt auf der Fehlerausgabe, der
// Rückgabewert bleibt 1. Ohne `--json` ändert sich nichts.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinFachlicherFehlerTraegtMitJSONSeinenCode(t *testing.T) {
	code, aus, fehler := fahre(t, "--json", "karten", "4711")

	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d, erwartet %d (%s)", code, CodeFachlich, fehler)
	}

	var gelesen struct {
		Fehler struct {
			Code    string `json:"code"`
			Vorgang string `json:"vorgang"`
		} `json:"fehler"`
	}

	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil {
		t.Fatalf("die Ausgabe ist kein JSON: %v\n%s", err, aus)
	}

	if gelesen.Fehler.Code != "nichtGefunden" || len(gelesen.Fehler.Vorgang) != 8 {
		t.Errorf("Code %q, Vorgang %q", gelesen.Fehler.Code, gelesen.Fehler.Vorgang)
	}

	if !strings.Contains(fehler, "nicht gefunden") || !strings.Contains(fehler, gelesen.Fehler.Vorgang) {
		t.Errorf("der Satz und die Kennung auf der Fehlerausgabe fehlen:\n%s", fehler)
	}
}

func TestOhneJSONBleibtDieAusgabeLeer(t *testing.T) {
	code, aus, _ := fahre(t, "karten", "4711")

	if code != CodeFachlich || aus != "" {
		t.Errorf("Rückgabewert %d, Ausgabe %q — ohne --json kommt nur der Satz", code, aus)
	}
}

// **Ein Wächter über den Weg** — gefunden, weil 53 Stellen und
// `klageUeberNaht` einen übersetzten Fehler nur als `Meldung` ausgaben und
// mit `--json` keinen Code nannten. Wer einen `app.Fehler` meldet, nimmt
// `klageUeberFehler`, wer einen `error` meldet, `klageUeber`.
//
// **Was er nicht hält**: eine Bedienungsklage (Rückgabewert 2) — sie ist
// keine Fehlerklasse —, und einen Fehler, den ein Befehl als Zeile einer
// Liste ausgibt (`sag`).
func TestKeinFachlicherFehlerGehtAmJSONVorbei(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	vorbei := regexp.MustCompile(`\.klage\("%s(\\n)?", [A-Za-z_][A-Za-z0-9_.]*\.Meldung\)`)

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") || datei == "ausgabe.go" {
			continue
		}

		roh, lesefehler := os.ReadFile(datei) //nolint:gosec // ein Pfad aus Glob im Paket
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		for _, fund := range vorbei.FindAllString(string(roh), -1) {
			t.Errorf("%s: %s — nimm klageUeberFehler, sonst fehlt der Code mit --json", datei, fund)
		}
	}
}

// **Auch ein belegter Bestand trägt seine Form** — Befund 4 des Zweitblicks
// über Phase 1: `klageUeber` kehrte mit seinem Rat zurück, bevor es die Form
// schrieb. Der Satz mit dem Rat bleibt.
func TestEinBelegterBestandTraegtMitJSONSeinenCode(t *testing.T) {
	var aus, fehler bytes.Buffer

	s := &schreiber{ziel: &aus, fehler: &fehler, alsJSON: true, vorgang: "a1b2c3d4"}
	s.klageUeber(fmt.Errorf("Migration beginnen: %w", domain.ErrBelegt))

	if !strings.Contains(fehler.String(), "gleich noch einmal") {
		t.Errorf("der Rat fehlt:\n%s", fehler.String())
	}

	if !strings.Contains(aus.String(), `"code":"belegt"`) || !strings.Contains(aus.String(), `"vorgang":"a1b2c3d4"`) {
		t.Errorf("mit --json fehlt die Form:\n%s", aus.String())
	}
}

// **Und eine Datei, die es nicht gibt** — `einstellungen vorschau` meldete den
// Lesefehler mit eigenem Satz und Rückgabewert 1, ohne Form.
func TestEineFehlendeDateiTraegtMitJSONIhreForm(t *testing.T) {
	code, aus, _ := fahre(t, "--json", "einstellungen", "vorschau", filepath.Join(t.TempDir(), "gibt-es-nicht.json"))

	if code != CodeFachlich || !strings.Contains(aus, `"fehler"`) || !strings.Contains(aus, `"vorgang"`) {
		t.Errorf("Rückgabewert %d, Ausgabe %q — erwartet die Form des Fehlers", code, aus)
	}
}
