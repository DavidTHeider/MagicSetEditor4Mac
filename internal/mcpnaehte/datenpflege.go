package mcpnaehte

// **Die Datenpflege über MCP** — Werkstück O8 von Phase 5.2
// (#1432).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// pflegeauskunft reicht die Vorschau an das Werkzeug.
type pflegeauskunft struct {
	db  *sql.DB
	ort string
}

func (p pflegeauskunft) Pflegevorschau(ctx context.Context) (mcp.Pflegevorschau, error) {
	antwort := app.PflegeAnwendungUeber(p.db, p.ort).Pflegevorschau(ctx)
	if antwort.Fehler != nil {
		return mcp.Pflegevorschau{}, antwort.Fehler.AlsFehler()
	}

	vorschau := mcp.Pflegevorschau{
		Erstellt: antwort.Erstellt, Gueltigbis: antwort.Gueltigbis,
		Befunde: make([]mcp.Pflegebefund, 0, len(antwort.Befunde)),
	}

	for _, befund := range antwort.Befunde {
		vorschau.Befunde = append(vorschau.Befunde, mcp.Pflegebefund{
			Art: befund.Art, Klasse: befund.Klasse, Anzahl: befund.Anzahl,
			Revision: befund.Revision, Beispiele: befund.Beispiele,
		})
	}

	return vorschau, nil
}
