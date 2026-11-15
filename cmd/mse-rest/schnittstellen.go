package main

// **Die Schnittstellen über den Dienst** — OPS-09 (#1907), nur lesend.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// schnittstellen liest über eine Anwendung ohne Bestand — der Stand steht im
// Zuhause dieses Prozesses.
type schnittstellen struct{}

func (schnittstellen) Schnittstellen(ctx context.Context) ([]rest.Schnittstelle, error) {
	antwort := app.NeueApp("").Schnittstellen(ctx)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	stand := make([]rest.Schnittstelle, 0, len(antwort.Schnittstellen))
	for _, eine := range antwort.Schnittstellen {
		stand = append(stand, rest.Schnittstelle(eine))
	}

	return stand, nil
}
