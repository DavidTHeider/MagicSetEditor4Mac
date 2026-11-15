package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Wahl ordnet die Liste nach nichts — und dann aufsteigend.**
//
//	: sort_cards_ascending(true)
//
// `src/data/settings.cpp`, Zeile 72, Fassung f9d9356 vom 2021-07-06. Werkstück
// 3 von Phase 3.7
// (#1168).
//
// **Die Vorgabe steht in der Domäne und nicht in `storage`.** Sie ist eine
// fachliche Aussage — „ohne Wahl gilt die Reihenfolge des Sets" —, und der
// Bestand ist nur der Ort, an dem sie fehlt.
func TestDieStandardkartenordnungOrdnetNachNichts(t *testing.T) {
	t.Parallel()

	ordnung := domain.StandardKartenordnung()

	if ordnung.Feld != "" || !ordnung.Aufsteigend {
		t.Errorf("%+v", ordnung)
	}
}
