package service_test

// **Aus `angeboten` schlägt nur die Gegenseite vor** — an der Stelle, durch
// die jede Fläche geht
// (#1975).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Initiator wird abgewiesen und schreibt nichts; die Gegenseite geht
// durch.**
func TestAusAngebotenSchlaegtNurDieGegenseiteVor(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()
	erste := angebotenerHandel(t, leitung)
	vorschlag := erste.Fortgeschrieben(mitMenge(erste, 3))
	vorher := ablage.geschrieben

	if _, fehler := leitung.Vorschlagen(t.Context(), vorschlag, "anna"); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Initiator schlug vor: %v", fehler)
	}

	if ablage.geschrieben != vorher {
		t.Errorf("die Abweisung hat geschrieben: %d statt %d", ablage.geschrieben, vorher)
	}

	if _, fehler := leitung.Vorschlagen(t.Context(), vorschlag, "bert"); fehler != nil {
		t.Errorf("die Gegenseite wurde abgewiesen: %v", fehler)
	}
}
