package main

// **Die Wege nach draußen im Dienst** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// **Eine Naht und keine Fachlogik**: welche Adresse zu welcher Karte gehört,
// entscheidet die Domäne; hier wird übersetzt — Feld für Feld.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// linkweg reicht die Wege an den Dienst.
type linkweg struct{ db *sql.DB }

func (l linkweg) Marktlinks(
	ctx context.Context, printing, name string,
) ([]rest.Marktlink, error) {
	antwort := app.PeerflaecheUeber(l.db).Marktlinks(ctx, printing, name)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	links := make([]rest.Marktlink, 0, len(antwort.Links))

	for _, link := range antwort.Links {
		links = append(links, rest.Marktlink{
			Anbieter: link.Anbieter, Art: link.Art,
			Herkunft: link.Herkunft, Adresse: link.Adresse,
		})
	}

	return links, nil
}
