// **Keine exportierte Funktion, die nur Tests rufen, ohne Grund** — SEC04
// (#2357),
// Befund Z-02: die Erreichbarkeit von der Fassade bis in die Domäne.
//
// # Was er prüft
//
// Jede exportierte Funktion **ohne Empfänger** in `domain`, `service`,
// `storage`, `app` und `fenster` wird außerhalb der Tests genommen — gerufen
// oder als Wert weitergereicht, im eigenen Paket bloß mit Namen, in einem
// anderen mit dem Paket davor. **Eine, die nur Tests nehmen, steht in
// [nurVonTestsGenommen] mit ihrer Klasse**, und eine, die wieder einen
// Verbraucher hat oder verschwand, muss aus der Liste.
//
// # Warum ein Syntaxbaum und kein Muster
//
// Der erste Messlauf zu diesem Wächter war ein Muster in Python, und er nahm
// `"https://" + LinkhostFuer(…)` für einen Kommentar: ab `//` fehlte die
// Zeile, und eine benutzte Funktion stand als tot da. Derselbe Fehler steht
// in `bindungsaufrufer_test.go` als bekannte Grenze; hier liest der Parser,
// was Code ist.
//
// # Was er nicht prüft
//
//   - **Methoden.** Ein Typ erfüllt eine Schnittstelle, ohne dass sein
//     Methodenname irgendwo steht; ob eine Methode gerufen wird, sagt erst die
//     Typprüfung, und die läuft hier nicht.
//   - **Transitiv toten Code.** Ein Verbraucher, den selbst nur Tests nehmen,
//     zählt: `domain.Sitzungsreferenz` gilt als genommen, solange
//     `domain.ZeileZuReferenz` sie ruft — gemessen an der Gegenprobe am
//     2026-10-16, die deshalb grün blieb. Eine Funktion mit **einem**
//     Verbraucher im Betrieb (`service.GriffZiehen`) wird rot, sobald er
//     fehlt.
//   - **Ob der Verbraucher erreicht wird.** Ein Ruf in einem Zweig, den
//     niemand betritt, zählt — dafür ist die Abdeckung da.
//   - **Gleichnamige Bezeichner im eigenen Paket.** Eine lokale Variable, die
//     wie die Funktion heißt, zählt als Verbraucher; der Fehler geht zur
//     sicheren Seite (weniger Funde), nicht zur lauten.
//   - **Build-Tags**: der Parser liest jede Datei, gleich unter welchem Tag —
//     ein Verbraucher nur in `abnahme` oder `fremdpakete` zählt mit.
package werkzeuge

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// verbraucherpakete sind die Schichten aus §2, deren Exporte hier gezählt
// werden — die Fachbibliotheken sind Blätter mit eigenem Vertrag.
var verbraucherpakete = []string{"domain", "service", "storage", "app", "fenster"}

func TestJedeExportierteFunktionHatEinenVerbraucher(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..")
	modul := modulpfad(t, wurzel)
	dateien := quelldateienOhneTests(t, wurzel)

	exportiert := map[string]string{}
	genommen := map[string]bool{}

	for _, pfad := range dateien {
		datei, fehler := parser.ParseFile(token.NewFileSet(), pfad, nil, parser.SkipObjectResolution)
		if fehler != nil {
			t.Fatalf("%s: %v", pfad, fehler)
		}

		paket := paketAusPfad(wurzel, pfad)
		if slices.Contains(verbraucherpakete, paket) {
			for _, decl := range datei.Decls {
				if f, ist := decl.(*ast.FuncDecl); ist && f.Recv == nil && f.Name.IsExported() {
					exportiert[paket+"."+f.Name.Name] = pfad
				}
			}
		}

		verbraucheSammeln(datei, paket, importiertePakete(datei, modul), genommen)
	}

	if len(exportiert) < 500 {
		t.Fatalf("nur %d exportierte Funktionen gefunden — der Weg trifft nicht", len(exportiert))
	}

	for name := range exportiert {
		_, benannt := nurVonTestsGenommen[name]

		switch {
		case !genommen[name] && !benannt:
			t.Errorf("%s (%s) nimmt außerhalb der Tests niemand — einen Verbraucher geben, "+
				"entfernen oder mit Klasse in nurVonTestsGenommen eintragen", name, exportiert[name])
		case genommen[name] && benannt:
			t.Errorf("%s hat einen Verbraucher und steht trotzdem in nurVonTestsGenommen", name)
		}
	}

	for name := range nurVonTestsGenommen {
		if _, gibt := exportiert[name]; !gibt {
			t.Errorf("%s steht in nurVonTestsGenommen, ist aber keine exportierte Funktion mehr", name)
		}
	}
}

