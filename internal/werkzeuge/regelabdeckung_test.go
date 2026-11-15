// Der Wächter über die Abdeckungssicht auf die Comprehensive Rules.
//
// # Wofür
//
// **Eine Zahl über den Stand der Regelumsetzung veraltet still**, und diese
// steht an drei Stellen: in der Sicht selbst, in ihrer Kopftabelle und im
// `README.md`. Wer ein Kapitel von `offen` auf `teilweise` setzt und die
// Kopfzahl vergisst, hat eine Datei, die sich selbst widerspricht — und das
// `README.md` sagt dann eine vierte Zahl.
//
// **Gezählt wird die Aufzählung, nicht abgeschrieben.** Dieselbe Bauart wie
// bei `faehigkeiten_test.go` und `kontraste_test.go`: die Aufzählung ist die
// Wahrheit, jede Kopfzahl ist ihre Zusammenfassung, und der Wächter hält sie
// gegeneinander.
//
// # Was er nicht hält
//
// **Ob ein Stand stimmt.** Dass CR 122 `teilweise` ist und nicht `offen`,
// belegt die Zeile mit ihrer Fundstelle; ob die Fundstelle das trägt, prüft
// ein Mensch. Der Wächter zählt Zeilen — er liest sie nicht.
//
// **Und ob die Sicht vollständig ist.** Dass **alle** Kapitel der CR darin
// vorkommen, kann er nicht wissen: die CR liegt nicht im Repositorium und
// darf es nicht (§11). Was er prüft, ist die innere Stimmigkeit — und die
// **Summe**, die in der Sicht steht und die ein Mensch am Dokument geprüft
// hat.
package werkzeuge_test

