// **Die Zahlen im Bestand sind gezählt** —
// #1291,
// Werkstück A.
//
// # Warum es diesen Wächter gibt
//
// `docs/MSE2BESTAND.md` nennt in einer Übersicht, wie viele Zeilen ein ✓, ein
// — und ein ? tragen. **Diese drei Zahlen standen unbewacht**: am 2026-08-30
// ist eine Marke der Tabelle absichtlich verfälscht worden, und der ganze Lauf
// blieb grün. Das ist genau die Gestalt, gegen die Phase 4.0 angetreten ist —
// keine Zahl ohne Wächter.
//
// **Er steht nicht hinter `fremdpakete`**, anders als `mse2bestand_test.go`:
// er liest nur diese eine Datei gegen sich selbst und braucht den fremden
// Quelltext nicht.
//
// # Was er hält
//
//   - Die drei Zahlen der Übersicht stimmen mit den Marken der Tabellen.
//   - Die Summe ist die Zeilenzahl — eine vierte Marke fiele sonst heraus.
//
// # Was er **nicht** hält
//
// **Ob eine Marke verdient ist.** Ob ein ✓ zu Recht dasteht, sagt kein
// Wächter; das liest ein Mensch (§10).
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// markenzeile trifft eine Zeile der Bestandstabellen: `| `befehl` | … | ✓ | …`.
//
// **Ein eigenes Muster neben `bestandszeile`** aus `mse2bestand_test.go`: das
// steht hinter `fremdpakete` und wäre hier nicht da.
var markenzeile = regexp.MustCompile("^\\| `[^`]+` \\|")

// bestandszahl trifft eine Zeile der Übersicht: `| ✓ | … | **44** |`.
var bestandszahl = regexp.MustCompile(`^\| (✓|—|\?) \| [^|]+ \| \*\*(\d+)\*\* \|`)

// TestDieBestandszahlenSindGezaehlt hält die Übersicht gegen die Tabellen.
func TestDieBestandszahlenSindGezaehlt(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "MSE2BESTAND.md"))
	if err != nil {
		t.Fatalf("den Bestand lesen: %v", err)
	}

	gezaehlt := map[string]int{"✓": 0, "—": 0, "?": 0}
	genannt := map[string]int{}
	zeilen := 0

	for _, zeile := range strings.Split(string(roh), "\n") {
		if treffer := bestandszahl.FindStringSubmatch(zeile); treffer != nil {
			zahl, umwandlungsfehler := strconv.Atoi(treffer[2])
			if umwandlungsfehler != nil {
				t.Fatalf("%q ist keine Zahl: %v", treffer[2], umwandlungsfehler)
			}

			genannt[treffer[1]] = zahl

			continue
		}

		if !markenzeile.MatchString(zeile) {
			continue
		}

		zeilen++

		spalten := strings.Split(zeile, "|")
		if len(spalten) < 5 {
			t.Errorf("eine Zeile ohne Marke: %s", zeile)

			continue
		}

		marke := strings.TrimSpace(spalten[4])
		if _, bekannt := gezaehlt[marke]; !bekannt {
			t.Errorf("unbekannte Marke %q in: %s", marke, zeile)

			continue
		}

		gezaehlt[marke]++
	}

	// **Ohne diese Zeile wäre der Lauf grün, wenn kein Muster greift** — die
	// Gegenprobe zu sich selbst.
	if zeilen == 0 {
		t.Fatal("keine einzige Bestandszeile gefunden — das Muster greift nicht mehr")
	}

	summe := 0

	for marke, will := range gezaehlt {
		summe += will

		if hat, genanntWorden := genannt[marke]; !genanntWorden {
			t.Errorf("die Übersicht nennt %q nicht mehr", marke)
		} else if hat != will {
			t.Errorf("die Übersicht nennt für %q die Zahl %d — gezählt sind %d",
				marke, hat, will)
		}
	}

	if summe != zeilen {
		t.Errorf("die Marken zählen %d, die Tabellen haben %d Zeilen", summe, zeilen)
	}
}
