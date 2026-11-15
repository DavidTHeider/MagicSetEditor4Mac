// Die Wortlisten der Vorrichtung
// (#1074).
//
// **Eine eigene Datei**, weil `vorrichtung_test.go` mit ihnen über die
// Zeilengrenze ginge (§2).
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Die Wortlisten der Vorrichtung** —
// #1074.
func (williger) Wortlisten(context.Context, int64, string) ([]rest.Wortbereich, error) {
	return []rest.Wortbereich{{
		Liste: "type-en", Text: "Creature", Gewaehlt: 0,
		Eintraege: []rest.Worteintrag{{Name: "Creature", Aktiv: true}},
	}}, nil
}

func (williger) Wortwahl(context.Context, int64, int64, string, int, []int) error {
	return nil
}

func (z *zaehlender) Wortlisten(context.Context, int64, string) ([]rest.Wortbereich, error) {
	z.rufe++

	return nil, nil
}

func (z *zaehlender) Wortwahl(context.Context, int64, int64, string, int, []int) error {
	z.rufe++

	return nil
}

// **Die Schlüsselwörter der Vorrichtung** —
// #1062.
func (williger) Schluesselwoerter(context.Context, int64) (
	rest.Schluesselwortliste, error,
) {
	return rest.Schluesselwortliste{
		Woerter: []rest.Schluesselwort{{ID: 2, Wort: "Beben", Gueltig: true}},
		Modi:    []string{"custom"},
	}, nil
}

func (williger) SchluesselwortAnlegen(context.Context, int64) (int64, error) {
	return 9, nil
}

func (williger) SchluesselwortSetzen(
	context.Context, int64, int64, string, string,
) error {
	return nil
}

func (williger) SchluesselwortLoeschen(context.Context, int64, int64) error {
	return nil
}

func (z *zaehlender) Schluesselwoerter(context.Context, int64) (
	rest.Schluesselwortliste, error,
) {
	z.rufe++

	return rest.Schluesselwortliste{}, nil
}

func (z *zaehlender) SchluesselwortAnlegen(context.Context, int64) (int64, error) {
	z.rufe++

	return 9, nil
}

func (z *zaehlender) SchluesselwortSetzen(
	context.Context, int64, int64, string, string,
) error {
	z.rufe++

	return nil
}

func (z *zaehlender) SchluesselwortLoeschen(context.Context, int64, int64) error {
	z.rufe++

	return nil
}

// **Und das Anlegen eines Sets** (#891).
func (z *zaehlender) SetAnlegen(context.Context, string, string, string) (int64, error) {
	z.rufe++

	return 1, nil
}

// **Die Packwege zählen mit** (#889, Befund 7 des Zweitblicks über Phase W).
func (z *zaehlender) Packarten(context.Context, int64) (rest.Packangebot, error) {
	z.rufe++

	return rest.Packangebot{}, nil
}

func (z *zaehlender) PackZiehen(context.Context, int64, string, string) (
	rest.Packziehung, error,
) {
	z.rufe++

	return rest.Packziehung{}, nil
}
