package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Aktualisierung hat drei Stufen, und die Vorgabe ist „aus"** —
// #1100.
//
// §16 Regel 6: nichts wird still übernommen. Ein Programm, das beim ersten
// Start ungefragt ins Netz greift, ist eines, das etwas will — deshalb ist
// „aus" nicht irgendein Vorgabewert, sondern die Zusage.
func TestDieAktualisierungIstVorgabegemaessAus(t *testing.T) {
	t.Parallel()

	if vorgabe := domain.EinstellungsVorgabe(domain.EinstellungAktualisierung); vorgabe != domain.AktualisierungAus {
		t.Errorf("die Vorgabe ist %q, nicht %q", vorgabe, domain.AktualisierungAus)
	}

	werte := domain.ErlaubteEinstellungswerte(domain.EinstellungAktualisierung)

	gewollt := []string{
		domain.AktualisierungAus,
		domain.AktualisierungMelden,
		domain.AktualisierungHolen,
	}

	if len(werte) != len(gewollt) {
		t.Fatalf("%d Stufen, erwartet sind %d: %v", len(werte), len(gewollt), werte)
	}

	for stelle, wert := range gewollt {
		if werte[stelle] != wert {
			t.Errorf("Stufe %d ist %q, erwartet %q", stelle, werte[stelle], wert)
		}
	}

	// **Sie steht in den Anwendungseinstellungen**, nicht an einem Lauf: sie
	// gilt für das Programm und nicht für eine Ausgabe.
	if bereich := domain.EinstellungsBereich(domain.EinstellungAktualisierung); bereich != domain.BereichAnwendung {
		t.Errorf("Bereich = %q", bereich)
	}
}
