package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Nullwert ist die strengste Art.**
//
// Wer einen Befund baut, ohne sich zu entscheiden, bekommt `Portierung` —
// andersherum verschwänden neue Befunde still in einer Sammelart
// (#513).
func TestDieVorgabeIstPortierung(t *testing.T) {
	var ohne domain.Befund

	if ohne.Art.Oder() != domain.ArtPortierung {
		t.Errorf("die Vorgabe ist %q, erwartet %q", ohne.Art.Oder(), domain.ArtPortierung)
	}

	for _, art := range []domain.Befundart{
		domain.ArtPortierung, domain.ArtVorlage, domain.ArtUmgebung,
	} {
		if art.Oder() != art {
			t.Errorf("%q wird zu %q", art, art.Oder())
		}
	}
}
