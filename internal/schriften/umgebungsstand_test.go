package schriften

// **Der Stand der Umgebung, lesbar** — PERF02
// (#2352):
// eine Problemprüfung gilt nur, solange Schriften und Pakete dieselben sind.
// Eine Datei, die dazukommt, ändert den Stand; ein bloßes Lesen nicht.

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDerStandFolgtSchriftenUndPaketen(t *testing.T) {
	t.Parallel()

	orte, pakete := t.TempDir(), t.TempDir()
	suche := Neue([]string{orte}).MitPaketen(pakete)

	vorher := suche.Stand()
	if vorher != suche.Stand() {
		t.Fatal("zweimal gelesen, zwei Stände")
	}

	// **Die Änderungszeit eines Verzeichnisses hat Sekunden** auf manchen
	// Dateisystemen — gesetzt statt gewartet.
	spaeter := time.Now().Add(time.Hour)
	if err := os.WriteFile(filepath.Join(orte, "neu.ttf"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chtimes(orte, spaeter, spaeter); err != nil {
		t.Fatal(err)
	}

	nachSchrift := suche.Stand()
	if nachSchrift == vorher {
		t.Error("eine neue Schrift änderte den Stand nicht")
	}

	if err := os.Chtimes(pakete, spaeter.Add(time.Hour), spaeter.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	if suche.Stand() == nachSchrift {
		t.Error("ein geändertes Paketverzeichnis änderte den Stand nicht")
	}
}
