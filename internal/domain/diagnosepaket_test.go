package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Fünf Inhaltsteile — die Übersicht zählt nicht mit.** Sie ist die Vorschau
// und nennt die anderen; stünde sie in ihrer eigenen Liste, müsste sie ihre
// Größe nennen, bevor sie geschrieben ist.
func TestDieDiagnoseteileSindFuenfUndOhneUebersicht(t *testing.T) {
	t.Parallel()

	teile := domain.Diagnoseteile()

	if len(teile) != 5 {
		t.Fatalf("%d Teile, erwartet 5: %v", len(teile), teile)
	}

	if teile[0] != domain.TeilFassung {
		t.Errorf("der erste Teil ist %q", teile[0])
	}
}

// **Jeder Teil hat einen eigenen Namen.** Zwei gleiche wären zwei Einträge
// unter einem Namen im Archiv.
func TestJederDiagnoseteilHatEinenEigenenNamen(t *testing.T) {
	t.Parallel()

	gesehen := map[domain.Diagnoseteil]bool{}

	for _, teil := range domain.Diagnoseteile() {
		if teil == "" {
			t.Error("ein Teil ohne Namen")
		}

		if gesehen[teil] {
			t.Errorf("der Teil %q steht zweimal", teil)
		}

		gesehen[teil] = true
	}
}
