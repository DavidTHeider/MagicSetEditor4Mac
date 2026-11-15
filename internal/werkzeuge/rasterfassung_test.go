package werkzeuge_test

// **Die Fassung des Rasterers steigt, wenn er sich ändert** — Werkstück G2
// von Phase 5.5
// (#1459).
//
// # Wofür es diesen Wächter gibt
//
// `render.Fassung` geht in den Schlüssel jedes gerasterten Spielbildes ein.
// Bleibt sie stehen, während sich der Rasterer ändert, zeigt der Tisch Bilder
// von gestern — und zwar **still**: sie sehen aus wie Bilder.
//
// Eine Zahl, an die sich jemand erinnern muss, ist keine Zusage (§10). Dieser
// Wächter macht sie zu einer: er hält den Quelltext von `internal/render`
// gegen einen festgehaltenen Abdruck.
//
// # Was er nicht hält
//
//   - **Ob die Änderung das Bild wirklich ändert.** Ein umbenannter Hilfsname
//     zwingt zur neuen Fassung, obwohl kein Pixel anders fällt. Das ist die
//     teure Richtung, und sie ist gewählt: eine Fassung, die zu oft steigt,
//     kostet einen Cache; eine, die zu selten steigt, zeigt falsche Bilder.
//   - **Kommentare zählen nicht.** Sie werden vor dem Abdruck entfernt —
//     sonst kostete jede erklärende Zeile einen Cachelauf, und das erzöge zum
//     Schweigen.
//   - **Testdateien zählen nicht.** Sie zeichnen nichts.

import (
	"crypto/sha256"
	"encoding/hex"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// rasterabdruck ist der Abdruck von `internal/render` zur Fassung darunter.
//
// **Wer den Rasterer ändert, ändert beides** — `render.Fassung` und diese
// Zeile. Die Meldung des Laufs nennt den neuen Abdruck.
const rasterabdruck = "1d6a24b4c254c8f8917df006e10cc3b1a9ff89545d4b7bf54f96ad3d897f4560"

// rasterfassung ist die Fassung, zu der [rasterabdruck] gehört.
const rasterfassung = "7"

// **Der Rasterer ist noch der, zu dem seine Fassung gehört.**
func TestDieRasterfassungPasstZumRasterer(t *testing.T) {
	t.Parallel()

	if render.Fassung != rasterfassung {
		t.Fatalf("render.Fassung ist %q, dieser Wächter kennt %q — wer die "+
			"Fassung hebt, trägt sie hier ein und erneuert den Abdruck",
			render.Fassung, rasterfassung)
	}

	gezaehlt, dateien := abdruckVonRender(t)

	if gezaehlt != rasterabdruck {
		t.Errorf("`internal/render` hat sich geändert (%d Dateien):\n"+
			"  Abdruck jetzt: %s\n"+
			"  eingetragen:   %s\n"+
			"Wer den Rasterer ändert, hebt `render.Fassung` **und** trägt den "+
			"neuen Abdruck hier ein — sonst zeigt der Spieltisch Bilder von "+
			"gestern (G2, #1459)", dateien, gezaehlt, rasterabdruck)
	}
}

// abdruckVonRender gibt den Abdruck des Rasterers und die Zahl seiner Dateien.
//
// **Über den Syntaxbaum und ohne Kommentare** — der Grund steht im Kopf.
func abdruckVonRender(t *testing.T) (abdruck string, dateien int) {
	t.Helper()

	eintraege, err := os.ReadDir(filepath.Join("..", "render"))
	if err != nil {
		t.Fatalf("`internal/render` lesen: %v", err)
	}

	namen := make([]string, 0, len(eintraege))

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if eintrag.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		namen = append(namen, name)
	}

	// **Eine leere Liste macht diesen Wächter hohl** — dieselbe Untergrenze
	// wie überall in diesem Verzeichnis (§10).
	if len(namen) < 5 {
		t.Fatalf("nur %d Quelldateien in `internal/render` — der Wächter "+
			"greift ins Leere", len(namen))
	}

	slices.Sort(namen)

	summe := sha256.New()

	for _, name := range namen {
		summe.Write([]byte(name))
		summe.Write(rasterquelle(t, filepath.Join("..", "render", name)))
	}

	return hex.EncodeToString(summe.Sum(nil)), len(namen)
}

// rasterquelle gibt eine Quelldatei ohne ihre Kommentare.
//
// **Eigen und nicht [ohneKommentare] daneben**: jene nimmt einen Text und
// streicht Kommentarzeilen heraus; diese liest den **Syntaxbaum** und druckt
// ihn neu — damit ändert auch eine andere Einrückung den Abdruck nicht.
func rasterquelle(t *testing.T, pfad string) []byte {
	t.Helper()

	stellen := token.NewFileSet()

	baum, err := parser.ParseFile(stellen, pfad, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	baum.Comments = nil

	var gedruckt strings.Builder
	if err := (&printer.Config{Mode: printer.RawFormat}).
		Fprint(&gedruckt, stellen, baum); err != nil {
		t.Fatalf("%s schreiben: %v", pfad, err)
	}

	return []byte(gedruckt.String())
}
