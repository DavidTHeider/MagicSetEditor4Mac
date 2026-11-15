package service

// **Was ein Zustandswechsel an den festgehaltenen Exemplaren bewirkt** —
// Befund 3 des Zweitblicks über Phase 5.6
// (#1477).
//
// # Warum das am Wechsel hängt und nicht am Schritt
//
// **Weil derselbe Wechsel auf zwei Wegen entsteht.** Wer hier `uebergeben`
// geht, geht einen **Schritt**; wer drüben geht, schickt eine **Nachricht**,
// und hier entsteht der Wechsel durch [domain.Handelsangebot.NimmtStand].
// Bis zum 2026-09-19 hing die Wirkung am Schritt — und damit vermerkte nur
// der Rechner, der die **zweite** Stimme selbst abgab, seine Zusagen als
// übergeben. Beim anderen blieb der Abgang unbestätigt, und `verbuchen`
// scheiterte mit Unterdeckung: **ein echter Tausch, bei dem beide etwas
// weggeben, kam nie zum Abschluss.**
//
// # Und warum die Abschlussschritte hier nicht vorkommen
//
// **Weil sie keinen Wechsel an den Zusagen bedeuten**, sondern an Bestand und
// Beleg — und weil sie ohnehin nicht über die Leitung reisen
// ([domain.Handelsangebot.NimmtStand]).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zusagenNachWechsel führt aus, was dieser Wechsel bedeutet.
//
//   - Nach `uebergabeOffen` sind die Exemplare über den Tisch gegangen —
//     und **erst dadurch** findet die Verbuchung ihre Abgangszeilen.
//   - Ein Ende ohne Übergabe gibt frei, was niemand mehr braucht.
//   - Alles andere lässt die Zusagen in Ruhe, und das ist die Mehrzahl.
func zusagenNachWechsel(
	ctx context.Context, zusagen Zusagenwirkung, vorgang string,
	vorher, nachher domain.Handelszustand,
) error {
	switch domain.ZusagenwirkungFuer(vorher, nachher) {
	case domain.ZusagenUebergeben:
		return zusagen.Uebergeben(ctx, vorgang)

	case domain.ZusagenFreigeben:
		return zusagen.Freigeben(ctx, vorgang)
	}

	return nil
}
