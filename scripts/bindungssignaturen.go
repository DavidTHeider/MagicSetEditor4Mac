//go:build ignore

// Die Signaturen der gebundenen Methoden — für den Vertragsrundgang (B01,
// #2220).
//
//	go run scripts/bindungssignaturen.go internal/app
//
// Gibt je Methode von `*Bindungen` Namen, Parameter (ohne den Kontext),
// Rückgabetyp und **ob sie einen Dateidialog öffnen kann** als JSON.
//
// **Der Dialog kommt aus dem Aufrufgraphen**, nicht aus dem Fähigkeitskatalog:
// `✋` in `docs/FAEHIGKEITEN.md` sagt, dass eine Fähigkeit an einer Fläche
// **fehlt**, nicht welche Bindung einen Dialog öffnet — fünf Dialogbindungen
// fielen damit durch (erster Vertragsrundgang, 2026-10-09). Hier zählt, wer
// `App.waehleDatei` erreicht — über Methoden von `*App` und `*Bindungen` und
// Funktionen des Pakets, **nach Namen** aufgelöst: eine Überschätzung, wo zwei
// Typen gleich benannte Methoden haben, keine Unterschätzung. **Aus dem Quelltext und nicht aus einer Kopie**: eine
// eingecheckte Liste liefe der nächsten Bindung hinterher, und der Rundgang
// riefe sie mit falschen Argumenten.
package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type parameter struct {
	Name string `json:"name"`
	Typ  string `json:"typ"`
}

type methode struct {
	Name      string      `json:"name"`
	Parameter []parameter `json:"parameter"`
	Rueckgabe string      `json:"rueckgabe"`
	Datei     string      `json:"datei"`
	Dialog    bool        `json:"dialog"`
}

// gerufen sammelt die Namen, die ein Rumpf ruft — `x.Name(…)` und `Name(…)`.
func gerufen(rumpf *ast.BlockStmt) []string {
	namen := []string{}

	if rumpf == nil {
		return namen
	}

	ast.Inspect(rumpf, func(knoten ast.Node) bool {
		if ruf, ist := knoten.(*ast.CallExpr); ist {
			switch f := ruf.Fun.(type) {
			case *ast.SelectorExpr:
				namen = append(namen, f.Sel.Name)
			case *ast.Ident:
				namen = append(namen, f.Name)
			}
		}

		return true
	})

	return namen
}

// erreicht sagt, ob von `name` aus über `kanten` das Ziel erreicht wird.
func erreicht(name, ziel string, kanten map[string][]string, besucht map[string]bool) bool {
	if name == ziel {
		return true
	}

	if besucht[name] {
		return false
	}

	besucht[name] = true

	for _, weiter := range kanten[name] {
		if erreicht(weiter, ziel, kanten, besucht) {
			return true
		}
	}

	return false
}

func ausdruck(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return ausdruck(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + ausdruck(t.X)
	case *ast.ArrayType:
		return "[]" + ausdruck(t.Elt)
	case *ast.MapType:
		return "map[" + ausdruck(t.Key) + "]" + ausdruck(t.Value)
	case *ast.InterfaceType:
		return "any"
	case *ast.Ellipsis:
		return "..." + ausdruck(t.Elt)
	default:
		return "?"
	}
}

func main() {
	menge := token.NewFileSet()

	dateien, fehler := filepath.Glob(filepath.Join(os.Args[1], "*.go"))
	if fehler != nil {
		panic(fehler)
	}

	alle := []methode{}
	kanten := map[string][]string{}
	bindung := map[string][]string{}

	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}

		baum, fehler := parser.ParseFile(menge, datei, nil, 0)
		if fehler != nil {
			panic(fehler)
		}

		for _, deklaration := range baum.Decls {
			f, ist := deklaration.(*ast.FuncDecl)
			if !ist {
				continue
			}

			empfaenger := ""
			if f.Recv != nil {
				empfaenger = ausdruck(f.Recv.List[0].Type)
			}

			// **Auch die inneren Methoden von `*Bindungen`** — `waehlenUndOeffnen`
			// öffnet den Dialog für drei Bindungen (zweiter Rundgang).
			if empfaenger == "" || empfaenger == "*App" || (empfaenger == "*Bindungen" && !f.Name.IsExported()) {
				kanten[f.Name.Name] = append(kanten[f.Name.Name], gerufen(f.Body)...)
			}

			if empfaenger != "*Bindungen" || !f.Name.IsExported() {
				continue
			}

			bindung[f.Name.Name] = gerufen(f.Body)

			m := methode{Name: f.Name.Name, Parameter: []parameter{}, Datei: filepath.Base(datei)}

			for _, p := range f.Type.Params.List {
				if typ := ausdruck(p.Type); typ != "context.Context" {
					for _, name := range p.Names {
						m.Parameter = append(m.Parameter, parameter{name.Name, typ})
					}
				}
			}

			if f.Type.Results != nil && len(f.Type.Results.List) > 0 {
				m.Rueckgabe = ausdruck(f.Type.Results.List[0].Type)
			}

			alle = append(alle, m)
		}
	}

	for i := range alle {
		for _, weiter := range bindung[alle[i].Name] {
			if erreicht(weiter, "waehleDatei", kanten, map[string]bool{}) {
				alle[i].Dialog = true

				break
			}
		}
	}

	sort.Slice(alle, func(i, j int) bool { return alle[i].Name < alle[j].Name })

	if fehler := json.NewEncoder(os.Stdout).Encode(alle); fehler != nil {
		panic(fehler)
	}
}
