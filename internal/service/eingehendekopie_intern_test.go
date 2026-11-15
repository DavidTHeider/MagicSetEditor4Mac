package service

// **Die eingehende Kopie** —
// #964.
//
// **Eine eigene Datei**, weil `bestandersetzen_intern_test.go` über die
// Zeilengrenze wuchs (§2). Der Schnitt folgt der Frage: dort **der Tausch**,
// hier **die Kopie, die ihm vorausgeht**.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDieEingehendeKopieMeldetIhreFehler(t *testing.T) {
	t.Parallel()

	t.Run("die Sicherung gibt es nicht", func(t *testing.T) {
		t.Parallel()

		ort := t.TempDir()

		_, err := eingehendeKopie(filepath.Join(ort, "gibt-es-nicht.db"),
			filepath.Join(ort, "bestand.db"))
		if err == nil {
			t.Fatal("erwartet war ein Fehler, bekommen: nil")
		}
	})

	t.Run("die Sicherung lässt sich nicht lesen", func(t *testing.T) {
		t.Parallel()

		ort := t.TempDir()

		// **`os.Open` auf ein Verzeichnis gelingt** — erst das Lesen
		// scheitert. Der Fall trifft also `kopieren` und nicht die Wache
		// davor.
		sicherung := filepath.Join(ort, "sicherung.db")
		vollesVerzeichnis(t, sicherung)

		_, err := eingehendeKopie(sicherung, filepath.Join(ort, "bestand.db"))
		if err == nil {
			t.Fatal("erwartet war ein Fehler, bekommen: nil")
		}
	})

	t.Run("die Kopie lässt sich nicht anlegen", func(t *testing.T) {
		t.Parallel()

		ort := t.TempDir()

		sicherung := filepath.Join(ort, "sicherung.db")
		legeAn(t, sicherung, "neu")

		gesperrt := filepath.Join(ort, "gesperrt")
		if err := os.Mkdir(gesperrt, 0o500); err != nil {
			t.Fatalf("das gesperrte Verzeichnis anlegen: %v", err)
		}

		t.Cleanup(func() {
			// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
			_ = os.Chmod(gesperrt, 0o700) //nolint:errcheck // nur zum Aufräumen
		})

		if _, err := eingehendeKopie(sicherung, filepath.Join(gesperrt, "bestand.db")); err == nil {
			t.Fatal("erwartet war ein Fehler, bekommen: nil")
		}
	})
}
