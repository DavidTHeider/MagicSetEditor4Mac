package mcpnaehte

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Der Platz über MCP** — Werkstück 4 von Phase 4.13
// (#1352).
//
// **Über den Ort und nicht über die Verbindung**: gezählt werden Dateien.
type platz struct{ ort string }

func (p platz) Platz(ctx context.Context) ([]mcp.Platzposten, error) {
	return alsPlatzposten(app.PlatzflaecheUeber(p.ort).Platz(ctx))
}

func (p platz) PlatzFreigeben(ctx context.Context, posten string) (
	mcp.Platzgeraeumt, error,
) {
	antwort := app.PlatzflaecheUeber(p.ort).PlatzFreigeben(ctx, posten)

	gezaehlt, fehler := alsPlatzposten(antwort)
	if fehler != nil {
		return mcp.Platzgeraeumt{}, fehler
	}

	return mcp.Platzgeraeumt{
		Posten: gezaehlt, Nachladefehler: alsNachladefehler(antwort.Nachladefehler),
	}, nil
}

// alsPlatzposten übersetzt die Antwort der Fassade.
func alsPlatzposten(antwort app.PlatzAntwort) ([]mcp.Platzposten, error) {
	if antwort.Fehler != nil {
		return nil, app.AlsFehler(antwort.Fehler)
	}

	posten := make([]mcp.Platzposten, 0, len(antwort.Posten))
	for _, einer := range antwort.Posten {
		posten = append(posten, mcp.Platzposten{
			Posten: einer.Posten, Dateien: einer.Dateien,
			Bytes: einer.Bytes, Raeumbar: einer.Raeumbar,
		})
	}

	return posten, nil
}
