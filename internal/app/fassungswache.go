package app

// **Ein laufender Prozess fragt vor dem Schreiben, ob er den Bestand noch
// kennt** — M02
// (#2346),
// Befund A-11 des Audits 0.9.0.
//
// Ein Dienst unter launchd überlebt den Tausch des Bündels und läuft mit dem
// alten Code weiter; die neue Fassung wandert den Bestand unter ihm. Bis M02
// schrieb er danach weiter, in ein Schema, das er nicht kennt. **Jetzt holt
// er vor jedem Ruf eine Freigabe** und hält sie bis zu dessen Ende
// (`storage.HalteSchemafreigabe`): solange wandert niemand, und wenn eine
// Wanderung dasteht, die er nicht kennt, ruft er nicht und **endet** — der
// eingerichtete Dienst mit einem Code, nach dem launchd ihn mit der neuen
// Fassung neu startet (`cmd/mse-rest/wiederanlauf.go`).
//
// **Was das nicht kann**: ein Dienst aus 0.8.0 kennt diese Frage nicht. Ihn
// hält jede Wurzel an, bevor sie wandert (`dienstnachdemaufstieg.go`).

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// fassungswacheMit gibt die Freigabe, die ein Dienst oder Werkzeugserver vor
// jedem Ruf holt und bis zu seinem Ende hält — `halten` ist
// `storage.HalteSchemafreigabe` ([Fassungswache]). `beenden` läuft beim
// ersten Treffer, einmal.
func fassungswacheMit(
	halten func(context.Context) (func(), error), beenden func(),
) func(context.Context) (func(), error) {
	var einmal sync.Once

	return func(ctx context.Context) (func(), error) {
		frei, fehler := halten(ctx)

		// **Ein Lesefehler ist kein Ja** — Zweitblick über Phase 2, Befund 3:
		// hier galt er als Zustimmung, und ein Prozess, der nicht wusste, ob
		// er den Bestand kennt, schrieb hinein. **Er endet dafür nicht**;
		// beendet wird nur, wer nachweislich veraltet ist.
		var neuer *domain.NeuereFassung
		if !errors.As(fehler, &neuer) {
			return frei, fehler
		}

		einmal.Do(func() {
			slog.WarnContext(ctx, "der Bestand stammt aus einer neueren Fassung — dieser Prozess schreibt nicht mehr und endet",
				"stand", neuer.Stand)
			beenden()
		})

		return nil, &domain.ProzessVeraltet{Fassung: neuer}
	}
}
