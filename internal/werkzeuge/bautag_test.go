package werkzeuge_test

// **Wann eine Testdatei zu einem Lauf gehört** — nach derselben Regel, nach
// der `go test` sie übersetzt (`go/build/constraint`), und nicht nach der
// Schreibweise ihrer ersten Zeile.
//
// Die Wächter über die Pläne in `docs/belege` lasen bis F04 (#2246) nur
// `//go:build <tag>` wörtlich: eine Datei mit `netzlauf || dienstdrift` stand
// für keinen von beiden hinter dem Tag, und eine hinter einem fremden Tag
// (`fremdpakete_test.go`) zählte für jeden Plan ihres Pakets mit.

import (
	"go/build/constraint"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// uebersetztMit sagt, ob `go test -tags <tag>` auf diesem Rechner die Datei
// übersetzt. Eine Datei ohne Ausdruck gehört immer dazu.
func uebersetztMit(quelle, tag string) bool {
	ausdruck := bauausdruck(quelle)
	if ausdruck == nil {
		return true
	}

	return ausdruck.Eval(func(name string) bool {
		return name == tag || name == "darwin" || name == "unix" ||
			name == "arm64" || name == "gc"
	})
}

// hinterTag sagt, ob die Datei **erst** mit dem Tag übersetzt wird — die
// Regel von `laufbeleg.Markierte`.
func hinterTag(quelle, tag string) bool {
	return uebersetztMit(quelle, tag) && !uebersetztMit(quelle, "")
}

// bauausdruck liest die Zeile `//go:build` vor `package`.
func bauausdruck(quelle string) constraint.Expr {
	for zeile := range strings.Lines(quelle) {
		zeile = strings.TrimSpace(zeile)
		if strings.HasPrefix(zeile, "package ") {
			return nil
		}

		if !constraint.IsGoBuild(zeile) {
			continue
		}

		ausdruck, fehler := constraint.Parse(zeile)
		if fehler != nil {
			return nil
		}

		return ausdruck
	}

	return nil
}

// **Jeder Driftlauf steht auch hinter `netzlauf`** (F04, #2246).
//
// Gelintet werden die Tags, die `scripts/ci-lint.sh` und
// `scripts/vorpruefung.sh` aufzählen, und `dienstdrift` steht nicht darin:
// er wählt nur aus, was hinter `netzlauf` ohnehin übersetzt und gelintet
// wird. Eine Datei nur hinter `dienstdrift` bliebe ungeprüft.
func TestDienstdriftStehtAuchHinterNetzlauf(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")
	gefunden := 0

	for _, teil := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join(wurzel, teil), func(pfad string, eintrag fs.DirEntry, fehler error) error {
			if fehler != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, "_test.go") {
				return fehler
			}

			roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Testdateien im Baum
			if lesefehler != nil {
				return lesefehler
			}

			if !hinterTag(string(roh), "dienstdrift") {
				return nil
			}

			gefunden++

			if !hinterTag(string(roh), "netzlauf") {
				t.Errorf("%s steht hinter dienstdrift und nicht hinter netzlauf — kein Lint übersetzt sie", pfad)
			}

			return nil
		})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", teil, fehler)
		}
	}

	// **Der Zuhörer** (§3): ohne eine einzige Datei prüfte die Schleife nichts.
	if gefunden == 0 {
		t.Fatal("keine Datei steht hinter dienstdrift — der Leser greift daneben")
	}
}
