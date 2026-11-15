// **Die unwilligen Antworten der Sammlung und der Werkzeuge** — Ausgeben,
// Einlesen, Besitz, Lagerorte, Dubletten, Wünsche, die Massenvorgänge, die
// Kartenskripte und die drei Auskünfte über die Anwendung.
//
// **Eine eigene Datei**, seit `unwilligeausfuhr_test.go` mit dem Marktplatz
// über die Zeilengrenze aus §2 lief (A4, #1443). Der Schnitt folgt der Frage:
// dort scheitert, **was die Platte verlässt oder auf sie kommt**, hier, **was
// die Sammlung und die Werkzeuge tun**.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (u unwilliger) KIModelle(context.Context) (rest.KIKatalogsicht, error) {
	return rest.KIKatalogsicht{}, u.fehler
}

func (u unwilliger) KIVorschlag(context.Context, rest.KIVorschlagsbezug) (
	rest.KIVorschlagssicht, error,
) {
	return rest.KIVorschlagssicht{}, u.fehler
}

func (u unwilliger) Anbieterstand(context.Context, string) (rest.Anbietersicht, error) {
	return rest.Anbietersicht{}, u.fehler
}

func (u unwilliger) Fremdpreisverlauf(context.Context, rest.Verlaufsbezug) (
	rest.Verlaufssicht, error,
) {
	return rest.Verlaufssicht{}, u.fehler
}

func (u unwilliger) SammlungAusgeben(context.Context, string) (string, error) {
	return "", u.fehler
}

func (u unwilliger) SammlungCSVEinlesen(context.Context, string, string, string) (any, error) {
	return nil, u.fehler
}

// **Die Sammlungssichten** (#1292) — alle vier scheitern.
func (u unwilliger) SammlungBesitz(context.Context, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) SammlungLagerorte(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) SammlungDubletten(context.Context, int) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) SammlungWuensche(context.Context) (any, error) {
	return nil, u.fehler
}

// **Die Massenvorgänge** (#1292) — alle drei scheitern.
func (u unwilliger) MasseAendern(context.Context, rest.Massenaenderung) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) MasseLoeschen(context.Context, []rest.Massenzeile) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) MassePreise(context.Context, int) (any, error) {
	return nil, u.fehler
}

// **Die Kartenskripte** (#1337).
func (u unwilliger) Kartenskripte(context.Context, int64) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) KartenskriptAusfuehren(context.Context, int64, string) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Metriken(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Meta(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) Vertraeglichkeit(context.Context) (any, error) {
	return nil, u.fehler
}

func (u unwilliger) KIModellPruefen(context.Context, rest.KIProbenbezug) (
	rest.KIProbensicht, error,
) {
	return rest.KIProbensicht{}, u.fehler
}

func (u unwilliger) KIModellVergessen(context.Context, rest.KIProbenbezug) error {
	return u.fehler
}

func (u unwilliger) KIGeraetemodellFuerSitzung(context.Context, bool) (bool, error) {
	return true, u.fehler
}

// KIBatch stolpert — A13.1 (#1452).
func (u unwilliger) KIBatch(context.Context, rest.KIBatchbezug) (
	rest.KIBatchsicht, error,
) {
	return rest.KIBatchsicht{}, u.fehler
}

// KIEinschaetzung stolpert — A13.2a (#1452).
func (u unwilliger) KIEinschaetzung(
	context.Context, rest.KIEinschaetzungsbezug,
) (any, error) {
	return nil, u.fehler
}
