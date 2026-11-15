package laufbeleg

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Markierte findet die Tests, die erst ein Tag in den Lauf bringt.
//
// # Erwartet ist, was der Tag einschließt
//
// **Eine Datei ist markiert, wenn sie mit dem Tag übersetzt wird und ohne
// ihn nicht** — entschieden von `go/build` und damit nach derselben Regel,
// nach der `go test` sie auswählt: Ausdruck, Betriebssystem, Dateiname. Eine
// Datei für `plan9` ist hier nicht markiert und eine mit `darwin || linux`
// auch nicht: sie liefe ohne den Tag ebenso.
//
// **Berechnet und nicht eingetragen**: eine Liste, die jemand pflegt, vergäße
// den nächsten Lauf, und der fehlte dann still (AGENTS.md §10).
//
// **Übergangen wird, was nicht zum Modul gehört** — ein Verzeichnis mit
// eigenem `go.mod`, `testdata`, `vendor`, `node_modules` und alles, was mit
// `.` oder `_` beginnt. So übergeht es auch `go` selbst.
func Markierte(wurzel, tag string) ([]Fall, error) {
	modul, fehler := Modul(wurzel)
	if fehler != nil {
		return nil, fehler
	}

	mit := build.Default
	mit.BuildTags = append([]string{tag}, build.Default.BuildTags...)
	ohne := build.Default

	faelle := []Fall{}

	fehler = filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil {
			return fehler
		}

		if eintrag.IsDir() {
			if pfad != wurzel && uebergehen(pfad, eintrag.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		neue, fehler := markierteIn(pfad, eintrag.Name(), mit, ohne)
		if fehler != nil {
			return fehler
		}

		relativ, _ := filepath.Rel(wurzel, filepath.Dir(pfad)) //nolint:errcheck // pfad liegt unter wurzel, WalkDir hat ihn von dort
		paket := modul

		if relativ != "." {
			paket += "/" + filepath.ToSlash(relativ)
		}

		for _, test := range neue {
			faelle = append(faelle, Fall{Paket: paket, Test: test})
		}

		return nil
	})
	if fehler != nil {
		return nil, fmt.Errorf("die markierten Fälle unter %s: %w", wurzel, fehler)
	}

	sort.Slice(faelle, func(i, j int) bool {
		if faelle[i].Paket != faelle[j].Paket {
			return faelle[i].Paket < faelle[j].Paket
		}

		return faelle[i].Test < faelle[j].Test
	})

	return faelle, nil
}

// Modul liest den Modulpfad aus dem `go.mod` der Wurzel.
func Modul(wurzel string) (string, error) {
	inhalt, fehler := os.ReadFile(filepath.Join(wurzel, "go.mod")) // #nosec G304 -- die Wurzel nennt der Aufruf
	if fehler != nil {
		return "", fmt.Errorf("kein go.mod unter %s: %w", wurzel, fehler)
	}

	for _, zeile := range strings.Split(string(inhalt), "\n") {
		if felder := strings.Fields(zeile); len(felder) >= 2 && felder[0] == "module" {
			return strings.Trim(felder[1], `"`), nil
		}
	}

	return "", fmt.Errorf("das go.mod unter %s nennt kein Modul", wurzel)
}

// uebergehen sagt, ob ein Verzeichnis nicht zum Modul gehört.
func uebergehen(pfad, name string) bool {
	switch {
	case strings.HasPrefix(name, "."), strings.HasPrefix(name, "_"),
		name == "testdata", name == "vendor", name == "node_modules":
		return true
	default:
		_, fehler := os.Stat(filepath.Join(pfad, "go.mod"))

		return fehler == nil
	}
}

// markierteIn gibt die Tests einer Datei, wenn erst der Tag sie einschließt.
func markierteIn(pfad, name string, mit, ohne build.Context) ([]string, error) {
	if !strings.HasSuffix(name, "_test.go") {
		return nil, nil
	}

	verzeichnis := filepath.Dir(pfad)

	drin, fehler := mit.MatchFile(verzeichnis, name)
	if fehler != nil {
		return nil, fehler
	}

	sonst, _ := ohne.MatchFile(verzeichnis, name) //nolint:errcheck // dieselbe Datei hat eben schon gelesen werden können
	if !drin || sonst {
		return nil, nil
	}

	return testsIn(pfad)
}

// testsIn gibt die Tests einer Datei — die Funktionen, die `go test` als
// Test ruft.
func testsIn(pfad string) ([]string, error) {
	datei, fehler := parser.ParseFile(token.NewFileSet(), pfad, nil, parser.SkipObjectResolution)
	if fehler != nil {
		return nil, fehler
	}

	testing := "testing"

	for _, eingebunden := range datei.Imports {
		if weg, _ := strconv.Unquote(eingebunden.Path.Value); weg == "testing" && eingebunden.Name != nil { //nolint:errcheck // der Zerteiler hat den Pfad eben als Zeichenkette gelesen
			testing = eingebunden.Name.Name
		}
	}

	tests := []string{}

	for _, deklaration := range datei.Decls {
		funktion, ist := deklaration.(*ast.FuncDecl)
		if ist && funktion.Recv == nil && istTestname(funktion.Name.Name) &&
			nimmtT(funktion.Type, testing) {
			tests = append(tests, funktion.Name.Name)
		}
	}

	return tests, nil
}

// istTestname folgt der Regel von `go test`: `Test`, dahinter nichts oder
// etwas, das nicht mit einem Kleinbuchstaben beginnt.
func istTestname(name string) bool {
	rest, ist := strings.CutPrefix(name, "Test")
	if !ist || rest == "" {
		return ist
	}

	zeichen, _ := utf8.DecodeRuneInString(rest)

	return !unicode.IsLower(zeichen)
}

// nimmtT sagt, ob eine Funktion genau ein `*testing.T` nimmt und nichts
// zurückgibt.
func nimmtT(typ *ast.FuncType, testing string) bool {
	if typ.Results != nil || len(typ.Params.List) != 1 || len(typ.Params.List[0].Names) > 1 {
		return false
	}

	zeiger, ist := typ.Params.List[0].Type.(*ast.StarExpr)
	if !ist {
		return false
	}

	auswahl, ist := zeiger.X.(*ast.SelectorExpr)
	if !ist {
		return false
	}

	paket, ist := auswahl.X.(*ast.Ident)

	return ist && paket.Name == testing && auswahl.Sel.Name == "T"
}
