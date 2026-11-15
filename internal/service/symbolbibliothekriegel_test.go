package service_test

// **Der Riegel vor jeder Datei** — DOM-02
// (#1697),
// nach Vermerk 0078.
//
// **Einsetzen, Nachziehen und Ausgeben schreiben Dateien**, und ein Leser hat
// nur eine lesende Verbindung (`query_only`) — die hält Dateien nicht auf.
// Nachziehen ändert gar keinen Wert im Bestand: ohne Riegel schriebe ein
// Leser die Symbole fremder Sets um. Gefragt wird deshalb **vor** der ersten
// Datei.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerRiegelHaeltJedeDateiAuf(t *testing.T) {
	bibliothek, _, werte := bibliotheksplatz(t)
	krone := kroneAufgenommen(t, bibliothek)

	if _, err := bibliothek.Speichern(t.Context(), krone.ID, verschoben()); err != nil {
		t.Fatalf("speichern: %v", err)
	}

	datei := werte.werte[1]["symbol"]
	bibliothek.MitPruefer(verweigerndeWache{})

	ziel := filepath.Join(t.TempDir(), "krone.mse-symbol")

	for name, ruf := range map[string]func() error{
		"Einsetzen":  func() error { _, e := bibliothek.Einsetzen(t.Context(), krone.ID, 2); return e },
		"Nachziehen": func() error { _, e := bibliothek.Nachziehen(t.Context(), krone.ID); return e },
		"Ausgeben":   func() error { return bibliothek.Ausgeben(t.Context(), krone.ID, ziel) },
	} {
		if err := ruf(); !errors.Is(err, domain.ErrErlaubnis) {
			t.Errorf("%s: %v", name, err)
		}
	}

	if _, err := os.Stat(ziel); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("die Ausgabe liegt doch da: %v", err)
	}

	if werte.werte[2]["symbol"] != "" || werte.werte[1]["symbol"] != datei {
		t.Errorf("die Werte: %v", werte.werte)
	}

	// **Die Liste bleibt offen** — lesen darf, wer nur liest.
	if _, err := bibliothek.Liste(t.Context(), ""); err != nil {
		t.Errorf("die Liste: %v", err)
	}
}
