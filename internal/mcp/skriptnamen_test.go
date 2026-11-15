// **Die Namen der Umgebung, für ein Sprachmodell** — Werkstück 1 von
// Phase 3.3 (#1164).
//
// # Warum das Werkzeug nicht dem Fenster vorbehalten bleibt
//
// Hier stand als Begründung, ein Sprachmodell habe „einen eigenen Editor mit
// eigener Vervollständigung". Das ist bei MSE-Script gerade **nicht** so:
// welche Funktionen eine Vorlage dazulegt, steht in keinem Trainingsstand —
// es steht im Paket auf diesem Rechner. §16 verlangt für eine Fähigkeit an
// nur einer Fläche eine Begründung, und §10 sagt, dass auch eine Ablehnung
// Deckung braucht. Diese hatte keine.
package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// rufeNamen sucht das Werkzeug und ruft es.
func rufeNamen(t *testing.T, konsole mcp.Skriptkonsole, argumente string) (string, error) {
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
		if werkzeug.Name == "lesen.skriptnamen" {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatal("es gibt kein Werkzeug lesen.skriptnamen")

	return "", nil
}

// **Das Werkzeug zählt auf, was die Umgebung kennt.**
func TestDasNamenswerkzeugZaehltAuf(t *testing.T) {
	t.Parallel()

	probe := &konsolenprobe{namen: []string{"card", "length", "set"}}

	antwort, fehler := rufeNamen(t, probe, `{"karteID": 7}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.namenskarte != 7 {
		t.Errorf("weitergereicht: %d", probe.namenskarte)
	}

	for _, name := range []string{"card", "length", "set"} {
		if !strings.Contains(antwort, name) {
			t.Errorf("%q fehlt in %s", name, antwort)
		}
	}
}

// **Die Fehlerwege des Werkzeugs.**
func TestDieFehlerwegeDesNamenswerkzeugs(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		konsole   mcp.Skriptkonsole
		argumente string
	}{
		"kein JSON":     {&konsolenprobe{}, `{`},
		"keine Kennung": {&konsolenprobe{}, `{"karteID": 0}`},
		"die Konsole gibt auf": {
			&konsolenprobe{fehler: errors.New("kein Bestand")},
			`{"karteID": 7}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeNamen(t, fall.konsole, fall.argumente); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}

// **Eine Umgebung ohne Namen gibt eine leere Liste, kein `null`.**
//
// Dieselbe Überlegung wie am Fenster: wer die Antwort liest, soll auf **eines**
// prüfen müssen.
func TestEineLeereUmgebungGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	antwort, fehler := rufeNamen(t, &konsolenprobe{namen: nil}, `{"karteID": 7}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if strings.Contains(antwort, "null") {
		t.Errorf("Antwort %s", antwort)
	}
}
