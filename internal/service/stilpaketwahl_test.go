package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Ausfuhr fragt ein Stilpaket nach dem Set** — QA-04-F3
// (#2039).
// Leer heißt der Stil des Sets, ein Name der eigene Stil einer Karte; beide im
// Spiel des Sets gesucht, wie MSE ein Stilpaket über den Namen und das Spiel
// lädt (`stilpaketnamen`).
func TestEinStilpaketNachDemSet(t *testing.T) {
	t.Parallel()

	waehler := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{{ID: 1, Game: "probe", Style: "probe-schlicht"}}},
		saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	)

	for paket, erwartet := range map[string]string{
		"":            "Schlichter Stil der Saat",
		"probe-kunst": "Stil der Saat mit dem Bild der Karte",
		"kunst":       "Stil der Saat mit dem Bild der Karte",
	} {
		stil, err := waehler.Stilpaket(t.Context(), 1, paket)
		if err != nil {
			t.Fatalf("%q: %v", paket, err)
		}

		if stil.VollerName != erwartet {
			t.Errorf("%q: %q, erwartet %q", paket, stil.VollerName, erwartet)
		}
	}

	if _, err := waehler.Stilpaket(t.Context(), 1, "gibt-es-nicht"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlender Stil: %v", err)
	}

	if _, err := waehler.Stilpaket(t.Context(), 2, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fehlendes Set: %v", err)
	}
}
