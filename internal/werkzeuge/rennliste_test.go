package werkzeuge_test

// **Wer ohne Race-Detektor läuft, hat keine Goroutine** — die Zusage hinter
// `OHNE_RENNEN` in `scripts/ci.sh`.
//
// # Warum es diesen Wächter gibt
//
// Am 2026-09-11 kostete `go test -race` über alles **42 Minuten**, und
// `internal/storage` allein 1357 Sekunden gegen 28,5 ohne Detektor — das
// 48-fache. Der Grund ist nachlesbar: SQLite läuft hier als WebAssembly
// (`ncruces/go-sqlite3` über wazero), und der Detektor instrumentiert jeden
// Speicherzugriff des Interpreters. Er misst wazero, nicht diesen Code.
//
// **Der Preis dafür ist eine Zusage**, und die hält niemand von Hand: sobald
// eines dieser Pakete eine Goroutine oder ein `sync.`-Mittel im Produktivcode
// bekommt, ist der Grund fort — und ein Wettlauf darin bliebe unentdeckt.
//
// # Was er nicht prüft
//
// **Ob ein Wettlauf im Testcode steckt.** Die Tests dieser Pakete laufen
// nebenläufig (`t.Parallel`), und ein Doppel mit gemeinsamem Zustand ist
// weiterhin möglich. Was er prüft, ist die Aussage über den **Produktivcode**,
// auf der die Ausnahme beruht.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rennfrei ist, was im Produktivcode eines ausgenommenen Pakets nicht stehen
// darf.
//
// **`go func` und `sync.`** — die zwei Gestalten, in denen Nebenläufigkeit
// anfängt. Ein `chan` allein genügt nicht: einer, den niemand aus einer
// zweiten Goroutine bedient, ist eine Warteschlange.
var rennfrei = regexp.MustCompile(`\bgo func\b|\bgo [a-zA-Z_][a-zA-Z0-9_.]*\(|\bsync\.`)

// ohneRennenAus liest die Liste aus `scripts/ci.sh`.
//
// **Aus dem Skript und nicht aus einer Kopie daneben**: eine zweite
// Aufzählung liefe auseinander, und der Wächter prüfte dann eine Liste, die
// nicht gilt (§10).
func ohneRennenAus(t *testing.T) []string {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join("..", "..", "scripts", "ci.sh"))
	if fehler != nil {
		t.Fatalf("ci.sh lesen: %v", fehler)
	}

	for zeile := range strings.SplitSeq(string(roh), "\n") {
		rest, gefunden := strings.CutPrefix(zeile, `OHNE_RENNEN="`)
		if !gefunden {
			continue
		}

		return strings.Fields(strings.TrimSuffix(rest, `"`))
	}

	t.Fatal("ci.sh nennt kein OHNE_RENNEN mehr — " +
		"ein Wächter, der nichts findet, hält nichts")

	return nil
}

func TestWerOhneDetektorLaeuftHatKeineGoroutine(t *testing.T) {
	t.Parallel()

	muster := ohneRennenAus(t)
	if len(muster) == 0 {
		t.Fatal("OHNE_RENNEN ist leer")
	}

	for _, eines := range muster {
		verzeichnis := filepath.Join("..", "..",
			strings.TrimSuffix(strings.TrimPrefix(eines, "./"), "/..."))

		t.Run(eines, func(t *testing.T) {
			t.Parallel()

			for _, fund := range goroutinenIn(t, verzeichnis) {
				t.Errorf("%s läuft ohne Race-Detektor und benutzt "+
					"Nebenläufigkeit: %s", eines, fund)
			}
		})
	}
}

// goroutinenIn nennt die Produktivdateien mit Nebenläufigkeit.
//
// **Rekursiv**, weil `./…/...` auch Unterpakete meint.
func goroutinenIn(t *testing.T, verzeichnis string) []string {
	t.Helper()

	var funde []string

	fehler := filepath.WalkDir(verzeichnis, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") ||
			strings.HasSuffix(pfad, "_test.go") {
			return nil
		}

		// #nosec G122,G304 -- ein Pfad, den WalkDir über diesem Baum liefert;
		// eine Symlink-Falle darin wäre bereits ein Fund an sich
		roh, lesefehler := os.ReadFile(pfad)
		if lesefehler != nil {
			return lesefehler
		}

		if stelle := rennfrei.FindString(ohneKommentare(string(roh))); stelle != "" {
			funde = append(funde, pfad+" ("+stelle+")")
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s durchsehen: %v", verzeichnis, fehler)
	}

	return funde
}
