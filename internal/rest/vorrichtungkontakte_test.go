package rest_test

// **Die Kontaktwege der drei Doppel** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine eigene Datei für alle drei**, weil `vorrichtungpeer_test.go`,
// `unwilligeausfuhr_test.go` und `vorrichtungzaehler_test.go` mit ihnen über
// die Zeilengrenze aus §2 gingen. Der Schnitt folgt der Frage: hier steht,
// was ein Doppel zu den **Kontakten** sagt — willig, unwillig, zählend.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// eineProbeliste ist die Antwort des willigen Doppels.
//
// **Mit den Zwecken darin**, weil sie in jeder echten Antwort stehen — ein
// Doppel, das sie wegließe, prüfte einen Vertrag, den es nicht gibt.
func eineProbeliste() rest.Peerkontakte {
	return rest.Peerkontakte{
		Kontakte: []rest.Peerkontakt{{
			Kennung:       "bwer",
			Fingerabdruck: "abcd-ef01",
			Stand:         "unbestaetigt",
			Rechte:        []string{},
		}},
		Zwecke: []string{"paarung", "teilen", "handel", "partie"},
	}
}

func (williger) Peerkontakte(context.Context) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (williger) PeerkontaktAufnehmen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (williger) PeerkontaktBestaetigen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (williger) PeerkontaktSperren(
	context.Context, string,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (williger) PeerkontaktEntsperren(
	context.Context, string,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (williger) PeerzweckErlauben(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (williger) PeerzweckEntziehen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (williger) PeerkontaktAdresse(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

func (williger) PeerkontaktNotiz(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return eineProbeliste(), nil
}

// **Und die Nachbarn** (N2, #1455) — eine Liste, die nichts belegt.
func (williger) Nachbarn(context.Context) (rest.Nachbarn, error) {
	return rest.Nachbarn{Nachbarn: []rest.Nachbar{{
		Kennung: "bnachbar", Fingerabdruck: "abcd-ef01",
		Name: "Annas Mac", Adresse: "192.168.1.7:41234",
	}}}, nil
}

// **Und die acht Kontaktwege scheitern wie alles andere.**

func (u unwilliger) Peerkontakte(context.Context) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) PeerkontaktAufnehmen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) PeerkontaktBestaetigen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) PeerkontaktSperren(
	context.Context, string,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) PeerkontaktEntsperren(
	context.Context, string,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) PeerzweckErlauben(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) PeerzweckEntziehen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) PeerkontaktAdresse(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) PeerkontaktNotiz(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	return rest.Peerkontakte{}, u.fehler
}

func (u unwilliger) Nachbarn(context.Context) (rest.Nachbarn, error) {
	return rest.Nachbarn{}, u.fehler
}

// **Und sie zählen mit.**

func (z *zaehlender) Peerkontakte(context.Context) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) PeerkontaktAufnehmen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) PeerkontaktBestaetigen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) PeerkontaktSperren(
	context.Context, string,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) PeerkontaktEntsperren(
	context.Context, string,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) PeerzweckErlauben(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) PeerzweckEntziehen(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) PeerkontaktAdresse(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) PeerkontaktNotiz(
	context.Context, string, string,
) (rest.Peerkontakte, error) {
	z.zaehle()

	return rest.Peerkontakte{}, nil
}

func (z *zaehlender) Nachbarn(context.Context) (rest.Nachbarn, error) {
	z.zaehle()

	return rest.Nachbarn{}, nil
}

func (williger) Sammlungsfreigaben(
	context.Context, string,
) (rest.Sammlungsfreigaben, error) {
	return rest.Sammlungsfreigaben{}, nil
}

func (williger) SammlungFreigeben(
	context.Context, rest.Sammlungsfreigabe,
) (rest.Sammlungsfreigaben, error) {
	return rest.Sammlungsfreigaben{}, nil
}

func (williger) SammlungsfreigabeWiderrufen(
	context.Context, string, string, string,
) (rest.Sammlungsfreigaben, error) {
	return rest.Sammlungsfreigaben{}, nil
}

func (williger) Freigabesicht(
	context.Context, string, string,
) ([]rest.Sichtkarte, error) {
	return []rest.Sichtkarte{}, nil
}

func (u unwilliger) Sammlungsfreigaben(
	context.Context, string,
) (rest.Sammlungsfreigaben, error) {
	return rest.Sammlungsfreigaben{}, u.fehler
}

func (u unwilliger) SammlungFreigeben(
	context.Context, rest.Sammlungsfreigabe,
) (rest.Sammlungsfreigaben, error) {
	return rest.Sammlungsfreigaben{}, u.fehler
}

func (u unwilliger) SammlungsfreigabeWiderrufen(
	context.Context, string, string, string,
) (rest.Sammlungsfreigaben, error) {
	return rest.Sammlungsfreigaben{}, u.fehler
}

func (u unwilliger) Freigabesicht(
	context.Context, string, string,
) ([]rest.Sichtkarte, error) {
	return []rest.Sichtkarte{}, u.fehler
}

func (z *zaehlender) Sammlungsfreigaben(
	context.Context, string,
) (rest.Sammlungsfreigaben, error) {
	z.zaehle()

	return rest.Sammlungsfreigaben{}, nil
}

func (z *zaehlender) SammlungFreigeben(
	context.Context, rest.Sammlungsfreigabe,
) (rest.Sammlungsfreigaben, error) {
	z.zaehle()

	return rest.Sammlungsfreigaben{}, nil
}

func (z *zaehlender) SammlungsfreigabeWiderrufen(
	context.Context, string, string, string,
) (rest.Sammlungsfreigaben, error) {
	z.zaehle()

	return rest.Sammlungsfreigaben{}, nil
}

func (z *zaehlender) Freigabesicht(
	context.Context, string, string,
) ([]rest.Sichtkarte, error) {
	z.zaehle()

	return []rest.Sichtkarte{}, nil
}
