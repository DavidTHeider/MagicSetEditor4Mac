package rest_test

// **Die Doppel der Handelsleitung** —
// #1637.
//
// **Sie stehen hier und nicht bei der Verbuchung**: dort geht es um Bestand
// und Beleg, hier um den Vertrag davor — der Schnitt folgt der Frage (§2).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// einProbevorgang ist der Stand, den die willigen Doppel zurückgeben.
func einProbevorgang() rest.Handelsvorgang {
	return rest.Handelsvorgang{
		Vorgang: "v-1", Revision: 1, Abdruck: "ab", Art: "tausch",
		Initiator: "bich", Parteien: []string{"bich", "bgegenueber"},
		Zustand:      "entwurf",
		Positionen:   []rest.Handelsposition{},
		Zustimmungen: []string{},
	}
}

func (williger) HandelEroeffnen(
	context.Context, rest.Handelswunsch,
) (rest.Handelsvorgang, error) {
	return einProbevorgang(), nil
}

func (williger) HandelVorschlagen(
	context.Context, string, []rest.Handelsposition,
) (rest.Handelsvorgang, error) {
	return einProbevorgang(), nil
}

func (williger) HandelStand(
	context.Context, string,
) (rest.Handelsvorgang, error) {
	return einProbevorgang(), nil
}

func (williger) HandelVorgaenge(
	context.Context, string,
) ([]rest.Handelsvorgang, error) {
	return []rest.Handelsvorgang{einProbevorgang()}, nil
}

func (williger) HandelSchicken(context.Context, string, string) (
	string, error,
) {
	return "n-1", nil
}

func (williger) HandelFesthalten(
	context.Context, string, int64, int, int64,
) (rest.Festgehalten, error) {
	return rest.Festgehalten{Zusagen: []rest.Handelszusage{{ExemplarID: 7, Menge: 1}}}, nil
}

func (williger) Handelszusagen(
	context.Context, string,
) ([]rest.Handelszusage, error) {
	return []rest.Handelszusage{}, nil
}

func (williger) HandelPostbestaetigung(context.Context, string) error {
	return nil
}

func (williger) HandelWiederaufnahme(
	context.Context, string, string, rest.Handelsfortschritt,
) (rest.Wiederaufnahme, error) {
	return rest.Wiederaufnahme{Nachzuliefern: []rest.Nachlieferung{}}, nil
}

func (u unwilliger) HandelEroeffnen(
	context.Context, rest.Handelswunsch,
) (rest.Handelsvorgang, error) {
	return rest.Handelsvorgang{}, u.fehler
}

func (u unwilliger) HandelVorschlagen(
	context.Context, string, []rest.Handelsposition,
) (rest.Handelsvorgang, error) {
	return rest.Handelsvorgang{}, u.fehler
}

func (u unwilliger) HandelStand(
	context.Context, string,
) (rest.Handelsvorgang, error) {
	return rest.Handelsvorgang{}, u.fehler
}

func (u unwilliger) HandelVorgaenge(
	context.Context, string,
) ([]rest.Handelsvorgang, error) {
	return nil, u.fehler
}

func (u unwilliger) HandelSchicken(context.Context, string, string) (
	string, error,
) {
	return "", u.fehler
}

func (u unwilliger) HandelFesthalten(
	context.Context, string, int64, int, int64,
) (rest.Festgehalten, error) {
	return rest.Festgehalten{}, u.fehler
}

func (u unwilliger) Handelszusagen(
	context.Context, string,
) ([]rest.Handelszusage, error) {
	return nil, u.fehler
}

func (u unwilliger) HandelPostbestaetigung(context.Context, string) error {
	return u.fehler
}

func (u unwilliger) HandelWiederaufnahme(
	context.Context, string, string, rest.Handelsfortschritt,
) (rest.Wiederaufnahme, error) {
	return rest.Wiederaufnahme{}, u.fehler
}

func (z *zaehlender) HandelEroeffnen(
	context.Context, rest.Handelswunsch,
) (rest.Handelsvorgang, error) {
	z.zaehle()

	return einProbevorgang(), nil
}

func (z *zaehlender) HandelVorschlagen(
	context.Context, string, []rest.Handelsposition,
) (rest.Handelsvorgang, error) {
	z.zaehle()

	return einProbevorgang(), nil
}

func (z *zaehlender) HandelStand(
	context.Context, string,
) (rest.Handelsvorgang, error) {
	z.zaehle()

	return einProbevorgang(), nil
}

func (z *zaehlender) HandelVorgaenge(
	context.Context, string,
) ([]rest.Handelsvorgang, error) {
	z.zaehle()

	return []rest.Handelsvorgang{einProbevorgang()}, nil
}

func (z *zaehlender) HandelSchicken(context.Context, string, string) (
	string, error,
) {
	z.zaehle()

	return "n-1", nil
}

func (z *zaehlender) HandelFesthalten(
	context.Context, string, int64, int, int64,
) (rest.Festgehalten, error) {
	z.zaehle()

	return rest.Festgehalten{Zusagen: []rest.Handelszusage{}}, nil
}

func (z *zaehlender) Handelszusagen(
	context.Context, string,
) ([]rest.Handelszusage, error) {
	z.zaehle()

	return []rest.Handelszusage{}, nil
}

func (z *zaehlender) HandelPostbestaetigung(context.Context, string) error {
	z.zaehle()

	return nil
}

func (z *zaehlender) HandelWiederaufnahme(
	context.Context, string, string, rest.Handelsfortschritt,
) (rest.Wiederaufnahme, error) {
	z.zaehle()

	return rest.Wiederaufnahme{Nachzuliefern: []rest.Nachlieferung{}}, nil
}
