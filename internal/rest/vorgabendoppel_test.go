package rest_test

// **Die Doppel der Vorgabennaht** — Werkstück 15 von Phase 4.12
// (#1351).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// probevorgaben ist der Stand, den das willige Doppel gibt.
func probevorgaben() []rest.Vorgabenzeile {
	return []rest.Vorgabenzeile{{
		Art: "deck", Mehrwertig: true, Werte: []string{"7"},
		Moeglich: []rest.Vorgabeangebot{{Wert: "7", Beschriftung: "Meins"}},
		Verwaist: []string{},
	}}
}

func (williger) Vorgaben(context.Context, string) ([]rest.Vorgabenzeile, error) {
	return probevorgaben(), nil
}

func (williger) VorgabeSetzen(_ context.Context, _, art string, werte []string) (
	rest.Vorgabengesetzt, error,
) {
	zeilen := probevorgaben()
	zeilen[0].Art = art
	zeilen[0].Werte = werte

	return rest.Vorgabengesetzt{Vorgaben: zeilen}, nil
}

func (u unwilliger) Vorgaben(context.Context, string) (
	[]rest.Vorgabenzeile, error,
) {
	return nil, u.fehler
}

func (u unwilliger) VorgabeSetzen(context.Context, string, string, []string) (
	rest.Vorgabengesetzt, error,
) {
	return rest.Vorgabengesetzt{}, u.fehler
}

// zaehlender zählt beide mit — sonst belegte „gültig" nur einen Status.

func (z *zaehlender) Vorgaben(ctx context.Context, profilID string) (
	[]rest.Vorgabenzeile, error,
) {
	z.zaehle()

	return z.williger.Vorgaben(ctx, profilID)
}

func (z *zaehlender) VorgabeSetzen(
	ctx context.Context, profilID, art string, werte []string,
) (rest.Vorgabengesetzt, error) {
	z.zaehle()

	// **Die Werte sind hier der Prüfgegenstand**: ein `PUT` ohne `werte` soll
	// als **leere Liste** ankommen und nicht als `nil` (Werkstück 15, #1351).
	z.werte = werte

	return z.williger.VorgabeSetzen(ctx, profilID, art, werte)
}
