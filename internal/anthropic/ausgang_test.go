package anthropic_test

// **Wie eine Antwort geendet hat** — Werkstück A8 von Phase 5.3
// (#1447).
//
// **Das Issue verlangt die Trennung ausdrücklich**: „unvollständig, Refusal
// und unbekannt getrennt behandeln."

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
)

// **Sieben Gründe, vier Klassen.**
//
// **Nachgelesen am 2026-09-11** unter
// <https://platform.claude.com/docs/en/api/handling-stop-reasons>.
func TestJederAbschlussgrundBekommtSeineKlasse(t *testing.T) {
	t.Parallel()

	for grund, erwartet := range map[string]anthropic.Ausgang{
		"end_turn":                      anthropic.AusgangFertig,
		"stop_sequence":                 anthropic.AusgangFertig,
		"max_tokens":                    anthropic.AusgangUnvollstaendig,
		"model_context_window_exceeded": anthropic.AusgangUnvollstaendig,
		"refusal":                       anthropic.AusgangAbgelehnt,
		"tool_use":                      anthropic.AusgangUnbekannt,
		"pause_turn":                    anthropic.AusgangUnbekannt,

		// **Ein Wert, den es heute nicht gibt** — die Zusage gegen Drift
		// (§16): er wird **nicht** als „fertig" gelesen, nur weil er nicht
		// in der Liste steht.
		"etwasneues": anthropic.AusgangUnbekannt,
		"":           anthropic.AusgangUnbekannt,
	} {
		t.Run(grund, func(t *testing.T) {
			t.Parallel()

			antwort := redeMit(t, fmt.Sprintf(
				`{"model":"claude-opus-5","content":[{"type":"text","text":"x"}],
				  "stop_reason":%q}`, grund))

			if antwort.Ausgang != erwartet {
				t.Errorf("%q gibt %q, erwartet %q", grund, antwort.Ausgang, erwartet)
			}

			// **Der rohe Grund reist mit** — die Klasse ist die Entscheidung,
			// der Grund die Auskunft fürs Protokoll.
			if antwort.Grund != grund {
				t.Errorf("der Grund ist %q", antwort.Grund)
			}
		})
	}
}

// **Eine Ablehnung nennt ihre Kategorie.**
//
// **Weil „das Modell wollte nicht" nicht weiterhilft.** Die Dokumentation
// sagt: „The response includes a `stop_details` object identifying the policy
// category that triggered it."
func TestEineAblehnungNenntIhreKategorie(t *testing.T) {
	t.Parallel()

	antwort := redeMit(t, `{"model":"claude-opus-5","content":[],
		"stop_reason":"refusal","stop_details":{"type":"gewalt"}}`)

	if antwort.Ausgang != anthropic.AusgangAbgelehnt {
		t.Fatalf("der Ausgang ist %q", antwort.Ausgang)
	}

	if antwort.Naeheres != "gewalt" {
		t.Errorf("das Nähere ist %q", antwort.Naeheres)
	}
}

// **Mehrere Textblöcke werden zusammengefügt.**
//
// # Warum nicht der erste genommen wird
//
// **Weil die Antwort eine Liste ist.** Bei einem nachdenkenden Modell steht
// vor dem Text ein `thinking`-Block; wer „das erste Element" nähme, bekäme
// die Überlegung statt der Antwort — und bei zwei Textblöcken die Hälfte.
func TestMehrereTextbloeckeWerdenZusammengefuegt(t *testing.T) {
	t.Parallel()

	antwort := redeMit(t, `{"model":"claude-opus-5","stop_reason":"end_turn",
		"content":[{"type":"thinking","thinking":"hm"},
		           {"type":"text","text":"{\"name\":"},
		           {"type":"text","text":" \"Blitzschlag\"}"}]}`)

	if antwort.Inhalt != `{"name": "Blitzschlag"}` {
		t.Errorf("der Inhalt ist %q", antwort.Inhalt)
	}
}

// redeMit stellt eine Frage an eine Gegenstelle mit fester Antwort.
func redeMit(t *testing.T, antwort string) anthropic.Antwort {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte(antwort)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	gegeben, fehler := klientFuer(server).Rede(t.Context(), anthropic.Frage{
		Alias: "claude-opus-5", Auftrag: "x",
	})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	return gegeben
}
