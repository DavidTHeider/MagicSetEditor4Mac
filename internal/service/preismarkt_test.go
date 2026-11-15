package service_test

// **Das Portfolio summiert Papier und Verkaufspreise** — Werkstück A3 von
// Phase 5.3 (#1442).
//
// # Warum das ein eigener Lauf ist
//
// Ein **Ankaufspreis** ist, was ein Händler zahlt — er liegt deutlich unter
// dem, was er verlangt. Ein **Onlinepreis** ist eine andere Ware. Wer beides
// mitsummierte, hätte eine Sammlung, die es nicht gibt; und weil beide Zahlen
// aussehen wie Preise, sähe es niemand.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestNurPapierUndVerkaufZaehlenInsPortfolio(t *testing.T) {
	t.Parallel()

	erhoben := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)

	faelle := map[string]struct {
		aendere func(*domain.Preisstand)
		zaehlt  bool
	}{
		"Papier und Verkauf": {func(*domain.Preisstand) {}, true},
		"ein Ankaufspreis": {
			func(p *domain.Preisstand) { p.Art = domain.PreisartAnkauf }, false,
		},
		"ein Onlinepreis": {
			func(p *domain.Preisstand) { p.Markt = domain.MarktOnline }, false,
		},
		"eine ungeprüfte Reihe": {
			func(p *domain.Preisstand) { p.Lage = domain.PreisUngeprueft; p.Wert = 0 },
			false,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stand := standVon(domain.PreisquelleScryfall, "eur", 400, erhoben)
			fall.aendere(&stand)

			ablage := &preisdoppel{staende: []domain.Preisstand{stand}}

			uebersicht, err := uebersichtUeber(t, ablage,
				[]*domain.Exemplar{stueck(probedruckkennung, 2, false)})
			if err != nil {
				t.Fatalf("Portfolio: %v", err)
			}

			if fall.zaehlt && len(uebersicht.Zeilen) != 1 {
				t.Fatalf("es sind %d Zeilen: %+v",
					len(uebersicht.Zeilen), uebersicht.Zeilen)
			}

			if !fall.zaehlt && len(uebersicht.Zeilen) != 0 {
				t.Errorf("gezählt wurde trotzdem: %+v", uebersicht.Zeilen)
			}
		})
	}
}
