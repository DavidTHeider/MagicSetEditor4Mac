package flaechen_test

// **Wer antwortet hier — über drei Flächen** — Werkstück O9 von Phase 5.2
// (#1433).
//
// **Was er hält**: dieselbe Fassung an jeder Fläche — und jede nennt sich
// selbst. Wer zwei Prozesse an einem Bestand betreibt, sieht sonst nicht,
// welchen er gerade fragt.
//
// **Was er nicht hält**: das Fenster. Es hat diese Fähigkeit nicht als eigenen
// Weg — der Grund steht in `docs/FAEHIGKEITEN.md`: dort steht dasselbe im
// Diagnosetext, und eine zweite Bindung daneben wäre eine ohne Aufrufer.

import (
	"net/http"
	"strings"
	"testing"
)

func TestJedeFlaecheNenntSichSelbstUndDieselbeFassung(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		aus, klage, code := kommandozeile(t, deckumgebung(t), "meta")

		switch {
		case code != 0:
			t.Fatalf("meta gab %d (%s)", code, klage)
		case !strings.Contains(aus, "flaeche\tkommandozeile"):
			t.Errorf("Ausgabe %q nennt die Fläche nicht", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		antwort := mussMCP(t, deckumgebung(t), "lesen.meta", map[string]any{})

		if !strings.Contains(antwort, `"flaeche": "mcp"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		if antwort := mussREST(t, an, http.MethodGet, "/meta", ""); !strings.Contains(
			antwort, `"flaeche":"dienst"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})
}
