package rest_test

// **Die Doppel der Problemprüfung** — PERF02
// (#2352).
// Der Zähler zählt jeden Ruf; ohne seine eigenen Methoden riefe er still den
// willigen, und die Probe »abgewiesen heißt nicht gerufen« sähe nichts.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Problemstand(context.Context, int64) (rest.Problemstand, error) {
	return rest.Problemstand{Urteil: "ungeprueft"}, nil
}

func (williger) ProblemePruefen(context.Context, int64) (rest.Problemstand, error) {
	return rest.Problemstand{
		Urteil: "aktuell", Sicher: true, Zahl: 1, Geprueft: 2, Gesamt: 2,
		Karten: []rest.Kartenproblem{{KarteID: "7", Befunde: []string{"die Schrift fehlt"}}},
	}, nil
}

func (u unwilliger) Problemstand(context.Context, int64) (rest.Problemstand, error) {
	return rest.Problemstand{}, u.fehler
}

func (u unwilliger) ProblemePruefen(context.Context, int64) (rest.Problemstand, error) {
	return rest.Problemstand{}, u.fehler
}

func (z *zaehlender) Problemstand(ctx context.Context, setID int64) (rest.Problemstand, error) {
	z.zaehle()

	return z.williger.Problemstand(ctx, setID)
}

func (z *zaehlender) ProblemePruefen(ctx context.Context, setID int64) (rest.Problemstand, error) {
	z.zaehle()

	return z.williger.ProblemePruefen(ctx, setID)
}
