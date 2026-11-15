package fenster

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sync"
	"testing"
)

func TestDieLetzteGroesseBleibtStehen(t *testing.T) {
	t.Parallel()

	var g letzteGroesse

	if b, h := g.abschliessen(); b != 0 || h != 0 {
		t.Fatalf("vor der ersten Meldung: %d × %d", b, h)
	}

	g = letzteGroesse{}
	g.melden(func() (int, int) { return 1100, 760 })
	g.melden(func() (int, int) { return 0, 0 })
	g.melden(func() (int, int) { return -1, 500 })

	if b, h := g.abschliessen(); b != 1100 || h != 760 {
		t.Errorf("gemerkt ist %d × %d, erwartet 1100 × 760 — eine leere Meldung überschrieb sie", b, h)
	}
}

// **Eine spätere Meldung überholt eine frühere, nicht umgekehrt** — Nachlauf
// über Phase 5 von 0.9.0, Befund 2. Wails ruft jeden Hörer in einer eigenen
// Goroutine; zwei Meldungen können ihre Größe in umgekehrter Reihenfolge
// zurückbekommen. Es zählt die Reihenfolge, in der sie **kamen**.
func TestEineFruehereMeldungUeberholtKeineSpaetere(t *testing.T) {
	t.Parallel()

	var g letzteGroesse

	frueh, _ := g.nummer()
	spaet, _ := g.nummer()

	g.setzen(spaet, 1200, 800)
	g.setzen(frueh, 1000, 700)

	if b, h := g.abschliessen(); b != 1200 || h != 800 {
		t.Errorf("gemerkt ist %d × %d, erwartet die spätere Meldung 1200 × 800", b, h)
	}
}

// **Nach dem Schluss wird nichts mehr gemerkt und nichts mehr gefragt** —
// Befund 2 desselben Nachlaufs: eine Meldung, die nach dem Lesen des
// Schlusses ankam, änderte den Stand, den niemand mehr las, und eine Frage
// danach erreichte ein Fenster, das schon fort war.
func TestNachDemSchlussWirdNichtsMehrGemerkt(t *testing.T) {
	t.Parallel()

	var g letzteGroesse

	g.melden(func() (int, int) { return 1100, 760 })
	unterwegs, _ := g.nummer()

	if b, h := g.abschliessen(); b != 1100 || h != 760 {
		t.Fatalf("der Schluss las %d × %d", b, h)
	}

	g.setzen(unterwegs, 1300, 900)

	gefragt := false
	g.melden(func() (int, int) { gefragt = true; return 1400, 950 })

	if gefragt {
		t.Error("nach dem Schluss wurde das Fenster noch gefragt")
	}

	if _, offen := g.nummer(); offen {
		t.Error("nach dem Schluss vergibt die Hülle noch Nummern")
	}

	if b, h := g.abschliessen(); b != 1100 || h != 760 {
		t.Errorf("nach dem Schluss steht %d × %d, erwartet 1100 × 760", b, h)
	}
}

func TestDieLetzteGroesseVertraegtGleichzeitigeMeldungen(t *testing.T) {
	t.Parallel()

	var (
		g      letzteGroesse
		gruppe sync.WaitGroup
	)

	for i := 1; i <= 50; i++ {
		gruppe.Go(func() {
			g.melden(func() (int, int) { return 1000 + i, 700 + i })
		})
	}

	gruppe.Wait()

	if b, h := g.abschliessen(); b-h != 300 {
		t.Errorf("Breite und Höhe stammen aus verschiedenen Meldungen: %d × %d", b, h)
	}
}

// **Die Größe wird nur über [letzteGroesse.melden] erfragt** — #2439 und
// Befund 2 des Nachlaufs über Phase 5. Auf dem Hauptfaden gefragt, aus einer
// eigenen Goroutine, kommt die Frage beim Schluss nie dran (der Kopf von
// `letztegroesse.go`); und eine Frage an `melden` vorbei hätte weder Nummer
// noch Schluss. Gelesen wird `wails.go` als Syntaxbaum, denn die Verdrahtung
// mit Wails läuft hier nicht: **jeder** Ausdruck `….Size` — ein Ruf, ein
// Methodenwert, ein Alias — muss das Argument von `groesse.melden` sein.
//
// **Ausgenommen ist [einpassen]**: es fragt die Größe beim Aufschlagen, um das
// Fenster in den Bildschirm zu passen, und merkt sie nicht — die Meldung
// danach geht über `melden`.
//
// **Was er nicht hält**: eine Größe aus einer anderen Datei des Pakets, oder
// über eine Schnittstelle, deren Methode anders heißt.
func TestDieGroesseWirdNurUeberMeldenErfragt(t *testing.T) {
	t.Parallel()

	satz := token.NewFileSet()

	baum, fehler := parser.ParseFile(satz, "wails.go", nil, 0)
	if fehler != nil {
		t.Fatal(fehler)
	}

	erlaubt := map[ast.Expr]bool{}
	gemeldet := 0

	ast.Inspect(baum, func(knoten ast.Node) bool {
		ruf, istRuf := knoten.(*ast.CallExpr)
		if !istRuf {
			return true
		}

		if wahl, istWahl := ruf.Fun.(*ast.SelectorExpr); istWahl && wahl.Sel.Name == "melden" && len(ruf.Args) == 1 {
			erlaubt[ruf.Args[0]] = true
			gemeldet++
		}

		return true
	})

	ast.Inspect(baum, func(knoten ast.Node) bool {
		if funktion, istFunktion := knoten.(*ast.FuncDecl); istFunktion && funktion.Name.Name == "einpassen" {
			return false
		}

		wahl, istWahl := knoten.(*ast.SelectorExpr)
		if istWahl && wahl.Sel.Name == "Size" && !erlaubt[wahl] {
			t.Errorf("%s: die Größe wird an `groesse.melden` vorbei erfragt", satz.Position(wahl.Pos()))
		}

		return true
	})

	if gemeldet < 2 {
		t.Errorf("wails.go meldet die Größe %d-mal, erwartet bei der Änderung und beim Einpassen", gemeldet)
	}
}
