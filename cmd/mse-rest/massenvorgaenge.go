// **Die Massenvorgänge des Dienstes** —
// #1292.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// massenvorgaenge führt die Vorgänge an die Naht von `internal/rest`.
type massenvorgaenge struct{ db *sql.DB }

func (m massenvorgaenge) MasseAendern(
	ctx context.Context, aenderung rest.Massenaenderung,
) (any, error) {
	return app.MassennahtUeber(m.db).Aendern(ctx, app.MassenaenderungDTO{
		Zeilen:   alsAppzeilen(aenderung.Zeilen),
		Menge:    aenderung.Menge,
		Zustand:  aenderung.Zustand,
		Lagerort: aenderung.Lagerort,
		Wunsch:   aenderung.Wunsch,
	})
}

func (m massenvorgaenge) MasseLoeschen(
	ctx context.Context, zeilen []rest.Massenzeile,
) (any, error) {
	return app.MassennahtUeber(m.db).Loeschen(ctx, alsAppzeilen(zeilen))
}

func (m massenvorgaenge) MassePreise(
	ctx context.Context, hoechstensStunden int,
) (any, error) {
	return app.MassennahtUeber(m.db).PreiseInMasse(ctx, hoechstensStunden)
}

// alsAppzeilen bildet die Zeilen ab.
func alsAppzeilen(zeilen []rest.Massenzeile) []app.MassenzeileDTO {
	gebaut := make([]app.MassenzeileDTO, 0, len(zeilen))
	for _, zeile := range zeilen {
		gebaut = append(gebaut, app.MassenzeileDTO{
			ID: zeile.ID, Fassung: zeile.Fassung,
		})
	}

	return gebaut
}
