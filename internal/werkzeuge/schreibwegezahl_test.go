package werkzeuge_test

// **Die Zahl über der Tabelle der Schreibwege stimmt mit ihr überein** —
// Werkstück 8 von Phase 4.12
// (#1351).
//
// # Warum das ein dritter Wächter ist und keine dritte Zeile im zweiten
//
// **Weil eine Tabelle anders gezählt wird als eine Aufzählung.**
// `TestDieUeberschriftenZaehlenIhreListe` liest Einträge zwischen Backticks;
// hier ist jede **Zeile** ein Eintrag, und die Kopfzeile und der Trennstrich
// zählen nicht mit.
//
// # Was er gefunden hat
//
// **Die Überschrift nannte 49, die Tabelle führte 63** — gemessen am
// 2026-09-05, bevor Werkstück 8 seine Zeile dazulegte. Das ist genau der
// Fall, den `schreibwege_test.go` in seinem Kopf beschreibt („die Überschrift
// nennt daneben eine Zahl, und die prüfte niemand") — und die Datei nannte
// dabei **zwei** Überschriften, nicht drei. Ein Wächter, der die halbe Zusage
// hält, macht die andere unsichtbar (§10).

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// tabellenzahl findet `## Die <Zahl> Schreibwege`.
var tabellenzahl = regexp.MustCompile(`## Die (\d+) Schreibwege`)

func TestDieUeberschriftZaehltIhreTabelle(t *testing.T) {
	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "SCHREIBWEGE.md"))
	if err != nil {
		t.Fatalf("SCHREIBWEGE.md lesen: %v", err)
	}

	treffer := tabellenzahl.FindStringSubmatch(string(roh))
	if treffer == nil {
		t.Fatal("keine Überschrift »## Die <Zahl> Schreibwege«")
	}

	genannt, umwandlungsfehler := strconv.Atoi(treffer[1])
	if umwandlungsfehler != nil {
		t.Fatalf("die Zahl ist keine: %v", umwandlungsfehler)
	}

	gezaehlt := schreibwegezeilen(string(roh), treffer[0])

	if genannt != gezaehlt {
		t.Errorf("die Überschrift nennt %d, die Tabelle führt %d Zeilen",
			genannt, gezaehlt)
	}

	// **Die Reichweite der Messung** (§10): eine Tabelle, die der Ausdruck
	// nicht mehr findet, wäre null — und null gleich null wäre grün.
	if gezaehlt < 40 {
		t.Errorf("nur %d Zeilen gefunden — liest der Ausdruck die Tabelle noch?",
			gezaehlt)
	}
}

// schreibwegezeilen zählt die Datenzeilen der Tabelle nach dieser Überschrift.
//
// **Eine Zeile beginnt mit »| `«** — die Kopfzeile beginnt mit »| Datei«, der
// Trennstrich mit »| ---«. Gezählt wird bis zur nächsten Überschrift.
func schreibwegezeilen(text, ueberschrift string) int {
	nach := text[strings.Index(text, ueberschrift)+len(ueberschrift):]

	gezaehlt := 0

	for _, zeile := range strings.Split(nach, "\n") {
		if strings.HasPrefix(zeile, "## ") {
			break
		}

		if strings.HasPrefix(zeile, "| `") {
			gezaehlt++
		}
	}

	return gezaehlt
}
