// **Die Sammlungsausfuhr des MCP-Servers** —
// #1292.
//
// **Ohne Wache** (0044): beide Wege fassen nur den Bestand an. Das Ausgeben
// schreibt keine Datei — es gibt Text zurück, und wohin der geht, entscheidet,
// wer ruft.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// sammlungsausfuhr erfüllt `mcp.Sammlungsausfuhr`.
type sammlungsausfuhr struct{ db *sql.DB }

func (s sammlungsausfuhr) Ausgeben(ctx context.Context, form string) (string, error) {
	return app.AusfuhrUeber(s.db).Ausgeben(ctx, form)
}

func (s sammlungsausfuhr) CSVEinlesen(ctx context.Context, text, quelle, zusammenfuehrung string) (
	mcp.Einfuhrbericht, error,
) {
	bericht, err := app.AusfuhrUeber(s.db).CSVEinlesen(ctx, text, quelle, zusammenfuehrung)
	if bericht == nil {
		return mcp.Einfuhrbericht{Angelegt: []int64{}, Befunde: []string{}}, err
	}

	return alsMCPBericht(bericht), err
}

// CSVVorschauen sagt, was das Einlesen täte — DOM-03.AK-01 (#1698).
func (s sammlungsausfuhr) CSVVorschauen(ctx context.Context, text string) (mcp.Sammelvorschau, error) {
	vorschau, err := app.AusfuhrUeber(s.db).CSVVorschauen(ctx, text)
	if err != nil {
		return mcp.Sammelvorschau{}, err
	}

	return alsMCPVorschau(vorschau), nil
}

// alsMCPBericht bildet den Bericht der Naht ab — für Liste und CSV.
func alsMCPBericht(bericht *app.Einfuhrbericht) mcp.Einfuhrbericht {
	return mcp.Einfuhrbericht{
		Angelegt: bericht.Angelegt, Erhoeht: bericht.Erhoeht, Ausgelassen: bericht.Ausgelassen,
		Stueck: bericht.Stueck, Befunde: bericht.Befunde,
	}
}

// alsMCPVorschau bildet die Vorschau der Naht ab.
func alsMCPVorschau(vorschau *app.Sammelvorschau) mcp.Sammelvorschau {
	posten := make([]mcp.Vorschauposten, 0, len(vorschau.Posten))

	for _, einer := range vorschau.Posten {
		posten = append(posten, mcp.Vorschauposten(einer))
	}

	return mcp.Sammelvorschau{Posten: posten, Befunde: vorschau.Befunde}
}
