package storage

// **Die Fehlerwege der festgehaltenen Sitzung** — ID-02 (#1701).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieFestgehalteneSitzungMeldetIhreFehler(t *testing.T) {
	ablage := NeueNutzerablage(offeneDatenbank(t))

	if _, fehler := ablage.SitzungLesen(probekontext(t)); fehler == nil {
		t.Error("ohne Tafel gelesen")
	}

	for _, zustand := range []domain.Sitzungszustand{domain.SitzungGesperrt, domain.SitzungAngemeldet} {
		if fehler := ablage.SitzungFesthalten(probekontext(t), zustand); fehler == nil {
			t.Errorf("%s ohne Tafel festgehalten", zustand)
		}
	}
}
