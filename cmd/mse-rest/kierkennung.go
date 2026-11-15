package main

// **Die Kartenerkennung am Dienst** — Werkstück A13.3 von Phase 5.3
// (#1452).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// KIErkennung liest ein Kartenfoto und nennt die passenden Drucke.
func (p *preise) KIErkennung(ctx context.Context, bezug rest.KIErkennungsbezug) (
	any, error,
) {
	antwort := p.kianwendung().KIKarteErkennen(ctx,
		app.KIErkennungsbezug{
			Pfad: bezug.Pfad, Anbieter: bezug.Anbieter, Alias: bezug.Alias,
		})
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort, nil
}
