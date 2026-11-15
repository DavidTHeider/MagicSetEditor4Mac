package rest_test

// **Die Doppel des Poststands** — Werkstück `T5` von Phase 5.6
// (#1470).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Handelspoststand(
	context.Context, string,
) (rest.Poststand, error) {
	return rest.Poststand{}, nil
}

func (u unwilliger) Handelspoststand(
	context.Context, string,
) (rest.Poststand, error) {
	return rest.Poststand{}, u.fehler
}

func (z *zaehlender) Handelspoststand(
	context.Context, string,
) (rest.Poststand, error) {
	z.zaehle()

	return rest.Poststand{}, nil
}
