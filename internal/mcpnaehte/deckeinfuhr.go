// **Die Decklisteneinfuhr des MCP-Servers** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// **Ohne Wache** (0044): `Vorschau` schreibt nichts, und `Uebernehmen`
// schreibt nur in den Bestand — es überschreibt keine Datei auf der Platte.
// Was es nicht tut, ist selbst entscheiden, welcher von 72 Drucken gemeint
// ist; das steht in `internal/mcp/deckeinfuhr.go`.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// deckeinfuhr erfüllt `mcp.Deckeinfuhr` über der Naht von `internal/app`.
type deckeinfuhr struct{ db *sql.DB }

func (d deckeinfuhr) Vorschau(
	ctx context.Context, deckID int64, text string,
) (mcp.Deckvorschau, error) {
	naht, err := app.DeckeinfuhrUeber(d.db)
	if err != nil {
		return mcp.Deckvorschau{}, err
	}

	vorschau, err := naht.Vorschau(ctx, deckID, text)
	if err != nil {
		return mcp.Deckvorschau{}, err
	}

	zeilen := make([]mcp.Deckvorschauzeile, 0, len(vorschau.Zeilen))
	for _, zeile := range vorschau.Zeilen {
		zeilen = append(zeilen, alsMCPVorschauzeile(zeile))
	}

	return mcp.Deckvorschau{
		DeckID: vorschau.DeckID, Stand: vorschau.Stand, Zeilen: zeilen,
		Befunde: vorschau.Befunde, Hauptdeck: vorschau.Hauptdeck,
		Sideboard: vorschau.Sideboard,
	}, nil
}

// alsMCPVorschauzeile bildet eine Zeile auf ihre Gestalt für MCP ab.
func alsMCPVorschauzeile(zeile app.Deckvorschauzeile) mcp.Deckvorschauzeile {
	treffer := make([]mcp.Listentreffer, 0, len(zeile.Treffer))
	for _, einer := range zeile.Treffer {
		treffer = append(treffer, mcp.Listentreffer{
			Printing: einer.Printing, Name: einer.Name,
			Setcode: einer.Setcode, Setname: einer.Setname,
			Nummer: einer.Nummer, Seltenheit: einer.Seltenheit,
		})
	}

	return mcp.Deckvorschauzeile{
		Bereich: zeile.Bereich, Menge: zeile.Menge, Name: zeile.Name,
		Setcode: zeile.Setcode, Nummer: zeile.Nummer,
		Printing: zeile.Printing, Treffer: treffer, Drucke: zeile.Drucke,
		Befund: zeile.Befund, Uebernehmbar: zeile.Uebernehmbar,
	}
}

func (d deckeinfuhr) Uebernehmen(
	ctx context.Context, deckID int64, stand string,
	zeilen []mcp.Deckeinfuhrzeile,
) (mcp.Deckeinfuhrbericht, error) {
	naht, err := app.DeckeinfuhrUeber(d.db)
	if err != nil {
		return mcp.Deckeinfuhrbericht{}, err
	}

	bestaetigt := make([]app.Deckeinfuhrzeile, 0, len(zeilen))
	for _, zeile := range zeilen {
		bestaetigt = append(bestaetigt, app.Deckeinfuhrzeile{
			Printing: zeile.Printing, Menge: zeile.Menge,
			Bereich: zeile.Bereich,
		})
	}

	bericht, err := naht.Uebernehmen(ctx, deckID, stand, bestaetigt)
	if err != nil {
		return mcp.Deckeinfuhrbericht{}, err
	}

	return mcp.Deckeinfuhrbericht{
		Gelegt: bericht.Gelegt, Stueck: bericht.Stueck, Stand: bericht.Stand,
	}, nil
}
