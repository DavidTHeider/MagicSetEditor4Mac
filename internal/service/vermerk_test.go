package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Der Vermerk nach einem Bezug (#56).
//
// **Eine eigene Datei**, weil `paketverwalter_test.go` an der Zeilengrenze
// steht (§2) — und weil es eine andere Frage ist: dort das Verwalten, hier das
// Festhalten, was ankam.

// **Was bezogen wurde, bekommt Herkunft und Prüfsumme** (#56).
//
// Über den **Ort** und nicht über einen Pfad je Paket: ein Bezug holt ein
// ganzes Repositorium, und die Summe gilt für alles darin.
func TestVermerkenSchreibtHerkunftUndPruefsummeAnAllePakete(t *testing.T) {
	wurzel := t.TempDir()
	bezogen := filepath.Join(wurzel, "M15")

	if err := os.MkdirAll(bezogen, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	paket(t, bezogen, "eins.mse-style", "mse version: 2.0.0\n")
	paket(t, bezogen, "zwei.mse-game", "mse version: 2.0.0\n")
	paket(t, wurzel, "danebenliegend.mse-style", "mse version: 2.0.0\n")

	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, wurzel, bezogen, wurzel)

	befunde, err := verwalter.Vermerken(t.Context(), service.Vermerk{
		Ort:        bezogen,
		Herkunft:   "https://example.invalid/pack@3.1.0",
		Pruefsumme: "sha256:abc",
	})
	if err != nil {
		t.Fatalf("vermerken: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	vermerkt := 0

	for _, gelesen := range katalog.pakete {
		switch gelesen.Name {
		case "eins", "zwei":
			vermerkt++

			if gelesen.Pruefsumme != "sha256:abc" ||
				gelesen.Herkunft != "https://example.invalid/pack@3.1.0" {
				t.Errorf("%q: %+v", gelesen.Name, gelesen)
			}

		case "danebenliegend":
			// **Was nicht unter dem Ort liegt, bekommt nichts.** Eine Summe an
			// einem Paket, das aus einem anderen Bezug stammt, wäre eine
			// Falschangabe.
			if gelesen.Pruefsumme != "" || gelesen.Herkunft != "" {
				t.Errorf("danebenliegend hat einen fremden Vermerk: %+v", gelesen)
			}
		}
	}

	if vermerkt != 2 {
		t.Errorf("%d von 2 Paketen vermerkt: %+v", vermerkt, katalog.pakete)
	}
}

// **Ein Fehler des Bestands kommt auch beim Vermerken durch.**
func TestVermerkenMeldetDenBestandsfehler(t *testing.T) {
	kaputt := errors.New("der Katalog schweigt")
	verwalter := service.NeuerPaketverwalter(&ablage{fehler: kaputt}, t.TempDir(), t.TempDir())

	if _, err := verwalter.Vermerken(t.Context(), service.Vermerk{}); !errors.Is(err, kaputt) {
		t.Errorf("Fehler: %v", err)
	}
}
