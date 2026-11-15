package mcpnaehte

// **Die Kartensuche für den MCP-Server** — Werkstück 1 von Phase 3.6
// (#1167).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// fremdsuche reicht die Suche an die Naht durch.
// **Sie hält den Bestand, seit es Dienstschalter gibt** (Phase 4.8): dort
// steht, ob Scryfall gefragt werden darf. Ohne ihn ginge ein abgeschalteter
// Dienst hier weiter hinaus, und die Zusage gälte nur im Fenster.
type fremdsuche struct{ db *sql.DB }

// Suche gibt die Treffer zu einer Frage.
//
// **Die Naht wird je Ruf gebaut**, wie bei jeder anderen Scryfall-Fähigkeit
// dieses Prozesses: es gibt keinen Start, an dem sie einmal entstünde.
func (f fremdsuche) Kartensuche(ctx context.Context, frage string) (mcp.Suchergebnis, error) {
	naht, err := app.Fremdsuchnaht(ctx, f.db)
	if err != nil {
		return mcp.Suchergebnis{}, err
	}

	ergebnis, suchfehler := naht.Suche(ctx, frage)
	if suchfehler != nil {
		return mcp.Suchergebnis{}, suchfehler
	}

	// **Feld für Feld und nicht über JSON** — dieselbe Regel wie bei der
	// Sammlung: ein Umweg über `Marshal` sähe kürzer aus und verlöre beim
	// ersten Feld, dessen Marken auseinanderlaufen, genau dieses Feld, still.
	treffer := make([]mcp.Suchtrefferang, 0, len(ergebnis.Treffer))
	for _, eines := range ergebnis.Treffer {
		treffer = append(treffer, mcp.Suchtrefferang{
			Kennung: eines.Kennung, Name: eines.Name,
			Setcode: eines.Setcode, Setname: eines.Setname,
			Nummer: eines.Nummer, Seltenheit: eines.Seltenheit,
			Manakosten: eines.Manakosten, Typzeile: eines.Typzeile,
			Bild: eines.Bild,
		})
	}

	return mcp.Suchergebnis{
		Treffer:      treffer,
		Weitere:      ergebnis.Weitere,
		AlterStunden: ergebnis.AlterStunden,
		Befunde:      ergebnis.Befunde,
	}, nil
}
