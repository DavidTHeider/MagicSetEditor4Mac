// **Der Arbeitskontext an der Kommandozeile** — Werkstück O1 von Phase 5.2
// (#1425),
// nach [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// kontextbefehl zeigt, wer gerade in welchem Profil als was arbeitet.
//
// **Neben `nutzer-wechseln` und `profil`, nicht darin**: jene bewegen die
// Zeiger, dieser liest den Stand — in einem Zug, damit kein Wechsel zwischen
// zwei Antworten fällt.
func kontextbefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) != 0 {
		an.klage("kontext nimmt keine Angaben.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.NutzerflaecheUeber(db).Arbeitskontext(ctx)
		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		leseart := ""
		if antwort.NurLesend {
			leseart = ", nur lesend"
		}

		an.sag("Es arbeitet der Nutzer %s im Profil %s des Nutzers %s als %s%s "+
			"(über %s, Revision %d).\n",
			antwort.NutzerID, antwort.ProfilID, antwort.EignerID, antwort.Rolle,
			leseart, antwort.Modus, antwort.Revision)

		return CodeGut
	})
}
