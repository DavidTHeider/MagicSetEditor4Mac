package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Manakosten in geschweiften Klammern werden zu MSEs Schreibweise**
// (#1005).
//
// # Warum das hier steht und nicht in einem Formatpaket
//
// Weil es **zwei** fremde Quellen gibt, die dieselbe Schreibweise benutzen:
// MTG JSON und Scryfall schreiben beide `{2}{R}`. Beide Pakete sind Blätter
// und kennen einander nicht (§2) — eine Kopie in jedem wäre genau der Fall,
// vor dem `msepaket/paket.go` warnt: „Zwei Kopien dieser Fallunterscheidung
// liefen an dem Tag auseinander, an dem eine von ihnen eine dritte Form
// lernte."
//
// # Woher die Regel stammt
//
// **Belegt, nicht geraten**: `add_cards_scripts` in `magic.mse-game` legt
// Karten mit `casting_cost: "W"` an — ohne geschweifte Klammern. Was in einer
// Klammer steht, bleibt zusammen: aus `{W/U}` wird `W/U`.
func TestManakostenOhneKlammern(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct{ roh, erwartet string }{
		"eine Farbe":               {"{W}", "W"},
		"Zahl und Farbe":           {"{2}{W}", "2W"},
		"zwei gleiche":             {"{W}{W}", "WW"},
		"Hybrid bleibt zusammen":   {"{W/U}", "W/U"},
		"Phyrexianisch":            {"{W/P}", "W/P"},
		"eine große Zahl":          {"{15}", "15"},
		"nichts":                   {"", ""},
		"schon ohne Klammern":      {"2W", "2W"},
		"eine unvollständige Form": {"{2", "2"},
		"Leerzeichen dazwischen":   {"{2} {W}", "2W"},
		"das leere Kostenzeichen":  {"{0}", "0"},
		"Farblos, nicht Generisch": {"{C}", "C"},
		"X":                        {"{X}{R}", "XR"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if ist := domain.ManakostenOhneKlammern(fall.roh); ist != fall.erwartet {
				t.Errorf("%q wurde zu %q, erwartet war %q", fall.roh, ist, fall.erwartet)
			}
		})
	}
}
