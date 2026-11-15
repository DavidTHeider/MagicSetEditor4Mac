//go:build ignore

// Die Stellen, an denen ein Zufallsmutant entstehen kann — B05
// (#2224).
//
//	go run scripts/mutationsorte.go internal/domain internal/service internal/storage
//
// Gibt je Stelle Datei, Byteversatz, Länge, Urtext, Ersatz und Art als JSON.
// **Nur Quelltext außerhalb der Tests**, und **nur Operatoren mit einem
// Gegenstück derselben Gestalt** — ein Mutant, der nicht übersetzt, sagt
// nichts über die Tests, und eine Art, die fast nie übersetzt, verschwendete
// die Stichprobe:
//
//   - Vergleiche: `==`↔`!=`, `<`↔`<=`, `>`↔`>=`,
//   - Verknüpfungen: `&&`↔`||`,
//   - Rechnung: `+`↔`-` (nur zwischen Zahlen, an der Stelle nicht prüfbar —
//     was nicht übersetzt, wird als solches gezählt),
//   - Wahrheitswerte: `true`↔`false`,
//   - Bedingungen: `if x` → `if !(x)`.
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

type stelle struct {
	Paket   string `json:"paket"`
	Datei   string `json:"datei"`
	Zeile   int    `json:"zeile"`
	Versatz int    `json:"versatz"`
	Laenge  int    `json:"laenge"`
	Urtext  string `json:"urtext"`
	Ersatz  string `json:"ersatz"`
	Art     string `json:"art"`
}

var gegenstueck = map[token.Token]token.Token{
	token.EQL: token.NEQ, token.NEQ: token.EQL,
	token.LSS: token.LEQ, token.LEQ: token.LSS,
	token.GTR: token.GEQ, token.GEQ: token.GTR,
	token.LAND: token.LOR, token.LOR: token.LAND,
	token.ADD: token.SUB, token.SUB: token.ADD,
}

func art(t token.Token) string {
	switch t {
	case token.LAND, token.LOR:
		return "verknuepfung"
	case token.ADD, token.SUB:
		return "rechnung"
	default:
		return "vergleich"
	}
}

func main() {
	menge := token.NewFileSet()
	alle := []stelle{}

	for _, paket := range os.Args[1:] {
		dateien, fehler := filepath.Glob(filepath.Join(paket, "*.go"))
		if fehler != nil {
			panic(fehler)
		}

		for _, datei := range dateien {
			if strings.HasSuffix(datei, "_test.go") {
				continue
			}

			quelle, fehler := os.ReadFile(datei) //nolint:gosec // ein Pfad aus dem eigenen Baum
			if fehler != nil {
				panic(fehler)
			}

			baum, fehler := parser.ParseFile(menge, datei, quelle, 0)
			if fehler != nil {
				panic(fehler)
			}

			neu := func(pos token.Pos, laenge int, ersatz, wie string) {
				p := menge.Position(pos)
				alle = append(alle, stelle{paket, datei, p.Line, p.Offset, laenge, string(quelle[p.Offset : p.Offset+laenge]), ersatz, wie})
			}

			ast.Inspect(baum, func(knoten ast.Node) bool {
				switch k := knoten.(type) {
				case *ast.BinaryExpr:
					if ziel, gibt := gegenstueck[k.Op]; gibt {
						// **Ein `+` zwischen Zeichenketten** hat kein Gegenstück —
						// ein Literal links oder rechts verrät es.
						if k.Op == token.ADD || k.Op == token.SUB {
							for _, seite := range []ast.Expr{k.X, k.Y} {
								if lit, ist := seite.(*ast.BasicLit); ist && lit.Kind == token.STRING {
									return true
								}
							}
						}

						neu(k.OpPos, len(k.Op.String()), ziel.String(), art(k.Op))
					}
				case *ast.Ident:
					if k.Name == "true" || k.Name == "false" {
						ersatz := map[string]string{"true": "false", "false": "true"}[k.Name]
						neu(k.NamePos, len(k.Name), ersatz, "wahrheitswert")
					}
				case *ast.IfStmt:
					anfang := menge.Position(k.Cond.Pos()).Offset
					ende := menge.Position(k.Cond.End()).Offset
					neu(k.Cond.Pos(), ende-anfang, "!("+string(quelle[anfang:ende])+")", "bedingung")
				}

				return true
			})
		}
	}

	sort.SliceStable(alle, func(i, j int) bool {
		if alle[i].Datei != alle[j].Datei {
			return alle[i].Datei < alle[j].Datei
		}

		return alle[i].Versatz < alle[j].Versatz
	})

	if fehler := json.NewEncoder(os.Stdout).Encode(alle); fehler != nil {
		panic(fehler)
	}
}
