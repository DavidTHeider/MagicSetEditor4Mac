// Das Lesen für den Wächter über das Abhängigkeitsinventar
// (`abhaengigkeiten_test.go`, ENG-04).
//
// Getrennt, weil beide zusammen über die Zeilengrenze aus AGENTS.md §2
// gingen. Der Schnitt folgt der Frage: dort steht, **was** gilt, hier, **wie**
// gelesen wird — das Inventar, die Importe beider Sprachen und die Actions.
package werkzeuge

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// inventardatei ist das Inventar aus ENG-04.AK-01.
const inventardatei = "docs/ABHAENGIGKEITEN.md"

// inventarzeile ist eine Zeile einer Tabelle des Inventars.
type inventarzeile struct {
	fassung string
	lizenz  string
	// verantwortet sind die Stellen in Codeschrift aus der Spalte
	// „Verantwortet" — Pakete, Dateien oder Verzeichnisse.
	verantwortet []string
}

// inventar liest die Tabelle unter einer Überschrift zweiter Ebene.
//
// **Fünf Spalten, oder die Zeile ist ein Befund**: Name, Fassung, Lizenz,
// Verantwortet, Wofür. Eine Zeile mit einer Spalte weniger fiele sonst still
// aus dem Vergleich — und das Inventar hätte eine Abhängigkeit, die niemand
// nachhält.
func inventar(t *testing.T, ueberschrift string) map[string]inventarzeile {
	t.Helper()

	var (
		zeilen      = map[string]inventarzeile{}
		imAbschnitt bool
		kopf        bool
	)

	for _, roheZeile := range strings.Split(string(projektdatei(t, inventardatei)), "\n") {
		if titel, ok := strings.CutPrefix(roheZeile, "## "); ok {
			imAbschnitt, kopf = strings.TrimSpace(titel) == ueberschrift, false

			continue
		}

		if !imAbschnitt || !strings.HasPrefix(strings.TrimSpace(roheZeile), "|") {
			continue
		}

		spalten := tabellenspalten(roheZeile)

		switch {
		case len(spalten) != 5:
			t.Fatalf("%s, %q: die Zeile hat %d Spalten statt fünf: %s",
				inventardatei, ueberschrift, len(spalten), roheZeile)
		case !kopf:
			kopf = true
		case strings.HasPrefix(spalten[0], "---"):
		default:
			name := ohneAuszeichnung(spalten[0])
			if _, doppelt := zeilen[name]; doppelt {
				t.Errorf("%s, %q: %s steht zweimal", inventardatei, ueberschrift, name)
			}

			zeilen[name] = inventarzeile{
				fassung:      ohneAuszeichnung(spalten[1]),
				lizenz:       ohneAuszeichnung(spalten[2]),
				verantwortet: inCodeschrift(spalten[3]),
			}
		}
	}

	if len(zeilen) == 0 {
		t.Fatalf("%s: unter %q steht keine Tabelle", inventardatei, ueberschrift)
	}

	return zeilen
}

// codeschrift findet die Stellen in Codeschrift einer Zelle.
var codeschrift = regexp.MustCompile("`([^`]+)`")

// inCodeschrift gibt die Stellen in Codeschrift einer Zelle.
func inCodeschrift(zelle string) []string {
	funde := codeschrift.FindAllStringSubmatch(zelle, -1)
	stellen := make([]string, 0, len(funde))

	for _, fund := range funde {
		stellen = append(stellen, fund[1])
	}

	return stellen
}

// direkteAnforderungen liest die Zeilen eines require-Blocks **ohne**
// `// indirect` — `gomodLesen` wirft den Kommentar weg und damit genau die
// Unterscheidung, um die es hier geht.
func direkteAnforderungen(t *testing.T, datei string) map[string]string {
	t.Helper()

	direkt := map[string]string{}
	imBlock := false

	for _, roheZeile := range strings.Split(string(projektdatei(t, datei)), "\n") {
		zeile := strings.TrimSpace(roheZeile)

		switch {
		case zeile == "require (":
			imBlock = true
		case zeile == ")":
			imBlock = false
		case strings.Contains(zeile, "// indirect"):
		case imBlock:
			if felder := strings.Fields(ohneKommentar(zeile)); len(felder) == 2 {
				direkt[felder[0]] = felder[1]
			}
		case strings.HasPrefix(zeile, "require "):
			if felder := strings.Fields(ohneKommentar(zeile)); len(felder) == 3 {
				direkt[felder[1]] = felder[2]
			}
		}
	}

	return direkt
}

// modulZu gibt das Modul, aus dem ein Paketpfad stammt — das längste
// passende Präfix, wie bei `modul.version`. Ohne Fund die leere Zeichenkette.
func modulZu(paketpfad string, module map[string]string) string {
	var bestes string

	for pfad := range module {
		if (paketpfad == pfad || strings.HasPrefix(paketpfad, pfad+"/")) && len(pfad) > len(bestes) {
			bestes = pfad
		}
	}

	return bestes
}

