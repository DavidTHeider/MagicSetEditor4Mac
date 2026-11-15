package main

// **Ein migrierter Bestand je Lauf, kopiert je Test** — QA-07b
// (#1968).
//
// # Warum
//
// **Jeder Test legte einen frischen Bestand an und migrierte ihn**, unter
// `-race` und mit Abdeckung, und `cmd/mse` war der lange Pol der Teststufe
// (`docs/PIPELINE.md`, QA-07). Die Wanderungen sind für jeden Test dieselben;
// was sich unterscheidet, ist erst, was der Test danach hineinlegt.
//
// # Was es nicht ändert
//
// **Was geprüft wird.** Die Datei ist Byte für Byte die, die eine Migration
// ergäbe — sie **ist** eine —, und `mitBestand` migriert weiterhin bei jedem
// Befehl; über einem fertigen Bestand ist das ein Blick in die Tafel der
// Wanderungen. Ein Test, der die Migration selbst prüft, legt seinen
// Bestand weiter selbst an.

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// musterordner hält den Musterbestand für diesen Lauf.
var musterordner string

// musterbestand migriert einmal je Lauf.
var musterbestand = sync.OnceValues(func() ([]byte, error) {
	ort := filepath.Join(musterordner, "muster.db")

	db, fehler := storage.Oeffnen(storage.MitProfil(context.Background(), 1), ort)
	if fehler != nil {
		return nil, fehler
	}

	if fehler := storage.Migriere(context.Background(), db, nil); fehler != nil {
		_ = db.Close()

		return nil, fehler
	}

	if fehler := db.Close(); fehler != nil {
		return nil, fehler
	}

	return os.ReadFile(ort) //nolint:gosec // die Datei dieses Laufs
})

// frischerBestand gibt den Ort eines migrierten Bestands, den nur dieser
// Test kennt.
func frischerBestand(t *testing.T) string {
	t.Helper()

	inhalt, fehler := musterbestand()
	if fehler != nil {
		t.Fatalf("der Musterbestand: %v", fehler)
	}

	ort := filepath.Join(t.TempDir(), "bestand.db")
	if fehler := os.WriteFile(ort, inhalt, 0o600); fehler != nil {
		t.Fatalf("den Musterbestand kopieren: %v", fehler)
	}

	return ort
}

// TestMain legt den Ordner des Musterbestands an und räumt ihn weg — **und
// gibt dem ganzen Lauf ein eigenes Zuhause** (A01, #2215).
//
// Ohne es lief `go test ./cmd/mse/` von Hand im echten: achtzehn Läufe
// fanden den Paketort unter dem Zuhause dieses Rechners, und seit
// `internal/app` einem Testlauf dort keinen schreibenden Ort mehr gibt,
// sagen sie es. `scripts/ci.sh` und die Vorprüfung setzen ein Zuhause
// ohnehin; dies ist der Boden für den Lauf von Hand.
func TestMain(m *testing.M) {
	ordner, fehler := os.MkdirTemp("", "mse-muster-")
	if fehler != nil {
		panic(fehler)
	}

	zuhause, fehler := os.MkdirTemp("", "mse-testzuhause-")
	if fehler != nil {
		panic(fehler)
	}

	// **Den Anmeldebund sperrt `internal/schluesselbund` einem Testlauf
	// ohnehin** (§3, an `os.Args[0]`); hier geht es um die Orte darunter.
	if fehler := os.Setenv("HOME", zuhause); fehler != nil {
		panic(fehler)
	}

	musterordner = ordner
	ergebnis := m.Run()

	_ = os.RemoveAll(ordner)  //nolint:errcheck // ein Ordner dieses Laufs; was bleibt, räumt das System
	_ = os.RemoveAll(zuhause) //nolint:errcheck // dito

	os.Exit(ergebnis)
}
