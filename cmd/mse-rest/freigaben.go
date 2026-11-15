// **Die Naht zu den Freigaben im Dienst** — Werkstück 11 von Phase 4.12
// (#1351).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// fremdfreigaben bedient die Wege über der Fassade.
type fremdfreigaben struct{ db *sql.DB }

func (f fremdfreigaben) Freigaben(ctx context.Context, profilID string) (
	rest.Freigabenstand, error,
) {
	return alsFreigabenstand(
		app.NutzerflaecheKopflos(f.db).Freigaben(ctx, profilID))
}

func (f fremdfreigaben) FreigabeSetzen(
	ctx context.Context, profilID, nutzerID, rolle string,
) (rest.Freigabenstand, error) {
	return alsFreigabenstand(app.NutzerflaecheKopflos(f.db).FreigabeSetzen(ctx,
		app.FreigabeangabenDTO{
			ProfilID: profilID, NutzerID: nutzerID, Rolle: rolle,
		}))
}

func (f fremdfreigaben) FreigabeLoeschen(
	ctx context.Context, profilID, nutzerID string,
) (rest.Freigabenstand, error) {
	return alsFreigabenstand(app.NutzerflaecheKopflos(f.db).FreigabeLoeschen(
		ctx, profilID, nutzerID))
}

// alsFreigabenstand bildet die Antwort der Fassade auf die des Dienstes ab.
func alsFreigabenstand(antwort app.FreigabenAntwort) (rest.Freigabenstand, error) {
	if antwort.Fehler != nil {
		return rest.Freigabenstand{}, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]rest.Freigabezeile, 0, len(antwort.Freigaben))
	for _, eine := range antwort.Freigaben {
		zeilen = append(zeilen, rest.Freigabezeile{
			NutzerID: eine.NutzerID, Name: eine.Name,
			Rolle: eine.Rolle, Angelegt: eine.Angelegt,
		})
	}

	return rest.Freigabenstand{
		Freigaben:      zeilen,
		Sichtbarkeit:   antwort.Sichtbarkeit,
		EigeneRolle:    antwort.EigeneRolle,
		Nachladefehler: alsNachladebefund(antwort.Nachladefehler),
	}, nil
}
