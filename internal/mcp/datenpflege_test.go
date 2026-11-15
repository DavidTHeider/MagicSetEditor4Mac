package mcp_test

// **Die Datenpflege über MCP** — Werkstück O8 von Phase 5.2
// (#1432).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type datenpflegeprobe struct {
	antwort mcp.Pflegevorschau
	fehler  error
}

func (p datenpflegeprobe) Pflegevorschau(context.Context) (mcp.Pflegevorschau, error) {
	return p.antwort, p.fehler
}

func TestOhneNahtGibtEsKeinPflegewerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.pflegevorschau" {
			t.Fatal("das Werkzeug stand ohne Naht da")
		}
	}
}

// **Der Befund kommt mit Klasse und Revision** — beides braucht, wer
// entscheiden soll.
func TestDasPflegewerkzeugNenntKlasseUndRevision(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Datenpflege: datenpflegeprobe{antwort: mcp.Pflegevorschau{
			Erstellt:   "2026-09-10T12:00:00Z",
			Gueltigbis: "2026-09-10T12:15:00Z",
			Befunde: []mcp.Pflegebefund{{
				Art: "moeglicheDubletten", Klasse: "nurInformieren",
				Anzahl: 5, Revision: "abc",
			}},
		}},
	}, "lesen.pflegevorschau")

	antwort, err := werkzeug.Ruf(t.Context(), nil)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	for _, teil := range []string{`"klasse": "nurInformieren"`, `"revision": "abc"`} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("Antwort %q nennt %q nicht", antwort, teil)
		}
	}
}

// **Ein Fehler kommt durch** — eine leere Vorschau sähe aus wie ein sauberer
// Bestand, und das ist die eine Auskunft, die niemals falsch sein darf.
func TestDasPflegewerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Datenpflege: datenpflegeprobe{fehler: errors.New("die Datenbank ist zu")},
	}, "lesen.pflegevorschau")

	if _, err := werkzeug.Ruf(t.Context(), nil); err == nil {
		t.Error("der Fehler ging verloren")
	}
}
