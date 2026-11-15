// **Der Symboleditor über MCP** — der Adapter
// (#1391).
//
// **Hier und nicht in `internal/mcp`**: der Schichtenwächter lässt dort weder
// `service` noch `app` zu, und das aus gutem Grund — was zusammengesetzt
// wird, gehört an eine Stelle.

package mcpnaehte

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// symbolwerkstatt bedient die Symbolwerkzeuge.
// **Mit dem Bestand**, seit die Palette dazukam: sie hängt an einer Karte
// und ihrer Vorlage, die übrigen Wege nicht.
type symbolwerkstatt struct{ db *sql.DB }

// Lesen holt das abgelegte Symbol eines Sets — im Profil des Rufs (#1815).
func (s symbolwerkstatt) SymbolLesen(ctx context.Context, setID int64, name string) (any, error) {
	naht, err := app.SymbolwerkstattUeber(s.db)
	if err != nil {
		return nil, err
	}

	return naht.Lesen(ctx, setID, name)
}

// Ablegen schreibt eine Zeichnung neben das Set.
//
// **Die Teile kommen als rohes JSON.** `internal/mcp` ist ein Blatt und kennt
// `app.SymbolteilDTO` nicht; die Gestalt wird hier gelesen, wo beide Seiten
// bekannt sind.
func (s symbolwerkstatt) SymbolAblegen(
	ctx context.Context, setID int64, teile json.RawMessage, unter string,
) (string, error) {
	naht, err := app.SymbolwerkstattUeber(s.db)
	if err != nil {
		return "", err
	}

	var gelesen []app.SymbolteilDTO
	if err := json.Unmarshal(teile, &gelesen); err != nil {
		return "", app.SymbolteileUnlesbar(err)
	}

	return naht.Ablegen(ctx, setID, gelesen, unter)
}

// Rechnen führt eine der Rechnungen aus — **ohne Naht und ohne Bestand**.
func (symbolwerkstatt) SymbolRechnen(was string, argumente json.RawMessage) (any, error) {
	return app.SymbolRechnen(was, argumente)
}

// Palette gibt, was ein Kartenfeld zum Einfügen anbietet — ohne die Bildchen.
//
// **Mit dem Kontext des Rufs** (#1815) — derselbe Grund wie am Dienst.
func (s symbolwerkstatt) Palette(ctx context.Context, karteID int64, feld string) (any, error) {
	werk, err := app.Zeichennaht(s.db)
	if err != nil {
		return nil, err
	}

	return werk.Palette(ctx, karteID, feld)
}