// verbraucheSammeln trägt ein, was eine Datei an Funktionen nimmt.
//
// **Erklärende Bezeichner zählen nicht**: der Name einer Funktion, eines
// Feldes oder eines Schlüssels in einem Verbundliteral steht da, ohne etwas
// zu nehmen.
func verbraucheSammeln(datei *ast.File, paket string, importe map[string]string, genommen map[string]bool) {
	var gehe func(ast.Node) bool

	gehe = func(knoten ast.Node) bool {
		switch k := knoten.(type) {
		case *ast.FuncDecl:
			if k.Recv != nil {
				ast.Inspect(k.Recv, gehe)
			}

			ast.Inspect(k.Type, gehe)

			if k.Body != nil {
				ast.Inspect(k.Body, gehe)
			}

			return false
		case *ast.SelectorExpr:
			if links, ist := k.X.(*ast.Ident); ist {
				if ziel, importiert := importe[links.Name]; importiert {
					genommen[ziel+"."+k.Sel.Name] = true

					return false
				}
			}

			ast.Inspect(k.X, gehe)

			return false
		case *ast.KeyValueExpr:
			if _, schluessel := k.Key.(*ast.Ident); !schluessel {
				ast.Inspect(k.Key, gehe)
			}

			ast.Inspect(k.Value, gehe)

			return false
		case *ast.Field:
			ast.Inspect(k.Type, gehe)

			return false
		case *ast.Ident:
			genommen[paket+"."+k.Name] = true
		}

		return true
	}

	for _, decl := range datei.Decls {
		ast.Inspect(decl, gehe)
	}
}

// importiertePakete gibt je Name im Quelltext das Paket direkt unter
// `internal/`, für das er steht — auch unter einem anderen Namen.
func importiertePakete(datei *ast.File, modul string) map[string]string {
	importe := map[string]string{}

	for _, imp := range datei.Imports {
		pfad, fehler := strconv.Unquote(imp.Path.Value)
		if fehler != nil {
			continue
		}

		paket, unter := strings.CutPrefix(pfad, modul+"/internal/")
		if !unter || strings.Contains(paket, "/") {
			continue
		}

		name := paket

		if imp.Name != nil {
			name = imp.Name.Name
		}

		importe[name] = paket
	}

	return importe
}

// paketAusPfad gibt das Paket unter `internal/`, in dem eine Datei liegt —
// leer für alles andere.
func paketAusPfad(wurzel, pfad string) string {
	relativ, fehler := filepath.Rel(wurzel, pfad)
	if fehler != nil {
		return ""
	}

	teile := strings.Split(filepath.ToSlash(relativ), "/")

	if len(teile) == 3 && teile[0] == "internal" {
		return teile[1]
	}

	return ""
}

// quelldateienOhneTests gibt jede Go-Datei des Repositoriums außer Tests,
// Werkzeugablagen und Prüfdaten.
func quelldateienOhneTests(t *testing.T, wurzel string) []string {
	t.Helper()

	var dateien []string

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() {
			switch eintrag.Name() {
			case ".go", ".git", "node_modules", "testdata", "graft", "frontend", ".huellenprobe":
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(pfad, ".go") && !strings.HasSuffix(pfad, "_test.go") {
			dateien = append(dateien, pfad)
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("Quelltext abgehen: %v", fehler)
	}

	return dateien
}

// modulpfad liest den Modulpfad aus `go.mod`.
func modulpfad(t *testing.T, wurzel string) string {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join(wurzel, "go.mod")) //nolint:gosec // Datei im Repositorium
	if fehler != nil {
		t.Fatal(fehler)
	}

	zeile, _, _ := strings.Cut(string(roh), "\n")

	return strings.TrimSpace(strings.TrimPrefix(zeile, "module"))
}
