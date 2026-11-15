package mcp_test

// **Der Platz über MCP** — Werkstück 4 von Phase 4.13
// (#1352).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// platzprobe merkt sich, was gerufen wurde.
type platzprobe struct {
	geraeumt  string
	fehler    error
	nachladen *mcp.Nachladefehler
}

func (p *platzprobe) Platz(context.Context) ([]mcp.Platzposten, error) {
	if p.fehler != nil {
		return nil, p.fehler
	}

	return []mcp.Platzposten{
		{Posten: "bestand", Dateien: 1, Bytes: 127_000_000},
		{Posten: "vorsicherungen", Dateien: 3, Bytes: 348_000_000, Raeumbar: true},
	}, nil
}

func (p *platzprobe) PlatzFreigeben(ctx context.Context, posten string) (
	mcp.Platzgeraeumt, error,
) {
	if p.fehler != nil {
		return mcp.Platzgeraeumt{}, p.fehler
	}

	p.geraeumt = posten

	if p.nachladen != nil {
		return mcp.Platzgeraeumt{Posten: []mcp.Platzposten{}, Nachladefehler: p.nachladen}, nil
	}

	gezaehlt, fehler := p.Platz(ctx)

	return mcp.Platzgeraeumt{Posten: gezaehlt}, fehler
}

// rufePlatz sucht ein Werkzeug und ruft es.
func rufePlatz(t *testing.T, name string, probe *platzprobe, argumente string) (
	string, error,
) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:     probenbestand(),
		Schreiber: probenbestand(),
		Ausfuhr:   ruhendeAusfuhr{},
		Auskunft:  stilleAuskunft{},
		Angaben:   neueAngabenprobe(),
		Stile:     neueStilprobe(),
		Bestand:   neueBestandsprobeWeit(),
		Platz:     probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("es gibt kein Werkzeug %q", name)

	return "", nil
}

// **Das lesende sagt, was liegt.**
func TestDasLesendePlatzwerkzeugZaehlt(t *testing.T) {
	t.Parallel()

	text, fehler := rufePlatz(t, "lesen.platz", &platzprobe{}, `{}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	for _, teil := range []string{"vorsicherungen", "348000000", "Raeumbar"} {
		if !strings.Contains(text, teil) {
			t.Errorf("die Antwort führt %q nicht: %s", teil, text)
		}
	}
}

// **Das schreibende räumt weg und gibt die Zählung danach.**
func TestDasSchreibendePlatzwerkzeugRaeumtWeg(t *testing.T) {
	t.Parallel()

	probe := &platzprobe{}

	text, fehler := rufePlatz(t, "schreiben.platz.freigeben", probe,
		`{"posten": "vorsicherungen"}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.geraeumt != "vorsicherungen" {
		t.Errorf("weggeräumt wurde %q", probe.geraeumt)
	}

	if !strings.Contains(text, "bestand") {
		t.Errorf("die Antwort ist nicht die neue Zählung: %s", text)
	}
}

// **Und beide reichen ihren Fehler durch** (§8).
func TestDiePlatzwerkzeugeReichenIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	klemmt := domain.ErrValidation

	for name, argumente := range map[string]string{
		"lesen.platz":               `{}`,
		"schreiben.platz.freigeben": `{"posten": "vorsicherungen"}`,
	} {
		if _, fehler := rufePlatz(t, name, &platzprobe{fehler: klemmt},
			argumente); !errors.Is(fehler, klemmt) {
			t.Errorf("%s gab %v", name, fehler)
		}
	}
}

// **Und ein Argument, das kein JSON ist, wird abgewiesen.**
func TestEinKaputtesArgumentAmPlatzWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := rufePlatz(t, "schreiben.platz.freigeben", &platzprobe{},
		`{"posten": 42}`); fehler == nil {
		t.Error("eine Zahl als Posten wurde angenommen")
	}
}
