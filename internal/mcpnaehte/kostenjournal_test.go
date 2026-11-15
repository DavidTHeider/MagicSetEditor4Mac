package mcpnaehte

// **Die Verdrahtung des Kostenjournals über MCP** — Werkstück O12 von Phase
// 5.2 (#1439).

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDieKostenauskunftErreichtIhreNaht(t *testing.T) {
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

	stand, err := auskunft.Kostenstand(probekontext(t), "anbieter", "monat")
	if err != nil {
		t.Fatalf("der Stand: %v", err)
	}

	if stand.Konto != "anbieter" || stand.Zeitraum != "monat" {
		t.Errorf("der Stand ist %+v", stand)
	}

	// **Ein erfundener Zeitraum kommt als Fehler** — aus dem Fachkern.
	if _, err := auskunft.Kostenstand(probekontext(t), "anbieter", "stunde"); err == nil {
		t.Error("ein erfundener Zeitraum kam durch")
	}
}
