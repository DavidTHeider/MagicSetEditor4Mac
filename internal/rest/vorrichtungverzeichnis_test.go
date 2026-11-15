package rest_test

// **Die Doppel für Liste und Entwürfe des Handels** — TRADE-01
// (#1715).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) HandelUebersicht(context.Context, rest.Handelsfilter) ([]rest.Handelsuebersichtszeile, error) {
	return nil, nil
}

func (williger) HandelEntwurfAendern(context.Context, rest.Handelswunsch) (rest.Handelsvorgang, error) {
	return einProbevorgang(), nil
}

func (williger) HandelKopieren(context.Context, string) (rest.Handelsvorgang, error) {
	return einProbevorgang(), nil
}

func (williger) HandelEntwurfLoeschen(context.Context, string) error { return nil }

func (williger) HandelArchivieren(context.Context, string, bool) error { return nil }

func (u unwilliger) HandelUebersicht(context.Context, rest.Handelsfilter) ([]rest.Handelsuebersichtszeile, error) {
	return nil, u.fehler
}

func (u unwilliger) HandelEntwurfAendern(context.Context, rest.Handelswunsch) (rest.Handelsvorgang, error) {
	return rest.Handelsvorgang{}, u.fehler
}

func (u unwilliger) HandelKopieren(context.Context, string) (rest.Handelsvorgang, error) {
	return rest.Handelsvorgang{}, u.fehler
}

func (u unwilliger) HandelEntwurfLoeschen(context.Context, string) error { return u.fehler }

func (u unwilliger) HandelArchivieren(context.Context, string, bool) error { return u.fehler }

func (z *zaehlender) HandelUebersicht(context.Context, rest.Handelsfilter) ([]rest.Handelsuebersichtszeile, error) {
	z.zaehle()

	return nil, nil
}

func (z *zaehlender) HandelEntwurfAendern(context.Context, rest.Handelswunsch) (rest.Handelsvorgang, error) {
	z.zaehle()

	return einProbevorgang(), nil
}

func (z *zaehlender) HandelKopieren(context.Context, string) (rest.Handelsvorgang, error) {
	z.zaehle()

	return einProbevorgang(), nil
}

func (z *zaehlender) HandelEntwurfLoeschen(context.Context, string) error {
	z.zaehle()

	return nil
}

func (z *zaehlender) HandelArchivieren(context.Context, string, bool) error {
	z.zaehle()

	return nil
}
