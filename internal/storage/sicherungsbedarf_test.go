package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// **Was eine Sicherung ungefähr braucht** — M02
// (#2346):
// die belegten Seiten, denn `VACUUM INTO` schreibt die freien nicht mit. Die
// Sicherung, die danach entsteht, ist hier nicht größer — für diesen Bestand
// gemessen, nicht für jeden zugesagt.
func TestSicherungsbedarfZaehltDieBelegtenSeiten(t *testing.T) {
	t.Parallel()

	db, _ := aufDerPlatte(t)
	ctx := probekontext(t)

	bedarf := Sicherungsbedarf(ctx, db)
	if bedarf <= 0 {
		t.Fatalf("ein migrierter Bestand braucht %d Bytes", bedarf)
	}

	ziel := filepath.Join(t.TempDir(), "sicherung.db")
	if fehler := Sichern(ctx, db, ziel); fehler != nil {
		t.Fatal(fehler)
	}

	geschrieben, fehler := os.Stat(ziel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if geschrieben.Size() > bedarf {
		t.Errorf("die Sicherung belegt %d Bytes, angesagt waren %d", geschrieben.Size(), bedarf)
	}

	// **Und wer es nicht erfährt, sagt 0** — kein Fehler: der Bedarf ist eine
	// Auskunft für die Einordnung eines anderen.
	if fehler := db.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if zu := Sicherungsbedarf(ctx, db); zu != 0 {
		t.Errorf("eine geschlossene Datenbank braucht %d Bytes", zu)
	}
}
