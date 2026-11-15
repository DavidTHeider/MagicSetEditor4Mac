package service_test

// **Nur ein Tausch bewegt Bestand** — TRADE-03.AK-02
// (#1717).
//
// Der Kopf des Vertrags sagt es („vier Verträge, und nur einer davon bewegt
// Exemplare"); durchgesetzt hat es niemand. Ein Teilen oder eine Deckleihe,
// die bis zur Übergabe kam, verbuchte wie ein Tausch.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestNurEinTauschBewegtBestand(t *testing.T) {
	t.Parallel()

	for _, art := range []domain.Handelsart{domain.HandelTeilen, domain.HandelDeckleihe, domain.HandelMarkt} {
		t.Run(string(art), func(t *testing.T) {
			t.Parallel()

			stelle, stand, bestand := eineVerbuchung(t)
			stelleEin(t, stand, bestand, true)

			angebot := stand.staende[derVorgang]
			angebot.Art = art
			stand.staende[derVorgang] = angebot

			if _, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler ist %v", fehler)
			}

			if len(bestand.gebucht) != 0 {
				t.Errorf("es wurde %dmal gebucht", len(bestand.gebucht))
			}
		})
	}

	if !domain.HandelTausch.BewegtBestand() {
		t.Error("ein Tausch bewegt keinen Bestand")
	}
}
