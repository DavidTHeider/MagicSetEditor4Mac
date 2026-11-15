//go:build fmlauf

// **Der Beleg des Modellvergleichs** — aus `modellvergleich_test.go`
// herausgelöst, damit jede Datei eine Frage beantwortet: dort wird gefragt,
// hier gezählt und geschrieben.
package app_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// messung ist, was ein Modell über alle Runden gab.
type messung struct {
	name   string
	kalt   time.Duration
	runden [][]ergebnis
}

// zahl zählt die Ergebnisse einer Art über alle Runden.
func (m messung) zahl(art string) int {
	gezaehlt := 0

	for _, runde := range m.runden {
		for _, eines := range runde {
			if eines.art == art {
				gezaehlt++
			}
		}
	}

	return gezaehlt
}

// alle ist die Zahl der gestellten Fragen.
func (m messung) alle() int { return len(m.runden) * len(m.runden[0]) }

// formtreu zählt die Antworten mit gültigem JSON und dem verlangten Feld —
// gleich, ob die Sprache gehalten wurde.
func (m messung) formtreu() int { return m.zahl(artFertig) + m.zahl(artSprache) }

// gleichlauf zählt die Fragen, auf die beide Runden wörtlich dasselbe gaben.
func (m messung) gleichlauf() int {
	gleich := 0

	for stelle := range m.runden[0] {
		if m.runden[0][stelle].wert == m.runden[1][stelle].wert {
			gleich++
		}
	}

	return gleich
}

// quantil gibt das q-Quantil der Dauern über alle Runden (nächster Rang).
func (m messung) quantil(q float64) time.Duration {
	dauern := []time.Duration{}

	for _, runde := range m.runden {
		for _, eines := range runde {
			dauern = append(dauern, eines.dauer)
		}
	}

	slices.Sort(dauern)
	stelle := int(q*float64(len(dauern))+0.999999) - 1

	return dauern[max(0, min(stelle, len(dauern)-1))]
}

// schreibeVergleich legt `docs/belege/modellvergleich.md` an.
func schreibeVergleich(t *testing.T, satz testsatz, fragen []frage, messungen []messung) {
	t.Helper()

	var text strings.Builder

	fmt.Fprintf(&text, `# Die Textmodelle an denselben Fragen

**%s** — AI-02.AK-03 (#1743),
Commit %s, %s, macOS %s, Testsatz %s.

    go test -count=1 -tags fmlauf -run TestDieTextmodelleImVergleich -v ./internal/app/

**Gefragt wird, wie die Anwendung fragt**: Anweisung `+"`KIKartenanweisung`"+`, die
Auftragssätze des Formulars, ein Schema mit dem einen Pflichtfeld — über
dieselben Klienten, die die Anwendung einhüllt. **Kartenkontext und Schema
sind nachgebaut** wie in `+"`fmguete.md`"+`: die Zahlen gelten für diesen Testsatz
in dieser Form. Jedes Modell bekommt erst eine Aufwärmfrage (sie zählt als
„erste Antwort"), dann %d Runden über %d Fragen.

| Modell | Fragen | formtreu | Sprache gehalten | leer | kein JSON | abgelehnt | abgeschnitten | Fehler | erste Antwort | p50 | p95 | Runde 1 = Runde 2 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
`, time.Now().Format("2006-01-02 15:04"), befehl(t, "git", "rev-parse", "--short", "HEAD"),
		befehl(t, "sysctl", "-n", "hw.model"), befehl(t, "sw_vers", "-productVersion"), satz.Fassung,
		runden, len(fragen))

	for _, m := range messungen {
		fmt.Fprintf(&text, "| %s | %d | %d | %d | %d | %d | %d | %d | %d | %.2f s | %.2f s | %.2f s | %d von %d |\n",
			m.name, m.alle(), m.formtreu(), m.zahl(artFertig), m.zahl(artLeer), m.zahl(artKaputt),
			m.zahl(artAbgelehnt), m.zahl(artAbgeschnitten), m.zahl(artFehler), m.kalt.Seconds(),
			m.quantil(0.5).Seconds(), m.quantil(0.95).Seconds(), m.gleichlauf(), len(fragen))
	}

	text.WriteString(`
**formtreu** heißt: gültiges JSON mit dem verlangten Feld, nicht leer.
**Sprache gehalten** ist die grobe Probe aus ` + "`fmguete_test.go`" + ` und zählt nur
formtreue Antworten. **Ob ein Vorschlag passt, entscheidet der Lauf nicht** —
die Spalte „passt?" füllt, wer bewertet, und nennt sich.

## Runde 1, Frage für Frage

`)

	kopf := "| Karte | Sprache | Feld |"
	linie := "| ---: | --- | --- |"

	for _, m := range messungen {
		kopf += " " + m.name + " | passt? |"
		linie += " --- | --- |"
	}

	text.WriteString(kopf + "\n" + linie + "\n")

	for stelle, eine := range fragen {
		fmt.Fprintf(&text, "| %d | %s | %s |", eine.karte, eine.sprache, eine.feld)

		for _, m := range messungen {
			fmt.Fprintf(&text, " %s | |", zelle(m.runden[0][stelle]))
		}

		text.WriteString("\n")
	}

	ziel := filepath.Join("..", "..", "docs", "belege", "modellvergleich.md")
	if fehler := os.WriteFile(ziel, []byte(text.String()), 0o600); fehler != nil {
		t.Errorf("Beleg schreiben: %v", fehler)
	}

	t.Log(text.String())
}

// zelle ist eine Antwort in einer Tabellenzelle — mit Ausgang, wenn sie
// nicht fertig war.
func zelle(eines ergebnis) string {
	wert := strings.NewReplacer("|", "/", "\n", " ").Replace(eines.wert)
	if eines.art != artFertig {
		wert = "*" + eines.art + "*: " + wert
	}

	return fmt.Sprintf("%s (%.1f s)", wert, eines.dauer.Seconds())
}

// befehl gibt die erste Zeile eines Programms, oder „?".
func befehl(t *testing.T, programm string, argumente ...string) string {
	t.Helper()

	ctx, abbrechen := context.WithTimeout(t.Context(), 10*time.Second)
	defer abbrechen()

	// #nosec G204 -- die Programme stehen oben als Literale
	aus, fehler := exec.CommandContext(ctx, programm, argumente...).Output() //nolint:gosec // s. o.
	if fehler != nil {
		return "?"
	}

	return strings.TrimSpace(string(aus))
}