import (
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// abdeckungsstaende sind die vier Stände, die eine Kapitelzeile tragen darf.
//
// **Vier und nicht drei**: „trifft nicht zu" ist eine Antwort und keine Lücke.
// Zweikopf-Riese und wechselnde Mannschaften liegen außerhalb des Umfangs, und
// das ist entschieden, nicht offen.
var abdeckungsstaende = []string{
	"umgesetzt", "teilweise", "offen", "trifft nicht zu",
}

// kapitelzeile findet „| **CR 122 — Counters** | teilweise | …".
//
// **Die zweite Spalte wird gegen die vier Stände gehalten und nicht einfach
// genommen**: dieselbe Gestalt tragen auch die Tabellen im Abschnitt über die
// Kapitel ohne Werkstück, und dort steht dort eine **Regelzahl**. Wer sie
// mitzählte, zählte Kapitel doppelt.
var kapitelzeile = regexp.MustCompile(`^\| \*\*CR (\d{3})[^|]*\| ([^|]+?) \|`)

// kopfzahl findet „| **offen** | **99** |" in der Zusammenfassung.
func kopfzahl(stand string) *regexp.Regexp {
	return regexp.MustCompile(
		`\|\s*\*\*` + regexp.QuoteMeta(stand) + `\*\*\s*\|\s*\*\*(\d+)\*\*\s*\|`)
}

// readmezahl findet „| offen | 99 |" im README.
func readmezahl(stand string) *regexp.Regexp {
	return regexp.MustCompile(
		`\|\s*` + regexp.QuoteMeta(stand) + `\s*\|\s*(\d+)\s*\|`)
}

func TestDieAbdeckungssichtZaehltSichSelbstRichtig(t *testing.T) {
	sicht := gelesen(t, filepath.Join("..", "..", "docs", "REGELABDECKUNG.md"))

	gezaehlt := zaehleKapitel(t, sicht)

	for _, stand := range abdeckungsstaende {
		genannt := zahlHinter(t, kopfzahl(stand), sicht,
			"docs/REGELABDECKUNG.md")

		if genannt != gezaehlt[stand] {
			t.Errorf("docs/REGELABDECKUNG.md nennt %d Kapitel als %q, "+
				"gezählt sind %d", genannt, stand, gezaehlt[stand])
		}
	}

	summe := zahlHinter(t, kopfzahl("Summe"), sicht, "docs/REGELABDECKUNG.md")

	if gesamt := summeUeber(gezaehlt); summe != gesamt {
		t.Errorf("docs/REGELABDECKUNG.md nennt %d Kapitel insgesamt, "+
			"gezählt sind %d", summe, gesamt)
	}
}

func TestDasReadmeNenntDenStandDerRegelumsetzung(t *testing.T) {
	sicht := gelesen(t, filepath.Join("..", "..", "docs", "REGELABDECKUNG.md"))
	liesmich := gelesen(t, filepath.Join("..", "..", "README.md"))

	gezaehlt := zaehleKapitel(t, sicht)

	for _, stand := range abdeckungsstaende {
		genannt := zahlHinter(t, readmezahl(stand), liesmich, "README.md")

		if genannt != gezaehlt[stand] {
			t.Errorf("README.md nennt %d Kapitel als %q, in der Sicht "+
				"sind es %d", genannt, stand, gezaehlt[stand])
		}
	}
}

// TestDerAbdeckungswaechterGreift ist die Gegenprobe (§10).
//
// **Ein Wächter, der nichts abweist, ist keiner.** Hier wird derselben
// Zählung eine Zeile untergeschoben, deren Stand sich geändert hat — und die
// Zahl muss sich mitbewegen.
func TestDerAbdeckungswaechterGreift(t *testing.T) {
	probe := strings.Join([]string{
		"| **CR 100 — General** | teilweise | `a.go:1` | — | M1 |",
		"| **CR 101 — Golden Rules** | offen | — | — | M9 |",
		"| **CR 707 — Copying** | 33 | trägt `wechselwirkungen` |",
	}, "\n")

	vorher := zaehleKapitel(t, probe)

	if vorher["teilweise"] != 1 || vorher["offen"] != 1 {
		t.Fatalf("die Probe zählt falsch: %v", vorher)
	}

	// **Die Zeile mit der Regelzahl darf nicht mitzählen** — sonst zählte
	// jedes Kapitel aus Abschnitt 7 ein zweites Mal.
	if summeUeber(vorher) != 2 {
		t.Fatalf("eine Zeile ohne Stand wurde mitgezählt: %v", vorher)
	}

	nachher := zaehleKapitel(t,
		strings.Replace(probe, "| offen |", "| umgesetzt |", 1))

	if nachher["offen"] != 0 || nachher["umgesetzt"] != 1 {
		t.Fatalf("ein geänderter Stand bewegte die Zählung nicht: %v", nachher)
	}
}

// zaehleKapitel zählt die Kapitelzeilen nach ihrem Stand.
func zaehleKapitel(t *testing.T, text string) map[string]int {
	t.Helper()

	gezaehlt := map[string]int{}
	gesehen := map[string]string{}

	for _, zeile := range strings.Split(text, "\n") {
		treffer := kapitelzeile.FindStringSubmatch(zeile)
		if treffer == nil {
			continue
		}

		stand := strings.TrimSpace(treffer[2])
		if !slices.Contains(abdeckungsstaende, stand) {
			continue
		}

		if vorher, schon := gesehen[treffer[1]]; schon {
			t.Fatalf("CR %s steht zweimal mit einem Stand — einmal als %q, "+
				"einmal als %q", treffer[1], vorher, stand)
		}

		gesehen[treffer[1]] = stand
		gezaehlt[stand]++
	}

	return gezaehlt
}

// zahlHinter liest die eine Zahl, die ein Muster fängt.
func zahlHinter(
	t *testing.T, muster *regexp.Regexp, text, woher string,
) int {
	t.Helper()

	treffer := muster.FindStringSubmatch(text)
	if treffer == nil {
		t.Fatalf("%s nennt keine Zahl für %q — die Kennzahlen fehlen",
			woher, muster.String())
	}

	zahl, fehler := strconv.Atoi(treffer[1])
	if fehler != nil {
		t.Fatalf("%s nennt %q, und das ist keine Zahl", woher, treffer[1])
	}

	return zahl
}

// summeUeber addiert die vier Stände.
func summeUeber(gezaehlt map[string]int) int {
	summe := 0
	for _, stand := range abdeckungsstaende {
		summe += gezaehlt[stand]
	}

	return summe
}
