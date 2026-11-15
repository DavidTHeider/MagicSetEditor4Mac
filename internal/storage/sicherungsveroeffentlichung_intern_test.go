package storage

// **Die Fehlerwege des Veröffentlichens** — BK-06a
// (#2130).
// Eine gewöhnliche Platte gibt sie im Ablauf von `Sichern` nicht her; hier
// werden sie einzeln betreten.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestEinScheiternderVerweisRaeumtDieNachbardateiWeg(t *testing.T) {
	verzeichnis := t.TempDir()
	zwischen := filepath.Join(verzeichnis, ".s.entsteht-1")

	if err := os.WriteFile(zwischen, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	// **Ein Ziel in einem Verzeichnis, das es nicht gibt** — `link(2)` meldet
	// etwas anderes als „schon da".
	err := veroeffentlichen(zwischen, filepath.Join(verzeichnis, "fehlt", "s.sqlite"))
	if err == nil || errors.Is(err, fs.ErrExist) {
		t.Fatalf("erwartet war ein Fehler beim Veröffentlichen, gekommen ist %v", err)
	}

	if _, fehlt := os.Stat(zwischen); !errors.Is(fehlt, fs.ErrNotExist) {
		t.Errorf("die Nachbardatei liegt noch da: %v", fehlt)
	}
}

func TestFestschreibenMeldetEinenFehlendenPfad(t *testing.T) {
	if err := festschreiben(filepath.Join(t.TempDir(), "fehlt")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("erwartet war ErrNotExist, gekommen ist %v", err)
	}
}
