package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der lesbare Bezug eines Exemplars** —
// #1200.
//
// Die Sammlungsseite zeigte und **verlangte** die Scryfall-Kennung:
// `8a1b0d5e-1e0f-…` statt „Lightning Bolt · M11 #149".
func TestDieBezeichnungEinesExemplars(t *testing.T) {
	for _, fall := range []struct {
		name     string
		exemplar domain.Exemplar
		will     string
	}{
		{
			name:     "ohne Namen bleibt die Kennung",
			exemplar: domain.Exemplar{Printing: "8a1b0d5e"},
			will:     "8a1b0d5e",
		},
		{
			name:     "mit allem der volle Bezug",
			exemplar: domain.Exemplar{Name: "Lightning Bolt", Setcode: "m11", Nummer: "149"},
			will:     "Lightning Bolt · M11 #149",
		},
		{
			name:     "ohne Nummer bis zum Setcode",
			exemplar: domain.Exemplar{Name: "Lightning Bolt", Setcode: "m11"},
			will:     "Lightning Bolt · M11",
		},
		{
			name:     "ohne Setcode nur der Name",
			exemplar: domain.Exemplar{Name: "Lightning Bolt"},
			will:     "Lightning Bolt",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if haben := fall.exemplar.Bezeichnung(); haben != fall.will {
				t.Errorf("Bezeichnung = %q, erwartet %q", haben, fall.will)
			}
		})
	}
}

// **Der Bezug wird geglättet, nicht erfunden.**
func TestDerBezugWirdGeglaettet(t *testing.T) {
	exemplar := domain.Exemplar{
		Printing: "abc", Menge: 1,
		Name: "  Lightning Bolt ", Setcode: " M11 ", Nummer: " 149 ",
	}

	exemplar.Normalisieren()

	if exemplar.Name != "Lightning Bolt" || exemplar.Setcode != "m11" ||
		exemplar.Nummer != "149" {
		t.Errorf("geglättet: %+v", exemplar)
	}
}
