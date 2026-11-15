package service_test

// **Ein Begleitordner ist privat, auch wenn es ihn schon gab** — Zweitblick
// über Phase 5, Befund 3 (SEC02, #2355): `os.MkdirAll(ordner, 0700)` ändert
// die Rechte eines bestehenden Ordners nicht und folgt einem Verweis.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinBestehenderBegleitordnerWirdPrivat(t *testing.T) {
	t.Parallel()

	ziel, _ := einAgentenziel(t)
	if err := os.MkdirAll(ziel.Begleitordner, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(ziel.Begleitordner, 0o777); err != nil { //nolint:gosec // G302: der Fall, gegen den geprüft wird
		t.Fatal(err)
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler != nil {
		t.Fatal(fehler)
	}

	if angaben, fehler := os.Lstat(ziel.Begleitordner); fehler != nil || angaben.Mode().Perm() != 0o700 {
		t.Errorf("der Begleitordner hat %v (%v), erwartet 0700", angaben.Mode().Perm(), fehler)
	}
}

func TestEinBegleitordnerAlsVerweisHaeltDenAgenten(t *testing.T) {
	t.Parallel()

	ziel, gerufen := einAgentenziel(t)
	fremd := filepath.Join(t.TempDir(), "fremd")

	if err := os.Mkdir(fremd, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(fremd, ziel.Begleitordner); err != nil {
		t.Fatal(err)
	}

	if _, fehler := service.SchnittstelleEinrichten(t.Context(), ziel); fehler == nil || len(*gerufen) != 0 {
		t.Errorf("durch einen Verweis eingerichtet: %v, %v", fehler, *gerufen)
	}

	if _, fehler := os.Stat(ziel.Datei); fehler == nil {
		t.Error("die plist liegt da, obwohl der Begleitordner ein Verweis ist")
	}
}

func TestEinPrivaterOrdnerNimmtKeineDatei(t *testing.T) {
	t.Parallel()

	datei := filepath.Join(t.TempDir(), "eine-datei")
	schreibeDatei(t, datei, "x")

	if fehler := service.PrivaterOrdner(datei); fehler == nil {
		t.Error("eine Datei galt als privater Ordner")
	}
}
