// Der Wächter über das Protokoll.
//
// # Zwei Zusagen, und die zweite braucht die erste
//
//  1. **`log/slog` steht nur an angemeldeten Stellen.** Ein Protokollaufruf
//     irgendwo im Baum ginge an der Regel vorbei, bevor sie greift.
//  2. **Jede Protokollzeile trägt eine feste Meldung und angemeldete
//     Merkmale.** Eine gerechnete Meldung — `"Karte " + name + " gespeichert"`
//     — käme durch [app.geschwaerzt] hindurch, denn der Handler sieht nur die
//     Merkmale, nicht wie die Meldung entstanden ist.
//
// # Was er nicht prüft
//
// **Ob der Wert zu seinem Namen passt.** Ein `ort`, unter dem ein Kartenname
// steht, kommt durch — hier wie im Handler. Eine Maschine unterscheidet einen
// getippten Kartennamen nicht von einem getippten Dateinamen; das ist eine
// Frage ans Review (§10).
//
// **Testdateien.** Dort steht absichtlich, was nicht durchkommen soll.
package werkzeuge

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// schreibmethoden sind die Aufrufe, die eine Zeile erzeugen.
var schreibmethoden = map[string]bool{
	"Debug": true, "Info": true, "Warn": true, "Error": true,
	"DebugContext": true, "InfoContext": true,
	"WarnContext": true, "ErrorContext": true,
	"With": true,
}

func TestNurAngemeldeteStellenProtokollieren(t *testing.T) {
	var fremd []string

	if err := filepath.WalkDir("..", func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() ||
			!strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
			return err
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus WalkDir unter ..
		if lesefehler != nil {
			return lesefehler
		}

		if strings.Contains(string(roh), `"log/slog"`) && !angemeldet(pfad) {
			fremd = append(fremd, pfad)
		}

		return nil
	}); err != nil {
		t.Fatalf("gehen: %v", err)
	}

	// **`main.go` liegt über `internal/` und wird eigens gelesen.**
	if roh, err := os.ReadFile("../../main.go"); err == nil &&
		strings.Contains(string(roh), `"log/slog"`) && !angemeldet("../../main.go") {
		fremd = append(fremd, "../../main.go")
	}

	if len(fremd) > 0 {
		t.Errorf("diese Dateien protokollieren, ohne angemeldet zu sein: %s\n"+
			"Eintragen in protokollstellen — mit Grund.", strings.Join(fremd, ", "))
	}
}

// angemeldet sagt, ob eine Datei protokollieren darf.
func angemeldet(pfad string) bool {
	for stelle := range protokollstellen {
		if filepath.Clean(stelle) == filepath.Clean(pfad) {
			return true
		}
	}

	return false
}

// TestJedeProtokollzeileIstErklaert liest die angemeldeten Stellen.
func TestJedeProtokollzeileIstErklaert(t *testing.T) {
	erlaubt := angemeldeteMerkmale(t)
	if len(erlaubt) < 5 {
		t.Fatalf("%d Merkmale gefunden — der Leser greift daneben: %v", len(erlaubt), erlaubt)
	}

	var befunde []string

	for pfad := range protokollstellen {
		baum, err := parser.ParseFile(token.NewFileSet(), pfad, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("%s lesen: %v", pfad, err)
		}

		ast.Inspect(baum, func(knoten ast.Node) bool {
			ruf, ist := knoten.(*ast.CallExpr)
			if !ist {
				return true
			}

			wahl, ist := ruf.Fun.(*ast.SelectorExpr)
			if !ist || !schreibmethoden[wahl.Sel.Name] {
				return true
			}

			befunde = append(befunde, pruefeRuf(pfad, wahl.Sel.Name, ruf.Args, erlaubt)...)

			return true
		})
	}

	if len(befunde) > 0 {
		t.Errorf("Protokollzeilen, die nicht erklärt sind:\n  %s",
			strings.Join(befunde, "\n  "))
	}
}

// pruefeRuf prüft Meldung und Merkmalsnamen eines einzelnen Aufrufs.
func pruefeRuf(pfad, methode string, args []ast.Expr, erlaubt map[string]bool) []string {
	var befunde []string

	// **`…Context` trägt den Kontext vor der Meldung** — und deshalb steht
	// die Meldung dort an zweiter Stelle.
	//
	// **Das stand hier schon, und zwar nur für die Merkmale**: die Meldung
	// wurde unbedingt an `args[0]` gesucht. Aufgefallen ist es nie, weil bis
	// zum Werkstück O9 (#1433) **keine einzige** Stelle die Kontextform
	// benutzte — der Wächter hielt die halbe Zusage, und die andere Hälfte
	// war unsichtbar (§10). Mit der Vorgangskennung wurden es neun Stellen,
	// und alle neun standen sofort als Befund da.
	meldung := 0
	if strings.HasSuffix(methode, "Context") {
		meldung = 1
	}

	// **`With` trägt keine Meldung**, alle anderen eine.
	erstesMerkmal := meldung
	if methode != "With" {
		if len(args) <= meldung {
			return nil
		}

		if !istZeichenkette(args[meldung]) {
			befunde = append(befunde,
				pfad+": die Meldung von "+methode+" ist kein Zeichenkettenliteral")
		}

		erstesMerkmal = meldung + 1
	}

	for stelle := erstesMerkmal; stelle < len(args); stelle += 2 {
		name, ist := zeichenkettenwert(args[stelle])
		if !ist {
			befunde = append(befunde,
				pfad+": ein Merkmalsname ist kein Zeichenkettenliteral")

			continue
		}

		if !erlaubt[name] {
			befunde = append(befunde, pfad+": das Merkmal "+strconv.Quote(name)+
				" ist nicht angemeldet")
		}
	}

	return befunde
}

func istZeichenkette(ausdruck ast.Expr) bool {
	_, ist := zeichenkettenwert(ausdruck)

	return ist
}

func zeichenkettenwert(ausdruck ast.Expr) (string, bool) {
	wert, ist := ausdruck.(*ast.BasicLit)
	if !ist || wert.Kind != token.STRING {
		return "", false
	}

	entpackt, err := strconv.Unquote(wert.Value)

	return entpackt, err == nil
}

// merkmalszeile trifft einen Eintrag der Liste in `protokollmerkmale.go`.
var merkmalszeile = regexp.MustCompile(`(?m)^\s*"([a-zäöüß]+)":\s*"`)

// angemeldeteMerkmale liest die Liste aus dem Quelltext.
//
// **Aus dem Quelltext und nicht über einen Export.** `erlaubteMerkmale` gehört
// zu `internal/app` und soll dort unexportiert bleiben; sie hier auszustellen
// hieße, eine innere Regel zur Schnittstelle zu machen, damit ein Test sie
// lesen kann.
func angemeldeteMerkmale(t *testing.T) map[string]bool {
	t.Helper()

	roh, err := os.ReadFile("../app/protokollmerkmale.go") //nolint:gosec // ein fester Pfad
	if err != nil {
		t.Fatalf("Merkmale lesen: %v", err)
	}

	namen := map[string]bool{}
	for _, treffer := range merkmalszeile.FindAllStringSubmatch(string(roh), -1) {
		namen[treffer[1]] = true
	}

	return namen
}
