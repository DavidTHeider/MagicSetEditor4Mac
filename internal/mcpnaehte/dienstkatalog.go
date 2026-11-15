package mcpnaehte

// **Der Dienstkatalog für den MCP-Server** — Werkstück A1 von Phase 5.3
// (#1440). Derselbe Weg wie im Fenster und in den zwei anderen Flächen:
// `app.DienstkatalognahtUeber`, je Ruf gebaut, damit ein umgelegter
// Schalter beim nächsten Ruf gilt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type fremdkatalog struct{ db *sql.DB }

func (f fremdkatalog) Dienstkatalog(ctx context.Context) ([]mcp.Dienstfaehigkeit, error) {
	eintraege, err := app.DienstkatalognahtUeber(f.db).Katalog(ctx)
	if err != nil {
		return nil, err
	}

	umgesetzt := make([]mcp.Dienstfaehigkeit, 0, len(eintraege))

	for _, eintrag := range eintraege {
		var nachweis *mcp.Dienstnachweis
		if eintrag.Nachweis != nil {
			nachweis = &mcp.Dienstnachweis{
				Lauf: eintrag.Nachweis.Lauf, Datum: eintrag.Nachweis.Datum,
				Fassung: eintrag.Nachweis.Fassung, Endpunkte: eintrag.Nachweis.Endpunkte,
				GiltFuerDieseFassung: eintrag.Nachweis.GiltFuerDieseFassung,
			}
		}

		umgesetzt = append(umgesetzt, mcp.Dienstfaehigkeit{
			Dienst: eintrag.Dienst, Frage: eintrag.Frage, Status: eintrag.Status,
			Grund: eintrag.Grund, Angabe: eintrag.Angabe,
			HatWahl: eintrag.HatWahl, Gewaehlt: eintrag.Gewaehlt,
			Wirkung: eintrag.Wirkung,
			Vertrag: mcp.Dienstvertrag{
				Basisadresse: eintrag.Vertrag.Basisadresse, Endpunkte: eintrag.Vertrag.Endpunkte,
				Zugang: eintrag.Vertrag.Zugang, Tarif: eintrag.Vertrag.Tarif,
				Grenze: eintrag.Vertrag.Grenze, Frist: eintrag.Vertrag.Frist,
				Bedingungen: eintrag.Vertrag.Bedingungen, Abgerufen: eintrag.Vertrag.Abgerufen,
				Beleg: eintrag.Vertrag.Beleg,
			},
			Nachweis: nachweis,
		})
	}

	return umgesetzt, nil
}
