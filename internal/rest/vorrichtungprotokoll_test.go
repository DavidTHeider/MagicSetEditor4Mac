package rest_test

// **Das Protokoll der drei Doppel** — TRADE-03.AK-03
// (#1717).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) HandelProtokoll(context.Context, string) ([]rest.Handelsprotokollzeile, error) {
	return []rest.Handelsprotokollzeile{{Revision: 1, Nach: "entwurf"}}, nil
}

func (u unwilliger) HandelProtokoll(context.Context, string) ([]rest.Handelsprotokollzeile, error) {
	return nil, u.fehler
}

func (z *zaehlender) HandelProtokoll(context.Context, string) ([]rest.Handelsprotokollzeile, error) {
	z.zaehle()

	return nil, nil
}
