package main

// **Jeder fachliche Fehler kommt mit `--json` als Form an** — Befund 4 des
// Zweitblicks über Phase 1 von 0.9.0 (F03, #2342), und Befund 3 des
// Nachlaufs darüber.
//
// # Warum ein zweiter Wächter
//
// `TestKeinFachlicherFehlerGehtAmJSONVorbei` sucht eine **Schreibweise** —
// `.klage("%s", ….Meldung)`. Daran vorbei gingen ein belegter Bestand
// (`klageUeber` kehrte vor der Form zurück) und sechs Stellen, die mit
// eigenem Satz und Rückgabewert 1 endeten.
//
// # Wege statt Vorkommen
//
// Der erste Anlauf zählte, ob **irgendwo** davor gemeldet wurde — auch in
// einem Zweig, der nicht betreten wird, in einer Schleife oder in einem
// Funktionsliteral, das niemand ruft. **Jetzt gilt eine Meldung nur als
// ganze Anweisung**, die vor der Rückgabe im selben oder in einem
// umschließenden Block steht: dann läuft jeder Weg zur Rückgabe über sie.
// Ein Funktionsliteral ist eine eigene Wurzel. **Eine 1 ist jede Konstante mit
// dem Wert 1, die Zahl selbst und jede Zuordnung**, an jeder Stelle einer
// Rückgabe; eine solche Konstante außerhalb einer Rückgabe — in einer
// Variablen — ist schon ein Befund, denn ihr Weg ist nicht mehr zu sehen.
//
// # Was er nicht hält
//
// **Die Ergebniswege** unten: sie geben ihre Antwort selbst aus, und der
// Rückgabewert sagt nur, dass sie keine zum Weiterlaufen ist (Vermerk 0033).
// Einen Helfer, der einen anderen meldenden Helfer ruft, verfolgt er nicht —
// gemeldet wird in den Helfern dieser Liste. Und ob ein gemeldeter Fehler den
// **richtigen** Code trägt; das hält `internal/flaechen`.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// melder sind die Rufe, die einen fachlichen Fehler samt Form melden.
var melder = map[string]bool{"klageUeber": true, "klageUeberFehler": true, "klageMitSatz": true}

// meldendeHelfer melden selbst, bevor sie falsch zurückgeben — gezählt nur
// als ganze, verneinte Bedingung: `if !kenntSetIm(…) { return … }`.
var meldendeHelfer = map[string]bool{"kenntSetIm": true}

// ergebniswege geben ihre Antwort selbst aus; der Rückgabewert ist ihr Urteil.
var ergebniswege = map[string]string{
	"fachcode": "Zurücksetzen: die Antwort steht auf der Ausgabe, und mit --json ein Fehler darin " +
		"unter `fehler` — dieselbe Form; ohne --json meldet klageUeberFehler. Kostenjournal: " +
		"ein unklarer Posten ist ein Befund der Antwort",
	"lagecode":     "Rücksprung: unklar ist kein Fehler, aber kein Weiterlaufen",
	"fassungscode": "Verträglichkeit: das Urteil steht auf der Ausgabe",
	"bilanzcode":   "Datenpflege: die Bilanz je Maßnahme steht auf der Ausgabe",
	"befundscode":  "Datenpflege und Rücksetzplan: die Befunde stehen auf der Ausgabe",
	"freigabecode": "Regelfreigabe: die Antwort steht auf der Ausgabe",
	"urteilscode":  "Deckauskunft: die fehlenden Karten stehen auf der Ausgabe; Paketprüfung mit --json: der Befund steht im Bericht (#2400)",
	"batchzeilen":  "KI-Batch: je Ziel eine Zeile mit seinem Ausgang",
	"zustandscode": "Gesundheit: nicht bereit ist die Antwort, und sie steht auf der Ausgabe",
}

// zuordnungen machen aus einem Fehler einen Rückgabewert — **gemeldet wird
// beim Aufrufer**, und dort hält der Wächter es wie bei `return CodeFachlich`.
var zuordnungen = map[string]bool{"mtgjsonCode": true, "paketCode": true, "vorlagenCode": true}

// rueckweg prüft die Rückgaben eines Pakets.
type rueckweg struct {
	menge  *token.FileSet
	einsen map[string]bool
	befund []string
}

