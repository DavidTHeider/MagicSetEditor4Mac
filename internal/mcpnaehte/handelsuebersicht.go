// **Die Liste der Handelsvorgänge des MCP-Servers** — TRADE-01
// (#1715).
//
// **Ohne Wache**, wie der Handelsstand: der Weg liest nur.

package mcpnaehte

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// Handelsuebersicht nennt die Vorgänge dieses Profils, gefiltert.
func (h handelsstand) Handelsuebersicht(
	ctx context.Context, filter mcp.Handelsfilter,
) ([]mcp.Handelsuebersichtszeile, error) {
	antwort := app.PeerflaecheUeber(h.db).HandelUebersicht(ctx, app.HandelsfilterDTO{
		Partner: filter.Partner, Lagen: filter.Lagen, Suche: filter.Suche, Marke: filter.Marke,
		Seit: filter.Seit, Bis: filter.Bis, Ordnung: filter.Ordnung, Archiv: filter.Archiv,
	})
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]mcp.Handelsuebersichtszeile, 0, len(antwort.Zeilen))

	for _, zeile := range antwort.Zeilen {
		partner := make([]mcp.Handelspartner, 0, len(zeile.Gegenueber))
		for _, gegenueber := range zeile.Gegenueber {
			partner = append(partner, mcp.Handelspartner{Kennung: gegenueber.Kennung, Name: gegenueber.Name})
		}

		zeilen = append(zeilen, mcp.Handelsuebersichtszeile{
			Vorgang: alsHandelsstand(zeile.Vorgang), Lage: zeile.Lage, Naechster: zeile.Naechster,
			Gegenueber: partner, Karten: zeile.Karten, Marken: zeile.Marken,
			Angelegt: zeile.Angelegt, Geaendert: zeile.Geaendert,
			Archiviert: zeile.Archiviert, Unbestaetigt: zeile.Unbestaetigt,
		})
	}

	return zeilen, nil
}
