package app

// **`spielen` und `netz` entlassen die Gäste am eigenen Tisch** — Z3 des
// Zweitblicks über Phase 6.12
// (#1905).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestAbschaltenEntlaesstDieGaeste(t *testing.T) {
	t.Parallel()

	for _, kennung := range []string{"spielen", "netz"} {
		t.Run(kennung, func(t *testing.T) {
			t.Parallel()

			anwendung := NeueApp("")
			anwendung.MitEinstellungen(&merkendeEinstellungen{werte: map[string]string{}})

			gast := &angenommenendoppel{}
			anwendung.partie.runde.Setze(einGastplatz(gast))

			anwendung.EinstellungSetzen(context.Background(),
				domain.SchalterEinstellung(kennung), domain.SchalterAus)

			gast.schloss.Lock()
			defer gast.schloss.Unlock()

			if !gast.geschlossen {
				t.Errorf("der Gast sitzt nach `%s` aus noch am Tisch", kennung)
			}
		})
	}
}