func TestJederFachlicheFehlerTraegtSeineForm(t *testing.T) {
	dateien, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	menge := token.NewFileSet()

	var baeume []*ast.File

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		baum, lesefehler := parser.ParseFile(menge, datei, nil, 0)
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		baeume = append(baeume, baum)
	}

	weg := rueckweg{menge: menge, einsen: einsenIn(baeume)}
	if !weg.einsen["CodeFachlich"] {
		t.Fatalf("CodeFachlich ist keine Konstante mit dem Wert 1 mehr — der Wächter liest daneben (%v)", weg.einsen)
	}

	gesehen := 0

	for _, baum := range baeume {
		gesehen += weg.pruefeDatei(baum)
	}

	for _, befund := range weg.befund {
		t.Error(befund)
	}

	if gesehen < 100 {
		t.Fatalf("nur %d Funktionen gelesen — der Wächter liest daneben", gesehen)
	}
}

// einsenIn sind die Konstanten mit dem Wert 1.
func einsenIn(baeume []*ast.File) map[string]bool {
	einsen := map[string]bool{}

	for _, baum := range baeume {
		ast.Inspect(baum, func(knoten ast.Node) bool {
			wert, ist := knoten.(*ast.ValueSpec)
			if !ist {
				return true
			}

			for i, name := range wert.Names {
				if i < len(wert.Values) && istEins(wert.Values[i]) {
					einsen[name.Name] = true
				}
			}

			return true
		})
	}

	return einsen
}

func istEins(ausdruck ast.Expr) bool {
	zahl, ist := ausdruck.(*ast.BasicLit)

	return ist && zahl.Kind == token.INT && zahl.Value == "1"
}

// pruefeDatei prüft jede Funktion der Datei und gibt, wie viele es waren.
func (w *rueckweg) pruefeDatei(baum *ast.File) int {
	gesehen := 0

	for _, deklaration := range baum.Decls {
		funktion, ist := deklaration.(*ast.FuncDecl)
		if !ist || funktion.Body == nil {
			continue
		}

		gesehen++

		if _, ausgenommen := ergebniswege[funktion.Name.Name]; ausgenommen || zuordnungen[funktion.Name.Name] {
			continue
		}

		w.pruefeWurzel(funktion.Name.Name, funktion.Body)
	}

	return gesehen
}

// pruefeWurzel prüft einen Rumpf und jedes Funktionsliteral darin als eigene
// Wurzel. **Ein Vergleich mit einer 1** — `code == CodeFachlich` — trägt sie
// nicht weiter und ist deshalb erlaubt.
func (w *rueckweg) pruefeWurzel(name string, rumpf *ast.BlockStmt) {
	inRueckgabe := map[*ast.Ident]bool{}
	w.block(name, rumpf.List, false, inRueckgabe)

	ast.Inspect(rumpf, func(knoten ast.Node) bool {
		switch k := knoten.(type) {
		case *ast.FuncLit:
			w.pruefeWurzel(name+" (Funktionsliteral)", k.Body)

			return false
		case *ast.BinaryExpr:
			if k.Op == token.EQL || k.Op == token.NEQ {
				for _, seite := range []ast.Expr{k.X, k.Y} {
					if name, ist := seite.(*ast.Ident); ist {
						inRueckgabe[name] = true
					}
				}
			}
		case *ast.Ident:
			if w.einsen[k.Name] && !inRueckgabe[k] {
				w.melde(name, k.Pos(), k.Name+" außerhalb einer Rückgabe — ihr Weg ist nicht mehr zu sehen")
			}
		}

		return true
	})
}

// block geht die Anweisungen eines Blocks durch; `gemeldet` sagt, ob jeder
// Weg hierher schon über eine Meldung lief — **oder mit `--json` schon
// zurückkehrte**: hinter `if an.alsJSON { …; return … }` ist nur noch Text,
// und dort braucht der Rückgabewert keine Form.
func (w *rueckweg) block(name string, anweisungen []ast.Stmt, gemeldet bool, inRueckgabe map[*ast.Ident]bool) {
	for _, anweisung := range anweisungen {
		w.anweisung(name, anweisung, gemeldet, inRueckgabe)

		gemeldet = gemeldet || istMeldung(anweisung) || nurNochText(anweisung)
	}
}

// nurNochText sagt, ob die Anweisung `if an.alsJSON { …; return … }` ist,
// ohne Sonst — danach läuft nur noch, wer kein JSON will.
func nurNochText(anweisung ast.Stmt) bool {
	wenn, ist := anweisung.(*ast.IfStmt)
	if !ist || wenn.Init != nil || wenn.Else != nil || len(wenn.Body.List) == 0 {
		return false
	}

	feld, ist := wenn.Cond.(*ast.SelectorExpr)
	if !ist || feld.Sel.Name != "alsJSON" {
		return false
	}

	_, endet := wenn.Body.List[len(wenn.Body.List)-1].(*ast.ReturnStmt)

	return endet
}

