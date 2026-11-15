package werkzeuge_test

// **Jede sprachabhängige Stelle steht im Inventar** — LOC-03.AK-01
// (#1689).
//
// # Wozu dieser Wächter da ist
//
// **Weil eine vergessene Stelle nicht auffällt.** Wer `localeCompare` ohne
// Nachdenken schreibt, bekommt eine Liste, die auf einem englischen Rechner
// anders sortiert ist als auf einem deutschen — und niemand sieht es, weil
// beide Läufe grün sind.
//
// # Was er nicht hält
//
// **Ob die Klasse stimmt.** Dass `weltbild.ts` technische Schlüssel sortiert
// und `paketsicht.ts` Menschentext, sieht nur ein Mensch; wer es
// verwechselt, bekommt eine Liste, die sich mit der Systemsprache ändert,
// ohne dass ein Lauf rot wird. Das steht auch im Dokument selbst.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// sprachabhaengig trifft die Formen, an denen Sprache ins Spiel kommt.
//
// **Und `toLocale…` gehört dazu** — LOC-03.AK-03
// (#1689).
// Hier standen nur `Intl`, `localeCompare` und `collate.New`, und damit hielt
// der Wächter die Hälfte seiner Zusage: gemessen am 2026-09-22 formatierte
// `Diagnosepaket.tsx` Größen mit `toLocaleString`, und zwei Suchen falteten
// Groß und Klein mit `toLocaleLowerCase` — alle drei standen nicht im
// Inventar, und der Lauf war grün (§10).
var sprachabhaengig = regexp.MustCompile(
	`Intl\.(NumberFormat|DateTimeFormat|Collator|getCanonicalLocales)|` +
		`localeCompare|collate\.New|toLocale(String|DateString|TimeString|LowerCase|UpperCase)\(`)

// **Jede Datei mit einer sprachabhängigen Stelle steht in
// `docs/LOKALISIERUNG.md`.**
func TestJedeSprachabhaengigeStelleStehtImInventar(t *testing.T) {
	t.Parallel()

	inventar := inventartext(t)

	fehlend := []string{}

	for _, datei := range sprachabhaengigeDateien(t) {
		if !strings.Contains(inventar, datei) {
			fehlend = append(fehlend, datei)
		}
	}

	if len(fehlend) > 0 {
		t.Errorf("diese Dateien benutzen Sprache und stehen nicht in "+
			"docs/LOKALISIERUNG.md:\n%s", strings.Join(fehlend, "\n"))
	}
}

// **Und keine Zeile im Inventar zeigt auf eine Datei, die es nicht gibt.**
//
// **Die Gegenprobe zur Zeile darüber** (§10): eine Liste, die nur wächst,
// führt nach dem ersten Umbau Stellen, die niemand mehr findet.
func TestJedeGenannteDateiGibtEs(t *testing.T) {
	t.Parallel()

	inventar := inventartext(t)

	genannt := regexp.MustCompile(
		"`((?:frontend/src|internal)/[A-Za-z0-9_/.]+\\.(?:go|ts|tsx))`")

	for _, treffer := range genannt.FindAllStringSubmatch(inventar, -1) {
		pfad := filepath.Join("..", "..", treffer[1])

		if _, fehler := os.Stat(pfad); fehler != nil {
			t.Errorf("das Inventar nennt %s — die Datei gibt es nicht",
				treffer[1])
		}
	}
}

// inventartext gibt das Inventar.
func inventartext(t *testing.T) string {
	t.Helper()

	roh, fehler := os.ReadFile("../../docs/LOKALISIERUNG.md")
	if fehler != nil {
		t.Fatalf("das Inventar lesen: %v", fehler)
	}

	return string(roh)
}

// sprachabhaengigeDateien sammelt jede Datei mit einer solchen Stelle.
func sprachabhaengigeDateien(t *testing.T) []string {
	t.Helper()

	gefunden := []string{}

	for _, wurzel := range []string{"../../frontend/src", "../../internal"} {
		fehler := filepath.Walk(wurzel,
			func(pfad string, angabe os.FileInfo, fehler error) error {
				if fehler != nil {
					return fehler
				}

				if !traegtSprache(t, pfad, angabe) {
					return nil
				}

				gefunden = append(gefunden,
					strings.TrimPrefix(filepath.ToSlash(pfad), "../../"))

				return nil
			})
		if fehler != nil {
			t.Fatalf("%s durchgehen: %v", wurzel, fehler)
		}
	}

	return gefunden
}

// traegtSprache sagt, ob eine Datei eine sprachabhängige Stelle hat.
//
// **Tests zählen nicht mit**: dort steht eine Stelle, um eine andere zu
// prüfen — und ein Inventar der Läufe wäre eines über sich selbst.
func traegtSprache(t *testing.T, pfad string, angabe os.FileInfo) bool {
	t.Helper()

	name := angabe.Name()

	if angabe.IsDir() || strings.Contains(pfad, "node_modules") ||
		strings.HasSuffix(name, "_test.go") ||
		strings.Contains(name, ".test.") {
		return false
	}

	if !strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, ".ts") &&
		!strings.HasSuffix(name, ".tsx") {
		return false
	}

	roh, fehler := os.ReadFile(pfad) //nolint:gosec // feste Wurzeln im Baum
	if fehler != nil {
		t.Fatalf("%s lesen: %v", pfad, fehler)
	}

	return sprachabhaengig.Match(roh)
}
