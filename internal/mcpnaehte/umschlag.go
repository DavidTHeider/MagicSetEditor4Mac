// **Der portable Umschlag am MCP-Server** —
// #1457.
//
// **Dieselbe kopflose Anwendung wie an der Kommandozeile**
// ([app.UmschlagflaecheUeber]) und nicht ein zweiter Weg daneben: was
// geschnürt wird, soll überall dasselbe sein (§16).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// umschlag erfüllt `mcp.Umschlagwerkstatt` über der Naht von `app`.
type umschlag struct{ db *sql.DB }

func (u umschlag) UmschlagSchnueren(ctx context.Context, art, kennung, ziel string) (
	mcp.Umschlagangaben, error,
) {
	return u.fahre(func(anwendung *app.App) app.UmschlagAntwort {
		return anwendung.UmschlagSchnueren(ctx, art, kennung, ziel)
	})
}

func (u umschlag) UmschlagAnsehen(ctx context.Context, pfad string) (
	mcp.Umschlagangaben, error,
) {
	return u.fahre(func(anwendung *app.App) app.UmschlagAntwort {
		return anwendung.UmschlagAnsehen(ctx, pfad)
	})
}

func (u umschlag) UmschlagUebernehmen(ctx context.Context, pfad string) (
	mcp.Umschlagannahme, error,
) {
	anwendung, fehler := app.UmschlagflaecheUeber(u.db)
	if fehler != nil {
		return mcp.Umschlagannahme{}, fehler
	}

	antwort := anwendung.UmschlagUebernehmen(ctx, pfad)
	if antwort.Fehler != nil {
		return mcp.Umschlagannahme{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Umschlagannahme{
		SetID:   antwort.SetID,
		Karten:  antwort.Karten,
		Befunde: antwort.Befunde,
	}, nil
}

// fahre baut die Anwendung, ruft und übersetzt — für die zwei Wege, die
// dieselbe Antwort geben.
func (u umschlag) fahre(rufe func(*app.App) app.UmschlagAntwort) (
	mcp.Umschlagangaben, error,
) {
	anwendung, fehler := app.UmschlagflaecheUeber(u.db)
	if fehler != nil {
		return mcp.Umschlagangaben{}, fehler
	}

	antwort := rufe(anwendung)
	if antwort.Fehler != nil {
		return mcp.Umschlagangaben{}, antwort.Fehler.AlsFehler()
	}

	angaben := mcp.Umschlagangaben{
		Urheber:       antwort.Urheber,
		Fingerabdruck: antwort.Fingerabdruck,
		Objekte:       make([]mcp.Umschlagobjekt, 0, len(antwort.Objekte)),
	}

	for _, objekt := range antwort.Objekte {
		angaben.Objekte = append(angaben.Objekte, mcp.Umschlagobjekt{
			Art:         objekt.Art,
			Bezeichnung: objekt.Bezeichnung,
			Groesse:     objekt.Groesse,
			Braucht:     objekt.Braucht,
		})
	}

	return angaben, nil
}
