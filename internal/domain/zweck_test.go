package domain_test

// **Die Aufzählung der Zwecke** — Werkstück N1 von Phase 5.4
// (#1454).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Zweck der Aufzählung gilt, und sonst keiner.**
func TestNurDieAufgezaehltenZweckeGelten(t *testing.T) {
	t.Parallel()

	if len(domain.DieZwecke) != 4 {
		t.Fatalf("es sind %d Zwecke", len(domain.DieZwecke))
	}

	for _, zweck := range domain.DieZwecke {
		if fehler := domain.PruefeZweck(zweck); fehler != nil {
			t.Errorf("%q kam nicht durch: %v", zweck, fehler)
		}
	}

	for _, fremd := range []domain.Zweck{"", "alles", "Paarung", "teilen "} {
		if fehler := domain.PruefeZweck(fremd); !errors.Is(
			fehler, domain.ErrValidation) {
			t.Errorf("%q kam durch: %v", fremd, fehler)
		}
	}
}
