// **Einladen und annehmen im Dienst** — NET-01
// (#1705).

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// KontakteinladungErstellen stellt eine Einladung dieses Profils aus.
func (k kontaktweg) KontakteinladungErstellen(
	ctx context.Context, name, zweck, adresse string,
) (rest.Kontakteinladung, error) {
	antwort := app.PeerflaecheUeber(k.db).KontakteinladungErstellenUnter(ctx, name, zweck, adresse)
	if antwort.Fehler != nil {
		return rest.Kontakteinladung{}, antwort.Fehler.AlsFehler()
	}

	return rest.Kontakteinladung(antwort.Einladung), nil
}

// KontakteinladungPruefen sagt, was eine Einladung ist.
func (k kontaktweg) KontakteinladungPruefen(
	ctx context.Context, code string,
) (rest.Einladungsvorschau, error) {
	antwort := app.PeerflaecheUeber(k.db).KontakteinladungPruefen(ctx, code)
	if antwort.Fehler != nil {
		return rest.Einladungsvorschau{}, antwort.Fehler.AlsFehler()
	}

	return rest.Einladungsvorschau(antwort.Vorschau), nil
}

// KontakteinladungAnnehmen nimmt eine Einladung an.
func (k kontaktweg) KontakteinladungAnnehmen(
	ctx context.Context, annahme rest.Einladungsannahme,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).KontakteinladungAnnehmen(
		ctx, annahme.Code, annahme.Name, annahme.Vergleichscode, annahme.Freigeben,
		annahme.MeinName))
}
