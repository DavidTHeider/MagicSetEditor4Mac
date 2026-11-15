package mcp_test

// **Das Doppel der Deckübernahme** — Werkstück 12 von Phase 4.12
// (#1351).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// uebernahmeprobe ist eine Übernahme im Arbeitsspeicher.
type uebernahmeprobe struct {
	// gesehen hält fest, wonach gefragt wurde.
	gesehen [2]string

	fehler error
}

func neueUebernahmeprobe() *uebernahmeprobe { return &uebernahmeprobe{} }

// DeckeZurUebernahme nennt die Decks von drüben — oder das Nein der Naht.
func (p *uebernahmeprobe) DeckeZurUebernahme(
	_ context.Context, quellProfil string,
) (any, error) {
	p.gesehen = [2]string{quellProfil, ""}

	if p.fehler != nil {
		return nil, p.fehler
	}

	return map[string]any{"decks": []map[string]string{
		{"id": "7", "name": "Nachtlauf"},
	}}, nil
}

func (p *uebernahmeprobe) DeckUebernehmen(
	_ context.Context, quellProfil, quellDeck string,
) (mcp.Uebernommenes, error) {
	p.gesehen = [2]string{quellProfil, quellDeck}

	if p.fehler != nil {
		return mcp.Uebernommenes{}, p.fehler
	}

	return mcp.Uebernommenes{
		DeckID: "99", Name: "Nachtlauf", Spiel: "probe", Plaetze: 2,
		Befunde: []string{fmt.Sprintf(
			"Das Format des Decks %s bleibt im anderen Profil.", quellDeck)},
	}, nil
}
