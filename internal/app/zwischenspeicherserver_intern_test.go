package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// TestEinServerHaeltDieGrenzeAbSeinemStart hält Nachlauf über Phase 6.4, N9:
// die Grenze galt nur beim Start des Fensters — ein `mse-rest`, das Wochen
// lief, räumte nie. **Dieselbe Pflege an seinem Start**, mit den
// Einstellungen des gezeigten Profils.
func TestEinServerHaeltDieGrenzeAbSeinemStart(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort := filepath.Join(t.TempDir(), "bestand.db")

	db, fehler := storage.Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if migriert := storage.Migriere(t.Context(), db, nil); migriert != nil {
		t.Fatal(migriert)
	}

	gezeigt, fehler := storage.MitGezeigtemProfil(t.Context(), db)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gesetzt := einstellungsverwalterUeber(db).Setzen(gezeigt, domain.ZwischenspeicherGrenze, "1"); gesetzt != nil {
		t.Fatal(gesetzt)
	}

	speicher, fehler := Scryfallverzeichnis()
	if fehler != nil {
		t.Fatal(fehler)
	}

	bilder := filepath.Join(speicher, "bilder")
	if fehler := os.MkdirAll(bilder, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	for stelle, name := range []string{"alt.png", "neu.png"} {
		pfad := filepath.Join(bilder, name)
		if fehler := os.WriteFile(pfad, make([]byte, 700<<10), 0o600); fehler != nil {
			t.Fatal(fehler)
		}

		zeit := time.Now().Add(time.Duration(stelle-2) * time.Hour)
		if fehler := os.Chtimes(pfad, zeit, zeit); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if fehler := ZwischenspeicherPflegen(t.Context(), db); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := os.Stat(filepath.Join(bilder, "alt.png")); !os.IsNotExist(fehler) {
		t.Errorf("das ältere Bild steht noch: %v", fehler)
	}

	if _, fehler := os.Stat(filepath.Join(bilder, "neu.png")); fehler != nil {
		t.Errorf("das neuere ist fort: %v", fehler)
	}
}

// **Ein Testlauf löscht nur in seiner Zwischenablage** — ein Lauf, der das
// Zuhause nicht umlenkt, darf den echten Speicher dieses Rechners nicht
// beschneiden.
func TestEinTestlaufRaeumtNurInSeinerZwischenablage(t *testing.T) {
	t.Parallel()

	if raeumbar("/Users/jemand/Library/Caches/MagicSetEditor4Mac/scryfall") {
		t.Error("der echte Speicher gilt als räumbar")
	}

	if !raeumbar(filepath.Join(t.TempDir(), "scryfall")) {
		t.Error("die Zwischenablage des Laufs gilt nicht als räumbar")
	}
}
