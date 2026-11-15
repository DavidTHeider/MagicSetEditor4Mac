package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Bilder liegen neben der Datenbank** —
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md).
//
// Unter `~/Library/Application Support/`, bei den Paketen und der Datenbank:
// es ist Bestand der Anwendung, kein Dokument der Nutzenden.
func TestDasBilderverzeichnisLiegtNebenDerDatenbank(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort, err := app.Bilderverzeichnis()
	if err != nil {
		t.Fatalf("Bilderverzeichnis: %v", err)
	}

	if filepath.Base(ort) != "beiwerk" {
		t.Errorf("Ort: %q", ort)
	}

	// **Und es wird nicht angelegt.** Es entsteht beim ersten Bild; es bei
	// jedem Start anzulegen hinterließe auf jedem Rechner einen leeren Ordner,
	// den die allermeisten nie füllen.
	if _, fehler := os.Stat(ort); fehler == nil {
		t.Error("das Verzeichnis ist beim Fragen schon entstanden")
	}
}

// **Ohne Benutzerverzeichnis wird gemeldet statt geraten.**
func TestOhneBenutzerverzeichnisGibtEsKeinBilderverzeichnis(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := app.Bilderverzeichnis(); err == nil {
		t.Error("kein Fehler")
	}
}
