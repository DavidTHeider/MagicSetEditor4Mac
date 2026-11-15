package domain_test

// **Die Lage der Leiste** — UX-14.AK-02
// (#1686):
// *„horizontale Anordnung ist ebenfalls möglich"*. Senkrecht bleibt die
// Vorgabe — ein frischer Bestand sieht aus, wie er immer aussah.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieLeisteStehtSenkrechtOderWaagerecht(t *testing.T) {
	t.Parallel()

	if vorgabe := domain.EinstellungsVorgabe(domain.EinstellungLeistenlage); vorgabe != domain.LeisteSenkrecht {
		t.Errorf("die Vorgabe ist %q", vorgabe)
	}

	for _, wert := range []string{domain.LeisteSenkrecht, domain.LeisteWaagerecht} {
		if fehler := domain.PruefeEinstellung(domain.EinstellungLeistenlage, wert); fehler != nil {
			t.Errorf("%q wird abgewiesen: %v", wert, fehler)
		}
	}

	if fehler := domain.PruefeEinstellung(
		domain.EinstellungLeistenlage, "schraeg"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("eine dritte Lage kam durch: %v", fehler)
	}
}
