package app

// **Die Peeridentität an den Flächen ohne Fenster** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Wofür es diesen Bauweg gibt
//
// **Weil Kommandozeile, MCP und Dienst je Ruf eine Anwendung bauen.** Sie
// haben kein Fenster, das eine hält, und ohne diesen Weg stünde die
// Identität dort leer da, während sie am Fenster steht — dieselbe Frage,
// zwei Antworten (§16).
//
// **Und er ist klein**: die Identität braucht den Bestand und den
// Schlüsselbund, sonst nichts. Wer ihn über [KIAnwendungUeber] baute, zöge
// vier KI-Anbieter und einen Suchindex mit herein, die kein Ruf hier
// anfasst.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// PeerflaecheUeber baut eine Anwendung, die nach der eigenen Identität
// gefragt werden kann.
func PeerflaecheUeber(db *sql.DB) *App {
	anwendung := NeueApp("")
	anwendung.db = db

	// **Mit den Einstellungen** — sonst gilt am Fernweg der Kommandozeile
	// kein eingestellter Proxy (Befund 1 des Zweitblicks über Phase D).
	anwendung.einstellungen = einstellungsverwalterUeber(db)

	return anwendung
}

// Ausweis gibt die Identität dieses Profils — oder sagt, warum es keine gibt.
//
// **Exportiert seit N6** (#1572):
// `mse-rest` braucht sie für sein TLS-Zeugnis, und eine zweite Naht dorthin
// wäre ein zweiter Weg zu derselben Identität — genau das, was diese Fläche
// verhindern soll.
func (a *App) Ausweis(ctx context.Context) (peer.Identitaet, error) {
	ctx = a.imProfil(ctx)

	return a.peerwerk().Ausweis(ctx)
}
