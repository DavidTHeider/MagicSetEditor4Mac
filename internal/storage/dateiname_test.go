// **Ein Dateipfad mit Sonderzeichen erreicht SQLite unzerschnitten**
// (Zweitblick über Phase S).
package storage_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **`AlsURI` kodiert genau die drei Zeichen, die SQLite nennt.**
func TestAlsURIKodiertWasSQLiteVerlangt(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"/ort/schlicht.db":  "file:/ort/schlicht.db",
		"/ort/mit?frage.db": "file:/ort/mit%3ffrage.db",
		"/ort/mit#raute.db": "file:/ort/mit%23raute.db",
		"/ort/mit%satz.db":  "file:/ort/mit%25satz.db",

		// **Die Reihenfolge zählt.** Wer erst `?` ersetzte und danach jedes
		// `%` verdoppelte, machte daraus `%253f`.
		"/ort/beides?%.db": "file:/ort/beides%3f%25.db",
	}

	for pfad, erwartet := range faelle {
		t.Run(pfad, func(t *testing.T) {
			t.Parallel()

			if got := storage.AlsURI(pfad); got != erwartet {
				t.Errorf("AlsURI(%q) = %q, erwartet war %q", pfad, got, erwartet)
			}
		})
	}
}

// **Und `Oeffnen` legt keine zweite Datei daneben.**
//
// Das ist der Schaden, um den es geht: SQLite scheitert an einem
// zerschnittenen Pfad nicht, sondern **legt an**, was davor steht.
func TestOeffnenLegtKeineDateiNebenDenPfad(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	pfad := filepath.Join(ort, "bestand?eins.db")

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	eintraege, err := os.ReadDir(ort)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for _, eintrag := range eintraege {
		if !strings.HasPrefix(eintrag.Name(), "bestand?eins.db") {
			t.Errorf("neben dem Bestand liegt %q — der Pfad wurde zerschnitten", eintrag.Name())
		}
	}
}
