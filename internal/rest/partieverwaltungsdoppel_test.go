package rest_test

// **Die vier Verwaltungswege an den drei Werkstattdoppeln** —
// PLAY-03.AK-05
// (#1711).
//
// **Eine eigene Datei**, damit die drei Vorrichtungen unter der
// Zeilengrenze aus §2 bleiben. Die Typen bleiben drei — Methoden dürfen
// daneben liegen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) PartieLoeschen(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	return rest.Partieverwaltungsangaben{Befunde: []string{}}, nil
}

func (williger) PartieKopieren(_ context.Context, kennung string) (
	rest.Partieverwaltungsangaben, error,
) {
	return rest.Partieverwaltungsangaben{
		Kennung: kennung + "-kopie", Befunde: []string{},
	}, nil
}

func (williger) PartieAusfuhr(_ context.Context, kennung string) (
	rest.Partieverwaltungsangaben, error,
) {
	return rest.Partieverwaltungsangaben{
		Kennung: kennung, Inhalt: `{"fassung":1}`, Befunde: []string{},
	}, nil
}

func (williger) PartieEinfuhr(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	return rest.Partieverwaltungsangaben{
		Kennung: "eingelesen", Befunde: []string{"ein fremdes Feld"},
	}, nil
}

func (u unwilliger) PartieLoeschen(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	return rest.Partieverwaltungsangaben{}, u.fehler
}

func (u unwilliger) PartieKopieren(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	return rest.Partieverwaltungsangaben{}, u.fehler
}

func (u unwilliger) PartieAusfuhr(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	return rest.Partieverwaltungsangaben{}, u.fehler
}

func (u unwilliger) PartieEinfuhr(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	return rest.Partieverwaltungsangaben{}, u.fehler
}

func (z *zaehlender) PartieLoeschen(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	z.rufe++

	return rest.Partieverwaltungsangaben{Befunde: []string{}}, nil
}

func (z *zaehlender) PartieKopieren(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	z.rufe++

	return rest.Partieverwaltungsangaben{Befunde: []string{}}, nil
}

func (z *zaehlender) PartieAusfuhr(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	z.rufe++

	return rest.Partieverwaltungsangaben{Befunde: []string{}}, nil
}

func (z *zaehlender) PartieEinfuhr(context.Context, string) (
	rest.Partieverwaltungsangaben, error,
) {
	z.rufe++

	return rest.Partieverwaltungsangaben{Befunde: []string{}}, nil
}
