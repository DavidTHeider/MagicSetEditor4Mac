package werkzeuge

// **Welche Flächen ein Lauf wirklich fährt** — Zweitblick über Phase C,
// Befund 3 (C04, #2229).
//
// Die Bilanz von `FLAECHENPARITAET.md` zählte, ob einer Klasse ein Lauf
// **zugeordnet** ist: 32 von 32. Gelesen fuhr kein einziger Ergebnislauf das
// Fenster — es kam nur in den Abweisungen vor. Eine Zahl, die eine Zuordnung
// zählt und wie Parität aussieht, ist die Zusage aus §10, die mehr verspricht,
// als ihr Rumpf hält.
//
// # Was er hält
//
// Je Klasse und Spalte die Flächen, die ihre Läufe **ansprechen**, gelesen aus
// dem Quelltext von `internal/flaechen`: von jedem genannten Lauf aus jede
// Funktion des Pakets, die er ruft oder weiterreicht, und darin die Helfer der
// Flächen — `fensterAus` (Fenster), `kommandozeile` (Kommandozeile), `mcp…`
// (MCP), `starteDienst…`, `dienstGelang` und `.ruf…` (Dienst). Der Abschnitt
// „Gemessene Flächen" muss je Klasse dasselbe sagen, und die Bilanz je Fläche
// muss stimmen; im roten Fall nennt er die Zeile, wie sie lauten soll.
//
// # Was er nicht hält
//
// **Ob ein Lauf an jeder angesprochenen Fläche dasselbe vergleicht** — eine
// Fläche, die nur zum Einrichten gerufen wird, zählt mit. Und `app.App` direkt
// (`PeerflaecheUeber`) ist **kein** Fenster: ohne `NeueBindungen` fehlen Tore,
// Rufkontext und die Abbildung von Wails.

import (
	"fmt"
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

// flaechenbuchstaben ordnet die Flächen, wie sie in der Tafel stehen.
var flaechenbuchstaben = []string{"F", "K", "M", "R"}

var (
	gemesseneZeile   = regexp.MustCompile(`(?m)^\| ([^|]+?) \| ((?:[FKMR] ?)+|—) \| ((?:[FKMR] ?)+|—) \| ((?:[FKMR] ?)+|—|entfällt) \|$`)
	flaechenbilanz   = regexp.MustCompile(`(?m)^\| (Fenster|Kommandozeile|MCP|Dienst) \| (\d+) \| (\d+) \| (\d+) \|$`)
	flaechenNamenVon = map[string]string{"F": "Fenster", "K": "Kommandozeile", "M": "MCP", "R": "Dienst"}
)

// flaechenDerLaeufe liest je Funktion in `internal/flaechen` die Flächen, die
// sie selbst oder eine Funktion des Pakets anspricht, die sie ruft oder
// weiterreicht.
func flaechenDerLaeufe(t *testing.T) func(lauf string) map[string]bool {
	t.Helper()

	ordner := filepath.Join("..", "flaechen")
	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", ordner, fehler)
	}

	dateien := token.NewFileSet()
	funktionen := map[string]*ast.FuncDecl{}

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".go") {
			continue
		}

		datei, lesefehler := parser.ParseFile(dateien, filepath.Join(ordner, eintrag.Name()), nil, parser.SkipObjectResolution)
		if lesefehler != nil {
			t.Fatalf("%s: %v", eintrag.Name(), lesefehler)
		}

		for _, deklaration := range datei.Decls {
			if funktion, istFunktion := deklaration.(*ast.FuncDecl); istFunktion && funktion.Recv == nil {
				funktionen[funktion.Name.Name] = funktion
			}
		}
	}

	if len(funktionen) == 0 {
		t.Fatal("in internal/flaechen keine Funktion gefunden — der Wächter greift daneben")
	}

	gemerkt := map[string]map[string]bool{}

	var flaechen func(name string, gesehen map[string]bool) map[string]bool

	flaechen = func(name string, gesehen map[string]bool) map[string]bool {
		if fertig, da := gemerkt[name]; da {
			return fertig
		}

		gefunden := map[string]bool{}
		funktion := funktionen[name]

		if funktion == nil || gesehen[name] {
			return gefunden
		}

		gesehen[name] = true

		ast.Inspect(funktion, func(knoten ast.Node) bool {
			switch k := knoten.(type) {
			case *ast.CallExpr:
				for _, buchstabe := range flaecheDesRufs(k) {
					gefunden[buchstabe] = true
				}
			case *ast.Ident:
				if k.Name != name && funktionen[k.Name] != nil {
					for buchstabe := range flaechen(k.Name, gesehen) {
						gefunden[buchstabe] = true
					}
				}
			}

			return true
		})

		gemerkt[name] = gefunden

		return gefunden
	}

	return func(lauf string) map[string]bool { return flaechen(lauf, map[string]bool{}) }
}