func (w *rueckweg) anweisung(name string, anweisung ast.Stmt, gemeldet bool, inRueckgabe map[*ast.Ident]bool) {
	switch a := anweisung.(type) {
	case *ast.ReturnStmt:
		w.rueckgabe(name, a, gemeldet, inRueckgabe)
	case *ast.BlockStmt:
		w.block(name, a.List, gemeldet, inRueckgabe)
	case *ast.IfStmt:
		w.wenn(name, a, gemeldet, inRueckgabe)
	case *ast.ForStmt:
		w.block(name, a.Body.List, gemeldet, inRueckgabe)
	case *ast.RangeStmt:
		w.block(name, a.Body.List, gemeldet, inRueckgabe)
	case *ast.SwitchStmt:
		w.faelle(name, a.Body, gemeldet || istMeldung(a.Init), inRueckgabe)
	case *ast.TypeSwitchStmt:
		w.faelle(name, a.Body, gemeldet || istMeldung(a.Init), inRueckgabe)
	case *ast.SelectStmt:
		w.faelle(name, a.Body, gemeldet, inRueckgabe)
	case *ast.LabeledStmt:
		w.anweisung(name, a.Stmt, gemeldet, inRueckgabe)
	}
}

func (w *rueckweg) wenn(name string, wenn *ast.IfStmt, gemeldet bool, inRueckgabe map[*ast.Ident]bool) {
	gemeldet = gemeldet || istMeldung(wenn.Init)

	w.block(name, wenn.Body.List, gemeldet || verneinterHelfer(wenn.Cond), inRueckgabe)

	switch sonst := wenn.Else.(type) {
	case *ast.BlockStmt:
		w.block(name, sonst.List, gemeldet, inRueckgabe)
	case *ast.IfStmt:
		w.wenn(name, sonst, gemeldet, inRueckgabe)
	}
}

func (w *rueckweg) faelle(name string, rumpf *ast.BlockStmt, gemeldet bool, inRueckgabe map[*ast.Ident]bool) {
	for _, fall := range rumpf.List {
		switch f := fall.(type) {
		case *ast.CaseClause:
			w.block(name, f.Body, gemeldet, inRueckgabe)
		case *ast.CommClause:
			w.block(name, f.Body, gemeldet, inRueckgabe)
		}
	}
}

// rueckgabe meldet eine 1 ohne Meldung davor — an jeder Stelle der Rückgabe.
func (w *rueckweg) rueckgabe(name string, rueckgabe *ast.ReturnStmt, gemeldet bool, inRueckgabe map[*ast.Ident]bool) {
	for _, ergebnis := range rueckgabe.Results {
		eins := istEins(ergebnis)

		switch e := ergebnis.(type) {
		case *ast.Ident:
			if w.einsen[e.Name] {
				inRueckgabe[e] = true
				eins = true
			}
		case *ast.CallExpr:
			ruf, ist := e.Fun.(*ast.Ident)
			eins = eins || ist && zuordnungen[ruf.Name]
		}

		if eins && !gemeldet {
			w.melde(name, rueckgabe.Pos(), "Rückgabewert 1 ohne gemeldeten Fehler — nimm klageUeber "+
				"oder klageMitSatz davor, sonst fehlt mit --json die Form")
		}
	}
}

func (w *rueckweg) melde(name string, stelle token.Pos, satz string) {
	w.befund = append(w.befund, w.menge.Position(stelle).String()+" ("+name+"): "+satz)
}

// istMeldung sagt, ob die Anweisung selbst ein Melder ist — `klageUeber(…)`
// oder `_ = an.klageMitSatz(…)`, nicht ein Ruf irgendwo darin.
func istMeldung(anweisung ast.Stmt) bool {
	var ausdruck ast.Expr

	switch a := anweisung.(type) {
	case *ast.ExprStmt:
		ausdruck = a.X
	case *ast.AssignStmt:
		if len(a.Rhs) == 1 {
			ausdruck = a.Rhs[0]
		}
	}

	ruf, ist := ausdruck.(*ast.CallExpr)

	return ist && rufname(ruf, melder)
}

// verneinterHelfer sagt, ob die Bedingung genau `!helfer(…)` ist — sonst
// liefe ein Weg vielleicht an ihm vorbei.
func verneinterHelfer(bedingung ast.Expr) bool {
	nicht, ist := bedingung.(*ast.UnaryExpr)
	if !ist || nicht.Op != token.NOT {
		return false
	}

	ruf, ist := nicht.X.(*ast.CallExpr)

	return ist && rufname(ruf, meldendeHelfer)
}

func rufname(ruf *ast.CallExpr, namen map[string]bool) bool {
	switch f := ruf.Fun.(type) {
	case *ast.SelectorExpr:
		return namen[f.Sel.Name]
	case *ast.Ident:
		return namen[f.Name]
	}

	return false
}
