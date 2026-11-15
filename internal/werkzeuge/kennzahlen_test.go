// Der Umfang dieses Projekts, gezählt statt behauptet.
//
// # Warum eine Zahl im README einen Wächter braucht
//
// **Weil sie sonst still veraltet.** Am 2026-08-29 stand in `AGENTS.md` noch
// „64 Bindungen", während 142 gezählt waren, und niemand hat es gesehen — eine
// Zahl in Fließtext steht unter keiner Aufsicht
// (#1287).
//
// **Dieser Absatz trägt sein Datum**, und das ist kein Zufall: der Wächter
// über die Flächenzahlen hat ihn beim ersten Lauf angestrichen, weil er die
// 64 als Behauptung über heute las. Genau dafür gibt es ihn.
// `flaechenzahlen_test.go` hält seither jede **Flächenzahl** im Repositorium
// gegen die Wirklichkeit; dieser hier hält den **Umfang**.
//
// # Warum mit Spielraum
//
// **Weil eine LoC-Zahl sich mit jedem Commit ändert.** Ein Wächter, der
// Gleichheit verlangte, wäre bei jedem zweiten Lauf rot — und ein Wächter, der
// dauernd rot ist, wird abgeschaltet. Er ist dann nicht streng, sondern
// abwesend.
//
// Geprüft wird deshalb die **Größenordnung**: weicht eine Zahl im README um
// mehr als [spielraum] ab, ist der Abschnitt überholt und wird neu erhoben.
//
// # Wie man ihn benutzt
//
//	make kennzahlen
//
// gibt die Tabelle aus, aus der der Abschnitt „Der Umfang, gezählt" entsteht.
//
// # Was er nicht prüft
//
//   - **Ob die Zahl etwas bedeutet.** 412.503 Codezeilen sagen nichts über
//     Güte; sie sagen, wie viel dasteht.
//   - **Die Komplexität des Frontends.** Der AST-Weg gilt für Go; für
//     TypeScript gäbe es keinen ohne neue Abhängigkeit.
//   - **Die Prosa daneben.** Dass „87 % liegen bei 1–5" auch als Satz stimmt,
//     ist eine Frage ans Review (§10).
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

// spielraum ist die erlaubte Abweichung einer Zahl im README.
//
// **Fünf Prozent** — genug für die Arbeit einiger Wochen, zu wenig für eine
// Zahl aus einer anderen Phase.
const spielraum = 0.05

// umfangszahlen sind die Angaben im README, die dieser Wächter nachrechnet.
//
// **Nur die tragenden.** Jede einzelne Zeile der Tabelle zu halten machte den
// Wächter zerbrechlich, ohne mehr zu sagen: wer die Gesamtzahl und das
// Verhältnis prüft, merkt jede Verschiebung, die zählt.
var umfangszahlen = []struct {
	name    string
	imText  *regexp.Regexp
	erheben func(erhebung) float64
}{
	{
		"die Gesamtzahl der Quelltextzeilen",
		regexp.MustCompile(`\*\*Quelltext gesamt\*\*\s*\|\s*\*\*([\d.]+)\*\*`),
		func(e erhebung) float64 { return float64(e.code) },
	},
	{
		"die Zeilen Dokumentation",
		regexp.MustCompile(`\*\*Dokumentation\*\*\s*\|\s*\*\*([\d.]+)\*\*`),
		func(e erhebung) float64 { return float64(e.doku) },
	},
	{
		"die Zeilen Go-Produktivcode",
		regexp.MustCompile(`\|\s*Go\s*\|\s*Produktiv\s*\|\s*([\d.]+)\s*\|`),
		func(e erhebung) float64 { return float64(e.nachArt["Go Produktiv"]) },
	},
	{
		"die Zeilen Go-Testcode",
		regexp.MustCompile(`\|\s*Go\s*\|\s*Test\s*\|\s*([\d.]+)\s*\|`),
		func(e erhebung) float64 { return float64(e.nachArt["Go Test"]) },
	},
	{
		"die Zahl der Go-Funktionen im Produktivcode",
		regexp.MustCompile(`\|\s*Funktionen\s*\|\s*([\d.]+)\s*\|`),
		func(e erhebung) float64 { return float64(e.funktionen) },
	},
	{
		"die mittlere Komplexität",
		regexp.MustCompile(`\|\s*Mittel\s*\|\s*\*\*([\d,]+)\*\*`),
		func(e erhebung) float64 { return e.mittel },
	},
}

