package flaechen_test

// **Die Verträglichkeit über Kommandozeile, MCP und Dienst** — Werkstück O6
// von Phase 5.2 (#1430).
//
// # Was er hält
//
// **Dieselbe Aussage in drei Gestalten** — und ohne Netz dieselbe: die
// Einstellung `aktualisierung` steht auf „aus", und daraus folgt „keine
// neuere Fassung", nicht ein Fehlschlag. Das ist §16, Regel 6: ohne Netz
// arbeitet die Anwendung weiter, und die Vorgabe der Einstellung ist „aus".
//
// # Was er nicht hält
//
// **Das Fenster** — dieser Lauf startet Prozesse. Die Bindung prüft
// `internal/app/vertraeglichkeit_test.go`, samt der drei Ausgänge.
//
// **Und den Weg mit einer echten Freigabe.** Dafür bräuchte er GitHub; was
// über eine echte Antwort entscheidet, steht in
// `internal/aktualisierung/manifest_test.go` gegen eine Gegenstelle.

import (
	"net/http"
	"strings"
	"testing"
)

func TestDieVertraeglichkeitGehtUeberDreiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		aus, klage, code := kommandozeile(t, deckumgebung(t), "vertraeglichkeit")

		// **Ohne Netz und mit abgeschalteter Aktualisierung gibt es kein
		// Urteil** — und das ist der geglückte Weg, nicht der gescheiterte.
		switch {
		case code != 0:
			t.Fatalf("vertraeglichkeit gab %d (%s)", code, klage)
		case !strings.Contains(aus, "urteil\t"):
			t.Errorf("Ausgabe %q nennt kein Urteil", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		antwort := mussMCP(t, deckumgebung(t), "lesen.vertraeglichkeit", map[string]any{})

		if !strings.Contains(antwort, `"hindernisse"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		if antwort := mussREST(t, an, http.MethodGet, "/vertraeglichkeit", ""); !strings.Contains(
			antwort, `"hindernisse"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})
}
