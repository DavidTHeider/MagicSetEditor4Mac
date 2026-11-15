package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Auch für Zeichnungen geht der Vorrat einmal aus** —
// #890.
//
// **Paketintern**, weil die Schranke es ist: sie steht bei 100.000, und ein
// Lauf, der sie wirklich ausschöpfte, schriebe hunderttausend Dateien.
// Dieselbe Bauart wie `bildzahl_intern_test.go` — und derselbe Grund: ein
// Zweig, den kein Lauf betritt, ist eine Behauptung.
func TestOhneFreienSymbolnamenGibtEsEinenFehler(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	// **Die Schranke sitzt an der Ablage**, nicht am Paket (#1163): sie war
	// ein Zustand, den alle Läufe teilen — und die Läufe dieses Pakets sind
	// nebenläufig.
	ablage := NeueSymbolablage(NeueBildablage(wurzel).MitHoechsterBildzahl(2))

	for range 2 {
		if _, err := ablage.Ablegen(7, msepaket.Symbolzeichnung{}, ""); err != nil {
			t.Fatalf("Ablegen: %v", err)
		}
	}

	if _, err := ablage.Ablegen(7, msepaket.Symbolzeichnung{}, ""); err == nil {
		t.Error("die dritte Zeichnung bekam noch einen Namen")
	}

	// **Und die beiden ersten stehen wirklich da** — sonst prüfte der Lauf
	// eine Schranke, die aus einem anderen Grund greift.
	for _, name := range []string{"symbol1.mse-symbol", "symbol2.mse-symbol"} {
		if _, err := os.Stat(filepath.Join(wurzel, "7", name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
