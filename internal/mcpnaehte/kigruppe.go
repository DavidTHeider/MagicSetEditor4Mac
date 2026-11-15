package mcpnaehte

// **Die atomare Gruppe an der MCP-Naht** — Werkstück A13.1b von Phase 5.3
// (#1452).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// KIGruppe übernimmt mehrere Vorschläge als eine Gruppe.
func (p *preise) KIGruppe(ctx context.Context, bezug mcp.KIGruppenbezug) (
	mcp.KIGruppensicht, error,
) {
	wahl := make([]app.KIGruppenwahlDTO, 0, len(bezug.Wahl))
	for _, eine := range bezug.Wahl {
		wahl = append(wahl, app.KIGruppenwahlDTO{
			Kennung: eine.Kennung, Felder: eine.Felder,
		})
	}

	antwort := p.kianwendung().KIGruppeUebernehmen(ctx, wahl)
	if antwort.Fehler != nil {
		return mcp.KIGruppensicht{}, antwort.Fehler.AlsFehler()
	}

	namen := make([]mcp.KIGruppenkartensicht, 0, len(antwort.Namen))
	for _, zeile := range antwort.Namen {
		namen = append(namen, mcp.KIGruppenkartensicht(zeile))
	}

	return mcp.KIGruppensicht{
		Karten: antwort.Karten, Felder: antwort.Felder, Namen: namen,
	}, nil
}
