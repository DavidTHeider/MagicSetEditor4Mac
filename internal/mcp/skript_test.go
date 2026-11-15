package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// konsolenprobe gibt ein festes Ergebnis — oder scheitert.
type konsolenprobe struct {
	fehler   error
	ausdruck string
	karteID  int64

	// Die zwei für `lesen.skriptnamen` (#1164).
	namen       []string
	namenskarte int64
}

func (k *konsolenprobe) Namen(_ context.Context, karteID int64) ([]string, error) {
	if k.fehler != nil {
		return nil, k.fehler
	}

	k.namenskarte = karteID

	return k.namen, nil
}

func (k *konsolenprobe) Auswerten(_ context.Context, karteID int64, ausdruck string) (
	mcp.Skriptergebnis, error,
) {
	if k.fehler != nil {
		return mcp.Skriptergebnis{}, k.fehler
	}

	k.karteID, k.ausdruck = karteID, ausdruck

	return mcp.Skriptergebnis{
		Quelltext: `"ab"`, Text: "ab", Art: "Zeichenkette", Befunde: []string{},
	}, nil
}

// rufeSkript sucht das Werkzeug und ruft es.
func rufeSkript(t *testing.T, konsole mcp.Skriptkonsole, argumente string) (string, error) {
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
		Saetze:      nil,
		Konsole:     konsole,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == "lesen.skript" {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatal("es gibt kein Werkzeug lesen.skript")

	return "", nil
}

// **Das Werkzeug wertet aus und gibt den Quelltext** —
// #1070.
//
// **Der Quelltext ist die Zusage**: ein Sprachmodell soll eine Zeichenkette
// von einem Namen unterscheiden können.
func TestDasSkriptwerkzeugWertetAus(t *testing.T) {
	t.Parallel()

	probe := &konsolenprobe{}

	antwort, fehler := rufeSkript(t, probe, `{"karteID": 7, "ausdruck": "\"a\" + \"b\""}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.karteID != 7 || probe.ausdruck != `"a" + "b"` {
		t.Errorf("weitergereicht: %d %q", probe.karteID, probe.ausdruck)
	}

	if !strings.Contains(antwort, `\"ab\"`) {
		t.Errorf("Antwort %s", antwort)
	}
}

// **Die Fehlerwege des Werkzeugs.**
func TestDieFehlerwegeDesSkriptwerkzeugs(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		konsole   mcp.Skriptkonsole
		argumente string
	}{
		"kein JSON":     {&konsolenprobe{}, `{`},
		"keine Kennung": {&konsolenprobe{}, `{"karteID": 0, "ausdruck": "1"}`},
		"die Konsole gibt auf": {
			&konsolenprobe{fehler: errors.New("kein Bestand")},
			`{"karteID": 7, "ausdruck": "1"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeSkript(t, fall.konsole, fall.argumente); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
