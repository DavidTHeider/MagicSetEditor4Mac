package main

// **Eine Datei ansehen von der Kommandozeile** — UX-13
// (#1685).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDieKommandozeileSiehtEineDateiAn(t *testing.T) {
	ort := t.TempDir()
	liste := filepath.Join(ort, "liste.txt")

	if fehler := os.WriteFile(liste, []byte("4 Blitzschlag\n"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	code, aus, _ := fahre(t, "datei-ansehen", liste)
	if code != CodeGut || !strings.Contains(aus, "art\ttext") || !strings.Contains(aus, "4 Blitzschlag") {
		t.Errorf("%d:\n%s", code, aus)
	}

	code, aus, _ = fahre(t, "--json", "datei-ansehen", liste)
	if code != CodeGut || !strings.Contains(aus, `"vorschau": "text"`) {
		t.Errorf("json %d:\n%s", code, aus)
	}

	paket := filepath.Join(ort, "magic.mse-game")
	if fehler := os.MkdirAll(paket, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.Symlink("/etc", filepath.Join(paket, "fremd")); fehler != nil {
		t.Fatal(fehler)
	}

	if code, aus, _ = fahre(t, "datei-ansehen", paket); code != CodeGut || !strings.Contains(aus, "eintrag\tfremd\t0\tunsicher") {
		t.Errorf("paket %d:\n%s", code, aus)
	}

	if code, _, fehler := fahre(t, "datei-ansehen", filepath.Join(ort, "fehlt")); code != CodeFachlich || fehler == "" {
		t.Errorf("fehlt: %d", code)
	}

	if code, _, _ := fahre(t, "datei-ansehen"); code != CodeBedienung {
		t.Errorf("ohne Pfad: %d", code)
	}
}
