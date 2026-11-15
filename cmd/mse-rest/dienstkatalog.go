package main

// **Der Dienstkatalog für den Dienst** — Werkstück A1 von Phase 5.3 (#1440).
// Dieselbe Naht wie bei MCP; die Abbildung steht zweimal, weil `rest` und
// `mcp` je eigene Sichten führen und keines das andere kennt (§2).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

type fremdkatalog struct{ db *sql.DB }

func (f fremdkatalog) Dienstkatalog(ctx context.Context) ([]rest.Dienstfaehigkeit, error) {
	eintraege, err := app.DienstkatalognahtUeber(f.db).Katalog(ctx)
	if err != nil {
		return nil, err
	}

	umgesetzt := make([]rest.Dienstfaehigkeit, 0, len(eintraege))

	for _, eintrag := range eintraege {
		var nachweis *rest.Dienstnachweis
		if eintrag.Nachweis != nil {
			nachweis = &rest.Dienstnachweis{
				Lauf: eintrag.Nachweis.Lauf, Datum: eintrag.Nachweis.Datum,
				Fassung: eintrag.Nachweis.Fassung, Endpunkte: eintrag.Nachweis.Endpunkte,
				GiltFuerDieseFassung: eintrag.Nachweis.GiltFuerDieseFassung,
			}
		}

		umgesetzt = append(umgesetzt, rest.Dienstfaehigkeit{
			Dienst: eintrag.Dienst, Frage: eintrag.Frage, Status: eintrag.Status,
			Grund: eintrag.Grund, Angabe: eintrag.Angabe,
			HatWahl: eintrag.HatWahl, Gewaehlt: eintrag.Gewaehlt,
			Wirkung: eintrag.Wirkung,
			Vertrag: rest.Dienstvertrag{
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
