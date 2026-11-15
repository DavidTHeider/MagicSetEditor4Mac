package mcp_test

// **Das Gesundheitswerkzeug** — Werkstück O9 von Phase 5.2
// (#1433).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// gesundheitsprobe antwortet, wie ihr gesagt wird.
type gesundheitsprobe struct {
	antwort mcp.Gesundheit
	fehler  error
}

func (g gesundheitsprobe) Gesundheit(context.Context) (mcp.Gesundheit, error) {
	return g.antwort, g.fehler
}

// **Ohne Naht gibt es das Werkzeug nicht** — eine Fläche, die nichts kann,
// soll auch nichts anbieten.
func TestOhneNahtGibtEsKeinGesundheitswerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.gesundheit" {
			t.Fatal("das Werkzeug stand ohne Naht da")
		}
	}
}

// **Mit Naht antwortet es — und trägt den fremden Ausfall mit.**
func TestDasGesundheitswerkzeugAntwortet(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Gesundheit: gesundheitsprobe{antwort: mcp.Gesundheit{
			Lebt: true, Bereit: true,
			Fremde: []mcp.Fremdzustand{
				{Dienst: "scryfall", Erreichbar: false, Grund: "kein Netz"},
			},
		}},
	}, "lesen.gesundheit")

	antwort, err := werkzeug.Ruf(t.Context(), nil)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	switch {
	case !strings.Contains(antwort, `"bereit": true`):
		t.Errorf("Antwort %q", antwort)
	case !strings.Contains(antwort, `"scryfall"`):
		t.Errorf("der fremde Dienst fehlt: %q", antwort)
	}
}

// **Und ein Fehler der Naht kommt als Fehler zurück.**
func TestDasGesundheitswerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Gesundheit: gesundheitsprobe{fehler: errors.New("nichts zu holen")},
	}, "lesen.gesundheit")

	if _, err := werkzeug.Ruf(t.Context(), nil); err == nil {
		t.Error("der Fehler ging verloren")
	}
}

// gesundheitswerkzeug sucht ein Werkzeug — und bricht ab, wenn es fehlt.
func gesundheitswerkzeug(t *testing.T, naehte mcp.Naehte, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(naehte) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}