// erhebung ist, was dieser Wächter aus dem Baum liest.
//
// **Quelltext und Dokumentation getrennt.** Zusammengezählt wäre die Summe
// größer und sagte weniger: 35.000 Zeilen Fließtext sind kein Programm.
type erhebung struct {
	code       int
	doku       int
	funktionen int
	mittel     float64

	// nachArt schlüsselt die Zeilen auf — der Schlüssel ist die Zeile, die im
	// README steht („Go Produktiv"), damit Tabelle und Erhebung **eine**
	// Quelle haben.
	//
	// **Sonst laufen sie auseinander**, und genau das ist am 2026-09-07
	// passiert: die Aufschlüsselung kam aus einer Erhebung von Hand, die
	// Gesamtzahl von hier — sie summierten sich auf 377.376 gegen 398.738.
	nachArt map[string]int
}

func TestDerUmfangImReadmeStimmtNoch(t *testing.T) {
	t.Parallel()

	gezaehlt := erhebe(t)

	t.Logf("Quelltext: %d Zeilen, Dokumentation: %d, Go-Funktionen: %d, "+
		"mittlere Komplexität: %.2f",
		gezaehlt.code, gezaehlt.doku, gezaehlt.funktionen, gezaehlt.mittel)

	for _, name := range arten {
		t.Logf("  %-22s %8d", name, gezaehlt.nachArt[name])
	}

	roh, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("README lesen: %v", err)
	}

	for _, zahl := range umfangszahlen {
		treffer := zahl.imText.FindSubmatch(roh)
		if treffer == nil {
			t.Errorf("%s steht nicht im README — der Abschnitt "+
				"„Der Umfang, gezählt"+`" ist fort oder anders geschrieben`,
				zahl.name)

			continue
		}

		steht := alsZahl(t, string(treffer[1]))
		soll := zahl.erheben(gezaehlt)

		if abweichung(steht, soll) > spielraum {
			t.Errorf("%s: im README steht %.2f, gezählt sind %.2f — "+
				"der Abschnitt ist überholt, `make kennzahlen` erhebt ihn neu",
				zahl.name, steht, soll)
		}
	}
}

// erhebe zählt Codezeilen und Komplexität über den Baum.
func erhebe(t *testing.T) erhebung {
	t.Helper()

	gezaehlt := erhebung{nachArt: map[string]int{}}

	summe := 0

	// **Derselbe Baum, den die Tabelle im README nennt** — ohne Generat,
	// Fremdes und die Werkzeugzwischenspeicher. `.` deckt auch `main.go`, die
	// Leitliniendateien und `docs/` ab.
	fremd := nichtGefuehrtIn(t)

	for _, ort := range []string{"../.."} {
		// **Jeder Fehlschlag heißt „zählt nicht mit".** Eine Datei, die sich
		// nicht lesen lässt, ist keine Zahl — und ein Wächter über den Umfang,
		// der an einer Datei scheitert, sagt nichts über den Umfang.
		//
		//nolint:errcheck,nilerr // siehe oben
		_ = filepath.Walk(ort, func(pfad string, angabe os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if fremd.uebergeht(pfad) {
				if angabe.IsDir() {
					return filepath.SkipDir
				}

				return nil
			}

			if angabe.IsDir() || uebergangen(pfad) {
				return nil
			}

			roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Pfad aus dem Baum selbst
			if lesefehler != nil {
				return nil //nolint:nilerr // siehe oben
			}

			zeilen := codezeilen(string(roh))

			if strings.HasSuffix(pfad, ".md") {
				gezaehlt.doku += zeilen
			} else {
				gezaehlt.code += zeilen
				gezaehlt.nachArt[art(pfad)] += zeilen
			}

			if strings.HasSuffix(pfad, ".go") && !strings.HasSuffix(pfad, "_test.go") {
				zahl, wert := komplexitaeten(pfad)
				gezaehlt.funktionen += zahl
				summe += wert
			}

			return nil
		})
	}

	if gezaehlt.funktionen > 0 {
		gezaehlt.mittel = float64(summe) / float64(gezaehlt.funktionen)
	}

	return gezaehlt
}

