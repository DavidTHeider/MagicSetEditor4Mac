package mcpnaehte

// **Der Verlauf einer Partie an der MCP-Naht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum er eine eigene Datei bekommt
//
// **Weil er eine eigene Frage beantwortet** (§2): `partie.go` übersetzt, was
// eine Partie **tut**, hier steht, was sie **war**.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// PartieVerlauf gibt den Ereignisstrom, wie ein Sitz ihn sehen darf.
//
// **Gefiltert hat `service`** — diese Naht übersetzt und prüft nichts nach.
func (p partien) PartieVerlauf(ctx context.Context, kennung, fuer string) (
	[]mcp.Partieereignis, error,
) {
	antwort := app.PartieflaecheUeber(p.db).PartieVerlauf(ctx, kennung, fuer)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return ereignisseAus(antwort.Verlauf), nil
}

// ereignisseAus übersetzt Ereignisse.
//
// **Eine Stelle für beide Wege, die welche geben** — der Befehl und die
// Rückschau. Zwei Kopien liefen beim ersten neuen Feld auseinander, und dann
// zeigte die eine Fläche etwas, das die andere verschwiege.
func ereignisseAus(strom []app.EreignisDTO) []mcp.Partieereignis {
	ereignisse := make([]mcp.Partieereignis, 0, len(strom))

	for _, ereignis := range strom {
		ereignisse = append(ereignisse, mcp.Partieereignis{
			Folge:       ereignis.Folge,
			Art:         ereignis.Art,
			Oeffentlich: ereignis.Oeffentlich,
			Privat:      ereignis.Privat,
			Kette:       ereignis.Kette,
		})
	}

	return ereignisse
}
