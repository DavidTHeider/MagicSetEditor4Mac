package app

// **Der Ordner der Ansage ist privat, auch wenn es ihn schon gab** — Zweitblick
// über Phase 5, Befund 3 (SEC02, #2355). `os.MkdirAll(ordner, 0700)` ändert
// die Rechte eines bestehenden Ordners nicht und folgt einem Verweis; die
// Zusage „0600 im 0700-Ordner" galt so nur für einen frischen.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDerOrdnerDerAnsageIstPrivat(t *testing.T) {
	t.Parallel()

	ansage := Dienstansage{Adresse: "127.0.0.1:4711", Marke: markeNummer(1)}

	t.Run("ein frischer Ordner", func(t *testing.T) {
		t.Parallel()

		ordner := filepath.Join(t.TempDir(), "schnittstellen")
		if err := AnsageInDatei(filepath.Join(ordner, "dienst.txt"), ansage); err != nil {
			t.Fatal(err)
		}

		rechteSind(t, ordner, 0o700)
	})

	t.Run("ein bestehender, der zu weit offen ist", func(t *testing.T) {
		t.Parallel()

		ordner := filepath.Join(t.TempDir(), "schnittstellen")
		if err := os.Mkdir(ordner, 0o750); err != nil {
			t.Fatal(err)
		}

		if err := os.Chmod(ordner, 0o777); err != nil { //nolint:gosec // G302: der Fall, gegen den geprüft wird
			t.Fatal(err)
		}

		if err := AnsageInDatei(filepath.Join(ordner, "dienst.txt"), ansage); err != nil {
			t.Fatal(err)
		}

		rechteSind(t, ordner, 0o700)
	})

	t.Run("ein Verweis auf einen anderen Ordner", func(t *testing.T) {
		t.Parallel()

		raum := t.TempDir()
		fremd := filepath.Join(raum, "fremd")

		if err := os.Mkdir(fremd, 0o750); err != nil {
			t.Fatal(err)
		}

		ordner := filepath.Join(raum, "schnittstellen")
		if err := os.Symlink(fremd, ordner); err != nil {
			t.Fatal(err)
		}

		if err := AnsageInDatei(filepath.Join(ordner, "dienst.txt"), ansage); err == nil {
			t.Error("die Ansage ging durch einen Verweis")
		}

		if eintraege, err := os.ReadDir(fremd); err != nil || len(eintraege) != 0 {
			t.Errorf("im Ziel des Verweises liegt etwas: %v (%v)", eintraege, err)
		}
	})
}

func rechteSind(t *testing.T, ordner string, erwartet os.FileMode) {
	t.Helper()

	angaben, err := os.Lstat(ordner)
	if err != nil {
		t.Fatal(err)
	}

	if angaben.Mode().Perm() != erwartet {
		t.Errorf("%s hat %o, erwartet %o", ordner, angaben.Mode().Perm(), erwartet)
	}
}
