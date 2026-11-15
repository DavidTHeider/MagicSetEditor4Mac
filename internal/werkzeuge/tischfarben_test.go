// **Die Farben des Tisches stehen in `KONTRASTE.md` mit ihrem Wert** —
// PLAY-03 (#1711).
//
// # Der Befund
//
// `docs/KONTRASTE.md` führt den Auswahlrahmen mit `#8a3d1e` (hell) und
// `#e8a05a` (dunkel). **`stil.css` führt `#2d1206` und `#fbf0dc`.**
//
// Die **Prosa unter der Tabelle** nennt dagegen die richtigen Zahlen: *„Heute
// sind es 3,03:1 im hellen und 3,07:1 im dunklen Erscheinungsbild."* Das sind
// exakt die Werte der **neuen** Farben gegen die Kante; die alten ergeben 1,32
// und 1,59 — und genau die nennt derselbe Absatz als den **gescheiterten**
// ersten Anlauf.
//
// Die Farbe wurde also gewechselt, die Schlussfolgerung nachgetragen, und die
// Tabelle darüber blieb stehen. Wer sie liest, liest vier Kontraste, die es
// nicht gibt.
//
// # Warum es niemand gesehen hat
//
// `kontraste_test.go` hält **Tailwind-Klassennamen** gegen die Datei —
// `--mse-tisch-*` sind Merkmale und keine Klassen, und er sagt das in seinem
// Kopf auch. Er hat also nicht versagt; es gab für diese sieben Farben gar
// keinen Wächter.
//
// # Was dieser hält
//
// Jeder Wert aus `--mse-tisch-*` in `frontend/src/stil.css` — hell wie dunkel
// — kommt in `docs/KONTRASTE.md` vor.
//
// # Was er **nicht** hält
//
//   - **Nicht, dass die Zahl daneben stimmt.** Nachmessen braucht einen
//     Browser, den die Pipeline nicht hat
//     ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)) — und
//     die Zahlen stehen von Hand erhoben da. Dieser Wächter hält nur, dass
//     über **diese** Farbe geredet wird.
//   - **Nicht, dass der genannte Grund der richtige ist.** Dieselbe Grenze wie
//     bei `kontraste_test.go`, und aus demselben Grund.
//   - **Nicht die fünf weiteren Tischbilder.** Sie stehen als Zahlen im
//     Quelltext, und `tischbilder.test.ts` rechnet jeden Satz nach — das ist
//     der bessere Wächter, und er ist schon da.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// tischmerkmale findet `--mse-tisch-…: #rrggbb;` in der Merkmalsdatei.
var tischmerkmale = regexp.MustCompile(`--mse-tisch-(\w+):\s*(#[0-9a-fA-F]{6})`)

// TestJedeTischfarbeStehtInKontraste hält die sieben gegen die Datei.
func TestJedeTischfarbeStehtInKontraste(t *testing.T) {
	t.Parallel()

	stil := lies(t, filepath.Join("frontend", "src", "stil.css"))
	kontraste := strings.ToLower(lies(t, filepath.Join("docs", "KONTRASTE.md")))

	gefunden := tischmerkmale.FindAllStringSubmatch(stil, -1)
	if len(gefunden) < 10 {
		t.Fatalf("nur %d Tischmerkmale gefunden — der Wächter greift daneben",
			len(gefunden))
	}

	for _, treffer := range gefunden {
		name, wert := treffer[1], strings.ToLower(treffer[2])

		if !strings.Contains(kontraste, wert) {
			t.Errorf("`--mse-tisch-%s` ist %s, und docs/KONTRASTE.md kennt "+
				"diesen Wert nicht — die Farbe ist gewechselt worden, und die "+
				"Messung daneben gilt einer anderen", name, wert)
		}
	}
}

// lies gibt den Inhalt einer Datei unterhalb der Wurzel.
func lies(t *testing.T, teile string) string {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join(wurzel, teile)) //nolint:gosec // fester Pfad
	if fehler != nil {
		t.Fatalf("%s lesen: %v", teile, fehler)
	}

	return string(roh)
}
