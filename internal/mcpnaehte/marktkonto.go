// **Das eigene Konto des MCP-Servers** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// **Ohne Wache** (0044): der Weg liest und fasst keine Datei an. Was er
// **nicht** kann, ist schreiben — nicht hier und nicht beim Anbieter.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// marktkonto erfüllt `mcp.Kontoauskunft` über der Naht von `internal/app`.
type marktkonto struct{ db *sql.DB }

func (m marktkonto) Marktkonto(
	ctx context.Context, quelle string, seite, menge int,
) (mcp.Kontostand, error) {
	antwort := app.PeerflaecheUeber(m.db).Marktkonto(ctx, quelle, seite, menge)
	if antwort.Fehler != nil {
		return mcp.Kontostand{}, antwort.Fehler.AlsFehler()
	}

	angebote := make([]mcp.Eigenangebot, 0, len(antwort.Stand.Angebote))
	for _, angebot := range antwort.Stand.Angebote {
		angebote = append(angebote, mcp.Eigenangebot{
			ID: angebot.ID, BlueprintID: angebot.BlueprintID,
			Name: angebot.Name, Menge: angebot.Menge,
			PreisCent: angebot.PreisCent, Waehrung: angebot.Waehrung,
			Eigennotiz: angebot.Eigennotiz,
		})
	}

	bestellungen := make([]mcp.Marktbestellung, 0,
		len(antwort.Stand.Bestellungen))
	for _, bestellung := range antwort.Stand.Bestellungen {
		bestellungen = append(bestellungen, mcp.Marktbestellung{
			ID: bestellung.ID, Code: bestellung.Code,
			Zustand:        bestellung.Zustand,
			UeberNullkonto: bestellung.UeberNullkonto,
			Rolle:          bestellung.Rolle, Stueck: bestellung.Stueck,
			Bezahlt: bestellung.Bezahlt, BewegtBestand: bestellung.BewegtBestand,
		})
	}

	return mcp.Kontostand{
		Name: antwort.Stand.Name, NutzerID: antwort.Stand.NutzerID,
		Angebote: angebote, Bestellungen: bestellungen,
		Befunde: antwort.Stand.Befunde,
	}, nil
}
