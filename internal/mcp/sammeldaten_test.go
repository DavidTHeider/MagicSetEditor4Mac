package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die beiden Sammeldatenwerkzeuge** —
// #888.
func TestSammeldatenwerkzeuge(t *testing.T) {
	t.Parallel()

	rufe := func(t *testing.T, name string, daten mcp.Fremdsammeldaten) (string, error) {
		t.Helper()

		for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
			Leser:       probenbestand(),
			Schreiber:   probenbestand(),
			Ausfuhr:     ruhendeAusfuhr{},
			Abgleich:    nil,
			Bilder:      nil,
			Auskunft:    stilleAuskunft{},
			Sammeldaten: daten,
			Anlage:      nil,
			Saetze:      nil,
			Konsole:     nil,
			Wortlisten:  nil,
			Woerter:     nil,
			Pack:        nil,
			Anleger:     nil,
			Angaben:     neueAngabenprobe(),
			Stile:       neueStilprobe(),
			Bestand:     neueBestandsprobeWeit(),
		}) {
			if werkzeug.Name == name {
				return werkzeug.Ruf(t.Context(), json.RawMessage(`{}`))
			}
		}

		t.Fatalf("es gibt kein Werkzeug %q", name)

		return "", nil
	}

	t.Run("das lesende sagt, was liegt", func(t *testing.T) {
		t.Parallel()

		text, fehler := rufe(t, "lesen.scryfall.sammeldaten", stilleSammeldaten{})
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		for _, teil := range []string{"oracle_cards", "Vorhanden", "24531600"} {
			if !strings.Contains(text, teil) {
				t.Errorf("die Antwort führt %q nicht: %s", teil, text)
			}
		}
	})

	t.Run("das schreibende lädt und meldet den Stand", func(t *testing.T) {
		t.Parallel()

		text, fehler := rufe(t, "schreiben.scryfall.sammeldaten", stilleSammeldaten{})
		if fehler != nil {
			t.Fatalf("Ruf: %v", fehler)
		}

		if !strings.Contains(text, "oracle_cards") {
			t.Errorf("die Antwort führt die Art nicht: %s", text)
		}
	})

	t.Run("beide reichen einen Fehler durch", func(t *testing.T) {
		t.Parallel()

		kaputt := stilleSammeldaten{fehler: domain.ErrNotFound}

		for _, name := range []string{
			"lesen.scryfall.sammeldaten", "schreiben.scryfall.sammeldaten",
		} {
			if _, fehler := rufe(t, name, kaputt); !errors.Is(fehler, domain.ErrNotFound) {
				t.Errorf("%s: Fehler = %v", name, fehler)
			}
		}
	})
}
