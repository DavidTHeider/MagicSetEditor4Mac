package werkzeuge

// **Der Einstellungsverwalter entsteht an einer Stelle** — OPS-02.AK-04
// (#1719):
// „alle Flächen verwenden denselben Pfadvertrag".
//
// # Was er hält
//
// **Außerhalb der Tests ruft genau eine Datei `NeuerEinstellungsverwalter`**
// — `internal/app/einstellungsverwalter.go`, die die Ordnerprüfung
// einsetzt. Eine zweite Baustelle bekäme sie nicht, und ihre Fläche nähme
// einen Ordner an, den die anderen abweisen. Bis OPS-02 waren es sechs.
//
// # Was er nicht hält
//
// **Ob die eine Stelle die Prüfung wirklich einsetzt** — das belegt
// `internal/app/ablageordner_test.go` am Fenster.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDerEinstellungsverwalterEntstehtAnEinerStelle(t *testing.T) {
	t.Parallel()

	stellen := []string{}

	for _, baum := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join(wurzel, baum), func(pfad string, eintrag fs.DirEntry, fehler error) error {
			if fehler != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
				return fehler
			}

			roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelltext dieses Repositoriums
			if lesefehler != nil {
				return lesefehler
			}

			if strings.Contains(string(roh), "service.NeuerEinstellungsverwalter(") {
				stellen = append(stellen, strings.TrimPrefix(pfad, wurzel+"/"))
			}

			return nil
		})
		if fehler != nil {
			t.Fatal(fehler)
		}
	}

	if len(stellen) != 1 || stellen[0] != "internal/app/einstellungsverwalter.go" {
		t.Errorf("NeuerEinstellungsverwalter steht in %v — gebaut wird über einstellungsverwalterUeber", stellen)
	}
}
