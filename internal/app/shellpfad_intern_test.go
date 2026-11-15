package app

// **Der PATH der Shell und nicht der des Fensters** — Befund 9 des Zweitblicks
// über Phase 6.4 (#1665):
// ein Programm, das über LaunchServices startet, bekommt nicht den PATH der
// Shell, und „liegt im PATH" hieß dann nichts.

import (
	"errors"
	"os/exec"
	"slices"
	"testing"
)

func TestDerPfadKommtAusDerShell(t *testing.T) {
	t.Parallel()

	var gefragt *exec.Cmd

	// **Eine Shell, die redet** — wie eine mit einer Begrüßung in ihrer
	// Einrichtung; gelesen wird nur, was zwischen den Marken steht.
	pfad := shellpfad(t.Context(), func(befehl *exec.Cmd) ([]byte, error) {
		gefragt = befehl

		return []byte("Willkommen\n" + pfadmarke + "/opt/homebrew/bin:/Users/ich/.local/bin" + pfadmarke), nil
	}, "/usr/bin")
	if pfad != "/opt/homebrew/bin:/Users/ich/.local/bin" {
		t.Errorf("%q", pfad)
	}

	if gefragt.Path != "/bin/zsh" || !slices.Contains(gefragt.Args, "-ilc") {
		t.Errorf("gefragt wurde %q %q", gefragt.Path, gefragt.Args)
	}

	// **Ohne Antwort bleibt der eigene.**
	if pfad := shellpfad(t.Context(), func(*exec.Cmd) ([]byte, error) {
		return nil, errors.New("keine Shell")
	}, "/usr/bin"); pfad != "/usr/bin" {
		t.Errorf("ohne Shell: %q", pfad)
	}
}

// **Ohne Marken bleibt der eigene** — Nachlauf über Phase 6.4, N7: eine
// Shell, die mit 0 endet und nichts sagt, gab einen leeren PATH. Und **die
// Frist bindet**: ein Hintergrundprozess aus `.zshrc`, der die Ausgabe offen
// hält, darf das Warten nicht verlängern.
func TestOhneMarkenBleibtDerEigenePfad(t *testing.T) {
	t.Parallel()

	var gefragt *exec.Cmd

	pfad := shellpfad(t.Context(), func(befehl *exec.Cmd) ([]byte, error) {
		gefragt = befehl

		return []byte("Willkommen, aber ohne Antwort"), nil
	}, "/usr/bin")
	if pfad != "/usr/bin" {
		t.Errorf("%q", pfad)
	}

	if gefragt.WaitDelay <= 0 {
		t.Error("ohne WaitDelay bindet die Frist nicht")
	}
}
