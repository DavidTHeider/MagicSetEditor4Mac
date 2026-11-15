package openai_test

// **Was OpenAI über seine Modelle sagt — und was nicht** — Werkstück A9 von
// Phase 5.3
// (#1448).
//
// # Der Fund, um den sich dieses Werkstück dreht
//
// **`/v1/models` nennt keine Fähigkeiten.** Nachgelesen am 2026-09-12: jedes
// Objekt trägt `id`, `object`, `created`, `owned_by` und ein optionales
// `shutdown_date` — und sonst nichts. Kein `capabilities`, kein Hinweis auf
// feste Form, Bilder oder Einbettung.
//
// Anthropic sagt es (A8), Ollama sagt es (A7). Hier sagt es niemand — und
// [0097](../../docs/entscheidungen/0097-was-ueber-ein-ki-modell-feststeht.md)
// ist dabei eindeutig: „unbekannt bleibt unbekannt; Modelllisten garantieren
// keine Capability."
//
// **Also entsteht aus dieser Liste kein einziger Katalogeintrag** — und je
// Modell ein Befund, der den Grund nennt. Wer aus `gpt-` schlösse, dass ein
// Modell Bilder liest, hätte genau das getan, was §15 verbietet.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

// **Die Modellliste macht Befunde und keine Einträge.**
func TestDieModelllisteMachtBefundeUndKeineEintraege(t *testing.T) {
	t.Parallel()

	server := eineGegenstelle(t)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	modelle, befunde, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(modelle) != 0 {
		t.Fatalf("aus einer Liste ohne Fähigkeiten entstanden Einträge: %+v", modelle)
	}

	if len(befunde) != 2 {
		t.Fatalf("es sind %d Befunde: %+v", len(befunde), befunde)
	}

	// **Der Befund nennt das Modell und den Grund** — nicht „nichts
	// gefunden". Wer ihn liest, soll wissen, was als Nächstes zu tun ist.
	for _, befund := range befunde {
		if !strings.Contains(befund.Meldung, "/v1/models") {
			t.Errorf("der Befund nennt den Weg nicht: %q", befund.Meldung)
		}

		if befund.Datei == "" {
			t.Errorf("der Befund nennt das Modell nicht: %+v", befund)
		}
	}
}

// eineGegenstelle antwortet wie die Schnittstelle von OpenAI.
func eineGegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "/models") {
				schreibeDatei(t, schreiben, "modelle.json")

				return
			}

			schreibeDatei(t, schreiben, "antwort.json")
		}))

	t.Cleanup(server.Close)

	return server
}

// schreibeDatei schickt eine Datei aus `testdaten`.
func schreibeDatei(t *testing.T, schreiben http.ResponseWriter, name string) {
	t.Helper()

	// **Der Name kommt aus dem Lauf, nicht von außen** — ein fester Ordner
	// im Repositorium, und jeder Aufrufer steht daneben.
	roh, fehler := os.ReadFile(filepath.Join("testdaten", name)) //nolint:gosec // siehe oben
	if fehler != nil {
		t.Fatalf("%s lesen: %v", name, fehler)
	}

	_, _ = schreiben.Write(roh) //nolint:errcheck // die Gegenstelle
}
