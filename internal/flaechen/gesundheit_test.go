package flaechen_test

// **Der Betriebszustand über Kommandozeile, MCP und Dienst** — Werkstück O9
// von Phase 5.2 (#1433).
//
// # Was er hält
//
// **Dieselbe Aussage in drei Gestalten**: ein Rückgabewert, ein JSON-Objekt
// und ein Statuscode. Wer sie auseinanderlaufen ließe, hätte drei Antworten
// auf eine Frage (§16).
//
// # Was er nicht hält
//
// **Das Fenster.** Er fährt es nicht (gemessen in `FLAECHENPARITAET.md`) —
// dieselbe Grenze wie bei `pack_test.go`, und sie steht hier, damit der Kopf
// nicht mehr verspricht als der Rumpf hält. Die Bindung prüft
// `internal/app/gesundheit_test.go`, samt ihrer Antwort.

import (
	"net/http"
	"strings"
	"testing"
)

// **Ein offener Bestand ist an allen drei Flächen bereit.**
func TestDerBetriebszustandGehtUeberDreiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		aus, klage, code := kommandozeile(t, umgebung, "gesundheit")

		switch {
		case code != 0:
			t.Fatalf("gesundheit gab %d (%s)", code, klage)
		case !strings.Contains(aus, "bereit\ttrue"):
			t.Errorf("Ausgabe %q sagt nicht, dass er bereit ist", aus)
		}

		// **Und als JSON dasselbe** — samt der leeren Fremdenliste (§3).
		roh, _, code := kommandozeile(t, umgebung, "gesundheit", "--json")

		switch {
		case code != 0:
			t.Fatalf("gesundheit --json gab %d", code)
		case !strings.Contains(roh, `"bereit": true`):
			t.Errorf("JSON %q", roh)
		case !strings.Contains(roh, `"fremde": []`):
			t.Errorf("die Fremdenliste war nicht leer, sondern: %q", roh)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		antwort := mussMCP(t, umgebung, "lesen.gesundheit", map[string]any{})

		if !strings.Contains(antwort, `"bereit": true`) {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		// **Der Dienst antwortet kompakt, die anderen eingerückt** — geprüft
		// wird die Aussage, nicht die Formatierung.
		if antwort := mussREST(t, an, http.MethodGet, "/gesundheit", ""); !strings.Contains(
			antwort, `"bereit":true`) {
			t.Errorf("Antwort %q", antwort)
		}

		// **Und `/bereit` sagt dasselbe im Statuscode** — die Form, die eine
		// Sonde liest.
		if code, _ := an.ruf(t, http.MethodGet, "/bereit", ""); code != http.StatusOK {
			t.Errorf("/bereit gab %d, erwartet 200", code)
		}
	})
}
