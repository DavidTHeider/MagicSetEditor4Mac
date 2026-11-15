package werkzeuge

// **Jede Senke unter `#nosec G304` steht im Inventar, und jede Zeile dort hat
// ihre Senke** — SEC-01
// (#2136).
//
// Gezählt wird je Datei und Funktion, nicht je Zeile: Zeilennummern wandern
// mit jeder Änderung darüber, und eine Funktion mit zwei Senken hat zwei
// Zeilen im Inventar.
//
// # Was er nicht hält
//
// **Ob die Klasse einer Senke stimmt.** Ein Aufrufer, der eine Senke einer
// Fläche öffnet, ändert ihre Klasse, und das sieht dieser Wächter nicht —
// dafür steht `pfadstrukturen_test.go` an den Flächen.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// g304unterdrueckung erkennt eine Senke — `#nosec G304` mit Begründung oder mit
// weiteren Regeln, nicht einen Verweis im Fließtext.
var g304unterdrueckung = regexp.MustCompile(`#nosec G304( --|,)`)

var g304zeile = regexp.MustCompile("^\\| `([^`]+\\.go)` · `(\\w+)` \\|")

func TestJedeG304SenkeStehtImInventar(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")

	roh, fehler := os.ReadFile(filepath.Join(wurzel, "docs", "G304INVENTAR.md")) //nolint:gosec // ein fester Ort im eigenen Repositorium
	if fehler != nil {
		t.Fatal(fehler)
	}

	var inventar []string

	for zeile := range strings.SplitSeq(string(roh), "\n") {
		if treffer := g304zeile.FindStringSubmatch(zeile); treffer != nil {
			inventar = append(inventar, treffer[1]+" · "+treffer[2])
		}
	}

	var senken []string

	for _, ordner := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join(wurzel, ordner), func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
				return err
			}

			gefunden, lesefehler := g304SenkenIn(pfad, wurzel)
			senken = append(senken, gefunden...)

			return lesefehler
		})
		if fehler != nil {
			t.Fatal(fehler)
		}
	}

	slices.Sort(inventar)
	slices.Sort(senken)

	for _, senke := range senken {
		if !slices.Contains(inventar, senke) {
			t.Errorf("%s trägt `#nosec G304` und steht nicht in docs/G304INVENTAR.md", senke)
		}
	}

	for _, zeile := range inventar {
		if !slices.Contains(senken, zeile) {
			t.Errorf("docs/G304INVENTAR.md nennt %s — dort steht keine Senke mehr", zeile)
		}
	}

	if !slices.Equal(inventar, senken) && !t.Failed() {
		t.Errorf("gleiche Namen, andere Zahl: Inventar %d, Quelltext %d", len(inventar), len(senken))
	}
}

// senkenIn gibt je Senke einer Datei „Datei · Funktion" — die Funktion, in
// der der Kommentar steht, oder die Variable, deren Wert eine ist.
func g304SenkenIn(pfad, wurzel string) ([]string, error) {
	menge := token.NewFileSet()

	datei, fehler := parser.ParseFile(menge, pfad, nil, parser.ParseComments)
	if fehler != nil {
		return nil, fehler
	}

	relativ, fehler := filepath.Rel(wurzel, pfad)
	if fehler != nil {
		return nil, fehler
	}

	var senken []string

	for _, gruppe := range datei.Comments {
		for _, kommentar := range gruppe.List {
			if !g304unterdrueckung.MatchString(kommentar.Text) {
				continue
			}

			senken = append(senken, filepath.ToSlash(relativ)+" · "+g304Umschliessend(datei, kommentar.Pos()))
		}
	}

	return senken, nil
}

// umschliessend nennt die Deklaration, in der eine Stelle steht.
func g304Umschliessend(datei *ast.File, stelle token.Pos) string {
	for _, deklaration := range datei.Decls {
		if stelle < deklaration.Pos() || stelle > deklaration.End() {
			continue
		}

		switch d := deklaration.(type) {
		case *ast.FuncDecl:
			return d.Name.Name
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if wert, ist := spec.(*ast.ValueSpec); ist && stelle >= wert.Pos() && stelle <= wert.End() {
					return wert.Names[0].Name
				}
			}
		}
	}

	return "?"
}
