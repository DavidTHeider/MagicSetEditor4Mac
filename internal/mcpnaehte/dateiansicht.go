// **Eine Datei ansehen am Werkzeugserver** — UX-13
// (#1685).

package mcpnaehte

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// dateiansicht erfüllt `mcp.Dateiansicht` über `app` — ohne Bilddaten.
type dateiansicht struct{}

func (dateiansicht) DateiAnsehen(_ context.Context, pfad string) (any, error) {
	antwort := app.DateiansichtAus(pfad)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
