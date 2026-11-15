package mcpnaehte

import (
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Befehle sind die Werkzeuge des MCP-Servers, wie die Kommandozentrale im
// Fenster sie ruft — Werkstück U12 von Phase 5.1 (#1423). **Dieselbe Liste,
// dieselben Rufe**: was sich hier ändert, ändert sich für das Modell mit.
//
// Der Urheber ist das Fenster: was die Kommandozentrale schreibt, steht im
// Verlauf als Arbeit am Fenster, nicht als die eines Servers.
func Befehle(db *sql.DB, pfad string) ([]app.Befehl, error) {
	schreiber, err := app.Schreibnaht(db, domain.UrheberFenster)
	if err != nil {
		return nil, err
	}

	werkzeuge := mcp.Werkzeuge(Alle(db, db, pfad, storage.NeuerSetLeser(db), schreiber))
	befehle := make([]app.Befehl, 0, len(werkzeuge))

	for _, werkzeug := range werkzeuge {
		befehle = append(befehle, app.Befehl{
			Name:       werkzeug.Name,
			Schreibend: werkzeug.Art == mcp.Schreibend,
			Idempotent: werkzeug.Idempotent,
			Erklaerung: werkzeug.Erklaerung,
			Schema:     werkzeug.Schema,
			Ruf:        werkzeug.Ruf,
		})
	}

	return befehle, nil
}
