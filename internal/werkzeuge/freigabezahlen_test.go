package werkzeuge_test

// **Die Zahlen im Freigabeblock sind der Kanonbeleg** — Abschluss von 0.6.0
// (#1660).
//
// # Der Anlass
//
// `freigabeblock_test.go` hält, dass der Block in `README.md` und
// `docs/RELEASE.md` gleich lautet — und sagt selbst, dass er **nicht** hält,
// ob seine Zahlen stimmen. Der Belegstand hielt den Block ausdrücklich nicht,
// weil er die Messung **jener** Freigabe beschrieb. Beides zusammen ließ ihn
// stehen: am Abschluss von 0.6.0 nannte er noch das Kanonmittel 20,056 vom
// 2026-09-20, während der Beleg 13,792 trug. Eine Release-Notiz daraus hätte
// einen Abstand angekündigt, den diese Fassung nicht mehr hat.
//
// # Was er hält
//
// **Der Block nennt den Tag des Belegs, und seine Tafel ist der Beleg** — auf
// die Stellen gerundet, die sie zeigt.
//
// # Was er nicht hält
//
// **Die Sätze um die Tafel** — den Anteil der Schrift, den Vergleich mit der
// vorigen Fassung. Die Zerlegung trägt der Beleg nicht als Kennzahl; sie
// steht im Rohprotokoll und in `docs/MESSMETHODIK.md`.

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDieTafelImFreigabeblockIstDerBeleg(t *testing.T) {
	t.Parallel()

	block := blockAus(t, "README.md")
	beleg := kanonbeleg(t)

	fest := func(name string) string { return deutsch(kennzahl(t, beleg, "mse2."+name), 3) }
	zufall := func(name string) string { return deutsch(kennzahl(t, beleg, "zufall."+name), 3) }

	for _, zeile := range []string{
		fmt.Sprintf("am **%s**", beleg.Ende.Format(time.DateOnly)),
		fmt.Sprintf("> | Karten | **%s** | **%s** |",
			deutsch(kennzahl(t, beleg, "mse2.stichprobe"), 0), deutsch(kennzahl(t, beleg, "zufall.stichprobe"), 0)),
		fmt.Sprintf("> | Kanonmittel | **%s** | **%s** |", fest("kanonmittel"), zufall("kanonmittel")),
		fmt.Sprintf("> | 95-%%-Band | — | [%s · %s] |", zufall("band.unten"), zufall("band.oben")),
		fmt.Sprintf("> | Median | %s | %s |", fest("median"), zufall("median")),
		fmt.Sprintf("> | Q1 · Q3 | %s · %s | %s · %s |", fest("q1"), fest("q3"), zufall("q1"), zufall("q3")),
		fmt.Sprintf("> | P99 | **%s** | %s |", fest("p99"), zufall("p99")),
		fmt.Sprintf("> | Minimum · Maximum | %s · %s | %s · %s |",
			fest("minimum"), fest("maximum"), zufall("minimum"), zufall("maximum")),
		fmt.Sprintf("**%s** Karten", deutsch(kennzahl(t, beleg, "mse2.grundgesamtheit"), 0)),
	} {
		if !strings.Contains(block, zeile) {
			t.Errorf("der Freigabeblock zeigt nicht, was der Beleg sagt — erwartet\n%s", zeile)
		}
	}
}
