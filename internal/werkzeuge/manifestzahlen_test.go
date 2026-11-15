// **Die Formatfassungen im Freigabeskript sind die aus dem Fachkern** —
// Befund 6 des Zweitblicks über Phase 5.2
// (#1435).
//
// # Warum es diesen Wächter gibt
//
// `scripts/releasemanifest.sh` schreibt `"einstellungen"` und `"protokoll"`
// als **Zahlen im Skript**. Der Fachkern führt dieselben Zahlen als Konstanten
// und vergleicht sie beim Wechsel (`domain.Vertraeglichkeit`). Zwei Orte für
// dieselbe Aussage laufen auseinander — und die im Skript ist die, die in
// jede Freigabe geht.
//
// # Was er nicht hält
//
// **Ob die Zahl stimmt.** Er hält sie gegeneinander; ob das Format sich
// geändert hat, ohne dass jemand die Zahl erhöhte, sieht er nicht. Das ist
// dieselbe Grenze wie bei jedem Wächter über eine Absicht.
package werkzeuge

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieFormatfassungenImFreigabeskriptStimmen(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join("..", "..", "scripts", "releasemanifest.sh")

	roh, fehler := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
	if fehler != nil {
		t.Fatalf("%s lesen: %v", pfad, fehler)
	}

	for name, erwartet := range map[string]int{
		"einstellungen": domain.AustauschFassung,
		"protokoll":     domain.Protokollfassung,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			muster := regexp.MustCompile(fmt.Sprintf(`%q:\s*(\d+)`, name))

			treffer := muster.FindStringSubmatch(string(roh))
			if treffer == nil {
				t.Fatalf("das Skript schreibt %q nicht — der Wächter liefe ins Leere", name)
			}

			gelesen, umwandlungsfehler := strconv.Atoi(treffer[1])
			if umwandlungsfehler != nil {
				t.Fatalf("%q ist keine Zahl: %v", treffer[1], umwandlungsfehler)
			}

			if gelesen != erwartet {
				t.Errorf("das Skript schreibt %d, der Fachkern sagt %d",
					gelesen, erwartet)
			}
		})
	}
}
