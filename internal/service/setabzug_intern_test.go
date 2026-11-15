package service

// **Was sich im Bauort nicht lesen lässt, geht nicht ins Archiv** — NET-04
// (#1708). Der Weg, den [Setschreiber.Abzug] von außen nicht betreten kann:
// unter einem eben angelegten Bauort scheitert ein Lesen nur, wenn jemand es
// scheitern lässt.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEinUnlesbarerBauortGibtNichts(t *testing.T) {
	t.Parallel()

	bau := t.TempDir()
	if fehler := os.WriteFile(filepath.Join(bau, "set"), []byte("x"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	klemmt := func(string) ([]byte, error) { return nil, errors.New("klemmt") }

	archiv := filepath.Join(t.TempDir(), "a.mse-set")

	if fehler := packen(bau, archiv, klemmt); fehler == nil {
		t.Error("mit einem klemmenden Leser kam ein Archiv zustande")
	}

	if _, fehler := os.Stat(archiv); fehler == nil {
		t.Error("es liegt ein Archiv da")
	}

	if _, fehler := einsammeln(filepath.Join(bau, "fehlt"), os.ReadFile); fehler == nil {
		t.Error("ein Bauort, den es nicht gibt, kam durch")
	}

	// **Die Gegenprobe**: mit dem echten Leser kommt die Datei an.
	if dateien, fehler := einsammeln(bau, os.ReadFile); fehler != nil || string(dateien["set"]) != "x" {
		t.Errorf("gelesen: %v, %v", dateien, fehler)
	}
}
