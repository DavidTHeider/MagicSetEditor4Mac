package werkzeuge

// **Wer die Zuteilung des Prozesses misst, läuft nicht parallel** — gefunden
// von der Pipeline über den Nachlauf von Phase 6.1.
//
// # Der Befund, für den es diesen Wächter gibt
//
// `runtime.ReadMemStats` zählt, was der **ganze Prozess** zugeteilt hat. Zwei
// Läufe aus dem Nachlauf maßen damit, ob eine Datei ganz gelesen wird, und
// riefen `t.Parallel()` — damit zählte jeder Test mit, der daneben lief. Die
// älteren Läufe dieser Art sagen es längst („Nicht parallel, und das ist
// Absicht", `msepaket/paketgrenzen_intern_test.go`); die neuen wussten es
// nicht.
//
// **Rot wurde die Pipeline an einem von ihnen aus einem zweiten Grund**, und
// der gehört dazugesagt: die Diagnose teilte unter `-race` 69 MiB für ein
// Fenster von 16 MiB zu, auch allein gemessen (behoben in
// `internal/app/diagnoseprotokoll.go`). Ohne diesen Wächter wäre das nicht
// zu unterscheiden gewesen — ein roter Lauf hätte ebenso gut den Nachbarn
// messen können.
//
// # Was er hält
//
// **Keine Testfunktion unter `internal/` und `cmd/`, die misst, ruft
// `Parallel()`** — auch nicht in einem ihrer Unterläufe. Gemessen hat, wer
// `runtime.ReadMemStats` ruft oder einen Helfer im selben Paket, der es tut.
// Parallele Tests warten, bis die nicht parallelen fertig sind; ein Lauf
// ohne `t.Parallel()` misst damit allein.
//
// # Was er nicht hält
//
// Einen Helfer in einem **anderen** Paket, und Goroutinen, die ein früherer
// Test liegen ließ.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestWerDieZuteilungMisstLaeuftNichtParallel(t *testing.T) {
	t.Parallel()

	pakete := map[string]map[string]*ast.File{}
	menge := token.NewFileSet()

	for _, wurzelverzeichnis := range []string{"internal", "cmd"} {
		fehler := filepath.WalkDir(filepath.Join(wurzel, wurzelverzeichnis),
			func(pfad string, eintrag fs.DirEntry, err error) error {
				if err != nil || eintrag.IsDir() || !strings.HasSuffix(pfad, "_test.go") {
					return err
				}

				baum, err := parser.ParseFile(menge, pfad, nil, parser.SkipObjectResolution)
				if err != nil {
					return err
				}

				ordner, err := filepath.Rel(wurzel, filepath.Dir(pfad))
				if err != nil {
					return err
				}

				if pakete[ordner] == nil {
					pakete[ordner] = map[string]*ast.File{}
				}

				pakete[ordner][filepath.Base(pfad)] = baum

				return nil
			})
		if fehler != nil {
			t.Fatal(fehler)
		}
	}

	var befunde []string

	for ordner, dateien := range pakete {
		for _, befund := range zuteilungsbefunde(dateien) {
			befunde = append(befunde, ordner+"/"+befund)
		}
	}

	sort.Strings(befunde)

	for _, befund := range befunde {
		t.Errorf("%s misst die Zuteilung des Prozesses und läuft parallel — "+
			"andere Tests zählen mit", befund)
	}
}

// TestDerZuteilungswaechterWirdRot ist die Gegenprobe (§10): eine messende
// Testfunktion mit `t.Parallel()`, eine mit einem parallelen Unterlauf und
// eine, die über einen Helfer misst.
func TestDerZuteilungswaechterWirdRot(t *testing.T) {
	t.Parallel()

	quelle := `package probe

func messen() uint64 {
	var stand runtime.MemStats
	runtime.ReadMemStats(&stand)
	return stand.TotalAlloc
}

func TestSelbst(t *testing.T) {
	t.Parallel()
	var stand runtime.MemStats
	runtime.ReadMemStats(&stand)
}

func TestUnterlauf(t *testing.T) {
	var stand runtime.MemStats
	runtime.ReadMemStats(&stand)
	t.Run("x", func(t *testing.T) { t.Parallel() })
}

func TestUeberHelfer(t *testing.T) {
	t.Parallel()
	_ = messen()
}

func TestAllein(t *testing.T) {
	_ = messen()
}

func TestOhneMessung(t *testing.T) {
	t.Parallel()
}
`

	baum, err := parser.ParseFile(token.NewFileSet(), "probe_test.go", quelle, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}

	befunde := zuteilungsbefunde(map[string]*ast.File{"probe_test.go": baum})
	erwartet := []string{
		"probe_test.go:TestSelbst",
		"probe_test.go:TestUeberHelfer",
		"probe_test.go:TestUnterlauf",
	}

	if strings.Join(befunde, ",") != strings.Join(erwartet, ",") {
		t.Errorf("gefunden %v, erwartet %v", befunde, erwartet)
	}
}

// zuteilungsbefunde nennt die messenden Testfunktionen eines Pakets, die
// parallel laufen — als `datei:Funktion`, sortiert.
func zuteilungsbefunde(dateien map[string]*ast.File) []string {
	// **Erst die Messer, dann wer sie ruft** — bis sich nichts mehr ändert.
	messer := map[string]bool{}

	for veraendert := true; veraendert; {
		veraendert = false

		for _, baum := range dateien {
			for _, erklaerung := range baum.Decls {
				funktion, ist := erklaerung.(*ast.FuncDecl)
				if !ist || funktion.Body == nil || messer[funktion.Name.Name] {
					continue
				}

				if misst(funktion.Body, messer) {
					messer[funktion.Name.Name] = true
					veraendert = true
				}
			}
		}
	}

	var befunde []string

	for name, baum := range dateien {
		for _, erklaerung := range baum.Decls {
			funktion, ist := erklaerung.(*ast.FuncDecl)
			if !ist || funktion.Body == nil || !strings.HasPrefix(funktion.Name.Name, "Test") {
				continue
			}

			if messer[funktion.Name.Name] && rufe(funktion.Body, "Parallel") {
				befunde = append(befunde, name+":"+funktion.Name.Name)
			}
		}
	}

	sort.Strings(befunde)

	return befunde
}

// misst sagt, ob ein Rumpf `runtime.ReadMemStats` ruft — oder einen der
// bekannten Messer.
func misst(rumpf *ast.BlockStmt, messer map[string]bool) bool {
	gefunden := false

	ast.Inspect(rumpf, func(knoten ast.Node) bool {
		ruf, ist := knoten.(*ast.CallExpr)
		if !ist {
			return !gefunden
		}

		switch ziel := ruf.Fun.(type) {
		case *ast.SelectorExpr:
			paket, istName := ziel.X.(*ast.Ident)
			gefunden = gefunden || istName && paket.Name == "runtime" && ziel.Sel.Name == "ReadMemStats"
		case *ast.Ident:
			gefunden = gefunden || messer[ziel.Name]
		}

		return !gefunden
	})

	return gefunden
}

// rufe sagt, ob ein Rumpf irgendwo `….name()` ruft — auch in einem Unterlauf.
func rufe(rumpf *ast.BlockStmt, name string) bool {
	gefunden := false

	ast.Inspect(rumpf, func(knoten ast.Node) bool {
		if ruf, ist := knoten.(*ast.CallExpr); ist {
			if auswahl, istAuswahl := ruf.Fun.(*ast.SelectorExpr); istAuswahl && auswahl.Sel.Name == name {
				gefunden = true
			}
		}

		return !gefunden
	})

	return gefunden
}
