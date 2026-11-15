package service

// **Eine Änderung zwischen Lesen und Schreiben gewinnt** — Z4 des
// Zweitblicks über Phase 6.12
// (#1905).
//
// Gelesen wurde einmal für den Stand und einmal für die Änderung, und
// geschrieben wurde ohne Vergleich: wer die Datei dazwischen änderte, verlor
// seine Änderung, und ein Widerruf konnte eine eben ergänzte Datei löschen.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// rennziele gibt je Art mit Datei ein Ziel samt dem Inhalt davor.
func rennziele(t *testing.T) map[string]Schnittstellenziel {
	t.Helper()

	ordner := t.TempDir()

	return map[string]Schnittstellenziel{
		"Block": {
			Name: "befehlszeile", Datei: filepath.Join(ordner, ".zprofile"), Vorhanden: true,
			Block: `export PATH="$HOME/.local/bin:$PATH"` + "\n",
		},
		"JSON": {
			Name: "claudedesktop", Art: ArtJSON, Datei: filepath.Join(ordner, "claude.json"), Vorhanden: true,
			Schluessel: "magicseteditor4mac", Programm: "/Programme/mse-mcp",
			Gedaechtnis: filepath.Join(ordner, "gedaechtnis.json"),
		},
	}
}

func TestEineAenderungZwischenLesenUndSchreibenGewinnt(t *testing.T) {
	t.Parallel()

	for name, ziel := range rennziele(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rennen(t, ziel, []byte("{}\n"))
		})
	}
}

// **Auch wo widerrufen die Datei wegnähme** — sie entstand mit dem
// Einrichten, und eine Änderung dazwischen hält das Löschen an.
func TestEineAenderungHaeltAuchDasLoeschenAn(t *testing.T) {
	t.Parallel()

	for name, ziel := range rennziele(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rennen(t, ziel, nil)
		})
	}
}

// rennen stellt einen fremden Schreiber zwischen Lesen und Schreiben —
// beim Einrichten und beim Widerrufen. `vorher` nil heißt: ohne Datei.
func rennen(t *testing.T, ziel Schnittstellenziel, vorher []byte) {
	t.Helper()

	legeVorher(t, ziel.Datei, vorher)

	fremd := []byte("{\"jemand\": \"anderes\"}\n")
	ziel.zwischendurch = func() {
		if fehler := os.WriteFile(ziel.Datei, fremd, 0o600); fehler != nil {
			t.Error(fehler)
		}
	}

	if _, fehler := SchnittstelleEinrichten(t.Context(), ziel); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("einrichten: %v", fehler)
	}

	if inhalt, _ := os.ReadFile(ziel.Datei); !bytes.Equal(inhalt, fremd) { //nolint:errcheck,gosec // die Datei des Laufs
		t.Errorf("die fremde Änderung ist fort: %s", inhalt)
	}

	// **Und beim Widerrufen dasselbe** — erst ohne Rennen einrichten,
	// dann dazwischen ändern.
	ziel.zwischendurch = nil
	legeVorher(t, ziel.Datei, vorher)

	if _, fehler := SchnittstelleEinrichten(t.Context(), ziel); fehler != nil {
		t.Fatalf("einrichten ohne Rennen: %v", fehler)
	}

	ziel.zwischendurch = func() {
		if fehler := os.WriteFile(ziel.Datei, fremd, 0o600); fehler != nil {
			t.Error(fehler)
		}
	}

	if _, fehler := SchnittstelleWiderrufen(t.Context(), ziel); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("widerrufen: %v", fehler)
	}

	if inhalt, _ := os.ReadFile(ziel.Datei); !bytes.Equal(inhalt, fremd) { //nolint:errcheck,gosec // die Datei des Laufs
		t.Errorf("die fremde Änderung ist beim Widerrufen fort: %s", inhalt)
	}
}

// legeVorher legt den Inhalt davor an — oder räumt die Datei weg.
func legeVorher(t *testing.T, datei string, vorher []byte) {
	t.Helper()

	if vorher == nil {
		if fehler := os.Remove(datei); fehler != nil && !errors.Is(fehler, os.ErrNotExist) {
			t.Fatal(fehler)
		}

		return
	}

	if fehler := os.WriteFile(datei, vorher, 0o600); fehler != nil {
		t.Fatal(fehler)
	}
}
