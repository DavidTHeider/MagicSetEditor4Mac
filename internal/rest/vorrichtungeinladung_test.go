package rest_test

// **Die Einladungswege der drei Doppel** — NET-01
// (#1705).
//
// **Eine eigene Datei**, weil `vorrichtungkontakte_test.go` mit ihnen über
// die Zeilengrenze aus §2 ginge — der Schnitt folgt derselben Frage wie dort.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) KontakteinladungErstellen(
	context.Context, string, string, string,
) (rest.Kontakteinladung, error) {
	return rest.Kontakteinladung{Code: "mse4mac-kontakt:1:probe.probe"}, nil
}

func (williger) KontakteinladungPruefen(
	context.Context, string,
) (rest.Einladungsvorschau, error) {
	return rest.Einladungsvorschau{Name: "Anna", Lage: "unbekannt"}, nil
}

func (williger) KontakteinladungAnnehmen(
	context.Context, rest.Einladungsannahme,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (u unwilliger) KontakteinladungErstellen(
	context.Context, string, string, string,
) (rest.Kontakteinladung, error) {
	return rest.Kontakteinladung{}, u.fehler
}

func (u unwilliger) KontakteinladungPruefen(
	context.Context, string,
) (rest.Einladungsvorschau, error) {
	return rest.Einladungsvorschau{}, u.fehler
}

func (u unwilliger) KontakteinladungAnnehmen(
	context.Context, rest.Einladungsannahme,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (z *zaehlender) KontakteinladungErstellen(
	context.Context, string, string, string,
) (rest.Kontakteinladung, error) {
	z.zaehle()

	return rest.Kontakteinladung{}, nil
}

func (z *zaehlender) KontakteinladungPruefen(
	context.Context, string,
) (rest.Einladungsvorschau, error) {
	z.zaehle()

	return rest.Einladungsvorschau{}, nil
}

func (z *zaehlender) KontakteinladungAnnehmen(
	context.Context, rest.Einladungsannahme,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}
