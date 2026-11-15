package app_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Rückfall auf die Sammeldatei** —
// #888.
func TestBulkrueckfall(t *testing.T) {
	t.Run("ohne Datei sagt er, wie sie herkommt", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		quelle := app.Bulkrueckfall()
		if quelle == nil {
			t.Fatal("ohne Datei kam gar keine Quelle")
		}

		_, _, fehler := quelle.KarteNachName(probekontext(t), "Prüfstein")
		if fehler == nil {
			t.Fatal("die fehlende Datei blieb ohne Fehler")
		}

		// **Die Meldung nennt den Befehl.** „Es liegt keine Sammeldatei
		// bereit" ohne das Wie ist eine Sackgasse — dieselbe Überlegung wie
		// bei `nichtSpeicherbar` in `internal/service`.
		if !strings.Contains(fehler.Error(), "scryfall-sammeldaten") {
			t.Errorf("Fehler = %v — er sagt nicht, wie die Datei herkommt", fehler)
		}
	})

	t.Run("mit Datei sucht er darin", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		pfad, fehler := app.Bulkpfad()
		if fehler != nil {
			t.Fatalf("Bulkpfad: %v", fehler)
		}

		if err := os.WriteFile(pfad, einePackung(t,
			`{"object":"card","id":"a","name":"Prüfstein","artist":"Wer auch immer"}`,
		), 0o600); err != nil {
			t.Fatalf("ablegen: %v", err)
		}

		karte, _, holfehler := app.Bulkrueckfall().KarteNachName(probekontext(t), "prüfstein")
		if holfehler != nil {
			t.Fatalf("KarteNachName: %v", holfehler)
		}

		if karte.Name != "Prüfstein" {
			t.Errorf("Karte = %+v", karte)
		}
	})

	t.Run("ohne Zuhause gibt es keinen Rückfall", func(t *testing.T) {
		t.Setenv("HOME", "")
		t.Setenv("MSE_BESTAND", "")

		if quelle := app.Bulkrueckfall(); quelle != nil {
			t.Error("ohne Zuhause kam eine Quelle")
		}
	})

	t.Run("ein belegter Ordner wird gemeldet", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		ort, fehler := app.Scryfallverzeichnis()
		if fehler != nil {
			t.Fatalf("Scryfallverzeichnis: %v", fehler)
		}

		if err := os.WriteFile(
			filepath.Join(ort, "sammeldaten"), []byte("keine Ordner"), 0o600); err != nil {
			t.Fatalf("belegen: %v", err)
		}

		if _, err := app.Bulkverzeichnis(); err == nil {
			t.Error("der belegte Ordner blieb ohne Fehler")
		}

		if _, err := app.Bulkpfad(); err == nil {
			t.Error("Bulkpfad meldete den belegten Ordner nicht")
		}

		// **Und der Rückfall gibt dann nichts** — statt auf einen Pfad zu
		// zeigen, den es nicht gibt.
		if quelle := app.Bulkrueckfall(); quelle != nil {
			t.Error("der belegte Ordner ergab trotzdem eine Quelle")
		}
	})

	t.Run("eine unlesbare Datei kommt als Fehler durch", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		pfad, fehler := app.Bulkpfad()
		if fehler != nil {
			t.Fatalf("Bulkpfad: %v", fehler)
		}

		if err := os.WriteFile(pfad, []byte("kein gzip"), 0o600); err != nil {
			t.Fatalf("ablegen: %v", err)
		}

		if _, _, err := app.Bulkrueckfall().KarteNachName(
			probekontext(t), "Prüfstein"); err == nil {
			t.Error("die unlesbare Datei blieb ohne Fehler")
		} else if !domainfehler(err) {
			t.Errorf("Fehler = %v", err)
		}
	})
}

// domainfehler sagt, ob es ein eingeordneter Fehler ist.
func domainfehler(fehler error) bool {
	return errors.Is(fehler, domain.ErrValidation) || errors.Is(fehler, domain.ErrNotFound)
}

// einePackung packt Zeilen als `.jsonl.gz`.
func einePackung(t *testing.T, zeilen ...string) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	schreiber := gzip.NewWriter(&gepackt)
	for _, zeile := range zeilen {
		if _, fehler := schreiber.Write([]byte(zeile + "\n")); fehler != nil {
			t.Fatalf("packen: %v", fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return gepackt.Bytes()
}
