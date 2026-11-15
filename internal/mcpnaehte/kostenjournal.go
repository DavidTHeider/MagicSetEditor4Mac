package mcpnaehte

// **Das Kostenjournal über MCP** — Werkstück O12 von Phase 5.2
// (#1439).
//
// **Nur der Stand**, und der Grund steht im Kopf von `mcp/kostenjournal.go`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// kostenauskunft reicht den Stand an das Werkzeug.
type kostenauskunft struct{ db *sql.DB }

func (k kostenauskunft) Kostenstand(ctx context.Context, konto, zeitraum string) (
	mcp.Haushaltsstand, error,
) {
	antwort := app.KostenAnwendungUeber(k.db).Kostenstand(ctx, konto, zeitraum)
	if antwort.Fehler != nil {
		return mcp.Haushaltsstand{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Haushaltsstand{
		Konto:      antwort.Konto,
		Zeitraum:   antwort.Zeitraum,
		Grenze:     betrag(antwort.Grenze),
		Reserviert: betrag(antwort.Reserviert),
		Verbraucht: betrag(antwort.Verbraucht),
		Unklar:     betrag(antwort.Unklar),
		Verfuegbar: betrag(antwort.Verfuegbar),
	}, nil
}

func betrag(aus app.BetragDTO) mcp.Betrag {
	return mcp.Betrag{Wert: aus.Wert, Waehrung: aus.Waehrung}
}
