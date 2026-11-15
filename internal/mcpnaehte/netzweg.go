// **Der Netzweg am MCP-Server** —
// #1571.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// netzweg erfüllt `mcp.Netzweg` über der Naht von `app`.
type netzweg struct{ db *sql.DB }

func (n netzweg) Netzstand(ctx context.Context) (mcp.Netzstandangaben, error) {
	antwort := app.NetzflaecheUeber(n.db).Netzstand(ctx)
	if antwort.Fehler != nil {
		return mcp.Netzstandangaben{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Netzstandangaben{
		ProxyHTTP:       netzfeld(antwort.ProxyHTTP),
		ProxyHTTPS:      netzfeld(antwort.ProxyHTTPS),
		ProxyAus:        netzfeld(antwort.ProxyAus),
		FristVerbindung: antwort.FristVerbindung,
		FristKopf:       antwort.FristKopf,
		FristStillstand: antwort.FristStillstand,
		ProxyBenutzt:    antwort.ProxyBenutzt,
	}, nil
}

// netzfeld übersetzt ein Feld samt Herkunft.
func netzfeld(feld app.NetzfeldDTO) mcp.Netzfeld {
	return mcp.Netzfeld{Wert: feld.Wert, Herkunft: feld.Herkunft}
}
