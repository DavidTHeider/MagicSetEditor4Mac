// **Der Netzweg am Dienst** —
// #1571.

package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// netzweg erfüllt `rest.Netzweg` über der Naht von `app`.
type netzweg struct{ db *sql.DB }

func (n netzweg) Netzstand(ctx context.Context) (rest.Netzstandangaben, error) {
	antwort := app.NetzflaecheUeber(n.db).Netzstand(ctx)
	if antwort.Fehler != nil {
		return rest.Netzstandangaben{}, antwort.Fehler.AlsFehler()
	}

	return rest.Netzstandangaben{
		ProxyHTTP:       netzfeld(antwort.ProxyHTTP),
		ProxyHTTPS:      netzfeld(antwort.ProxyHTTPS),
		ProxyAus:        netzfeld(antwort.ProxyAus),
		FristVerbindung: antwort.FristVerbindung,
		FristKopf:       antwort.FristKopf,
		FristStillstand: antwort.FristStillstand,
		ProxyBenutzt:    antwort.ProxyBenutzt,
	}, nil
}

// netzfeld übersetzt ein Feld samt Herkunft.
func netzfeld(feld app.NetzfeldDTO) rest.Netzfeld {
	return rest.Netzfeld{Wert: feld.Wert, Herkunft: feld.Herkunft}
}
