package storage_test

// **Ein neuer Bestand gehört seinem Eigner** —
// #2508,
// gefunden beim Erststart aus dem Abbild (HAND6): `bestand.db` entstand mit
// `0644`, ein eingesetzter mit `0600` (Vermerk 0174).

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// rechteBeimSchreiben öffnet den Bestand, schreibt einmal und gibt die
// Rechte der Datei und ihrer Begleiter, **solange er offen ist**.
func rechteBeimSchreiben(t *testing.T, pfad string) map[string]fs.FileMode {
	t.Helper()

	db, fehler := storage.Oeffnen(probekontext(t), pfad)
	if fehler != nil {
		t.Fatalf("Oeffnen: %v", fehler)
	}

	t.Cleanup(func() {
		_ = db.Close() //nolint:errcheck // nur ein Lauf
	})

	if _, fehler := db.ExecContext(probekontext(t), "CREATE TABLE probe (wert TEXT)"); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	rechte := map[string]fs.FileMode{}

	for _, name := range []string{"", "-wal", "-shm"} {
		info, fehler := os.Stat(pfad + name)
		if fehler != nil {
			t.Fatalf("%s fehlt: %v", filepath.Base(pfad+name), fehler)
		}

		rechte[name] = info.Mode().Perm()
	}

	return rechte
}

func TestEinNeuerBestandGehoertSeinemEigner(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	for name, modus := range rechteBeimSchreiben(t, pfad) {
		if modus != 0o600 {
			t.Errorf("bestand.db%s entstand mit %04o", name, modus)
		}
	}
}

// **Ein vorhandener behält seine Rechte** — kein `chmod` fremder Dateien
// (Vermerk 0174), und die Begleiter folgen ihm, wie bei SQLite selbst.
func TestEinVorhandenerBestandBehaeltSeineRechte(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")
	//nolint:gosec // G306: gerade diese Rechte, die nicht 0600 sind, prüft der Lauf
	if fehler := os.WriteFile(pfad, nil, 0o640); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	//nolint:gosec // G302: dito, auch gegen die umask
	if fehler := os.Chmod(pfad, 0o640); fehler != nil {
		t.Fatalf("Rechte: %v", fehler)
	}

	for name, modus := range rechteBeimSchreiben(t, pfad) {
		if modus != 0o640 {
			t.Errorf("bestand.db%s hat %04o statt 0640", name, modus)
		}
	}
}
