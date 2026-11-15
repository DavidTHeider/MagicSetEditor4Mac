package app

// **Die Sicht eines Gegenübers über die Leitung anfragen** — P11
// (#2379),
// Befund C-15: was ein Gegenüber zum Handeln freigibt, kam bisher nur als
// Datei oder eingefügtes JSON herein.
//
// # Dieselbe Leitung wie die Zustellung
//
// **`externSenden`, eine Peerkennung, eine Sitzung mit dem Zweck `handel`** —
// wie [App.Handelszustellung]: die Anfrage geht an einen fremden Rechner,
// und was hinausgeht, ist mit der Identität dieses Profils unterschrieben.
// Die Antwort wird **nicht gespeichert**; sie ist eine Auskunft für diesen
// Blick.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// HandelSichtAnfragen fragt ein Gegenüber nach dem, was es diesem Profil zum
// Handeln freigibt.
func (a *App) HandelSichtAnfragen(ctx context.Context, kennung string) FreigabesichtAntwort {
	ctx = a.imProfil(ctx)

	nein := func(fehler error) FreigabesichtAntwort {
		return FreigabesichtAntwort{Zeilen: []SichtkarteDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	if fehler := a.darf(ctx, domain.WirkungExternSenden); fehler != nil {
		return nein(fehler)
	}

	if peer.Fingerabdruck(peer.Kennung(kennung)) == "" {
		return nein(domain.MitGrund(domain.ErrValidation, nil, "%q ist keine Peerkennung", kennung))
	}

	griff, fehler := a.sitzungsverwaltung().Oeffne(ctx, peer.Kennung(kennung), domain.ZweckHandel)
	if fehler != nil {
		return nein(fehler)
	}

	defer schliesseHandelssitzung(ctx, griff)

	sicht, fehler := service.NeuerSichtfrager(a.peerwerk()).Fragen(ctx, handelsstrecke{griff: griff}, kennung)
	if fehler != nil {
		return nein(fehler)
	}

	return FreigabesichtAntwort{Zeilen: alsSichtkarten(sicht)}
}
