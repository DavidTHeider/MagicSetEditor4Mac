package mcpnaehte

// **Die Verweise für den MCP-Server** — Werkstück 2 von Phase 3.6
// (#1167).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// fremdverweise reicht die Verweise an die Naht durch.
type fremdverweise struct{ db *sql.DB }

// VerweisSetzen bezieht eine Karte auf einen Druck.
func (f fremdverweise) VerweisSetzen(
	ctx context.Context, karteID int64, printing string,
) (mcp.Verweisangabe, error) {
	naht, err := app.Verweisflaechennaht(f.db)
	if err != nil {
		return mcp.Verweisangabe{}, err
	}

	verweis, setzfehler := naht.Setzen(ctx, karteID, printing)
	if setzfehler != nil {
		return mcp.Verweisangabe{}, setzfehler
	}

	return alsVerweisangabe(verweis), nil
}

// Verweis nennt den Druck, auf den eine Karte verweist.
func (f fremdverweise) Verweis(ctx context.Context, karteID int64) (
	*mcp.Verweisangabe, error,
) {
	naht, err := app.Verweisflaechennaht(f.db)
	if err != nil {
		return nil, err
	}

	verweis, lesefehler := naht.Fuer(ctx, karteID)
	if lesefehler != nil || verweis == nil {
		return nil, lesefehler
	}

	angabe := alsVerweisangabe(*verweis)

	return &angabe, nil
}

// VerweisLoesen nimmt den Verweis zurück.
func (f fremdverweise) VerweisLoesen(ctx context.Context, karteID int64) error {
	naht, err := app.Verweisflaechennaht(f.db)
	if err != nil {
		return err
	}

	return naht.Loesen(ctx, karteID)
}

// alsVerweisangabe übersetzt Feld für Feld.
//
// **Und nicht über JSON** — dieselbe Regel wie bei der Suche: ein Umweg über
// `Marshal` sähe kürzer aus und verlöre beim ersten Feld, dessen Marken
// auseinanderlaufen, genau dieses Feld, still.
func alsVerweisangabe(verweis app.KartenverweisDTO) mcp.Verweisangabe {
	return mcp.Verweisangabe{
		KarteID: verweis.KarteID, Printing: verweis.Printing,
		DruckID: verweis.DruckID, Vorlaeufig: verweis.Vorlaeufig,
		Name: verweis.Name, Setcode: verweis.Setcode, Nummer: verweis.Nummer,
		Adresse: verweis.Adresse, Gesetzt: verweis.Gesetzt,
	}
}
