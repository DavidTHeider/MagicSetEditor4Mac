package rest_test

// **Die Doppel der Freigabennaht** — Werkstück 11 von Phase 4.12
// (#1351).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// probefreigaben ist der Stand, den das willige Doppel gibt.
func probefreigaben() rest.Freigabenstand {
	return rest.Freigabenstand{
		Freigaben: []rest.Freigabezeile{{
			NutzerID: "2", Name: "Zweiter", Rolle: "leser",
			Angelegt: "2026-09-06T00:00:00Z",
		}},
		Sichtbarkeit: "privat",
		EigeneRolle:  "verwalter",
	}
}

func (williger) Freigaben(context.Context, string) (rest.Freigabenstand, error) {
	return probefreigaben(), nil
}

func (williger) FreigabeSetzen(_ context.Context, _, nutzerID, rolle string) (
	rest.Freigabenstand, error,
) {
	stand := probefreigaben()
	stand.Freigaben[0].NutzerID = nutzerID
	stand.Freigaben[0].Rolle = rolle

	return stand, nil
}

func (williger) FreigabeLoeschen(context.Context, string, string) (
	rest.Freigabenstand, error,
) {
	stand := probefreigaben()
	stand.Freigaben = []rest.Freigabezeile{}

	return stand, nil
}

func (u unwilliger) Freigaben(context.Context, string) (
	rest.Freigabenstand, error,
) {
	return rest.Freigabenstand{}, u.fehler
}

func (u unwilliger) FreigabeSetzen(context.Context, string, string, string) (
	rest.Freigabenstand, error,
) {
	return rest.Freigabenstand{}, u.fehler
}

func (u unwilliger) FreigabeLoeschen(context.Context, string, string) (
	rest.Freigabenstand, error,
) {
	return rest.Freigabenstand{}, u.fehler
}

// zaehlender zählt die drei mit — sonst belegte „gültig" nur einen Status.

func (z *zaehlender) Freigaben(ctx context.Context, profilID string) (
	rest.Freigabenstand, error,
) {
	z.zaehle()

	return z.williger.Freigaben(ctx, profilID)
}

func (z *zaehlender) FreigabeSetzen(
	ctx context.Context, profilID, nutzerID, rolle string,
) (rest.Freigabenstand, error) {
	z.zaehle()

	return z.williger.FreigabeSetzen(ctx, profilID, nutzerID, rolle)
}

func (z *zaehlender) FreigabeLoeschen(
	ctx context.Context, profilID, nutzerID string,
) (rest.Freigabenstand, error) {
	z.zaehle()

	return z.williger.FreigabeLoeschen(ctx, profilID, nutzerID)
}
