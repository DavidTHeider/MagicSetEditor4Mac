package main

// **Die Gegenproben des Wächters über die Wege** — Befund 3 des Nachlaufs
// über Phase 1 von 0.9.0. Je Gestalt, an der der erste Anlauf vorbeisah, ein
// kleines Stück Quelltext, und ob der Wächter es anstreicht.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestDerWaechterSiehtJedenWeg(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		rumpf    string
		befunden bool
	}{
		"gemeldet in einem Zweig, der nicht betreten wird": {"if b { an.klageUeber(f) }\n\treturn CodeFachlich", true},
		"gemeldet in einer Schleife":                       {"for range xs { an.klageUeber(f) }\n\treturn CodeFachlich", true},
		"gemeldet in einem Literal, das niemand ruft":      {"go func() { an.klageUeber(f) }()\n\treturn CodeFachlich", true},
		"gemeldet in einer Bedingung":                      {"if x && an.klageUeber(f) == nil { }\n\treturn CodeFachlich", true},
		"eine Rückgabe direkt im Fall":                     {"switch { case b: return CodeFachlich }\n\treturn 0", true},
		"eine Rückgabe im sonst-wenn":                      {"if a { return 0 } else if b { return CodeFachlich }\n\treturn 0", true},
		"zwei Werte":                                       {"return CodeFachlich, true", true},
		"die Zahl selbst":                                  {"return 1", true},
		"eine Variable":                                    {"code := CodeFachlich\n\treturn code", true},
		"eine Zuordnung ohne Meldung":                      {"return paketCode(f)", true},
		"ein Literal, das selbst 1 gibt":                   {"g := func() int { return CodeFachlich }\n\treturn g()", true},
		"eine Bedingung, die am Helfer vorbeiläuft":        {"if a && !kenntSetIm(x) { return CodeFachlich }\n\treturn 0", true},

		"gemeldet davor":                       {"an.klageUeber(f)\n\treturn CodeFachlich", false},
		"gemeldet davor, zwei Werte":           {"an.klageMitSatz(f, \"s\")\n\treturn \"\", CodeFachlich", false},
		"gemeldet im umschließenden Block":     {"an.klageUeber(f)\n\tif x { return CodeFachlich }\n\treturn 0", false},
		"gemeldet im Fall":                     {"switch { case b: an.klageUeber(f); return CodeFachlich }\n\treturn 0", false},
		"der verneinte Helfer":                 {"if !kenntSetIm(x) { return CodeFachlich }\n\treturn 0", false},
		"ein Vergleich":                        {"if code == CodeFachlich { return 0 }\n\treturn 0", false},
		"hinter dem JSON-Weg":                  {"if an.alsJSON { return 0 }\n\treturn CodeFachlich", false},
		"gemeldet im Literal, das selbst gibt": {"g := func() int { an.klageUeber(f); return CodeFachlich }\n\treturn g()", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quelle := "package main\n\nconst CodeFachlich = 1\n\nfunc probe() int {\n\t" + fall.rumpf + "\n}\n"
			menge := token.NewFileSet()

			baum, err := parser.ParseFile(menge, "probe.go", quelle, 0)
			if err != nil {
				t.Fatalf("%v\n%s", err, quelle)
			}

			weg := rueckweg{menge: menge, einsen: einsenIn([]*ast.File{baum})}
			weg.pruefeDatei(baum)

			if befunden := len(weg.befund) > 0; befunden != fall.befunden {
				t.Errorf("angestrichen %v, erwartet %v: %v\n%s", befunden, fall.befunden, weg.befund, quelle)
			}
		})
	}
}
