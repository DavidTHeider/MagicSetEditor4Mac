// **Die Verdrahtung des Kostenjournals** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Was dieser Lauf hält
//
// Dass die Wurzel den Weg **erreicht** — nicht, was er tut. Der Arbeitsablauf
// über alle Flächen steht in `internal/flaechen`.
package main

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieDreiKostenwegeErreichenIhreNaht(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	db, fehler := storage.Oeffnen(probekontext(t), filepath.Join(zuhause, "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	auskunft := kostenauskunft{db: db}

	if _, err := auskunft.Kostenstand(probekontext(t), "anbieter", "tag"); err != nil {
		t.Errorf("der Stand: %v", err)
	}

	if _, err := auskunft.Kostenposten(probekontext(t), "anbieter", "tag"); err != nil {
		t.Errorf("die Posten: %v", err)
	}

	if _, err := auskunft.KostenbudgetSetzen(
		probekontext(t), "anbieter", "tag", 5000, "EUR"); err != nil {
		t.Errorf("das Budget: %v", err)
	}

	// **Ein erfundener Zeitraum kommt als Fehler zurück** — aus dem Fachkern
	// und nicht aus der Verdrahtung.
	if _, err := auskunft.Kostenstand(probekontext(t), "anbieter", "stunde"); err == nil {
		t.Error("ein erfundener Zeitraum kam durch")
	}

	if _, err := auskunft.Kostenposten(probekontext(t), "anbieter", "stunde"); err == nil {
		t.Error("ein erfundener Zeitraum kam bei den Posten durch")
	}

	if _, err := auskunft.KostenbudgetSetzen(
		probekontext(t), "anbieter", "stunde", 1, "EUR"); err == nil {
		t.Error("ein erfundener Zeitraum kam beim Budget durch")
	}
}
