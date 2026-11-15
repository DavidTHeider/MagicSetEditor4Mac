package app

// **Der Werkzeugort hinter einem Link** — gefunden von der Hüllenprobe
// `ops09b` am 2026-10-01 (#1961).
//
// `mse` steht nach OPS-03 als Link in `~/.local/bin`, und `os.Executable`
// gibt unter macOS den Pfad, über den gestartet wurde — gemessen am selben
// Tag mit einem Programm, einem Link darauf und einem Aufruf über `PATH`:
// alle drei nennen den Link. Der Ort von `mse-mcp` war damit für die
// Kommandozeile `~/.local/bin` und für das Fenster das Bündel, und
// `mse schnittstellen` nannte „veraltet", was das Fenster „eingerichtet"
// nannte.

import (
	"os"
	"path/filepath"
	"testing"
)

// **Ein Link führt zum Ort des Programms**, nicht zum Ort des Links.
func TestDerWerkzeugortFolgtDemLink(t *testing.T) {
	t.Parallel()

	buendel := t.TempDir()
	programm := filepath.Join(buendel, "mse")

	if fehler := os.WriteFile(programm, []byte("#!/bin/sh\n"), 0o700); fehler != nil { //nolint:gosec // ein Programm des Laufs
		t.Fatal(fehler)
	}

	link := filepath.Join(t.TempDir(), "mse")
	if fehler := os.Symlink(programm, link); fehler != nil {
		t.Fatal(fehler)
	}

	ort, fehler := werkzeugortNeben(func() (string, error) { return link, nil })

	erwartet, _ := filepath.EvalSymlinks(buendel) //nolint:errcheck // der Ordner des Laufs steht
	if fehler != nil || ort != erwartet {
		t.Errorf("über den Link: %q, %v — erwartet %q", ort, fehler, erwartet)
	}
}

// **Ein Pfad, der sich nicht auflösen lässt, gilt, wie er ist** —
// `os.Executable` nennt unter macOS ein Programm, das läuft; scheitert das
// Auflösen trotzdem, ist sein Ordner die bessere Antwort als keine.
func TestEinUnaufloesbarerPfadGiltWieErIst(t *testing.T) {
	t.Parallel()

	ort, fehler := werkzeugortNeben(func() (string, error) { return "/gibt/es/nicht/mse", nil })
	if fehler != nil || ort != "/gibt/es/nicht" {
		t.Errorf("%q, %v", ort, fehler)
	}
}
