package main

// **Die eigene Peeridentität im Dienst** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine Naht und keine Fachlogik**: welcher der vier Stände gilt, entscheidet
// `internal/service`; hier wird übersetzt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// peerweg reicht die Identität an den Dienst.
type peerweg struct{ db *sql.DB }

func (p peerweg) Peeridentitaet(ctx context.Context) (rest.Peeridentitaet, error) {
	return alsRestpeer(app.PeerflaecheUeber(p.db).Peeridentitaet(ctx))
}

func (p peerweg) PeeridentitaetSicherstellen(ctx context.Context) (rest.Peeridentitaet, error) {
	return alsRestpeer(app.PeerflaecheUeber(p.db).PeeridentitaetSicherstellen(ctx))
}

func (p peerweg) PeeridentitaetRotieren(ctx context.Context) (rest.Peeridentitaet, error) {
	return alsRestpeer(app.PeerflaecheUeber(p.db).PeeridentitaetRotieren(ctx))
}

// alsRestpeer macht aus der Antwort der Anwendung die des Dienstes.
func alsRestpeer(antwort app.PeeridentitaetAntwort) (rest.Peeridentitaet, error) {
	if antwort.Fehler != nil {
		return rest.Peeridentitaet{}, antwort.Fehler.AlsFehler()
	}

	return rest.Peeridentitaet{
		Stand:               antwort.Stand,
		Kennung:             antwort.Kennung,
		Fingerabdruck:       antwort.Fingerabdruck,
		Seit:                antwort.Seit,
		Vorher:              antwort.Vorher,
		VorherFingerabdruck: antwort.VorherFingerabdruck,
		Befunde:             antwort.Befunde,
	}, nil
}
