// **Der Text, den man sieht** — Werkstück U5 von Phase 5.1
// (#1416).
package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestSichttextNimmtMarkenUndWeicheTrennerHeraus(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct{ roh, sicht string }{
		"ohne Marken":                 {"Grizzly Bears", "Grizzly Bears"},
		"Marken fallen weg":           {"<sym>R</sym> something <i>(note)</i>", "R something (note)"},
		"geschütztes Kleiner":         {"1 \u0001 2", "1 < 2"},
		"geschütztes Leerzeichen":     {"a\ueb00b", "a b"},
		"weicher Trenner verborgen":   {"Haste<sep-soft>, </sep-soft>Trample", "HasteTrample"},
		"zwei weiche Trenner":         {"a<sep-soft>x</sep-soft>b<sep-soft>y</sep-soft>c", "abc"},
		"Trenner ohne Ende bricht ab": {"a<sep-soft>x b", "a"},
		"offene Marke bis zum Ende":   {"a<b", "a"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if ist := domain.Sichttext(fall.roh); ist != fall.sicht {
				t.Errorf("Sichttext(%q) = %q, erwartet %q", fall.roh, ist, fall.sicht)
			}
		})
	}
}
