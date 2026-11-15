// **Ein eigener Wegwerfort je Bau** — gefunden, als `make ci` drei Kopien
// derselben Saat nebeneinander anlegte (FMT-02,
// #1691).
//
// Kopie und Neuanlage bauten ihr Paket unter einem **festen** Namen im
// gemeinsamen Temp-Verzeichnis — `mse4-kopie-<ziel>` und `mse4-anlage-<name>`.
// Zwei Prozesse mit demselben Ziel räumten einander den Ordner weg; im
// Betrieb sind das Fenster und Kommandozeile, im Test drei Flächen zugleich.
// Ein fremder Rest unter dem alten Namen belegt dasselbe deterministisch.
package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEinFremderRestAmWegwerfortHaeltNichtsAuf(t *testing.T) {
	ort, bau := t.TempDir(), t.TempDir()

	for _, rest := range []string{"mse4-kopie-probe-kopie.mse-game", "mse4-anlage-eigen.mse-game"} {
		if err := os.WriteFile(filepath.Join(bau, rest), []byte("fremd"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	verwalter := NeuerPaketverwalter(&kopierablage{}, ort, filepath.Join("..", "..", "saat"), ort)
	verwalter.bau = bau

	if _, err := verwalter.Auffrischen(t.Context()); err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if _, err := verwalter.Kopieren(t.Context(), "mse-game", "probe", ""); err != nil {
		t.Errorf("kopieren: %v", err)
	}

	if _, err := verwalter.Anlegen(t.Context(), Neuanlage{Art: "mse-game", Name: "eigen", Vollname: "Eigen"}); err != nil {
		t.Errorf("anlegen: %v", err)
	}

	// **Und hinterher liegt nur der fremde Rest da** — der eigene Ort ist fort.
	if eintraege, err := os.ReadDir(bau); err != nil || len(eintraege) != 2 {
		t.Errorf("%d Einträge am Wegwerfort (%v), erwartet die zwei fremden", len(eintraege), err)
	}
}

// **Und ein Schreibfehler in der Kopie wird gemeldet** — über den Griff; der
// feste Pfad, an dem ihn ein Test früher erzwang, ist fort.
func TestEinSchreibfehlerInDerKopieWirdGemeldet(t *testing.T) {
	ort := t.TempDir()

	verwalter := NeuerPaketverwalter(&kopierablage{}, ort, filepath.Join("..", "..", "saat"), ort)
	verwalter.bau = t.TempDir()
	verwalter.ablegen = func(string, []byte) error { return errors.New("die Platte klemmt") }

	if _, err := verwalter.Auffrischen(t.Context()); err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if _, err := verwalter.Kopieren(t.Context(), "mse-game", "probe", ""); err == nil ||
		!strings.Contains(err.Error(), "die Kopie schreiben") {
		t.Errorf("ein Schreibfehler kam durch: %v", err)
	}
}
