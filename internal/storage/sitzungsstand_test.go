package storage_test

// **Eine Sperre überlebt den Neustart** — ID-02
// (#1701):
// sonst höbe ⌘Q jede Sperre auf.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineSperreWirdFestgehalten(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if zustand, fehler := ablage.SitzungLesen(t.Context()); fehler != nil || zustand != domain.SitzungAngemeldet {
		t.Errorf("ohne Eintrag: %q %v", zustand, fehler)
	}

	for _, zustand := range []domain.Sitzungszustand{
		domain.SitzungGesperrt, domain.SitzungAbgemeldet, domain.SitzungAngemeldet,
	} {
		if fehler := ablage.SitzungFesthalten(t.Context(), zustand); fehler != nil {
			t.Fatal(fehler)
		}

		if gelesen, fehler := ablage.SitzungLesen(t.Context()); fehler != nil || gelesen != zustand {
			t.Errorf("%s: %q %v", zustand, gelesen, fehler)
		}
	}
}
