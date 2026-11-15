// **Der portable Umschlag am Dienst** —
// #1457.
//
// **Dieselbe kopflose Anwendung wie an der Kommandozeile und am MCP-Server**
// ([app.UmschlagflaecheUeber]) und nicht ein dritter Weg daneben: was
// geschnürt wird, soll überall dasselbe sein (§16).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// umschlag erfüllt `rest.Umschlagwerkstatt` über der Naht von `app`.
type umschlag struct{ db *sql.DB }

func (u umschlag) UmschlagSchnueren(ctx context.Context, art, kennung, ziel string) (
	rest.Umschlagangaben, error,
) {
	return u.fahre(func(anwendung *app.App) app.UmschlagAntwort {
		return anwendung.UmschlagSchnueren(ctx, art, kennung, ziel)
	})
}

func (u umschlag) UmschlagAnsehen(ctx context.Context, pfad string) (
	rest.Umschlagangaben, error,
) {
	return u.fahre(func(anwendung *app.App) app.UmschlagAntwort {
		return anwendung.UmschlagAnsehen(ctx, pfad)
	})
}

func (u umschlag) UmschlagUebernehmen(ctx context.Context, pfad string) (
	rest.Umschlagannahme, error,
) {
	anwendung, fehler := app.UmschlagflaecheUeber(u.db)
	if fehler != nil {
		return rest.Umschlagannahme{}, fehler
	}

	antwort := anwendung.UmschlagUebernehmen(ctx, pfad)
	if antwort.Fehler != nil {
		return rest.Umschlagannahme{}, antwort.Fehler.AlsFehler()
	}

	return rest.Umschlagannahme{
		SetID:   antwort.SetID,
		Karten:  antwort.Karten,
		Befunde: antwort.Befunde,
	}, nil
}

// fahre baut die Anwendung, ruft und übersetzt — für die zwei Wege, die
// dieselbe Antwort geben.
func (u umschlag) fahre(rufe func(*app.App) app.UmschlagAntwort) (
	rest.Umschlagangaben, error,
) {
	anwendung, fehler := app.UmschlagflaecheUeber(u.db)
	if fehler != nil {
		return rest.Umschlagangaben{}, fehler
	}

	antwort := rufe(anwendung)
	if antwort.Fehler != nil {
		return rest.Umschlagangaben{}, antwort.Fehler.AlsFehler()
	}

	angaben := rest.Umschlagangaben{
		Urheber:       antwort.Urheber,
		Fingerabdruck: antwort.Fingerabdruck,
		Objekte:       make([]rest.Umschlagobjekt, 0, len(antwort.Objekte)),
	}

	for _, objekt := range antwort.Objekte {
		angaben.Objekte = append(angaben.Objekte, rest.Umschlagobjekt{
			Art:         objekt.Art,
			Bezeichnung: objekt.Bezeichnung,
			Groesse:     objekt.Groesse,
			Braucht:     objekt.Braucht,
		})
	}

	return angaben, nil
}
