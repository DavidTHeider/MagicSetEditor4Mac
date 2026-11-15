package domain_test

// **Wer gegangen ist, stellt nichts mehr** — CR 800.4a (Fassung 2026-09-25),
// #1979.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein gegangener Sitz ist kein Absender mehr** — und die Verbliebenen
// bleiben es.
func TestEinGegangenerSitzIstKeinAbsenderMehr(t *testing.T) {
	t.Parallel()

	partie := eineDreierpartie(t).Scheidetaus("sitz-2", domain.GrundAufgabe)

	var befund *domain.FieldError

	fehler := partie.PruefeAbsender("sitz-2")
	if !errors.Is(fehler, domain.ErrConflict) || !errors.As(fehler, &befund) ||
		befund.Key != domain.SchluesselSitzGegangen {
		t.Errorf("der Gegangene: %v", fehler)
	}

	for _, sitz := range []string{"sitz-1", "sitz-3"} {
		if fehler := partie.PruefeAbsender(sitz); fehler != nil {
			t.Errorf("%s ist noch im Spiel: %v", sitz, fehler)
		}
	}

	if !partie.Gegangen("sitz-2") || partie.Gegangen("sitz-3") {
		t.Error("Gegangen nennt den falschen Sitz")
	}
}