// arten sind die Zeilen der Tabelle im README, in ihrer Reihenfolge.
var arten = []string{
	"Go Produktiv", "Go Test",
	"TypeScript Produktiv", "TypeScript Test",
	"Übriges",
}

// art ordnet eine Datei einer Zeile der Tabelle zu.
func art(pfad string) string {
	switch {
	case strings.HasSuffix(pfad, "_test.go"):
		return "Go Test"

	case strings.HasSuffix(pfad, ".go"):
		return "Go Produktiv"

	case strings.HasSuffix(pfad, ".ts") || strings.HasSuffix(pfad, ".tsx"):
		if strings.Contains(pfad, ".test.") || strings.Contains(pfad, "/test/") {
			return "TypeScript Test"
		}

		return "TypeScript Produktiv"
	}

	return "Übriges"
}

// uebergangen sagt, was nicht mitzählt — Generat und Fremdes.
func uebergangen(pfad string) bool {
	for _, teil := range []string{
		// **`/build/` ganz und nicht nur `/build/bin/`** — am 2026-09-07 lagen
		// unter `build/freigabe/` die zwei Stücklisten der Freigabe, rund
		// 20.000 Zeilen JSON, und der Wächter zählte sie als „Übriges". Die
		// Zahl im README trug damit einen Bauausstoß, der am nächsten Tag
		// nicht mehr dalag; gefunden am 2026-09-08, als der Wächter ohne eine
		// geänderte Zeile Quelltext rot wurde (#1412).
		"/wailsjs/", "/node_modules/", "/.go/", "/dist/", "/build/",
		"/graft/", "/.git/", "/coverage/", "/.cursor/", "/.claude/",
	} {
		if strings.Contains(pfad, teil) {
			return true
		}
	}

	for _, endung := range []string{
		".go", ".ts", ".tsx", ".sql", ".sh", ".css", ".mjs", ".json", ".md",
		".html", ".yml",
	} {
		if strings.HasSuffix(pfad, endung) {
			return false
		}
	}

	return true
}

// codezeilen zählt, was weder leer noch Kommentar ist.
func codezeilen(inhalt string) int {
	zahl := 0

	for _, zeile := range strings.Split(inhalt, "\n") {
		gestutzt := strings.TrimSpace(zeile)
		if gestutzt != "" && !strings.HasPrefix(gestutzt, "//") &&
			!strings.HasPrefix(gestutzt, "#") && !strings.HasPrefix(gestutzt, "--") {
			zahl++
		}
	}

	return zahl
}

// komplexitaeten gibt Zahl und Summe der Funktionen einer Go-Datei.
func komplexitaeten(pfad string) (zahl, summe int) {
	satz := token.NewFileSet()

	baum, err := parser.ParseFile(satz, pfad, nil, 0)
	if err != nil {
		return 0, 0
	}

	for _, deklaration := range baum.Decls {
		fn, ist := deklaration.(*ast.FuncDecl)
		if !ist || fn.Body == nil {
			continue
		}

		zahl++
		summe += komplexitaet(fn.Body)
	}

	return zahl, summe
}

// komplexitaet zählt die Verzweigungen eines Rumpfes.
func komplexitaet(rumpf ast.Node) int {
	wert := 1

	ast.Inspect(rumpf, func(knoten ast.Node) bool {
		switch v := knoten.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause, *ast.CommClause:
			wert++
		case *ast.BinaryExpr:
			if v.Op == token.LAND || v.Op == token.LOR {
				wert++
			}
		}

		return true
	})

	return wert
}

// alsZahl liest eine Zahl in deutscher Schreibweise — 412.503 und 3,12.
func alsZahl(t *testing.T, roh string) float64 {
	t.Helper()

	sauber := strings.ReplaceAll(strings.ReplaceAll(roh, ".", ""), ",", ".")

	wert, err := strconv.ParseFloat(sauber, 64)
	if err != nil {
		t.Fatalf("%q ist keine Zahl: %v", roh, err)
	}

	return wert
}

// abweichung gibt den Abstand zweier Zahlen als Anteil.
func abweichung(steht, soll float64) float64 {
	if soll == 0 {
		return 0
	}

	unterschied := steht - soll
	if unterschied < 0 {
		unterschied = -unterschied
	}

	return unterschied / soll
}
