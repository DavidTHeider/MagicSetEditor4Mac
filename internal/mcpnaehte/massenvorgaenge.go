// **Die Massenvorgänge des MCP-Servers** —
// #1292.
//
// **Ohne Wache auf die Platte** (0044): alle drei fassen nur den Bestand an.
// Was sie **nicht** tun, ist eine Fassung erfinden — jede Zeile bringt ihre
// mit, und eine veraltete wird übergangen statt überschrieben.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// massenvorgaenge erfüllt `mcp.Massenvorgaenge`.
type massenvorgaenge struct{ db *sql.DB }

func (m massenvorgaenge) Aendern(
	ctx context.Context, aenderung mcp.Massenaenderung,
) (mcp.Massenbericht, error) {
	bericht, err := app.MassennahtUeber(m.db).Aendern(ctx, app.MassenaenderungDTO{
		Zeilen:   alsAppzeilen(aenderung.Zeilen),
		Menge:    aenderung.Menge,
		Zustand:  aenderung.Zustand,
		Lagerort: aenderung.Lagerort,
		Wunsch:   aenderung.Wunsch,
	})

	return alsMCPbericht(bericht), err
}

func (m massenvorgaenge) Loeschen(
	ctx context.Context, zeilen []mcp.Massenzeile,
) (mcp.Massenbericht, error) {
	bericht, err := app.MassennahtUeber(m.db).Loeschen(ctx, alsAppzeilen(zeilen))

	return alsMCPbericht(bericht), err
}

func (m massenvorgaenge) PreiseInMasse(
	ctx context.Context, hoechstensStunden int,
) (mcp.Preisschub, error) {
	schub, err := app.MassennahtUeber(m.db).PreiseInMasse(ctx, hoechstensStunden)
	if schub == nil {
		return mcp.Preisschub{Befunde: []string{}}, err
	}

	return mcp.Preisschub{
		Geholt: schub.Geholt, Offen: schub.Offen,
		Fehlgeschlagen: schub.Fehlgeschlagen,
		Fertig:         schub.Fertig, Steht: schub.Steht,
		Befunde: schub.Befunde,
	}, err
}

// alsAppzeilen bildet die Zeilen ab.
func alsAppzeilen(zeilen []mcp.Massenzeile) []app.MassenzeileDTO {
	gebaut := make([]app.MassenzeileDTO, 0, len(zeilen))
	for _, zeile := range zeilen {
		gebaut = append(gebaut, app.MassenzeileDTO{
			ID: zeile.ID, Fassung: zeile.Fassung,
		})
	}

	return gebaut
}

// alsMCPbericht bildet den Bericht ab — **auch einen leeren**.
//
// Ein `nil` käme als `null` heraus, wo die Gegenseite ein Objekt erwartet.
func alsMCPbericht(bericht *app.MassenberichtDTO) mcp.Massenbericht {
	if bericht == nil {
		return mcp.Massenbericht{Befunde: []string{}, Vorher: []mcp.Massenstand{}}
	}

	vorher := make([]mcp.Massenstand, 0, len(bericht.Vorher))
	for _, stand := range bericht.Vorher {
		vorher = append(vorher, mcp.Massenstand(stand))
	}

	return mcp.Massenbericht{
		Geaendert: bericht.Geaendert, Befunde: bericht.Befunde, Vorher: vorher,
	}
}
