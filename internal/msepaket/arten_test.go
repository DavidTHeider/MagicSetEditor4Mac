package msepaket_test

// **Die sieben Paketarten als Liste** — FMT-01
// (#1690).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Jede Art der Liste liest sich aus ihrem Namen zurück, und es sind
// sieben** — wer eine achte aufnimmt, nimmt sie hier und in der
// Operationsmatrix auf (`docs/FAEHIGKEITEN.md`).
func TestDieSiebenArten(t *testing.T) {
	t.Parallel()

	arten := msepaket.Arten()

	if len(arten) != 7 {
		t.Fatalf("%d Arten: %v", len(arten), arten)
	}

	for _, art := range arten {
		if msepaket.ArtAusName("probe."+string(art)) != art {
			t.Errorf("%q liest sich nicht aus seinem Namen zurück", art)
		}
	}

	// **Das Set zuerst** — es ist die Datei, mit der gearbeitet wird; die
	// übrigen sind Vorlagen dafür.
	if arten[0] != msepaket.ArtSet || slices.Contains(arten, msepaket.ArtUnbekannt) {
		t.Errorf("Reihenfolge oder Inhalt: %v", arten)
	}
}
