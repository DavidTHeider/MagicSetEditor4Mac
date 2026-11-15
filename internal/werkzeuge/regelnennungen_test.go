package werkzeuge_test

// **Wie viele Regelnummern der Quelltext nennt** — der Abschnitt „Eine Ebene
// feiner: die Regeln selbst" im `README.md`.
//
// # Wogegen dieser Wächter steht
//
// **Gegen eine Zahl, die niemand nachzählt.** Sie ist der einzige Wert im
// Readme, der eine Aussage über die **Tiefe** der Regelumsetzung macht — und
// eine Zahl in Fließtext veraltet still. Denselben Fehler hat dieses Projekt
// bei den Bindungen gemacht: „64", als es 142 waren (§16).
//
// # Was er **nicht** hält
//
// **Ob eine genannte Regel wirklich trägt.** Er zählt Nennungen und liest
// sie nicht. Genau das steht auch im Readme neben der Zahl: sie ist ein Maß
// für **Belegdichte** und keines für Fortschritt.
//
// **Und er zählt keine Zitate in der Dokumentation** — nur Quelltext. Eine
// Nummer in einer Markdown-Datei belegt kein Verhalten.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// regelnennung findet „CR 608.2c" und „CR 205" — mit und ohne Unterregel.
var regelnennung = regexp.MustCompile(`CR (\d{3})(\.\d+[a-z]*)?`)

// regelnennungszahl findet die Zahl aus der Tabelle des Abschnitts.
var regelnennungszahl = regexp.MustCompile(
	`\*\*Am Code genannte Regelnummern\*\*\s*\|\s*\*\*([\d.]+)\*\*`)

func TestDieGenannteRegelzahlStimmt(t *testing.T) {
	t.Parallel()

	nummern := map[string]bool{}

	for _, wurzel := range []string{"..", filepath.Join("..", "..", "cmd")} {
		sammleNennungen(t, wurzel, nummern)
	}

	// **Nur der Quelltext zählt** — eine Nummer in einer Markdown-Datei
	// belegt kein Verhalten.
	gezaehlt := len(nummern)

	fund := regelnennungszahl.FindStringSubmatch(lies(t, "README.md"))
	if fund == nil {
		t.Fatal("das README nennt die Zahl der Regelnennungen nicht mehr — " +
			"sie stand dort, und ein Wächter, der nichts findet, hält nichts")
	}

	genannt, _ := strconv.Atoi(strings.ReplaceAll(fund[1], ".", "")) //nolint:errcheck // ein Nichttreffer gibt 0 und ist ungleich

	if genannt != gezaehlt {
		t.Errorf("das README nennt %d Regelnummern, gezählt sind %d — "+
			"nachzählen mit `make kennzahlen` und die Tabelle richten",
			genannt, gezaehlt)
	}

	// **Und die Kapitelzahl daneben**: sie steht in derselben Tabelle und
	// liefe sonst mit ihr auseinander.
	kapitel := map[string]bool{}
	for nummer := range nummern {
		kapitel[nummer[:3]] = true
	}

	if !strings.Contains(lies(t, "README.md"),
		"aus denen mindestens eine genannt ist   |      "+
			strconv.Itoa(len(kapitel))) {
		t.Errorf("das README nennt nicht %d Kapitel", len(kapitel))
	}
}

// sammleNennungen trägt jede Regelnummer aus den Go-Dateien unter `wurzel`
// ein.
func sammleNennungen(t *testing.T, wurzel string, nummern map[string]bool) {
	t.Helper()

	gelaufen := filepath.Walk(wurzel,
		func(pfad string, info os.FileInfo, fehler error) error {
			if fehler != nil || info.IsDir() ||
				!strings.HasSuffix(pfad, ".go") {
				return nil //nolint:nilerr // ein unlesbarer Zweig ist kein Fund
			}

			roh, gelesen := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus dem Baum
			if gelesen != nil {
				return nil //nolint:nilerr // siehe oben
			}

			for _, fund := range regelnennung.FindAllStringSubmatch(
				string(roh), -1) {
				nummern[fund[1]+fund[2]] = true
			}

			return nil
		})
	if gelaufen != nil {
		t.Fatalf("den Quelltext durchgehen: %v", gelaufen)
	}
}
