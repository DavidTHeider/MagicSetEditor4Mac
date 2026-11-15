package rest_test

// **Die Doppel der Verbuchung** — Werkstück `T6` von Phase 5.6
// (#1471).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Handelsschritt(
	context.Context, string, rest.Schrittwunsch,
) (rest.Handelsstand, error) {
	return rest.Handelsstand{}, nil
}

func (williger) Handelsjournal(
	context.Context, string,
) ([]rest.Journalzeile, error) {
	return []rest.Journalzeile{}, nil
}

func (u unwilliger) Handelsschritt(
	context.Context, string, rest.Schrittwunsch,
) (rest.Handelsstand, error) {
	return rest.Handelsstand{}, u.fehler
}

func (u unwilliger) Handelsjournal(
	context.Context, string,
) ([]rest.Journalzeile, error) {
	return nil, u.fehler
}

func (z *zaehlender) Handelsschritt(
	context.Context, string, rest.Schrittwunsch,
) (rest.Handelsstand, error) {
	z.zaehle()

	return rest.Handelsstand{}, nil
}

func (z *zaehlender) Handelsjournal(
	context.Context, string,
) ([]rest.Journalzeile, error) {
	z.zaehle()

	return []rest.Journalzeile{}, nil
}

func (williger) Handelsbeleg(context.Context, string) (rest.Handelsstand, error) {
	return rest.Handelsstand{}, nil
}

func (u unwilliger) Handelsbeleg(context.Context, string) (rest.Handelsstand, error) {
	return rest.Handelsstand{}, u.fehler
}

func (z *zaehlender) Handelsbeleg(context.Context, string) (rest.Handelsstand, error) {
	z.zaehle()

	return rest.Handelsstand{}, nil
}
