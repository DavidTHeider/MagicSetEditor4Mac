// **Eine Protokollzeile ohne Kontext hat keine Vorgangskennung** — Werkstück
// O9 von Phase 5.2
// (#1433).
//
// # Was dieser Wächter hält
//
// `internal/app.geschwaerzt` stempelt die Kennung des Vorgangs auf jede Zeile,
// deren Kontext eine trägt. Sie kommt dorthin über `slog`s `…Context`-Form —
// `ErrorContext(ctx, …)` statt `Error(…)`. Wer die kurze Form nimmt, **obwohl
// ein Kontext in Reichweite ist**, schreibt eine Zeile, die niemand mehr einem
// Ruf zuordnen kann; und weil sie trotzdem im Protokoll steht, fällt es keinem
// Test auf.
//
// # Was er nicht hält
//
// **Ob der Kontext der richtige ist.** Eine Funktion mit zwei Kontexten
// bekäme hier keinen Befund, und ein `context.Background()` mitten im Rumpf
// erst recht nicht. Er prüft die **Gestalt** — dass die Form, die eine
// Kennung tragen kann, auch benutzt wird —, nicht die Herkunft.
//
// **Und Stellen ohne Kontext prüft er gar nicht.** Ein `slog.Default().Error`
// in einer Funktion ohne `context.Context` ist erlaubt: sie hat keinen
// Vorgang, an den sich etwas hängen ließe (der Start, das Aufräumen, ein
// Hintergrundlauf).
package werkzeuge_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// kurzeFormen sind die Protokollmethoden ohne Kontext.
var kurzeFormen = map[string]bool{
	"Error": true, "Warn": true, "Info": true, "Debug": true, "Log": true,
}

func TestJedeProtokollzeileMitKontextNimmtDieKontextform(t *testing.T) {
	t.Parallel()

	befunde := make([]string, 0, 2)

	for _, verzeichnis := range []string{"../../internal", "../../cmd"} {
		befunde = append(befunde, kurzformenIn(t, verzeichnis)...)
	}

	if len(befunde) != 0 {
		t.Errorf("%d Protokollzeile(n) ohne Kontextform:\n%s",
			len(befunde), strings.Join(befunde, "\n"))
	}
}

// kurzformenIn sucht in einem Verzeichnisbaum.
func kurzformenIn(t *testing.T, wurzel string) []string {
	t.Helper()

	var befunde []string

	menge := token.NewFileSet()

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil {
			return fehler
		}

		if eintrag.IsDir() || !strings.HasSuffix(pfad, ".go") ||
			strings.HasSuffix(pfad, "_test.go") {
			return nil
		}

		datei, lesefehler := parser.ParseFile(menge, pfad, nil, 0)
		if lesefehler != nil {
			return lesefehler
		}

		for _, erklaerung := range datei.Decls {
			funktion, istFunktion := erklaerung.(*ast.FuncDecl)
			if !istFunktion || funktion.Body == nil {
				continue
			}

			if kontextname(funktion) != "" {
				befunde = append(befunde, kurzformenImRumpf(menge, pfad, funktion)...)
			}
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s lesen: %v", wurzel, fehler)
	}

	return befunde
}

// kontextname gibt den Namen des Kontextparameters — leer, wenn es keinen
// benutzbaren gibt (`_` zählt nicht: er ist ausdrücklich weggeworfen).
func kontextname(funktion *ast.FuncDecl) string {
	for _, feld := range funktion.Type.Params.List {
		auswahl, istAuswahl := feld.Type.(*ast.SelectorExpr)
		if !istAuswahl || auswahl.Sel.Name != "Context" {
			continue
		}

		paket, istPaket := auswahl.X.(*ast.Ident)
		if !istPaket || paket.Name != "context" {
			continue
		}

		for _, name := range feld.Names {
			if name.Name != "_" {
				return name.Name
			}
		}
	}

	return ""
}

// kurzformenImRumpf sammelt die Rufe ohne Kontext.
func kurzformenImRumpf(menge *token.FileSet, pfad string, funktion *ast.FuncDecl) []string {
	var befunde []string

	ast.Inspect(funktion.Body, func(knoten ast.Node) bool {
		ruf, istRuf := knoten.(*ast.CallExpr)
		if !istRuf {
			return true
		}

		auswahl, istAuswahl := ruf.Fun.(*ast.SelectorExpr)
		if !istAuswahl || !kurzeFormen[auswahl.Sel.Name] || !ausSlog(auswahl.X) {
			return true
		}

		stelle := menge.Position(ruf.Pos())
		befunde = append(befunde, "  "+pfad+":"+strconv.Itoa(stelle.Line)+
			" — slog…"+auswahl.Sel.Name+"(…) statt "+auswahl.Sel.Name+"Context(ctx, …)")

		return true
	})

	return befunde
}

// ausSlog sagt, ob der Empfänger `slog` oder `slog.Default()` ist.
func ausSlog(empfaenger ast.Expr) bool {
	if paket, istPaket := empfaenger.(*ast.Ident); istPaket {
		return paket.Name == "slog"
	}

	ruf, istRuf := empfaenger.(*ast.CallExpr)
	if !istRuf {
		return false
	}

	auswahl, istAuswahl := ruf.Fun.(*ast.SelectorExpr)
	if !istAuswahl || auswahl.Sel.Name != "Default" {
		return false
	}

	paket, istPaket := auswahl.X.(*ast.Ident)

	return istPaket && paket.Name == "slog"
}