// flaecheDesRufs erkennt einen Ruf an den Helfer einer Fläche.
func flaecheDesRufs(ruf *ast.CallExpr) []string {
	switch funktion := ruf.Fun.(type) {
	case *ast.Ident:
		switch {
		case funktion.Name == "fensterAus":
			return []string{"F"}
		case funktion.Name == "kommandozeile":
			return []string{"K"}
		case strings.HasPrefix(funktion.Name, "mcp"):
			return []string{"M"}
		case strings.HasPrefix(funktion.Name, "starteDienst"), funktion.Name == "dienstGelang":
			return []string{"R"}
		}
	case *ast.SelectorExpr:
		if strings.HasPrefix(funktion.Sel.Name, "ruf") {
			return []string{"R"}
		}
	}

	return nil
}

// flaechenText schreibt eine Menge von Flächen, wie die Tafel sie nennt.
func flaechenText(menge map[string]bool) string {
	teile := []string{}

	for _, buchstabe := range flaechenbuchstaben {
		if menge[buchstabe] {
			teile = append(teile, buchstabe)
		}
	}

	if len(teile) == 0 {
		return "—"
	}

	return strings.Join(teile, " ")
}

func TestDieFlaechenJederKlasseSindGemessen(t *testing.T) {
	t.Parallel()

	paritaet := lies(t, "docs/FLAECHENPARITAET.md")
	flaechenVon := flaechenDerLaeufe(t)

	steht := map[string][]string{}
	for _, zeile := range gemesseneZeile.FindAllStringSubmatch(paritaet, -1) {
		steht[strings.TrimSpace(zeile[1])] = []string{zeile[2], zeile[3], zeile[4]}
	}

	if len(steht) == 0 {
		t.Fatal("der Abschnitt „Gemessene Flächen“ fehlt oder hat seine Gestalt verloren")
	}

	bilanz := map[string][3]int{}
	klassen := 0

	for _, zeile := range flaechenparitaetszeile.FindAllStringSubmatch(paritaet, -1) {
		if strings.HasPrefix(strings.TrimSpace(zeile[6]), "**nur am Fenster**") {
			continue
		}

		klasse := strings.TrimSpace(zeile[1])
		klassen++
		will := make([]string, 3)

		for stelle, spalte := range []int{3, 4, 5} {
			if stelle == 2 && strings.HasPrefix(strings.TrimSpace(zeile[5]), "entfällt") {
				will[stelle] = "entfällt"

				continue
			}

			menge := map[string]bool{}
			for _, lauf := range paritaetsname.FindAllStringSubmatch(zeile[spalte], -1) {
				for buchstabe := range flaechenVon(lauf[1]) {
					menge[buchstabe] = true
				}
			}

			will[stelle] = flaechenText(menge)

			for buchstabe := range menge {
				zaehler := bilanz[buchstabe]
				zaehler[stelle]++
				bilanz[buchstabe] = zaehler
			}
		}

		if !slices.Equal(steht[klasse], will) {
			t.Errorf("%s: gemessen sind %v, die Tafel nennt %v — die Zeile lautet:\n| %s | %s | %s | %s |",
				klasse, will, steht[klasse], klasse, will[0], will[1], will[2])
		}

		delete(steht, klasse)
	}

	for klasse := range steht {
		t.Errorf("„Gemessene Flächen“ nennt %q — keine Klasse mit Läufen", klasse)
	}

	pruefeFlaechenbilanz(t, paritaet, bilanz, klassen)
}

// pruefeFlaechenbilanz hält die Bilanz je Fläche gegen die Messung.
func pruefeFlaechenbilanz(t *testing.T, paritaet string, bilanz map[string][3]int, klassen int) {
	t.Helper()

	zeilen := flaechenbilanz.FindAllStringSubmatch(paritaet, -1)
	if len(zeilen) != len(flaechenbuchstaben) {
		t.Fatalf("die Bilanz je Fläche hat %d Zeilen, erwartet %d", len(zeilen), len(flaechenbuchstaben))
	}

	for _, zeile := range zeilen {
		var buchstabe string
		for kurz, lang := range flaechenNamenVon {
			if lang == zeile[1] {
				buchstabe = kurz
			}
		}

		gezaehlt := bilanz[buchstabe]
		if got := fmt.Sprintf("%s %s %s", zeile[2], zeile[3], zeile[4]); got != fmt.Sprintf("%d %d %d", gezaehlt[0], gezaehlt[1], gezaehlt[2]) {
			t.Errorf("Bilanz %s: die Tafel nennt %s, gemessen sind %d %d %d (von %d Klassen) — die Zeile lautet:\n| %s | %d | %d | %d |",
				zeile[1], got, gezaehlt[0], gezaehlt[1], gezaehlt[2], klassen, zeile[1], gezaehlt[0], gezaehlt[1], gezaehlt[2])
		}
	}
}
