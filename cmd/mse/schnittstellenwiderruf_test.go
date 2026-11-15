package main

// **`mse schnittstellen widerrufen`** — OPS03
// (#2349):
// die Außerbetriebnahme beginnt auch ohne Fenster. Eingerichtet wird weiter
// nur am Fenster.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestSchnittstellenWiderrufenNimmtEigenesUndLaesstFremdes(t *testing.T) {
	ort, _, _ := mitKarte(t)
	heim := os.Getenv("HOME")
	codex := filepath.Join(heim, ".codex", "config.toml")

	if err := os.MkdirAll(filepath.Dir(codex), 0o700); err != nil { // #nosec G703 -- das Wegwerfzuhause dieses Laufs
		t.Fatal(err)
	}

	fremd := "[mcp_servers.anderer]\ncommand = \"/opt/anderer\"\n"
	if err := os.WriteFile(codex, []byte(fremd), 0o600); err != nil { // #nosec G703 -- dito
		t.Fatal(err)
	}

	// **Eingerichtet am Fenster** — dieselbe Anwendung über dasselbe Zuhause.
	if antwort := app.NeueApp("").SchnittstelleEinrichten(probekontext(t), "codex"); antwort.Fehler != nil {
		t.Fatalf("einrichten: %+v", antwort.Fehler)
	}

	code, aus, fehler := fahreAn(t, ort, "schnittstellen", "widerrufen")
	if code != CodeGut || !strings.Contains(aus, "codex\twiderrufen") {
		t.Fatalf("Rückgabewert %d, Ausgabe %q (%s)", code, aus, fehler)
	}

	if gelesen, err := os.ReadFile(codex); err != nil || string(gelesen) != fremd { //nolint:gosec // eine Datei des Laufs
		t.Errorf("die Datei ist nicht wie vorher: %q, %v", gelesen, err)
	}

	// **Ein zweites Mal ist nichts zu tun** — und kein Fehler.
	if code, aus, _ := fahreAn(t, ort, "schnittstellen", "widerrufen"); code != CodeGut ||
		!strings.Contains(aus, "codex\tunveraendert") {
		t.Errorf("ein zweites Mal: %d, %q", code, aus)
	}

	// **Ein Ziel allein** — und eines, das es nicht gibt.
	if code, aus, _ := fahreAn(t, ort, "schnittstellen", "widerrufen", "codex"); code != CodeGut ||
		!strings.Contains(aus, "codex\toffen") {
		t.Errorf("ein Ziel: %d, %q", code, aus)
	}

	if code, _, _ := fahreAn(t, ort, "schnittstellen", "widerrufen", "gibtsnicht"); code == CodeGut {
		t.Error("ein unbekanntes Ziel wurde angenommen")
	}

	// **Einrichten bleibt am Fenster.**
	for _, falsch := range [][]string{{"schnittstellen", "einrichten"}, {"schnittstellen", "widerrufen", "a", "b"}} {
		if code, _, _ := fahreAn(t, ort, falsch...); code != CodeBedienung {
			t.Errorf("%v: Rückgabewert %d, erwartet %d", falsch, code, CodeBedienung)
		}
	}
}

func TestSchnittstellenWiderrufenAlsJSON(t *testing.T) {
	ort, _, _ := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "--json", "schnittstellen", "widerrufen")
	if code != CodeGut || !strings.Contains(aus, `"ausgaenge"`) || !strings.Contains(aus, `"befehle"`) {
		t.Errorf("Rückgabewert %d, Ausgabe %q (%s)", code, aus, fehler)
	}

	if code, aus, _ := fahreAn(t, ort, "--json", "schnittstellen", "widerrufen", "codex"); code != CodeGut ||
		!strings.Contains(aus, `"schnittstellen"`) {
		t.Errorf("ein Ziel: %d, %q", code, aus)
	}
}

// **Ohne Zuhause scheitert jeder Weg fachlich** — und nennt es, auch mit
// `--json`.
func TestSchnittstellenOhneZuhauseScheitern(t *testing.T) {
	ort, _, _ := mitKarte(t)
	t.Setenv("HOME", "")

	for _, argumente := range [][]string{
		{"schnittstellen"},
		{"schnittstellen", "widerrufen"},
		{"schnittstellen", "widerrufen", "codex"},
		{"--json", "schnittstellen", "widerrufen"},
	} {
		if code, _, fehler := fahreAn(t, ort, argumente...); code != CodeFachlich || fehler == "" && argumente[0] != "--json" {
			t.Errorf("%v: Rückgabewert %d (%q)", argumente, code, fehler)
		}
	}
}
