package mcpnaehte

// **Die Kartenerkennung an der MCP-Naht** — Werkstück A13.3 von Phase 5.3
// (#1452).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// KIErkennung liest ein Kartenfoto und nennt die passenden Drucke.
func (p *preise) KIErkennung(ctx context.Context, bezug mcp.KIErkennungsbezug) (
	mcp.KIErkennungssicht, error,
) {
	antwort := p.kianwendung().KIKarteErkennen(ctx, app.KIErkennungsbezug{
		Pfad: bezug.Pfad, Anbieter: bezug.Anbieter, Alias: bezug.Alias,
	})
	if antwort.Fehler != nil {
		return mcp.KIErkennungssicht{}, antwort.Fehler.AlsFehler()
	}

	beobachtungen := make([]mcp.KIBeobachtungsicht, 0, len(antwort.Beobachtungen))
	for _, eine := range antwort.Beobachtungen {
		beobachtungen = append(beobachtungen, mcp.KIBeobachtungsicht(eine))
	}

	kandidaten := make([]mcp.KIBildkandidatsicht, 0, len(antwort.Kandidaten))
	for _, einer := range antwort.Kandidaten {
		kandidaten = append(kandidaten, mcp.KIBildkandidatsicht{
			Kennung: einer.Kennung, Name: einer.Name,
			Setcode: einer.Setcode, Setname: einer.Setname,
			Nummer: einer.Nummer, Seltenheit: einer.Seltenheit,
			Gestuetzt: einer.Gestuetzt,
		})
	}

	return mcp.KIErkennungssicht{
		Beobachtungen: beobachtungen,
		NichtGelesen:  antwort.NichtGelesen,
		Befunde:       alsBelegbefundsichten(antwort.Befunde),
		Kandidaten:    kandidaten,
		Drucke:        antwort.Drucke, Weitere: antwort.Weitere,
		Eindeutig: antwort.Eindeutig, Alter: antwort.Alter,
		Anbieter: antwort.Anbieter, Modell: antwort.Modell,
	}, nil
}
