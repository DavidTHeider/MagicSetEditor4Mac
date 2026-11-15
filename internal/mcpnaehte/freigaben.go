// **Die Naht zu den Freigaben** — Werkstück 11 von Phase 4.12
// (#1351).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// freigaben bedient die Werkzeuge über einem Bestand.
type freigaben struct{ db *sql.DB }

func (f freigaben) Freigaben(ctx context.Context, profilID string) (
	mcp.Freigabenstand, error,
) {
	return alsFreigabenstand(
		app.NutzerflaecheKopflos(f.db).Freigaben(ctx, profilID))
}

func (f freigaben) FreigabeSetzen(
	ctx context.Context, profilID, nutzerID, rolle string,
) (mcp.Freigabenstand, error) {
	return alsFreigabenstand(app.NutzerflaecheKopflos(f.db).FreigabeSetzen(ctx,
		app.FreigabeangabenDTO{
			ProfilID: profilID, NutzerID: nutzerID, Rolle: rolle,
		}))
}

func (f freigaben) FreigabeLoeschen(
	ctx context.Context, profilID, nutzerID string,
) (mcp.Freigabenstand, error) {
	return alsFreigabenstand(app.NutzerflaecheKopflos(f.db).FreigabeLoeschen(
		ctx, profilID, nutzerID))
}

// alsFreigabenstand bildet die Antwort der Fassade auf die des Werkzeugs ab.
func alsFreigabenstand(antwort app.FreigabenAntwort) (mcp.Freigabenstand, error) {
	if antwort.Fehler != nil {
		return mcp.Freigabenstand{}, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]mcp.Freigabezeile, 0, len(antwort.Freigaben))
	for _, eine := range antwort.Freigaben {
		zeilen = append(zeilen, mcp.Freigabezeile{
			NutzerID: eine.NutzerID, Name: eine.Name,
			Rolle: eine.Rolle, Angelegt: eine.Angelegt,
		})
	}

	return mcp.Freigabenstand{
		Freigaben:      zeilen,
		Sichtbarkeit:   antwort.Sichtbarkeit,
		EigeneRolle:    antwort.EigeneRolle,
		Nachladefehler: alsNachladefehler(antwort.Nachladefehler),
	}, nil
}
