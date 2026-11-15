package main

// **Der Poststand im Dienst** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// **Eine Naht und keine Fachlogik**: was ein Stand ist, entscheidet der
// Bestand; hier wird übersetzt — Feld für Feld und nicht über `Marshal`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// postweg reicht den Poststand an den Dienst.
type postweg struct{ db *sql.DB }

func (p postweg) Handelspoststand(
	ctx context.Context, kennung string,
) (rest.Poststand, error) {
	antwort := app.PeerflaecheUeber(p.db).Handelspoststand(ctx, kennung)
	if antwort.Fehler != nil {
		return rest.Poststand{}, antwort.Fehler.AlsFehler()
	}

	return rest.Poststand{
		Kennung: antwort.Stand.Kennung, Offen: antwort.Stand.Offen,
		HoechsteVersuche: antwort.Stand.HoechsteVersuche,
	}, nil
}
