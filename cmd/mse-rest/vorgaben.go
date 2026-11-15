// **Die Naht zu den Profilvorgaben im Dienst** — Werkstück 15 von Phase 4.12
// (#1351).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// fremdvorgaben bedient die Wege über der Fassade.
type fremdvorgaben struct{ db *sql.DB }

func (f fremdvorgaben) Vorgaben(ctx context.Context, profilID string) (
	[]rest.Vorgabenzeile, error,
) {
	antwort := app.NutzerflaecheKopflos(f.db).Profilvorgaben(ctx, profilID)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return alsVorgabenzeilen(antwort), nil
}

func (f fremdvorgaben) VorgabeSetzen(
	ctx context.Context, profilID, art string, werte []string,
) (rest.Vorgabengesetzt, error) {
	antwort := app.NutzerflaecheKopflos(f.db).ProfilvorgabeSetzen(ctx,
		app.VorgabeangabenDTO{ProfilID: profilID, Art: art, Werte: werte})
	if antwort.Fehler != nil {
		return rest.Vorgabengesetzt{}, antwort.Fehler.AlsFehler()
	}

	return rest.Vorgabengesetzt{
		Vorgaben:       alsVorgabenzeilen(antwort),
		Nachladefehler: alsNachladebefund(antwort.Nachladefehler),
	}, nil
}

// alsVorgabenzeilen bildet die Antwort der Fassade auf die des Dienstes ab.
func alsVorgabenzeilen(antwort app.VorgabenAntwort) []rest.Vorgabenzeile {
	zeilen := make([]rest.Vorgabenzeile, 0, len(antwort.Vorgaben))

	for _, stand := range antwort.Vorgaben {
		moeglich := make([]rest.Vorgabeangebot, 0, len(stand.Moeglich))
		for _, wahl := range stand.Moeglich {
			moeglich = append(moeglich, rest.Vorgabeangebot{
				Wert: wahl.Wert, Beschriftung: wahl.Beschriftung,
			})
		}

		zeilen = append(zeilen, rest.Vorgabenzeile{
			Art:        stand.Art,
			Mehrwertig: stand.Mehrwertig,
			Werte:      stand.Werte,
			Moeglich:   moeglich,
			Verwaist:   stand.Verwaist,
		})
	}

	return zeilen
}
