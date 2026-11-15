// **Eine Datei ansehen am Dienst** — UX-13
// (#1685).

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// dateiansicht erfüllt `rest.Dateiansicht` über `app` — ohne Bilddaten.
type dateiansicht struct{}

func (dateiansicht) DateiAnsehen(_ context.Context, pfad string) (any, error) {
	antwort := app.DateiansichtAus(pfad)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
