package rest_test

// **Die Doppel der Platznaht** — Werkstück 4 von Phase 4.13
// (#1352).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// probeplatz ist der Stand, den das willige Doppel gibt.
func probeplatz() rest.Platzstand {
	return rest.Platzstand{Posten: []rest.Platzzeile{
		{Posten: "bestand", Dateien: 1, Bytes: 4096, Raeumbar: false},
		{Posten: "vorsicherungen", Dateien: 3, Bytes: 12288, Raeumbar: true},
	}}
}

func (williger) Platz(context.Context) (rest.Platzstand, error) {
	return probeplatz(), nil
}

func (williger) PlatzFreigeben(_ context.Context, posten string) (
	rest.Platzstand, error,
) {
	stand := probeplatz()

	// **Nach dem Wegräumen bleibt eine** — die jüngste.
	for stelle := range stand.Posten {
		if stand.Posten[stelle].Posten == posten {
			stand.Posten[stelle].Dateien = 1
			stand.Posten[stelle].Bytes = 4096
		}
	}

	return stand, nil
}

func (u unwilliger) Platz(context.Context) (rest.Platzstand, error) {
	return rest.Platzstand{}, u.fehler
}

func (u unwilliger) PlatzFreigeben(context.Context, string) (
	rest.Platzstand, error,
) {
	return rest.Platzstand{}, u.fehler
}

func (z *zaehlender) Platz(ctx context.Context) (rest.Platzstand, error) {
	z.zaehle()

	return z.williger.Platz(ctx)
}

func (z *zaehlender) PlatzFreigeben(ctx context.Context, posten string) (
	rest.Platzstand, error,
) {
	z.zaehle()

	return z.williger.PlatzFreigeben(ctx, posten)
}
