package mcp_test

// **Turnierdatei und Bericht über MCP** — DOM-04.AK-03
// (#1699).
//
// Die drei Rufe der Probe stehen hier, damit `turniere_test.go` unter der
// Zeilengrenze aus §2 bleibt; geprüft werden sie in dessen Tabelle.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func (p *turnierprobe) TurnierAusgabe(_ context.Context, turnierID int64) (any, error) {
	p.turnierID = turnierID

	return p.antwort("TurnierAusgabe")
}

func (p *turnierprobe) TurnierBericht(_ context.Context, turnierID int64) (any, error) {
	p.turnierID = turnierID

	return p.antwort("TurnierBericht")
}

func (p *turnierprobe) TurnierEinlesen(_ context.Context, inhalt string) (any, error) {
	p.inhalt = inhalt

	return p.antwort("TurnierEinlesen")
}

// **Ohne Text gibt es nichts einzulesen** — und die Naht wird gar nicht
// gefragt.
func TestEinlesenOhneTextIstEinNein(t *testing.T) {
	t.Parallel()

	probe := &turnierprobe{}

	_, err := rufeTurnierwerkzeug(t, probe, "schreiben.turnier.einlesen", `{"inhalt":""}`)
	if !errors.Is(err, domain.ErrValidation) || probe.gerufen != "" {
		t.Errorf("Fehler %v, gerufen %q", err, probe.gerufen)
	}
}
