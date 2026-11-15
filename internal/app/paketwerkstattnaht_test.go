package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Naht liest und schreibt eine Paketdatei** —
// #1164.
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Dieser Lauf geht durch
// **denselben** Weg, den Kommandozeile, MCP und Dienst nehmen.
func TestDieWerkstattnahtLiestUndSchreibt(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	paket := paketImZuhause(t)

	naht, fehler := app.WerkstattnahtUeber()
	if fehler != nil {
		t.Fatalf("Naht: %v", fehler)
	}

	dateien, listenfehler := naht.Dateien(probekontext(t), paket)
	if listenfehler != nil {
		t.Fatalf("Dateien: %v", listenfehler)
	}

	if !slices.Contains(dateien, "style") {
		t.Fatalf("style fehlt in %v", dateien)
	}

	inhalt, stand, lesefehler := naht.Lesen(probekontext(t), paket, "style")
	if lesefehler != nil {
		t.Fatalf("Lesen: %v", lesefehler)
	}

	neuerStand, schreibfehler := naht.Schreiben(probekontext(t), paket, "style", inhalt+"x\n", stand)
	if schreibfehler != nil {
		t.Fatalf("Schreiben: %v", schreibfehler)
	}

	if neuerStand == stand {
		t.Error("der Stand blieb derselbe")
	}

	// **Der alte Stand trägt nicht mehr.** Ohne diese Zusicherung prüfte der
	// Lauf, dass geschrieben wird — nicht, dass der Schutz greift.
	if _, nochmal := naht.Schreiben(probekontext(t), paket, "style", "y\n", stand); !errors.Is(
		nochmal, domain.ErrConflict) {
		t.Errorf("der alte Stand kam durch: %v", nochmal)
	}
}

// **Ohne Zuhause kommt die Naht nicht zusammen.**
func TestOhneZuhauseKommtDieWerkstattnahtNichtZusammen(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.WerkstattnahtUeber(); fehler == nil {
		t.Fatal("kein Fehler ohne Zuhause")
	}
}

// paketImZuhause legt ein Paket im Paketverzeichnis dieses Laufs an.
func paketImZuhause(t *testing.T) string {
	t.Helper()

	wurzel, err := app.Paketverzeichnis()
	if err != nil {
		t.Fatalf("Paketverzeichnis: %v", err)
	}

	name := "probe.mse-style"

	if fehler := os.MkdirAll(filepath.Join(wurzel, name), 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(filepath.Join(wurzel, name, "style"),
		[]byte("mse version: 2.0.0\ngame: probe\n"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	return name
}
