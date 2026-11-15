package domain_test

// **Welche Schnittstelle die Suche nimmt — oder automatisch** — NET-02.AK-03
// (#1706).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinSchnittstellennameWirdGeprueft(t *testing.T) {
	t.Parallel()

	for _, gut := range []string{"", "en0", "en10", "bridge100", "utun3", "lo0"} {
		if fehler := domain.PruefeEinstellung(domain.NetzSchnittstelle, gut); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", gut, fehler)
		}
	}

	for _, schlecht := range []string{"0en", "en 0", "en0;rm", "eine-viel-zu-lange-schnittstelle", "en0\n"} {
		if fehler := domain.PruefeEinstellung(domain.NetzSchnittstelle, schlecht); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%q ging durch: %v", schlecht, fehler)
		}
	}

	// **Leer ist die Automatik**, und sie gilt ohne Wahl.
	if vorgabe := domain.EinstellungsVorgabe(domain.NetzSchnittstelle); vorgabe != "" {
		t.Errorf("die Vorgabe ist %q", vorgabe)
	}

	// **Eine Angabe dieses Geräts** — eine Schnittstelle heißt auf einem
	// anderen Rechner anders.
	if ebene, da := domain.EbeneVon(domain.NetzSchnittstelle); !da || ebene != domain.EbeneGeraet {
		t.Errorf("die Ebene ist %q (%v)", ebene, da)
	}
}
