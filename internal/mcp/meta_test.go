package mcp_test

// **Das Metawerkzeug** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type metaprobe struct {
	antwort mcp.Meta
	fehler  error
}

func (m metaprobe) Meta(context.Context) (mcp.Meta, error) { return m.antwort, m.fehler }

func TestOhneNahtGibtEsKeinMetawerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.meta" {
			t.Fatal("das Werkzeug stand ohne Naht da")
		}
	}
}

func TestDasMetawerkzeugNenntFassungUndFlaeche(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Meta: metaprobe{antwort: mcp.Meta{
			Fassung: "0.4.0", Plattform: "darwin/arm64",
			Go: "go1.26.6", Flaeche: "mcp",
			Stand: "0123abcd", Gebaut: "2026-09-23T00:00:00Z", Veraendert: true,
			Bindungsvertrag: "a1b2c3d4e5f6",
		}},
	}, "lesen.meta")

	antwort, err := werkzeug.Ruf(t.Context(), nil)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	// **Bau und Bindungsvertrag gehen mit** — ENG-06.AK-04 (#1731): dieselben
	// Angaben wie an Kommandozeile und Dienst.
	for _, teil := range []string{
		`"fassung": "0.4.0"`, `"flaeche": "mcp"`, `"stand": "0123abcd"`,
		`"veraendert": true`, `"bindungsvertrag": "a1b2c3d4e5f6"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("Antwort %q nennt %q nicht", antwort, teil)
		}
	}
}

func TestDasMetawerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Meta: metaprobe{fehler: errors.New("nichts zu sagen")},
	}, "lesen.meta")

	if _, err := werkzeug.Ruf(t.Context(), nil); err == nil {
		t.Error("der Fehler ging verloren")
	}
}
