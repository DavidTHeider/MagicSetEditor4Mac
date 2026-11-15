// **Plan, Ersetzen und Zurück für den MCP-Server** — DOM-01.AK-02 und AK-03
// (#1696).
//
// Gerechnet wird in `service`, übersetzt in `app` ([app.Paketweg.Plan]);
// hier wird nur die Gestalt des Plans in die von `internal/mcp` gebracht.

package mcpnaehte

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// PaketPlan sagt vor einer Änderung, was sie bewirkt.
func (p pflege) PaketPlan(ctx context.Context, vorgang, art, name, quelle string) (mcp.Paketplan, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return mcp.Paketplan{}, err
	}

	plan, fehler := weg.Plan(ctx, vorgang, art, name, quelle)
	if fehler != nil {
		return mcp.Paketplan{}, fehler
	}

	return mcp.Paketplan{
		Vorgang: plan.Vorgang, Paket: plan.Paket, Quelle: plan.Quelle, Von: plan.Von, Zu: plan.Zu,
		Geaendert: plan.Geaendert, Moeglich: plan.Moeglich, Sets: plan.Sets,
		Konflikte: alsKonflikte(plan.Konflikte),
	}, nil
}

// PaketErsetzen legt eine neue Fassung über die vorhandene.
func (p pflege) PaketErsetzen(ctx context.Context, quelle string, bestaetigt bool) (mcp.Paketzeile, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return mcp.Paketzeile{}, err
	}

	paket, fehler := weg.Ersetzen(ctx, quelle, bestaetigt)
	if fehler != nil {
		return mcp.Paketzeile{}, fehler
	}

	return alsPaketzeile(paket), nil
}

// PaketZurueck tauscht die aktuelle Fassung gegen die vorige.
func (p pflege) PaketZurueck(ctx context.Context, art, name string) (mcp.Paketzeile, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return mcp.Paketzeile{}, err
	}

	paket, fehler := weg.Zurueck(ctx, art, name)
	if fehler != nil {
		return mcp.Paketzeile{}, fehler
	}

	return alsPaketzeile(paket), nil
}

// PaketBedarf nennt, was einem Set an Paketen fehlt — DOM-01.AK-04.
func (p pflege) PaketBedarf(ctx context.Context, setID int64) ([]mcp.Paketkonflikt, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	befunde, fehler := weg.Setbedarf(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	return alsKonflikte(befunde), nil
}

// alsKonflikte übersetzt die Befunde der Anwendung — **eine Stelle** für den
// Plan und den Bedarf.
func alsKonflikte(befunde []app.AbhaengigkeitDTO) []mcp.Paketkonflikt {
	konflikte := make([]mcp.Paketkonflikt, 0, len(befunde))
	for _, befund := range befunde {
		konflikte = append(konflikte, mcp.Paketkonflikt{
			Paket: befund.Paket, Braucht: befund.Braucht, Art: befund.Art,
			Verlangt: befund.Verlangt, Da: befund.Da,
		})
	}

	return konflikte
}

// SetVorschau sagt, was das Einfahren einer Datei täte — FMT-03.AK-01
// (#1692), über denselben Bestand, in den `schreiben.set.einfahren` schriebe.
func (p pflege) SetVorschau(ctx context.Context, pfad string) (any, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	return weg.Setvorschau(ctx, p.db, pfad)
}
