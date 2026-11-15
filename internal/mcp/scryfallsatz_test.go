package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// satzprobe erfüllt `mcp.Fremdsaetze`.
type satzprobe struct {
	code   string
	spiel  string
	fehler error
}

func (s *satzprobe) Einlesen(_ context.Context, code, spiel string) (mcp.Satz, error) {
	s.code, s.spiel = code, spiel

	if s.fehler != nil {
		return mcp.Satz{}, s.fehler
	}

	return mcp.Satz{SetID: 9, Karten: 103, Befunde: []string{}}, nil
}

// rufeSatz sucht das Werkzeug und ruft es.
func rufeSatz(t *testing.T, saetze mcp.Fremdsaetze, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      saetze,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == "schreiben.scryfall.set" {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatal("es gibt kein Werkzeug schreiben.scryfall.set")

	return "", nil
}

// **Das Werkzeug liest einen Satz ein und meldet die Setkennung** —
// #1033.
func TestDasSatzwerkzeugLiestEin(t *testing.T) {
	probe := &satzprobe{}

	antwort, fehler := rufeSatz(t, probe, `{"code": "lea", "spiel": "magic"}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.code != "lea" || probe.spiel != "magic" {
		t.Errorf("die Naht bekam %q/%q", probe.code, probe.spiel)
	}

	if !strings.Contains(antwort, "103") {
		t.Errorf("Antwort = %s — die Kartenzahl fehlt", antwort)
	}
}

// **Ein unleserliches Argument wird abgewiesen.**
func TestDasSatzwerkzeugWeistUnleserlicheArgumenteAb(t *testing.T) {
	if _, fehler := rufeSatz(t, &satzprobe{}, `{"code": 7}`); fehler == nil {
		t.Fatal("kein Fehler")
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestDasSatzwerkzeugReichtSeinenFehlerDurch(t *testing.T) {
	probe := &satzprobe{fehler: errors.New("Scryfall antwortet nicht")}

	if _, fehler := rufeSatz(t, probe, `{"code": "lea", "spiel": "magic"}`); fehler == nil {
		t.Fatal("kein Fehler")
	}
}
