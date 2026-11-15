// **Die Naht zu den Profilvorgaben** — Werkstück 15 von Phase 4.12
// (#1351).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// vorgaben bedient die Werkzeuge über einem Bestand.
type vorgaben struct{ db *sql.DB }

// Vorgaben nennt die Vorgaben eines Profils samt ihrem Angebot.
func (v vorgaben) Vorgaben(ctx context.Context, profilID string) (
	[]mcp.Vorgabenzeile, error,
) {
	antwort := app.NutzerflaecheKopflos(v.db).Profilvorgaben(ctx, profilID)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return alsVorgabenzeilen(antwort), nil
}

// VorgabeSetzen ersetzt die Werte einer Art und gibt den ganzen Stand.
func (v vorgaben) VorgabeSetzen(
	ctx context.Context, profilID, art string, werte []string,
) (mcp.Vorgabengesetzt, error) {
	antwort := app.NutzerflaecheKopflos(v.db).ProfilvorgabeSetzen(ctx,
		app.VorgabeangabenDTO{ProfilID: profilID, Art: art, Werte: werte})
	if antwort.Fehler != nil {
		return mcp.Vorgabengesetzt{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Vorgabengesetzt{
		Vorgaben:       alsVorgabenzeilen(antwort),
		Nachladefehler: alsNachladefehler(antwort.Nachladefehler),
	}, nil
}

// alsVorgabenzeilen bildet die Antwort der Fassade auf die der Werkzeuge ab.
func alsVorgabenzeilen(antwort app.VorgabenAntwort) []mcp.Vorgabenzeile {
	zeilen := make([]mcp.Vorgabenzeile, 0, len(antwort.Vorgaben))

	for _, stand := range antwort.Vorgaben {
		moeglich := make([]mcp.Vorgabeangebot, 0, len(stand.Moeglich))
		for _, wahl := range stand.Moeglich {
			moeglich = append(moeglich, mcp.Vorgabeangebot{
				Wert: wahl.Wert, Beschriftung: wahl.Beschriftung,
			})
		}

		zeilen = append(zeilen, mcp.Vorgabenzeile{
			Art:        stand.Art,
			Mehrwertig: stand.Mehrwertig,
			Werte:      stand.Werte,
			Moeglich:   moeglich,
			Verwaist:   stand.Verwaist,
		})
	}

	return zeilen
}
