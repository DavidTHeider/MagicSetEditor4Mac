// Die Probleme eines ganzen Sets für den Dienst — PERF02
// (#2352).
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// probleme führt die zwei Endpunkte an den Problemweg. **Der Weg bleibt je
// Bestand** ([app.ProblemwegUeber]), und mit ihm der letzte Stand: was ein
// `POST` prüfte, nennt das nächste `GET`.
type probleme struct{ db *sql.DB }

func (p probleme) Problemstand(ctx context.Context, setID int64) (rest.Problemstand, error) {
	weg, err := app.ProblemwegUeber(p.db)
	if err != nil {
		return rest.Problemstand{}, err
	}

	stand, err := weg.Stand(ctx, setID)

	return problemstandFuerREST(stand), err
}

func (p probleme) ProblemePruefen(ctx context.Context, setID int64) (rest.Problemstand, error) {
	weg, err := app.ProblemwegUeber(p.db)
	if err != nil {
		return rest.Problemstand{}, err
	}

	stand, err := weg.Pruefen(ctx, setID, nil)

	return problemstandFuerREST(stand), err
}

// problemstandFuerREST übersetzt den Stand in die Gestalt des Dienstes.
func problemstandFuerREST(stand app.ProblemstandDTO) rest.Problemstand {
	karten := make([]rest.Kartenproblem, 0, len(stand.Karten))
	for _, karte := range stand.Karten {
		karten = append(karten, rest.Kartenproblem{KarteID: karte.KarteID, Befunde: karte.Befunde})
	}

	return rest.Problemstand{
		Urteil: stand.Urteil, Zahl: stand.Zahl, Sicher: stand.Sicher, Geprueft: stand.Geprueft,
		Gesamt: stand.Gesamt, Beendet: stand.Beendet, Alter: stand.Alter, Karten: karten,
	}
}
