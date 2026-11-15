package domain_test

// **Woher ein Deck oder ein Exemplar kam** — Werkstück 12 von Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Zwei Arten, und eine erfundene wird abgewiesen.**
func TestEineHerkunftsartWirdGeprueft(t *testing.T) {
	t.Parallel()

	for _, art := range domain.Herkunftsarten() {
		if fehler := art.Pruefe(); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", art, fehler)
		}
	}

	if fehler := domain.Herkunftsart("geerbt").Pruefe(); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die erfundene Art kam durch: %v", fehler)
	}

	// **Und die leere auch** — anders als bei der Sichtbarkeit gibt es hier
	// keinen Vorgabewert: eine Herkunft ohne Art ist keine Auskunft.
	if fehler := domain.Herkunftsart("").Pruefe(); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die leere Art kam durch: %v", fehler)
	}
}

// **Eine Herkunft aus einem Profil nennt das Profil.**
func TestEineHerkunftWirdGeprueft(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		herkunft domain.Herkunft
		feld     string
	}{
		"ohne Art": {
			domain.Herkunft{Quelle: "Ada"}, "art",
		},
		"ohne Quelle": {
			domain.Herkunft{Art: domain.HerkunftProfil, QuellProfil: 2}, "quelle",
		},
		"aus einem Profil ohne Kennung": {
			domain.Herkunft{Art: domain.HerkunftProfil, Quelle: "Ada"}, "quellProfil",
		},
		"aus einer Einfuhr mit Kennung": {
			domain.Herkunft{
				Art: domain.HerkunftEinfuhr, Quelle: "sammlung.csv", QuellProfil: 2,
			}, "quellProfil",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var feld *domain.FieldError
			if err := fall.herkunft.Validate(); !errors.As(err, &feld) ||
				feld.Field != fall.feld {
				t.Errorf("Fehler ist %v, erwartet am Feld %q", err, fall.feld)
			}
		})
	}
}

// **Und zwei vollständige gehen durch.**
//
// Die Gegenprobe zu den vier Läufen darüber: ohne sie wäre eine Prüfung, die
// **alles** abweist, ebenso grün.
func TestEineVollstaendigeHerkunftGehtDurch(t *testing.T) {
	t.Parallel()

	beide := []domain.Herkunft{
		{Art: domain.HerkunftProfil, Quelle: "Ada", QuellProfil: 2},
		{Art: domain.HerkunftEinfuhr, Quelle: "sammlung.csv"},
	}

	for _, eine := range beide {
		if err := eine.Validate(); err != nil {
			t.Errorf("%+v wurde abgewiesen: %v", eine, err)
		}
	}
}
