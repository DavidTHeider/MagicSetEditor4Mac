// **Der Profilwechsel an der Kommandozeile** — Werkstück 3 von Phase 4.12
// (#1351).

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// profilbefehl nennt das gezeigte Profil — oder wechselt.
//
// # Ein Befehl und nicht zwei
//
// **`mse profil` fragt, `mse profil <kennung>` stellt um.** Das ist die Form,
// die `git branch` und `git switch` auseinanderhalten — hier ist sie einer,
// weil die Frage ohne die Antwort nicht vorkommt: wer wechseln will, sieht
// vorher nach, wo er steht.
//
// **Die Liste steht in `mse nutzer`**, samt Zugehörigkeit. Ein `mse profile`
// daneben zeigte dieselben Zeilen ohne ihren Nutzer.
func profilbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) > 2 {
		an.klage("profil nimmt eine Kennung und darf einen Nachweis tragen.\n%s",
			hilfe)

		return CodeBedienung
	}

	// **Der Nachweis ist der zweite Faktor** — Werkstück 7 (#1351): ein Code
	// aus der Authenticator-App oder ein Wiederherstellungscode. **Leer heißt
	// „über den Sensor"**, und wer keinen Faktor hat, gibt gar nichts an.
	//
	// **Anders als das Kennwort steht er im Argument und nicht auf der
	// Standardeingabe**, und das ist eine Entscheidung: ein Einmalkennwort
	// gilt neunzig Sekunden. Was danach in `~/.zsh_history` steht, ist kein
	// Geheimnis mehr, sondern eine Zahl.
	nachweis := ""
	if len(rest) == 2 {
		nachweis = rest[1]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		flaeche := app.NutzerflaecheUeber(db)

		antwort := flaeche.GezeigtesProfil(ctx)
		if len(rest) >= 1 {
			antwort = flaeche.ProfilWechseln(ctx, rest[0], nachweis)
		}

		if antwort.Fehler != nil {
			an.klageUeberFehler(antwort.Fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort)
		}

		an.sag("%s\t%s\n", antwort.Profil.ID, antwort.Profil.Name)

		return an.abschluss()
	})
}
