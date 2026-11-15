package netz_test

// **Jeder Weg zum `Do` fragt vorher** — der Nachzügler zu Phase 4.8
// (#1321),
// aufgefallen in A3 von Phase 5.3
// (#1442).
//
// # Warum es diesen Wächter gibt
//
// Am Feld `erlaubnis` stand „wird vor **jeder** Anfrage gefragt", und in
// `internal/werkzeuge/netzerlaubnis_test.go` stand ein Wächter, der jede
// Bauzeile in `internal/app` auf `.MitErlaubnis(` prüft. Beides zusammen las
// sich wie eine geschlossene Zusage.
//
// **Sie war es nicht.** `netz.Strom` rief `klient.unten.Do` ohne zu fragen —
// ein Dienst mit Schalter, der strömt, ging hinaus, obwohl er aus war. Der
// Wächter drüben konnte es nicht sehen: er prüft, ob eine **Erlaubnis
// gesetzt** wird, nicht, ob sie **gelesen** wird.
//
// Das ist genau der Satz aus §10: ein Wächter, der die halbe Zusage hält,
// macht die andere unsichtbar. Dieser hier hält die andere Hälfte.
//
// # Was er hält
//
// Jede Funktion in `internal/netz`, die `unten.Do(` ruft, ruft in derselben
// Funktion `darfFragen(`.
//
// # Was er nicht hält
//
// **Die Reihenfolge.** Er sieht beide Aufrufe, nicht welcher zuerst kommt —
// dass die Erlaubnis **vor** der Drossel steht, prüft [TestAuchEinStromFragtVorher]
// über die Wirkung, nicht dieser Lauf über den Text.

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// funktionsanfang erkennt den Beginn einer Funktion.
var funktionsanfang = regexp.MustCompile(`^func (?:\([^)]*\) )?([A-Za-z_][A-Za-z0-9_]*)`)

func TestJederWegZumDoFragtVorher(t *testing.T) {
	t.Parallel()

	eintraege, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("internal/netz lesen: %v", err)
	}

	gesehen := 0

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		gesehen += pruefeAufDo(t, name)
	}

	// **Die Gegenprobe zur Aufzählung**: findet der Wächter gar kein `Do`,
	// ist er grün und sagt nichts. Zwei sind es — `holen` für [netz.Hole] und
	// [netz.SendeJSON], und [netz.Strom].
	if gesehen != 2 {
		t.Errorf("in internal/netz stehen %d Aufrufe `unten.Do(`, erwartet 2 — "+
			"kommt einer dazu, gehört er in diese Zahl und in die Überlegung, "+
			"ob er die Erlaubnis liest", gesehen)
	}
}

// pruefeAufDo sieht jede Funktion einer Datei an und gibt die Zahl der `Do`.
func pruefeAufDo(t *testing.T, name string) int {
	t.Helper()

	roh, err := os.ReadFile(name) //nolint:gosec // ein Dateiname aus dem eigenen Verzeichnis
	if err != nil {
		t.Fatalf("%s lesen: %v", name, err)
	}

	gefunden := 0

	for _, koerper := range funktionen(string(roh)) {
		if !strings.Contains(koerper.text, "unten.Do(") {
			continue
		}

		gefunden++

		if !strings.Contains(koerper.text, "darfFragen(") {
			t.Errorf("%s: %s ruft `unten.Do(` ohne `darfFragen(` — ein Weg "+
				"nach draußen, den ein abgeschalteter Dienst nicht aufhält",
				name, koerper.name)
		}
	}

	return gefunden
}

// koerper ist eine Funktion mit ihrem Namen.
type koerper struct {
	name string
	text string
}

// funktionen zerlegt eine Datei in ihre Funktionen.
//
// **Über den Text und nicht über `go/ast`**: was gesucht wird, ist ein
// Aufruf in derselben Funktion, und dafür genügen die Zeilen zwischen zwei
// Köpfen. Ein Zerteiler brächte hier nichts, was der Wächter nutzt.
func funktionen(quelle string) []koerper {
	var alle []koerper

	laufend := koerper{name: "(außerhalb)"}

	for _, zeile := range strings.Split(quelle, "\n") {
		if kopf := funktionsanfang.FindStringSubmatch(zeile); kopf != nil {
			alle = append(alle, laufend)
			laufend = koerper{name: kopf[1]}
		}

		laufend.text += zeile + "\n"
	}

	return append(alle, laufend)
}
