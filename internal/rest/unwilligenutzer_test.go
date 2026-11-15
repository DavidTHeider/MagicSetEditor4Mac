// **Die unwilligen Antworten zu Nutzern, Profilen und Konten** — Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, seit `unwilligeausfuhr_test.go` mit den drei
// Kontoarten aus Werkstück 6 über die Zeilengrenze aus §2 lief. Der Schnitt
// folgt der Frage: dort scheitert, was die Platte verlässt oder auf sie kommt,
// hier, wer den Bestand benutzt.
package rest_test

import (
	"context"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (u unwilliger) Nutzer(context.Context) ([]rest.Nutzer, error) {
	return nil, u.fehler
}

func (u unwilliger) NutzerAnlegen(context.Context, string, string) (rest.Nutzer, error) {
	return rest.Nutzer{}, u.fehler
}

func (u unwilliger) ProfilAnlegen(context.Context, string, string) (rest.Profil, error) {
	return rest.Profil{}, u.fehler
}

func (u unwilliger) GezeigtesProfil(context.Context) (rest.Profil, error) {
	return rest.Profil{}, u.fehler
}

func (u unwilliger) ProfilWechseln(context.Context, string, string) (rest.Profil, error) {
	return rest.Profil{}, u.fehler
}

func (u unwilliger) Anbieterkonten(context.Context, string) ([]rest.Konto, error) {
	return nil, u.fehler
}

func (u unwilliger) KontoLoesen(context.Context, string, string) error { return u.fehler }

func (u unwilliger) NutzerAendern(context.Context, rest.Nutzerangaben) (
	rest.Nutzer, error,
) {
	return rest.Nutzer{}, u.fehler
}

func (u unwilliger) ProfilAendern(context.Context, rest.Profilangaben) (
	rest.Profil, error,
) {
	return rest.Profil{}, u.fehler
}

func (u unwilliger) NutzerLoeschen(context.Context, string, string) error { return u.fehler }
func (u unwilliger) ProfilLoeschen(context.Context, string, string) error { return u.fehler }

func (u unwilliger) KontoEintragen(context.Context, string, string, string) (
	rest.Nutzer, error,
) {
	return rest.Nutzer{}, u.fehler
}

func (u unwilliger) KontoAnmelden(context.Context, string, string, string) (rest.Nutzer, error) {
	return rest.Nutzer{}, u.fehler
}

func (u unwilliger) KontoAnonym(context.Context) (rest.Nutzer, error) {
	return rest.Nutzer{}, u.fehler
}

// **Der zweite Faktor** — Werkstück 7 (#1351).
func (u unwilliger) Faktoren(context.Context, string) (rest.Faktorstand, error) {
	return rest.Faktorstand{}, u.fehler
}

func (u unwilliger) FaktorEinrichten(context.Context, string) (
	rest.Faktoreinrichtung, error,
) {
	return rest.Faktoreinrichtung{}, u.fehler
}

func (u unwilliger) FaktorBestaetigen(context.Context, string, string) error {
	return u.fehler
}

func (u unwilliger) BiometrieEinrichten(context.Context, string) error { return u.fehler }

func (u unwilliger) FaktorEntfernen(context.Context, string, string) error {
	return u.fehler
}

// **Das Profilbild** — Werkstück 8 von Phase 4.12 (#1351).
func (u unwilliger) NutzerbildSetzen(context.Context, string, string, string) (
	rest.Nutzer, error,
) {
	return rest.Nutzer{}, u.fehler
}

func (u unwilliger) NutzerbildEntfernen(context.Context, string, string) (
	rest.Nutzer, error,
) {
	return rest.Nutzer{}, u.fehler
}

// **Die Einstellungen und die Auskunft** — Werkstück 8 desselben Issues.
func (u unwilliger) Nutzereinstellungen(context.Context, string) (
	[]rest.Einstellungswahl, error,
) {
	return nil, u.fehler
}

func (u unwilliger) NutzereinstellungSetzen(context.Context, string, string, string) (
	[]rest.Einstellungswahl, error,
) {
	return nil, u.fehler
}

func (u unwilliger) Personendaten(context.Context, string) (json.RawMessage, error) {
	return nil, u.fehler
}

// **Das Abzeichen eines Profils** — Werkstück 9 desselben Issues.
func (u unwilliger) ProfilbadgeSetzen(context.Context, string, string, string) (
	rest.Profil, error,
) {
	return rest.Profil{}, u.fehler
}

func (u unwilliger) ProfilbadgeEntfernen(context.Context, string, string) (
	rest.Profil, error,
) {
	return rest.Profil{}, u.fehler
}

func (u unwilliger) WerArbeitet(context.Context) (string, error) { return "", u.fehler }

func (u unwilliger) Arbeitskontext(context.Context) (rest.Arbeitskontext, error) {
	return rest.Arbeitskontext{}, u.fehler
}

func (u unwilliger) NutzerWechseln(context.Context, string, string) (rest.Profil, error) {
	return rest.Profil{}, u.fehler
}
