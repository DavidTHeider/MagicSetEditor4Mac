// **Der Bestandsort einer Fläche ohne Fenster** (#886).
package app_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ohne Umgebung gilt der Ort der Anwendung.**
func TestOhneUmgebungGiltDerVorgabeort(t *testing.T) {
	t.Setenv(app.Bestandsumgebung, "")

	ort, err := app.BestandsortFuerFlaechen()
	if err != nil {
		t.Fatalf("Ort: %v", err)
	}

	vorgabe, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Vorgabe: %v", err)
	}

	if ort != vorgabe {
		t.Errorf("der Ort ist %q, erwartet war %q", ort, vorgabe)
	}
}

// **Mit Umgebung gilt sie.**
func TestMitUmgebungGiltSie(t *testing.T) {
	eigener := filepath.Join(t.TempDir(), "eigen.db")
	t.Setenv(app.Bestandsumgebung, eigener)

	ort, err := app.BestandsortFuerFlaechen()
	if err != nil {
		t.Fatalf("Ort: %v", err)
	}

	if ort != eigener {
		t.Errorf("der Ort ist %q, erwartet war %q", ort, eigener)
	}
}

// **Ein leerer Wert ist ein vergessenes `export`, kein Wunsch.**
//
// Ohne diese Unterscheidung entstünde eine Datei namens „" im
// Arbeitsverzeichnis — und niemand fände sie.
func TestEinLeererWertGiltNicht(t *testing.T) {
	t.Setenv(app.Bestandsumgebung, "")

	ort, err := app.BestandsortFuerFlaechen()
	if err != nil {
		t.Fatalf("Ort: %v", err)
	}

	if ort == "" {
		t.Error("der leere Wert wurde übernommen")
	}
}

// **Ohne Benutzerverzeichnis und ohne Umgebung gibt es keinen Ort.**
func TestOhneBeidesGibtEsKeinenOrt(t *testing.T) {
	t.Setenv(app.Bestandsumgebung, "")
	t.Setenv("HOME", "")

	if _, err := app.BestandsortFuerFlaechen(); err == nil {
		t.Error("es kam ein Ort ohne Benutzerverzeichnis")
	}
}
