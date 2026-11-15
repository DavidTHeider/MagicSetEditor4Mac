// **Eine eigene Kopie eines Pakets an der Kommandozeile** — FMT-02.AK-04
// (#1691).
//
// Einordnung nach §16: **erweiternd** — ein Befehl kommt hinzu.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// paketkopierbefehl legt eine eigene Kopie an — Art, Name und höchstens ein
// neuer Name.
func paketkopierbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 || len(rest) > 3 {
		an.klage("paket-kopieren braucht Art und Namen und darf einen neuen Namen tragen.\n%s", hilfe)

		return CodeBedienung
	}

	neu := ""
	if len(rest) == 3 {
		neu = rest[2]
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		weg, err := app.PaketnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		paket, fehler := weg.Kopieren(ctx, rest[0], rest[1], neu)
		if fehler != nil {
			an.klageUeber(fehler)

			return paketCode(fehler)
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, einbindungAntwort{Paket: alsPaketDTO(paket)})
		}

		an.sag("%s\t%s\t%s\t%s\n", paket.Art, paket.Name, paket.Version, paket.Pfad)

		return an.abschluss()
	})
}
