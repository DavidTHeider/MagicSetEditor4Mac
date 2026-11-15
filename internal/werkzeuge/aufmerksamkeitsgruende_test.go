// **Der Grund an einem Set steht in jedem Katalog** — LOC
// (#1963).
//
// # Der Befund
//
// Die Startseite zeigte die Gründe unter „Braucht Aufmerksamkeit" als den
// deutschen Satz aus Go, auch im englischen Fenster — „set, Zeile 12" und
// den Befund des Zerteilers. Seit #1963 reisen Schlüssel und Werte mit. Die
// Befunde des Zerteilers hält [TestJedeSkriptmeldungStehtInBeidenKatalogen],
// die der Feldart [TestDieFeldartsaetzeHabenIhrePlatzhalter]; der Grund an
// einem Set ist ein eigener Schlüssel, [domain.GrundVorbehalt], und den hält
// dieser Lauf — in allen acht Sprachen, denn der Wächter über die
// Zusatzsprachen vergleicht nur benannte Platzhalter und sieht `{0}` nicht.
//
// # Was er nicht hält
//
// **Ob das Fenster den Schlüssel nachschlägt.** Das hält
// `frontend/src/seiten/aufmerksamkeit.test.tsx`.
package werkzeuge_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerGrundEinesSetsStehtInJedemKatalog(t *testing.T) {
	t.Parallel()

	satz := regexp.MustCompile(`(?m)^\s+` + domain.GrundVorbehalt + `:\s*\n?\s*'([^']*)'`)

	for _, sprache := range []string{"de", "en", "es", "fr", "it", "ja", "ru", "zh"} {
		datei := "frontend/src/sprache/" + sprache + "/seitenstart.ts"
		inhalt := lies(t, datei)

		// **Im Block `gruende` und nirgends sonst** — derselbe Name an einer
		// anderen Stelle der Datei zählte sonst mit.
		anfang := strings.Index(inhalt, "\n    gruende: {")
		if anfang < 0 {
			t.Errorf("%s hat keinen Block gruende", datei)

			continue
		}

		block, _, _ := strings.Cut(inhalt[anfang:], "\n    },")

		treffer := satz.FindStringSubmatch(block)
		if treffer == nil {
			t.Errorf("%s kennt den Grund %q nicht", datei, domain.GrundVorbehalt)

			continue
		}

		// **Datei und Zeile** — so viele Werte setzt `setHinweise` in
		// `internal/app/aufmerksamkeit.go`.
		if !strings.Contains(treffer[1], "{0}") || !strings.Contains(treffer[1], "{1}") {
			t.Errorf("%s: %q nennt Datei oder Zeile nicht", datei, treffer[1])
		}
	}
}
