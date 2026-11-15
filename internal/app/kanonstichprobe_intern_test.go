package app

import (
	"strings"
	"testing"
)

// **Der Vergleich Karte für Karte verlangt dieselbe Stichprobe** — Befund 2
// des Zweitblicks über E/F.
//
// Bis dahin entschieden nur die schlechter gewordenen **vergleichbaren**
// Karten. Eine leere frühere Wertedatei oder zwei Läufe über verschiedene
// Karten bestanden, ohne dass eine einzige Karte gehalten worden wäre — und
// 0188 verlangt den Vergleich über dieselbe gezogene Stichprobe.
func TestDerKartenvergleichVerlangtDieselbeStichprobe(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name      string
		vergleich kartenvergleich
		enthaelt  []string
	}{
		{name: "dieselbe Stichprobe, nichts schlechter", vergleich: kartenvergleich{verglichen: 3, gleich: 3}},
		{name: "keine einzige Karte vergleichbar", vergleich: kartenvergleich{}, enthaelt: []string{"keine Karte"}},
		{
			name:      "eine Karte nur im neuen Lauf",
			vergleich: kartenvergleich{verglichen: 2, gleich: 2, ohneVorher: 1},
			enthaelt:  []string{"nicht dieselbe", "1 ohne früheren Wert"},
		},
		{
			name:      "eine Karte nur im früheren Lauf",
			vergleich: kartenvergleich{verglichen: 2, gleich: 2, ohneNachher: 1},
			enthaelt:  []string{"nicht dieselbe", "1 nicht mehr gemessen"},
		},
		{
			name:      "eine Karte schlechter",
			vergleich: kartenvergleich{verglichen: 2, gleich: 1, schlechter: []kartenwert{{Name: "eins"}}},
			enthaelt:  []string{"1 Karten sind um mehr als 0.01 schlechter"},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gemeldet := strings.Join(fall.vergleich.abweichungen(0.01, "vorher.tsv"), "\n")

			if len(fall.enthaelt) == 0 && gemeldet != "" {
				t.Errorf("gemeldet: %s", gemeldet)
			}

			for _, teil := range fall.enthaelt {
				if !strings.Contains(gemeldet, teil) {
					t.Errorf("%q fehlt in %q", teil, gemeldet)
				}
			}
		})
	}
}
