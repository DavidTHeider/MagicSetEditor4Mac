package main

// **Der Dienstkatalog an der Kommandozeile** — Werkstück A1 von Phase 5.3
// (#1440).
//
// `mse dienstkatalog` nennt je Zeile Dienst, Frage, Status und Grund; mit
// `--json` kommt alles — Vertrag und Nachweis — in derselben Gestalt wie
// über MCP und den Dienst.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

type dienstkatalogAntwort struct {
	Eintraege []app.DienstfaehigkeitDTO `json:"eintraege"`
}

func dienstkatalogbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) != 0 {
		an.klage("dienstkatalog nimmt keine Argumente.\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		eintraege, err := app.DienstkatalognahtUeber(db).Katalog(ctx)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, dienstkatalogAntwort{Eintraege: eintraege})
		}

		for _, eintrag := range eintraege {
			an.sag("%s\t%s\t%s\t%s\n", eintrag.Dienst, eintrag.Frage, eintrag.Status, eintrag.Grund)

			if eintrag.Nachweis != nil {
				an.sag("\tgelaufen am %s mit Fassung %s (%s)\n",
					eintrag.Nachweis.Datum, eintrag.Nachweis.Fassung, eintrag.Nachweis.Lauf)
			}
		}

		return an.abschluss()
	})
}
