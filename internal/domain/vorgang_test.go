package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestNeueVorgangskennungIstAchtZeichenUndVerschieden(t *testing.T) {
	t.Parallel()

	gesehen := map[domain.Vorgangskennung]bool{}

	for range 200 {
		kennung := domain.NeueVorgangskennung()

		if len(kennung) != 8 {
			t.Fatalf("Länge = %d, erwartet 8: %q", len(kennung), kennung)
		}

		for _, zeichen := range kennung {
			ziffer := zeichen >= '0' && zeichen <= '9'
			buchstabe := zeichen >= 'a' && zeichen <= 'f'

			if !ziffer && !buchstabe {
				t.Fatalf("kein Hexzeichen in %q: %q", kennung, zeichen)
			}
		}

		gesehen[kennung] = true
	}

	// **Nicht „alle verschieden"**: bei 200 Ziehungen aus 2³² ist eine
	// Doppelung möglich, wenn auch unwahrscheinlich — ein Test, der sie
	// verbietet, flackert. Was hier zählt, ist, dass nicht **eine** Kennung
	// zweihundertmal herauskommt.
	if len(gesehen) < 190 {
		t.Fatalf("nur %d verschiedene Kennungen aus 200", len(gesehen))
	}
}

func TestAusgaengeSindZwei(t *testing.T) {
	t.Parallel()

	if got := domain.Ausgaenge(); len(got) != 2 ||
		got[0] != domain.AusgangGewaehrt || got[1] != domain.AusgangAbgewiesen {
		t.Fatalf("Ausgaenge() = %v", got)
	}
}
