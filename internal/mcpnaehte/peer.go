package mcpnaehte

// **Die eigene Peeridentität am MCP-Server** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Nur die Auskunft.** Warum das Anlegen und das Ablösen hier nicht stehen,
// sagt der Kopf von `internal/mcp/peer.go`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// peerauskunft reicht die Identität an das Werkzeug.
type peerauskunft struct{ db *sql.DB }

func (p peerauskunft) Peeridentitaet(ctx context.Context) (mcp.Peeridentitaet, error) {
	antwort := app.PeerflaecheUeber(p.db).Peeridentitaet(ctx)
	if antwort.Fehler != nil {
		return mcp.Peeridentitaet{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Peeridentitaet{
		Stand:               antwort.Stand,
		Kennung:             antwort.Kennung,
		Fingerabdruck:       antwort.Fingerabdruck,
		Seit:                antwort.Seit,
		Vorher:              antwort.Vorher,
		VorherFingerabdruck: antwort.VorherFingerabdruck,
		Befunde:             antwort.Befunde,
	}, nil
}

// Peer gibt die Naht für den Werkzeugsatz.
func Peer(db *sql.DB) mcp.Peerauskunft { return peerauskunft{db: db} }