// goImporteure ordnet jedem Modul die Verzeichnisse zu, deren **Programm**-
// Dateien es importieren — ohne `_test.go`, denn was nur ein Test braucht,
// geht nicht ins Programm.
//
// Gelesen wird mit `go/parser` und nur der Kopf der Datei: eine Suche nach
// dem Modulpfad träfe auch die Wächter, die ihn als Zeichenkette nennen.
func goImporteure(t *testing.T, module map[string]string) map[string]map[string]bool {
	t.Helper()

	importeure := map[string]map[string]bool{}
	dateien := token.NewFileSet()

	err := fs.WalkDir(os.DirFS(wurzel), ".", func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() {
			return verzeichnisUebergehen(pfad)
		}

		// **Ein Verweis ist keine Quelle** — C06 (#2231): in einem
		// Arbeitsbaum ist `.go` ein Verweis auf den Hauptbaum, endet auf
		// „.go" und wurde als Datei gelesen.
		if !eintrag.Type().IsRegular() {
			return nil
		}

		if !strings.HasSuffix(pfad, ".go") || strings.HasSuffix(pfad, "_test.go") {
			return nil
		}

		kopf, err := parser.ParseFile(dateien, pfad, projektdatei(t, pfad), parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, einfuhr := range kopf.Imports {
			paket, err := strconv.Unquote(einfuhr.Path.Value)
			if err != nil {
				return err
			}

			if modul := modulZu(paket, module); modul != "" {
				if importeure[modul] == nil {
					importeure[modul] = map[string]bool{}
				}

				importeure[modul][path.Dir(pfad)] = true
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Importe nicht lesbar: %v", err)
	}

	return importeure
}

// verzeichnisUebergehen lässt aus, was nicht zum Wurzelmodul gehört oder
// kein Go-Quelltext dieses Projekts ist.
func verzeichnisUebergehen(pfad string) error {
	name := path.Base(pfad)

	switch {
	case pfad == ".":
		return nil
	case strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_"),
		name == "testdata", name == "node_modules",
		pfad == "tools", pfad == "frontend", pfad == "graft":
		return fs.SkipDir
	}

	return nil
}

// npmEinfuhr findet einen Paketnamen in einer Einfuhr — `from 'x'`,
// `import 'x'`, `import('x')`, `@import 'x'` und `@plugin 'x'` in CSS.
// Relative Pfade fängt das erste Zeichen ab.
var npmEinfuhr = regexp.MustCompile(`(?:\bfrom\s*|\bimport\s*\(?\s*|@import\s+|@plugin\s+|\brequire\(\s*)['"]([^'"./][^'"]*)['"]`)

// npmPaketname kürzt einen Einfuhrpfad auf den Paketnamen:
// `@mantine/core/styles.css` → `@mantine/core`, `vitest/config` → `vitest`.
func npmPaketname(einfuhr string) string {
	teile := strings.Split(einfuhr, "/")
	if strings.HasPrefix(einfuhr, "@") && len(teile) > 1 {
		return teile[0] + "/" + teile[1]
	}

	return teile[0]
}

// npmImporteure ordnet jedem Paketnamen die Dateien unter `frontend/` zu, die
// es einführen — Programm, Tests und Konfiguration.
func npmImporteure(t *testing.T) map[string][]string {
	t.Helper()

	importeure := map[string][]string{}

	err := fs.WalkDir(os.DirFS(wurzel), "frontend", func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := path.Base(pfad)

		switch {
		case eintrag.IsDir() && (name == "node_modules" || name == "dist" || name == "coverage"):
			return fs.SkipDir
		case eintrag.IsDir():
			return nil
		}

		switch path.Ext(name) {
		case ".ts", ".tsx", ".js", ".mjs", ".cjs", ".css":
		default:
			return nil
		}

		gesehen := map[string]bool{}

		for _, fund := range npmEinfuhr.FindAllStringSubmatch(string(projektdatei(t, pfad)), -1) {
			if paket := npmPaketname(fund[1]); !gesehen[paket] {
				gesehen[paket] = true
				importeure[paket] = append(importeure[paket], pfad)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Einfuhren der Oberfläche nicht lesbar: %v", err)
	}

	return importeure
}

// actionbindung ist eine Zeile `uses: <action>@<commit> # <fassung>`.
var actionbindung = regexp.MustCompile(`^\s*-?\s*uses:\s*([^@\s]+)@([0-9a-f]{40})\s+#\s*(v\S+)\s*$`)

// actionsLesen gibt je Action die Fassung aus dem Kommentar und die Datei, in
// der sie steht. Eine `uses:`-Zeile, die nicht an einen Commit gebunden ist,
// ist ein Befund.
func actionsLesen(t *testing.T) (fassungen, orte map[string]string) {
	t.Helper()

	fassungen, orte = map[string]string{}, map[string]string{}

	dateien, err := fs.Glob(os.DirFS(wurzel), ".github/workflows/*.yml")
	if err != nil || len(dateien) == 0 {
		t.Fatalf("keine Workflow-Datei gefunden: %v", err)
	}

	for _, datei := range dateien {
		for _, zeile := range strings.Split(string(projektdatei(t, datei)), "\n") {
			if !strings.Contains(zeile, "uses:") {
				continue
			}

			fund := actionbindung.FindStringSubmatch(zeile)
			if fund == nil {
				t.Errorf("%s: %q ist nicht an einen Commit mit Fassung gebunden (§1)",
					datei, strings.TrimSpace(zeile))

				continue
			}

			fassungen[fund[1]], orte[fund[1]] = fund[3], datei
		}
	}

	return fassungen, orte
}
