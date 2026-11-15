package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Abweisung eines Bestands vor der Grundlinie** (J02, #2266, Vermerk
// 0189): er nennt, was ihm fehlt, bis wohin die Grundlinie reicht — und den
// Weg über 0.7.0.
func TestAeltereFassungNenntWasFehltUndDenWeg(t *testing.T) {
	t.Parallel()

	abweisung := &domain.AeltereFassung{
		Fehlend:    []string{"0097_faktorhaushalt.sql"},
		Grundlinie: "0097_faktorhaushalt.sql",
	}

	meldung := abweisung.Error()

	for _, teil := range []string{"0097_faktorhaushalt.sql", "0.7.0"} {
		if !strings.Contains(meldung, teil) {
			t.Errorf("die Meldung nennt %q nicht: %q", teil, meldung)
		}
	}

	if !errors.Is(abweisung, domain.ErrSicherung) {
		t.Error("die Abweisung gilt nicht als Sicherungsfehler")
	}
}

// **Fehlen viele, nennt sie die Spanne statt jede einzelne** — einem Bestand
// aus 0.1.0 fehlen 96 Schritte, und eine Meldung mit 96 Dateinamen liest
// niemand.
func TestAeltereFassungFasstVieleZusammen(t *testing.T) {
	t.Parallel()

	abweisung := &domain.AeltereFassung{
		Fehlend:    []string{"0002_b.sql", "0003_c.sql", "0004_d.sql", "0005_e.sql"},
		Grundlinie: "0005_e.sql",
	}

	meldung := abweisung.Error()

	for _, teil := range []string{"fehlen 4 Wanderungen", "von 0002_b.sql bis 0005_e.sql"} {
		if !strings.Contains(meldung, teil) {
			t.Errorf("die Meldung nennt %q nicht: %q", teil, meldung)
		}
	}

	if strings.Contains(meldung, "0003_c.sql") {
		t.Errorf("die Meldung zählt jede einzelne auf: %q", meldung)
	}
}
