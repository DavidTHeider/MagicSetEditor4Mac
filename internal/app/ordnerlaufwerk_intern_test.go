package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// TestEinLaufwerksverweisFuehrtNichtInsZuhause hält Nachlauf über Phase 6.4,
// N1: `/Volumes/Macintosh HD` ist auf jedem Mac ein Verweis auf `/`. Über
// ihn kam `~/Library/LaunchAgents` als „Laufwerk" an der Ordnerregel vorbei
// — geprüft wurde der Pfad, wie er dastand, und nicht, wohin er führt.
func TestEinLaufwerksverweisFuehrtNichtInsZuhause(t *testing.T) {
	wurzel := t.TempDir()
	heim := filepath.Join(wurzel, "Users", "ich")
	agenten := filepath.Join(heim, "Library", "LaunchAgents")

	if fehler := os.MkdirAll(agenten, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	t.Setenv("HOME", heim)

	laufwerke := filepath.Join(t.TempDir(), "Volumes")
	if fehler := os.MkdirAll(laufwerke, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.Symlink(wurzel, filepath.Join(laufwerke, "Macintosh HD")); fehler != nil {
		t.Fatal(fehler)
	}

	ueberDienst := MitDienstmodus(t.Context(), domain.ModusDienst)
	durchDasLaufwerk := filepath.Join(laufwerke, "Macintosh HD", "Users", "ich", "Library", "LaunchAgents")

	if fehler := pruefeOrdnerUeber(ueberDienst, durchDasLaufwerk, laufwerke); fehler == nil {
		t.Error("über den Laufwerksverweis kam ~/Library/LaunchAgents durch")
	}

	// **Ein echtes Laufwerk bleibt offen.**
	stick := filepath.Join(laufwerke, "Stick")
	if fehler := os.MkdirAll(stick, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := pruefeOrdnerUeber(ueberDienst, stick, laufwerke); fehler != nil {
		t.Errorf("ein echtes Laufwerk: %v", fehler)
	}
}
