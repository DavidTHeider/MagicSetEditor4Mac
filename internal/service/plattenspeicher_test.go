package service_test

// **Ein Zwischenspeicher auf der Platte hat eine Grenze und lässt sich
// leeren** — OPS-08.AK-03
// (#1725):
// Belegung nennen, die ältesten Dateien über der Grenze wegnehmen, alles
// wegnehmen — und dabei **nur**, was das Muster als Eintrag erkennt, nie eine
// fremde Datei im selben Ordner.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinPlattenspeicherHaeltSeineGrenzeUndLeertSich(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	jetzt := time.Date(2026, time.September, 26, 12, 0, 0, 0, time.UTC)

	for stelle, name := range []string{"alt.json", "mitte.json", "neu.json", "notiz.txt"} {
		pfad := filepath.Join(ort, name)
		if fehler := os.WriteFile(pfad, []byte(strings.Repeat("x", 100)), 0o600); fehler != nil {
			t.Fatal(fehler)
		}

		zeit := jetzt.Add(time.Duration(stelle) * time.Minute)
		if fehler := os.Chtimes(pfad, zeit, zeit); fehler != nil {
			t.Fatal(fehler)
		}
	}

	speicher := service.Plattenspeicher{
		Ordner: ort, Eintrag: func(name string) bool { return strings.HasSuffix(name, ".json") },
	}

	if belegung := speicher.Belegung(); belegung.Eintraege != 3 || belegung.Bytes != 300 {
		t.Fatalf("Belegung %+v", belegung)
	}

	// **Die Uhr der Platte ordnet** — der älteste geht zuerst.
	if weg := speicher.Beschneiden(200); weg != 1 {
		t.Errorf("%d weggenommen", weg)
	}

	if _, fehler := os.Stat(filepath.Join(ort, "alt.json")); !os.IsNotExist(fehler) {
		t.Errorf("der älteste steht noch da: %v", fehler)
	}

	if weg := speicher.Leeren(); weg != 2 || speicher.Belegung().Eintraege != 0 {
		t.Errorf("%d geleert, danach %+v", weg, speicher.Belegung())
	}

	if _, fehler := os.Stat(filepath.Join(ort, "notiz.txt")); fehler != nil {
		t.Errorf("die fremde Datei ist fort: %v", fehler)
	}

	// **Ein Ordner, den es nicht gibt, ist ein leerer Speicher** — kein Fehler.
	leer := service.Plattenspeicher{Ordner: filepath.Join(ort, "fehlt"), Eintrag: func(string) bool { return true }}
	if leer.Belegung().Eintraege != 0 || leer.Beschneiden(0) != 0 || leer.Leeren() != 0 {
		t.Error("ein fehlender Ordner ist kein leerer Speicher")
	}
}
