package main

// **Die Nachbarn im Dienst** — Werkstück N2 von Phase 5.4
// (#1455).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// nachbarweg hängt `internal/app` an [rest.Nachbarnaht].
type nachbarweg struct{ db *sql.DB }

func (n nachbarweg) Nachbarn(ctx context.Context) (rest.Nachbarn, error) {
	antwort := app.PeerflaecheUeber(n.db).Nachbarn(ctx)
	if antwort.Fehler != nil {
		return rest.Nachbarn{}, antwort.Fehler.AlsFehler()
	}

	nachbarn := make([]rest.Nachbar, 0, len(antwort.Nachbarn))

	for _, nachbar := range antwort.Nachbarn {
		nachbarn = append(nachbarn, rest.Nachbar{
			Kennung:       nachbar.Kennung,
			Fingerabdruck: nachbar.Fingerabdruck,
			Name:          nachbar.Name,
			Adresse:       nachbar.Adresse,
			Merkadresse:   nachbar.Merkadresse,
			Bekannt:       nachbar.Bekannt,
			Stand:         nachbar.Stand,

			GemerkteAdresse:  nachbar.GemerkteAdresse,
			AdresseGeaendert: nachbar.AdresseGeaendert,
		})
	}

	return rest.Nachbarn{
		Nachbarn: nachbarn, Verworfen: antwort.Verworfen,
	}, nil
}
