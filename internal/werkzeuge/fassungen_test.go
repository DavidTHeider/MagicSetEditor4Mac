package werkzeuge_test

// **Jede Fassung steht in `docs/FASSUNGEN.md`, und zwar mit ihrem Wert** —
// ENG-06 (#1731).
//
// # Was er prüft
//
//  1. **Jede Konstante, die eine Fassung trägt**, steht dort mit Paket, Namen
//     und Wert in derselben Zeile. Eine Fassung trägt, wessen Name auf
//     »fassung« endet — **und jede Kennung der Gestalt `mse4mac/<wort>/<zahl>`**:
//     sie heißt nicht so und bricht genauso, wenn sie sich ändert.
//  2. **Keine Zeile nennt eine Fassung, die es nicht mehr gibt.**
//  3. **Jeder genannte Lauf existiert.** Eine Belegspalte mit einem
//     erfundenen Namen sähe aus wie ein Beleg — dieselbe Klasse wie eine
//     erfundene Regelnummer (§11).
//
// # Was er nicht prüft
//
//   - **Ob die Spalten das Verhalten richtig beschreiben.** Dafür stehen die
//     Läufe daneben; ob ein genannter Lauf das Behauptete auch prüft, sieht nur
//     ein Mensch.
//   - **Ein Format ohne Fassung.** Was keine Konstante hat, findet er nicht;
//     die Datei führt die bekannten unter »Ohne eigene Fassung«, und eine neue
//     fehlt dort still.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// keineFassung sind Namen, die auf »fassung« enden und keine sind — mit
// dem, was sie stattdessen sind.
var keineFassung = map[string]string{
	"domain.RuecksprungAppfassung":             "eine Art des Rücksprungs",
	"domain.TeilFassung":                       "der Name eines Teils im Diagnosepaket",
	"domain.SchluesselBestandAeltereFassung":   "ein Fehlerschlüssel",
	"domain.SchluesselBestandNeuereFassung":    "ein Fehlerschlüssel",
	"domain.SchluesselSicherungAeltereFassung": "ein Fehlerschlüssel",
	"domain.SchluesselSicherungFremdeFassung":  "ein Fehlerschlüssel",
	"domain.SchluesselSicherungNeuereFassung":  "ein Fehlerschlüssel",
	"domain.SchluesselUmschlagFassung":         "ein Fehlerschlüssel",
}

// protokollkennung ist die Gestalt einer Fassung, die nicht so heißt.
var protokollkennung = regexp.MustCompile(`^/?mse4mac/[a-z]+/\d+$`)

// genannteKonstante ist ein `paket.Name` in Backticks.
var genannteKonstante = regexp.MustCompile("`([a-z0-9]+\\.[A-Za-z0-9]+)`")

// genannterLauf ist ein Testname in Backticks.
var genannterLauf = regexp.MustCompile("`(Test[A-Z][A-Za-z0-9_]*)`")

func TestJedeFassungStehtInDerUebersicht(t *testing.T) {
	roh, fehler := os.ReadFile(filepath.Join("..", "..", "docs", "FASSUNGEN.md"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	zeilen := strings.Split(string(roh), "\n")
	fassungen := fassungskonstanten(t)

	for name, wert := range fassungen {
		zeile := zeileMit(zeilen, "`"+name+"`")

		switch {
		case zeile == "":
			t.Errorf("%s = %s steht nicht in docs/FASSUNGEN.md", name, wert)
		case !strings.Contains(zeile, "`"+wert+"`"):
			t.Errorf("%s nennt dort nicht seinen Wert `%s`: %s", name, wert, zeile)
		}
	}

	for _, treffer := range genannteKonstante.FindAllStringSubmatch(string(roh), -1) {
		name := treffer[1]
		if _, bekannt := fassungen[name]; !bekannt && traegtFassung(name) {
			t.Errorf("docs/FASSUNGEN.md nennt %s — die Konstante gibt es nicht", name)
		}
	}

	laeufe := alleTestnamen(t)

	for _, treffer := range genannterLauf.FindAllStringSubmatch(string(roh), -1) {
		if !laeufe[treffer[1]] {
			t.Errorf("docs/FASSUNGEN.md nennt %s als Beleg — den Lauf gibt es nicht", treffer[1])
		}
	}
}

// traegtFassung sagt, ob ein Name nach einer Fassung aussieht.
func traegtFassung(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), "fassung")
}

// zeileMit gibt die erste Tabellenzeile, die den Text enthält.
//
// **Eine Tabellenzeile und nicht irgendeine**: im Fließtext darf eine Fassung
// genannt werden, ohne ihren Wert zu tragen — gezählt wird ihre Zeile.
func zeileMit(zeilen []string, text string) string {
	if stelle := slices.IndexFunc(zeilen, func(zeile string) bool {
		return strings.HasPrefix(zeile, "|") && strings.Contains(zeile, text)
	}); stelle >= 0 {
		return zeilen[stelle]
	}

	return ""
}

// fassungskonstanten liest jede Fassung aus dem Quelltext — Name auf Wert.
func fassungskonstanten(t *testing.T) map[string]string {
	t.Helper()

	gefunden := map[string]string{}

	for _, wurzel := range []string{"internal", "cmd"} {
		for _, pfad := range quelldateien(t, filepath.Join("..", "..", wurzel)) {
			datei, fehler := parser.ParseFile(token.NewFileSet(), pfad, nil, 0)
			if fehler != nil {
				t.Fatalf("%s lesen: %v", pfad, fehler)
			}

			for name, wert := range konstantenIn(datei) {
				if _, ausgenommen := keineFassung[name]; ausgenommen {
					continue
				}

				if traegtFassung(name) || protokollkennung.MatchString(wert) {
					gefunden[name] = wert
				}
			}
		}
	}

	return gefunden
}

// konstantenIn gibt jede Konstante einer Datei mit wörtlichem Wert.
func konstantenIn(datei *ast.File) map[string]string {
	konstanten := map[string]string{}

	for _, deklaration := range datei.Decls {
		gruppe, ist := deklaration.(*ast.GenDecl)
		if !ist || gruppe.Tok != token.CONST {
			continue
		}

		for _, angabe := range gruppe.Specs {
			werte, ist := angabe.(*ast.ValueSpec)
			if !ist {
				continue
			}

			for stelle, name := range werte.Names {
				if stelle >= len(werte.Values) {
					continue
				}

				if wort, ist := werte.Values[stelle].(*ast.BasicLit); ist {
					konstanten[datei.Name.Name+"."+name.Name] = wortlaut(wort)
				}
			}
		}
	}

	return konstanten
}

// wortlaut gibt den Wert ohne Anführungszeichen.
func wortlaut(wort *ast.BasicLit) string {
	if wort.Kind != token.STRING {
		return wort.Value
	}

	if wert, fehler := strconv.Unquote(wort.Value); fehler == nil {
		return wert
	}

	return wort.Value
}
