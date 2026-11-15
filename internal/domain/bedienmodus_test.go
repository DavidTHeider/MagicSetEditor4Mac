package domain_test

// **Der Bedienmodus** — PLAY-02.AK-01
// (#1710).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Im Anfängermodus wird nichts abgewählt** — und die Ablehnung nennt den
// Weg hinaus.
func TestDerAnfaengermodusLaesstKeineAbwahlZu(t *testing.T) {
	t.Parallel()

	wahl := domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichCombat},
	}

	fehler := wahl.PasstZumModus(domain.ModusAnfaenger)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "freien Modus") {
		t.Errorf("die Ablehnung nennt den Ausweg nicht: %v", fehler)
	}
}

// **Ohne Abwahl geht der Anfängermodus durch** — er ist der Normalfall.
func TestDerAnfaengermodusOhneAbwahlGehtDurch(t *testing.T) {
	t.Parallel()

	if fehler := (domain.Regelwahl{}).PasstZumModus(
		domain.ModusAnfaenger); fehler != nil {
		t.Errorf("der leere Anfängermodus wurde abgewiesen: %v", fehler)
	}
}

// **Im freien Modus geht sie** — sonst gäbe es ihn nicht.
func TestDerFreieModusLaesstDieAbwahlZu(t *testing.T) {
	t.Parallel()

	wahl := domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichCombat},
	}

	if fehler := wahl.PasstZumModus(domain.ModusFrei); fehler != nil {
		t.Errorf("der freie Modus wies die Abwahl ab: %v", fehler)
	}
}

// **Ein Modus, den es nicht gibt, wird abgewiesen** — mit ihm im Satz.
func TestEinUnbekannterModusWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	fehler := (domain.Regelwahl{}).PasstZumModus("profi")

	if !errors.Is(fehler, domain.ErrValidation) ||
		!strings.Contains(fehler.Error(), "profi") {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Beide Modi stehen in der Aufzählung, und beide sind welche.**
func TestBeideBedienmodiSindWelche(t *testing.T) {
	t.Parallel()

	modi := domain.Bedienmodi()
	if len(modi) != 2 {
		t.Fatalf("es sind %d Modi", len(modi))
	}

	for _, modus := range modi {
		if !domain.IstBedienmodus(modus) {
			t.Errorf("%q steht in der Aufzählung und gilt nicht", modus)
		}
	}

	if domain.IstBedienmodus("") {
		t.Error("der leere Name gilt als Modus")
	}
}
