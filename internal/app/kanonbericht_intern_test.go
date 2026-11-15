//go:build fremdpakete

// **Wie der Kanonlauf seine Zahlen ins Protokoll schreibt** — die Berichtsseite
// zu `kanonverteilung_intern_test.go` und `kanonstratum_intern_test.go`.
//
// # Warum getrennt und hinter dem Tag
//
// **Die Rechnung läuft in der Pipeline, der Bericht nicht.** Verteilung und
// geschichteter Schätzer werden bei **jedem** Lauf geprüft — an einer Reihe von
// 1 bis 10 und an einem Beispiel, das man nachrechnen kann. Sie brauchen keine
// fremden Pakete, und eine Statistik, die nur hinter einem Tag geprüft wird,
// ist eine Statistik, die niemand prüft.
//
// **Diese Zeilen dagegen ruft nur der Kanonlauf**, und der steht hinter
// `fremdpakete`. Ohne Tag hielte `golangci-lint` sie für tot — zu Recht: im
// Bau ohne Tag ruft sie niemand.
package app

import (
	"fmt"
	"slices"
	"testing"
)

// alsZeilen gibt die Verteilung als Tabelle für das Protokoll.
func (v verteilung) alsZeilen(name string) []string {
	return []string{
		fmt.Sprintf("  %s — %d Karten", name, v.anzahl),
		fmt.Sprintf("    Mittel      %8.3f     Streuung   %10.3f", v.mittel, v.streuung),
		fmt.Sprintf("    Median      %8.3f     Standardab.%10.3f", v.median, v.abweichung),
		fmt.Sprintf("    Minimum     %8.3f     Maximum    %10.3f", v.kleinster, v.groesster),
		fmt.Sprintf("    Q1 (25 %%)   %8.3f     Q3 (75 %%)  %10.3f", v.q1, v.q3),
		fmt.Sprintf("    P90         %8.3f     P95        %10.3f", v.p90, v.p95),
		fmt.Sprintf("    P99         %8.3f     IQA        %10.3f", v.p99, v.q3-v.q1),
	}
}

// meldeVerteilung schreibt die Kennzahlen ins Protokoll des Laufs.
func meldeVerteilung(t *testing.T, name string, sortiert []float64) verteilung {
	t.Helper()

	gerechnet := verteilungAus(sortiert)
	for _, zeile := range gerechnet.alsZeilen(name) {
		t.Log(zeile)
	}

	return gerechnet
}

// kohorte sagt, wie eine Auswahl entstand — und damit, was aus ihr folgt.
type kohorte int

// Die zwei Kohorten aus QA-02 (#1734).
const (
	// festeKohorte ist dieselbe Auswahl in jedem Lauf — für den Verlauf.
	festeKohorte kohorte = iota
	// zufallsKohorte ist einmal gezogen und gespeichert — für die Schätzung.
	zufallsKohorte
)

// alsZeilen gibt die geschichtete Schätzung als Tabelle für das Protokoll.
//
// **Standardfehler und Band nur für die Zufallsstichprobe** — beide setzen
// eine voraus, und die feste Auswahl ist keine (QA-02.AK-03, `MESSMETHODIK.md`).
// Bis zum Zweitblick über Phase 6.0 standen sie auch unter der festen.
func (k kanonschaetzung) alsZeilen(art kohorte) []string {
	schaetzung := []string{
		fmt.Sprintf("    Standardfehler  %10.4f", k.standardfehler),
		fmt.Sprintf("    95-%%-Band       [%.4f, %.4f]", k.unten, k.oben),
	}
	if art == festeKohorte {
		schaetzung = []string{"    (feste Auswahl — nur beschreibend: sie ist keine Zufallsstichprobe, und ein Vertrauensbereich gilt für sie nicht)"}
	}

	return slices.Concat([]string{
		fmt.Sprintf("  Voller Kanon — geschichtete Stichprobe über %d Sets", k.schichten),
		fmt.Sprintf("    Grundgesamtheit N   %8d Karten", k.ganz),
		fmt.Sprintf("    Stichprobe      n   %8d Karten  (%.2f %%)",
			k.gezogen, 100*float64(k.gezogen)/float64(k.ganz)),
		"",
		fmt.Sprintf("    Kanonmittel     %10.4f", k.mittel),
	}, schaetzung, []string{
		"",
		fmt.Sprintf("    Median          %10.4f     Standardabw. %10.4f",
			k.median, k.standardabweichung),
		fmt.Sprintf("    Minimum         %10.4f     Maximum      %10.4f",
			k.kleinster, k.groesster),
		fmt.Sprintf("    Q1 (25 %%)       %10.4f     Q3 (75 %%)    %10.4f", k.q1, k.q3),
		fmt.Sprintf("    IQA             %10.4f     Streuung     %10.4f",
			k.q3-k.q1, k.streuung),
		fmt.Sprintf("    P90             %10.4f     P95          %10.4f", k.p90, k.p95),
		fmt.Sprintf("    P99             %10.4f", k.p99),
	})
}

