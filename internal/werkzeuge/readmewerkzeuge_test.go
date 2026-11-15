// Die README nennt die MCP-Werkzeuge — dieser Wächter hält sie gegen den Code.
//
// # Der Anlass
//
// Sie sagte bis zum 2026-08-23
// (#900):
//
//	`lesen.sets`, `lesen.karten`, `lesen.felder`, `lesen.suche` — und ein
//	einziges schreibendes, `schreiben.feld`. Es gibt keines, das eine Karte
//	oder ein Set löscht.
//
// Es gibt **fünf** schreibende, und eines davon löscht seit
// #621. Der
// Satz war richtig, als er geschrieben wurde — und er ist eine **zweite
// Aufzählung** neben der Werkzeugliste, also die, die niemand nachführt (§9).
//
// Dieselbe Klasse wie die falsche Anweisung des Servers
// (#896):
// dort ist sie abgeleitet worden, hier lässt sie sich nicht ableiten — eine
// README ist Fließtext. Also wird sie **gehalten**.
//
// # Was er nicht prüft
//
//   - **Ob die Beschreibung stimmt.** Er hält Namen gegeneinander, nicht Sätze.
//     Ob „liest denselben Bestand wie das Fenster" heute noch gilt, sieht nur
//     jemand, der hinsieht.
//   - **Die Reihenfolge.** Ein Name, wo immer er im Abschnitt steht, zählt.
//   - **Andere Dokumente.** `docs/ARCHITEKTUR.md` nennt die Werkzeuge nicht
//     einzeln; täte es das, gehörte es hierher.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// werkzeugnamen gibt die Namen, die es wirklich gibt.
//
// **`nil` für alle drei Nähte**: gefragt wird nach den Namen, nicht nach dem,
// was die Werkzeuge tun. Kein `Ruf` wird ausgeführt.
func werkzeugnamen(t *testing.T) []string {
	t.Helper()

	// **Aus dem Quelltext und nicht aus `mcp.Werkzeuge(mcp.Naehte{})`**
	// (#1291).
	//
	// Jene Liste ist **kürzer als die Fläche**: mehrere Nähte dürfen `nil`
	// sein, und dann gibt es ihre Werkzeuge nicht — mit einer leeren
	// Nahtsammlung kamen 67 von 92 zurück. Ein Werkzeug hinter einer solchen
	// Naht stand damit in **keiner** Zeile der README, und der Wächter war
	// zufrieden.
	//
	// Das ist §10: **die Reichweite einer Messung ist die Reichweite ihrer
	// Aussage.** Gemessen am 2026-08-30 beim Einbau der Paketpflege — ihre
	// vier Werkzeuge fehlten in der README, und dieser Lauf blieb grün.
	//
	// **Und der Quelltext wird hier eigens gelesen** statt über
	// `ausQuelltext`: jene Hilfe steht in `package werkzeuge`, diese Datei in
	// `package werkzeuge_test`, und über die Paketgrenze reicht keine von
	// beiden hinüber.
	var namen []string

	eintraege, err := os.ReadDir("../mcp")
	if err != nil {
		t.Fatalf("internal/mcp lesen: %v", err)
	}

	muster := regexp.MustCompile(`"((?:lesen|schreiben)\.[a-z][a-z0-9.]*)"`)

	for _, eintrag := range eintraege {
		if eintrag.IsDir() || !strings.HasSuffix(eintrag.Name(), ".go") ||
			strings.HasSuffix(eintrag.Name(), "_test.go") {
			continue
		}

		roh, err := os.ReadFile(filepath.Join("..", "mcp", eintrag.Name()))
		if err != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name(), err)
		}

		for _, treffer := range muster.FindAllStringSubmatch(string(roh), -1) {
			if !slices.Contains(namen, treffer[1]) {
				namen = append(namen, treffer[1])
			}
		}
	}

	return namen
}

// mcpAbschnitt schneidet den Abschnitt „Der MCP-Server" aus der README.
//
// **Nur der Abschnitt und nicht die ganze Datei.** Ein Werkzeugname darf
// woanders vorkommen, ohne dass die Aufzählung damit vollständig wäre.
func mcpAbschnitt(t *testing.T) string {
	t.Helper()

	inhalt := lies(t, "README.md")

	_, nach, gefunden := strings.Cut(inhalt, "## Der MCP-Server")
	if !gefunden {
		t.Fatal("die README hat keinen Abschnitt „Der MCP-Server" + "\" mehr")
	}

	abschnitt, _, gefunden := strings.Cut(nach, "\n## ")
	if !gefunden {
		return nach
	}

	return abschnitt
}

// **Die README nennt jedes Werkzeug, das es gibt — und keines, das es nicht gibt.**
func TestDieREADMENenntJedesMCPWerkzeug(t *testing.T) {
	abschnitt := mcpAbschnitt(t)

	vorhanden := werkzeugnamen(t)
	if len(vorhanden) == 0 {
		t.Fatal("kein einziges Werkzeug — hat sich mcp.Werkzeuge bewegt?")
	}

	for _, name := range vorhanden {
		if !strings.Contains(abschnitt, "`"+name+"`") {
			t.Errorf("die README nennt %s nicht", name)
		}
	}

	for _, genannt := range namenImText(abschnitt) {
		if !enthaelt(vorhanden, genannt) {
			t.Errorf("die README nennt %s — das Werkzeug gibt es nicht", genannt)
		}
	}
}

// **Und sie verneint nichts, was es gibt.**
//
// Ein eigener Fall, weil er der Anlass ist: die Aufzählung kann vollständig
// sein und der Satz daneben trotzdem das Gegenteil behaupten.
func TestDieREADMEVerneintKeinVorhandenesWerkzeug(t *testing.T) {
	abschnitt := mcpAbschnitt(t)

	verneinungen := map[string]string{
		"schreiben.karte.loeschen": "Es gibt keines, das eine",
	}

	for werkzeug, verneinung := range verneinungen {
		if enthaelt(werkzeugnamen(t), werkzeug) && strings.Contains(abschnitt, verneinung) {
			t.Errorf("die README sagt %q — %s gibt es aber", verneinung, werkzeug)
		}
	}
}

// namenImText liest die Werkzeugnamen aus einem Abschnitt.
//
// **Über das Präfix in Zurückstrichen**: so schreibt die README einen
// Bezeichner, und so ist er von Fließtext zu unterscheiden.
func namenImText(text string) []string {
	var namen []string

	for _, stueck := range strings.Split(text, "`") {
		if strings.HasPrefix(stueck, "lesen.") || strings.HasPrefix(stueck, "schreiben.") {
			namen = append(namen, stueck)
		}
	}

	return namen
}

func enthaelt(liste []string, gesucht string) bool {
	for _, eintrag := range liste {
		if eintrag == gesucht {
			return true
		}
	}

	return false
}
