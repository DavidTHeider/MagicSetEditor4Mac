package werkzeuge_test

// **Was Go dem Fenster meldet, hört das Fenster unter demselben Namen** —
// Zweitblick über Phase 6.11 (#1672, zu Z1).
//
// # Warum es diesen Wächter gibt
//
// `usePartiegast.ts` versprach in seinem Kopf, „ein Vertragstest" halte
// die Namen der Partieereignisse auf beiden Seiten zusammen. **Es gab
// keinen.** Ein Name, der auf einer Seite anders hieße, fiele still aus:
// `EventsOn` meldet einen Namen, auf den niemand hört, ohne Fehler.
//
// # Was er hält
//
// - Jeder Name, den das Fenster als `export const ereignis… = '…'`
//   führt, steht als Zeichenkette im Go-Quelltext unter `internal/`.
// - Jeder Name, den Go als `Ereignis… = "…"` führt, steht als
//   Zeichenkette im Quelltext des Fensters.
//
// # Was er nicht hält
//
// **Ob die Nutzlast passt** — Reihenfolge und Gestalt der Werte. Das prüfen
// die Läufe an der Naht (`partietischlauf_intern_test.go`, die Hüllenprobe).
// Und ob ein Name, der auf beiden Seiten steht, auch gesendet und gehört
// wird: er sieht Zeichenketten, keine Aufrufe.

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	fensterereignis = regexp.MustCompile(`export const ereignis[A-Za-z]+ = '([^']+)'`)
	goereignis      = regexp.MustCompile(`(?m)^\s*Ereignis[A-Za-z]+\s*=\s*"([^"]+)"`)
)

func TestFensterUndGoNennenDieselbenEreignisse(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")
	gosrc := quelltextUnter(t, filepath.Join(wurzel, "internal"), ".go")
	fenstersrc := quelltextUnter(t, filepath.Join(wurzel, "frontend", "src"), ".ts", ".tsx")

	fenster := namenIn(fensterereignis, fenstersrc)
	gonamen := namenIn(goereignis, gosrc)

	if len(fenster) < 10 || len(gonamen) < 10 {
		t.Fatalf("zu wenige Namen gefunden — Fenster %d, Go %d: liest der Wächter die richtigen Dateien?",
			len(fenster), len(gonamen))
	}

	for _, name := range fenster {
		if !strings.Contains(gosrc, `"`+name+`"`) {
			t.Errorf("das Fenster hört auf %q, Go meldet es nirgends", name)
		}
	}

	for _, name := range gonamen {
		if !strings.Contains(fenstersrc, `'`+name+`'`) {
			t.Errorf("Go meldet %q, das Fenster hört es nirgends", name)
		}
	}
}

// quelltextUnter liest alle Quelltexte mit einer der Endungen — ohne Tests.
func quelltextUnter(t *testing.T, ort string, endungen ...string) string {
	t.Helper()

	var alles strings.Builder

	fehler := filepath.WalkDir(ort, func(pfad string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil || eintrag.IsDir() {
			return fehler
		}

		name := eintrag.Name()
		if strings.Contains(name, "_test.") || strings.Contains(name, ".test.") {
			return nil
		}

		for _, endung := range endungen {
			if strings.HasSuffix(name, endung) {
				roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // unter dem Repositorium, aufgezählt
				if lesefehler != nil {
					return lesefehler
				}

				alles.Write(roh)
				alles.WriteByte('\n')
			}
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s lesen: %v", ort, fehler)
	}

	return alles.String()
}

// namenIn gibt die erste Gruppe jedes Treffers.
func namenIn(muster *regexp.Regexp, text string) []string {
	alle := muster.FindAllStringSubmatch(text, -1)
	namen := make([]string, 0, len(alle))

	for _, treffer := range alle {
		namen = append(namen, treffer[1])
	}

	return namen
}
