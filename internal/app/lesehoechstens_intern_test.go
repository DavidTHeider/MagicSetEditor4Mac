package app

// **Marke und Spur werden begrenzt gelesen** — F04
// (#2343):
// eine Marke trägt vier kurze Felder, und eine Spur geht nur mit ihrem
// Anfang ins Diagnosepaket.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLeseHoechstensNimmtNurDenAnfang(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "lang.txt")
	if err := os.WriteFile(pfad, []byte(strings.Repeat("x", 100)), 0o600); err != nil {
		t.Fatal(err)
	}

	roh, err := leseHoechstens(pfad, 10)
	if err != nil || len(roh) != 10 {
		t.Errorf("gelesen %d Bytes, Fehler %v — erwartet 10", len(roh), err)
	}
}

func TestLeseHoechstensMeldetEineFehlendeDatei(t *testing.T) {
	t.Parallel()

	if _, err := leseHoechstens(filepath.Join(t.TempDir(), "fehlt.txt"), 10); err == nil {
		t.Error("eine fehlende Datei las sich ohne Fehler")
	}
}
