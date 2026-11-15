// **Die zweite Tür bleibt schmal** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Was er hält
//
// `app.Nutzerfehler` ist die Tür ohne Vorgangskennung. Sie darf **nur dort**
// benutzt werden, wo gar kein `context.Context` in Reichweite ist — in einer
// reinen Rechnung, einer Eingabeprüfung, einem Hilfsschritt ohne Ruf. Wer
// einen Kontext hat, nimmt `app.Uebersetze(ctx, …)`.
//
// # Warum genau diese Regel und keine über den Inhalt
//
// **Weil sie die Gefahr trifft.** Die Gefahr ist nicht, dass jemand die
// falsche Fehlerklasse wählt — sie ist, dass jemand die **kürzere** Tür nimmt,
// obwohl die längere offen stand: dann verliert ein Fehler aus dem Unterbau
// den einzigen Weg zu seinem Grund, denn der Grund erreicht die Oberfläche
// absichtlich nie (§8). Diese Regel ist syntaktisch prüfbar und lässt sich
// nicht mit einer Umformulierung umgehen; eine Regel über die Fehlerklasse
// ließe sich mit einem `fmt.Errorf("%w", err)` in einer Zeile aushebeln.
//
// # Was er **nicht** hält
//
// **Ob eine kontextlose Funktion wirklich nur Eingaben prüft.** `teilweise`
// in `symbolgruppe.go` reicht den Fehler einer **übergebenen** Umformung
// durch; dass die drei Umformungen dahinter ausschließlich
// `domain.ErrValidation` geben, steht in `internal/service/symbolteile.go`
// und ist eine Frage ans Review (§10), nicht an diesen Wächter.
//
// **Und er sieht `internal/app` an, sonst nichts.** Außerhalb ist
// `Nutzerfehler` bisher nicht gerufen; käme ein Ruf dazu, prüfte ihn hier
// niemand. Das steht hier, damit es niemand für geprüft hält.
package werkzeuge_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNutzerfehlerStehtNurWoKeinKontextIst(t *testing.T) {
	t.Parallel()

	verzeichnis := "../app"

	eintraege, fehler := os.ReadDir(verzeichnis)
	if fehler != nil {
		t.Fatalf("%s lesen: %v", verzeichnis, fehler)
	}

	menge := token.NewFileSet()

	var befunde []string

	gezaehlt := 0

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if eintrag.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		pfad := filepath.Join(verzeichnis, name)

		datei, lesefehler := parser.ParseFile(menge, pfad, nil, 0)
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", pfad, lesefehler)
		}

		for _, erklaerung := range datei.Decls {
			funktion, istFunktion := erklaerung.(*ast.FuncDecl)
			if !istFunktion || funktion.Body == nil {
				continue
			}

			gefunden, mitKontext := nutzerfehlerrufe(menge, funktion)
			gezaehlt += len(gefunden)

			if mitKontext {
				for _, zeile := range gefunden {
					befunde = append(befunde, "  "+pfad+":"+zeile+
						" — hier ist ein Kontext in Reichweite")
				}
			}
		}
	}

	// **Die Reichweite der Messung** (§10): ein Wächter, der nichts findet,
	// ist grün und sagt nichts.
	if gezaehlt == 0 {
		t.Fatal("kein einziger Ruf von Nutzerfehler gefunden — der Wächter misst nichts")
	}

	if len(befunde) != 0 {
		t.Errorf("%d Ruf(e) von Nutzerfehler mit Kontext in Reichweite "+
			"(nimm app.Uebersetze(ctx, …)):\n%s",
			len(befunde), strings.Join(befunde, "\n"))
	}
}

// nutzerfehlerrufe gibt die Zeilen der Rufe und ob die Funktion einen Kontext
// führt.
func nutzerfehlerrufe(menge *token.FileSet, funktion *ast.FuncDecl) ([]string, bool) {
	var zeilen []string

	ast.Inspect(funktion.Body, func(knoten ast.Node) bool {
		ruf, istRuf := knoten.(*ast.CallExpr)
		if !istRuf {
			return true
		}

		if bezeichner, ist := ruf.Fun.(*ast.Ident); ist && bezeichner.Name == "Nutzerfehler" {
			zeilen = append(zeilen, strconv.Itoa(menge.Position(ruf.Pos()).Line))
		}

		return true
	})

	return zeilen, kontextname(funktion) != ""
}
