package mcpnaehte

// **Wie der MCP-Prozess sich ausweist** — Werkstück O2 von Phase 5.2 (#1426).

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDasTokenKommtAusDateiOderUmgebung(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	sicher := filepath.Join(ort, "token")
	offen := filepath.Join(ort, "offen")

	if err := os.WriteFile(sicher, []byte(" mse_1_geheim\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := os.WriteFile(offen, []byte("mse_1_geheim"), 0o644); err != nil { //nolint:gosec // die Probe braucht eine offene Datei
		t.Fatalf("schreiben: %v", err)
	}

	umgebung := func(werte map[string]string) func(string) string {
		return func(name string) string { return werte[name] }
	}

	faelle := map[string]struct {
		werte    map[string]string
		erwartet string
		fehlt    string
	}{
		"nichts gesetzt: Altbetrieb": {map[string]string{}, "", ""},
		"die Variable":               {map[string]string{Tokenvariable: " mse_1_x "}, "mse_1_x", ""}, // #nosec G101 -- Probewert
		"die Datei, beschnitten":     {map[string]string{Tokendateivariable: sicher}, "mse_1_geheim", ""},
		"die Datei geht vor": {
			map[string]string{Tokendateivariable: sicher, Tokenvariable: "mse_1_x"}, "mse_1_geheim", "",
		},
		"eine offene Datei wird abgewiesen": {
			map[string]string{Tokendateivariable: offen}, "", "für andere lesbar",
		},
		"eine fehlende Datei": {
			map[string]string{Tokendateivariable: filepath.Join(ort, "gibtsnicht")}, "", "ansehen",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			klartext, fehler := TokenAusUmgebung(umgebung(fall.werte), os.ReadFile, Dateirechte)

			if fall.fehlt != "" {
				if fehler == nil || !strings.Contains(fehler.Error(), fall.fehlt) {
					t.Fatalf("Fehler %v, erwartet %q", fehler, fall.fehlt)
				}

				return
			}

			if fehler != nil || klartext != fall.erwartet {
				t.Errorf("%q, %v — erwartet %q", klartext, fehler, fall.erwartet)
			}
		})
	}
}

// **Eine leere Datei und eine unlesbare sind Fehler**, nicht Altbetrieb: wer
// eine Datei nennt, will einen Ausweis.
func TestEineLeereOderUnlesbareTokendateiIstEinFehler(t *testing.T) {
	t.Parallel()

	leer := filepath.Join(t.TempDir(), "leer")
	if err := os.WriteFile(leer, []byte("\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	getenv := func(name string) string {
		if name == Tokendateivariable {
			return leer
		}

		return ""
	}

	if _, fehler := TokenAusUmgebung(getenv, os.ReadFile, Dateirechte); fehler == nil ||
		!strings.Contains(fehler.Error(), "leer") {
		t.Errorf("leere Datei: %v", fehler)
	}

	kaputt := errors.New("die Platte klemmt")
	unlesbar := func(string) ([]byte, error) { return nil, kaputt }
	rechte := func(string) (fs.FileMode, error) { return 0o600, nil }

	if _, fehler := TokenAusUmgebung(getenv, unlesbar, rechte); !errors.Is(fehler, kaputt) {
		t.Errorf("unlesbare Datei: %v", fehler)
	}
}
