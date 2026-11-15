package app

// **Das Tor ist die erste Wirkung jeder Bindung** — Nachlauf über ID-02,
// Befund 1 (#1701).
//
// # Was dieser Wächter hält
//
//   - Jede exportierte Methode von [Bindungen] ruft in ihrer **ersten**
//     Anweisung [Bindungen.app] oder [Bindungen.tor] — außer denen in
//     [sitzungsfrei], und die rufen [Bindungen.frei] und nie das Tor.
//   - Das Feld `anwendung` erreichen nur die drei Zugänge und
//     [NeueBindungen].
//
// # Was er nicht hält
//
// **Ob die erste Anweisung vor dem Tor etwas tut** — `x := rechne(b.app())`
// ruft `rechne` nach dem Tor, `x := rechne(y, b.app())` davor. Eine
// Bindung reicht durch und rechnet nicht (`bindungen.go`, „Ab hier"); dass
// sie es in der ersten Zeile doch täte, sieht nur ein Mensch.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// zugaenge erreichen das Feld.
var zugaenge = map[string]bool{"app": true, "tor": true, "frei": true, "NeueBindungen": true}

func TestJedeBindungRuftZuerstDasTor(t *testing.T) {
	methoden := bindungsmethoden(t)

	// **Ein Helfer zählt wie das Tor**, wenn er es selbst zuerst ruft —
	// `dienstStarten` hinter `RESTStarten`.
	tore := map[string]bool{"app": true, "tor": true}

	for _, methode := range methoden {
		if !methode.funktion.Name.IsExported() && !zugaenge[methode.funktion.Name.Name] &&
			len(methode.funktion.Body.List) > 0 && ruft(methode.funktion.Body.List[0], tore) {
			tore[methode.funktion.Name.Name] = true
		}
	}

	geprueft := 0

	for _, methode := range methoden {
		if !zugaenge[methode.funktion.Name.Name] && greiftAufsFeld(methode.funktion.Body) {
			t.Errorf("%s: %s greift am Tor vorbei auf das Feld", methode.datei, methode.funktion.Name.Name)
		}

		if methode.funktion.Name.IsExported() {
			geprueft++
			torPruefen(t, methode.datei, methode.funktion, tore)
		}
	}

	if geprueft < 400 {
		t.Errorf("nur %d Bindungen gelesen — der Wächter liest am Quelltext vorbei", geprueft)
	}
}

// bindungsmethode ist eine Methode von *Bindungen mit ihrer Datei.
type bindungsmethode struct {
	datei    string
	funktion *ast.FuncDecl
}

// bindungsmethoden liest alle Methoden von *Bindungen aus dem Quelltext.
func bindungsmethoden(t *testing.T) []bindungsmethode {
	t.Helper()

	dateien, fehler := filepath.Glob("*.go")
	if fehler != nil {
		t.Fatal(fehler)
	}

	menge := token.NewFileSet()

	var methoden []bindungsmethode

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		baum, lesefehler := parser.ParseFile(menge, datei, nil, 0)
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		for _, deklaration := range baum.Decls {
			if funktion, ist := deklaration.(*ast.FuncDecl); ist && funktion.Body != nil && anBindungen(funktion) {
				methoden = append(methoden, bindungsmethode{datei: datei, funktion: funktion})
			}
		}
	}

	return methoden
}

// torPruefen hält eine Bindung gegen die Regel.
func torPruefen(t *testing.T, datei string, funktion *ast.FuncDecl, tore map[string]bool) {
	t.Helper()

	name := funktion.Name.Name

	if sitzungsfrei[name] {
		if ruft(funktion.Body, tore) || !ruft(funktion.Body, map[string]bool{"frei": true}) {
			t.Errorf("%s: %s steht in sitzungsfrei und muss über frei() gehen", datei, name)
		}

		return
	}

	if ruft(funktion.Body, map[string]bool{"frei": true}) {
		t.Errorf("%s: %s ruft frei() und steht nicht in sitzungsfrei", datei, name)
	}

	if len(funktion.Body.List) == 0 || !ruft(funktion.Body.List[0], tore) {
		t.Errorf("%s: %s ruft das Tor nicht in der ersten Anweisung", datei, name)
	}
}

// anBindungen sagt, ob eine Funktion eine Methode von *Bindungen ist.
func anBindungen(funktion *ast.FuncDecl) bool {
	if funktion.Recv == nil || len(funktion.Recv.List) != 1 {
		return false
	}

	stern, ist := funktion.Recv.List[0].Type.(*ast.StarExpr)
	if !ist {
		return false
	}

	typ, ist := stern.X.(*ast.Ident)

	return ist && typ.Name == "Bindungen"
}

// ruft sagt, ob unter einem Knoten `b.<name>(…)` für einen der Namen
// gerufen wird.
func ruft(knoten ast.Node, namen map[string]bool) (gefunden bool) {
	ast.Inspect(knoten, func(n ast.Node) bool {
		if ruf, ist := n.(*ast.CallExpr); ist {
			if auswahl, ist := ruf.Fun.(*ast.SelectorExpr); ist {
				empfaenger, ist := auswahl.X.(*ast.Ident)
				gefunden = ist && empfaenger.Name == "b" && namen[auswahl.Sel.Name]
			}
		}

		return !gefunden
	})

	return gefunden
}

// greiftAufsFeld sagt, ob unter einem Knoten `.anwendung` gelesen wird.
func greiftAufsFeld(knoten ast.Node) (gefunden bool) {
	ast.Inspect(knoten, func(n ast.Node) bool {
		if auswahl, ist := n.(*ast.SelectorExpr); ist && auswahl.Sel.Name == "anwendung" {
			if empfaenger, ist := auswahl.X.(*ast.Ident); ist && empfaenger.Name == "b" {
				gefunden = true
			}
		}

		return !gefunden
	})

	return gefunden
}
