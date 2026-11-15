package domain_test

// **Ein Schritt gilt der Revision, die ein Mensch sah** — #2496, Rest aus
// Vermerk 0200, Punkt 12.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinSchrittGiltDerGesehenenRevision(t *testing.T) {
	t.Parallel()

	heute := einTausch(verbuchbar)
	heute.Revision = 2

	vorher := heute
	vorher.Revision = 1

	for _, fall := range []struct {
		name    string
		gesehen string
		will    error
	}{
		{"dieselbe Revision geht", heute.Abdruck(), nil},
		{"ohne Angabe wird nicht gefragt", "", nil},
		{"eine ältere ist überholt", vorher.Abdruck(), domain.ErrUeberholt},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if fehler := heute.PruefeGesehen(fall.gesehen); !errors.Is(fehler, fall.will) || (fall.will == nil) != (fehler == nil) {
				t.Errorf("es kam %v, erwartet %v", fehler, fall.will)
			}
		})
	}
}
