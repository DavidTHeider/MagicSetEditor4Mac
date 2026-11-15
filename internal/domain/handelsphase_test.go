package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJederZustandHatSeinePhase(t *testing.T) {
	t.Parallel()

	erwartet := map[domain.Handelszustand]domain.Handelsphase{
		domain.HandelEntwurf:            domain.HandelsphaseAngebot,
		domain.HandelAngeboten:          domain.HandelsphaseAngebot,
		domain.HandelVerhandlung:        domain.HandelsphaseAngebot,
		domain.HandelAngenommen:         domain.HandelsphaseZusagen,
		domain.HandelVorbereitet:        domain.HandelsphaseZusagen,
		domain.HandelUebergabeOffen:     domain.HandelsphaseVerbuchen,
		domain.HandelLokalVerbucht:      domain.HandelsphaseVerbuchen,
		domain.HandelAbgleichOffen:      domain.HandelsphaseVerbuchen,
		domain.HandelBeidseitigVerbucht: domain.HandelsphaseVerbuchen,
		domain.HandelKlaerung:           domain.HandelsphaseVerbuchen,
		domain.HandelAbgelehnt:          domain.HandelsphaseEnde,
		domain.HandelAbgebrochen:        domain.HandelsphaseEnde,
		domain.HandelAbgelaufen:         domain.HandelsphaseEnde,
	}

	// **Jeder Zustand und keiner doppelt** — ein neuer Zustand ohne Phase
	// fiele sonst still in die leere.
	if len(erwartet) != len(domain.Handelszustaende()) {
		t.Fatalf("%d Zustände, %d erwartet", len(domain.Handelszustaende()), len(erwartet))
	}

	for _, zustand := range domain.Handelszustaende() {
		if phase := zustand.Phase(); phase != erwartet[zustand] {
			t.Errorf("%s: Phase %q, erwartet %q", zustand, phase, erwartet[zustand])
		}
	}

	if phase := domain.Handelszustand("neu").Phase(); phase != domain.HandelsphaseEnde {
		t.Errorf("ein unbekannter Zustand: %q", phase)
	}
}

func TestDiePhasenStehenInIhrerReihenfolge(t *testing.T) {
	t.Parallel()

	if len(domain.DieHandelsphasen) != 5 ||
		domain.DieHandelsphasen[0] != domain.HandelsphaseGegenueber ||
		domain.DieHandelsphasen[4] != domain.HandelsphaseEnde {
		t.Fatalf("Phasen: %v", domain.DieHandelsphasen)
	}
}