// kennzahlen gibt die Schätzung als Zeilen für den Beleg — QA-01
// (#1733).
//
// **`internal/laufbeleg` liest jede Zeile `kennzahl name=wert`**, und der
// Wächter über das README hält die Tabelle dort gegen den Beleg. Bis dahin
// stand sie da, weil jemand sie aus einem Protokoll abgeschrieben hatte.
//
// **Die feste Kohorte bekommt nur, was sie beschreibt** — ohne
// Standardfehler und Band, aus demselben Grund wie in [kanonschaetzung.alsZeilen].
func (k kanonschaetzung) kennzahlen(praefix string, art kohorte) []string {
	type kennwert struct {
		name string
		wert float64
	}

	werte := []kennwert{{"kanonmittel", k.mittel}}
	if art == zufallsKohorte {
		werte = append(werte, kennwert{"standardfehler", k.standardfehler},
			kennwert{"band.unten", k.unten}, kennwert{"band.oben", k.oben})
	}

	werte = append(werte, []kennwert{
		{"median", k.median},
		{"q1", k.q1},
		{"q3", k.q3},
		{"p90", k.p90},
		{"p95", k.p95},
		{"p99", k.p99},
		{"minimum", k.kleinster},
		{"maximum", k.groesster},
		{"standardabweichung", k.standardabweichung},
		{"streuung", k.streuung},
		{"iqa", k.q3 - k.q1},
	}...)

	zeilen := make([]string, 0, 2+len(werte))
	zeilen = append(zeilen,
		fmt.Sprintf("kennzahl %s.grundgesamtheit=%d", praefix, k.ganz),
		fmt.Sprintf("kennzahl %s.stichprobe=%d", praefix, k.gezogen))

	for _, einer := range werte {
		zeilen = append(zeilen, fmt.Sprintf("kennzahl %s.%s=%.4f", praefix, einer.name, einer.wert))
	}

	return zeilen
}

// meldeKanon schreibt die Schätzung der **festen Kohorte** ins Protokoll —
// als Tabelle für Menschen und als Kennzahlen `mse2.*` für den Beleg.
func meldeKanon(t *testing.T, schichten []schicht) kanonschaetzung {
	t.Helper()

	return meldeKanonAls(t, "mse2", festeKohorte, schichten)
}

// meldeKanonAls schreibt eine Schätzung unter einem Präfix — `zufall` für die
// Zufallsstichprobe (QA-02, #1734): zwei Kohorten, zwei Namensräume, und ein
// Beleg, der beide getrennt trägt (AK-04).
func meldeKanonAls(t *testing.T, praefix string, art kohorte, schichten []schicht) kanonschaetzung {
	t.Helper()

	geschaetzt := schaetzeKanon(schichten)
	for _, zeile := range append(geschaetzt.alsZeilen(art), geschaetzt.kennzahlen(praefix, art)...) {
		t.Log(zeile)
	}

	return geschaetzt
}
