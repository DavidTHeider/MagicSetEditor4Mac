package openai_test

// **Die Responses-API und ihre vier Ausgänge** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum die Ausgänge hier mehr Formen haben als bei Anthropic
//
// **Weil zwei Stellen sie tragen.** Anthropic sagt alles in `stop_reason`;
// hier stehen `status` und `incomplete_details.reason` **nebeneinander**, und
// eine Ablehnung steht an einer **dritten**: als Inhaltsblock vom Typ
// `refusal`. Wer nur `status` läse, hielte eine Ablehnung für ein Ergebnis —
// `status` ist dabei `completed`.

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

// **Eine Frage mit Schema kommt als feste Form zurück.**
func TestEineFrageMitSchemaKommtAlsFesteForm(t *testing.T) {
	t.Parallel()

	var gesehen map[string]any

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			_ = json.NewDecoder(gestellt.Body).Decode(&gesehen) //nolint:errcheck // die Gegenstelle

			schreibeDatei(t, schreiben, "antwort.json")
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	antwort, fehler := klient.Rede(t.Context(), openai.Frage{
		Alias: "gpt-probe", System: "Sei knapp.", Auftrag: "Schreibe.",
		Schema: map[string]any{"type": "object"},
	})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if antwort.Ausgang != openai.AusgangFertig ||
		!strings.Contains(antwort.Inhalt, "Blitzschlag") {
		t.Errorf("die Antwort ist %+v", antwort)
	}

	if antwort.Modell != "gpt-probe" {
		t.Errorf("das Modell ist %q", antwort.Modell)
	}

	// **Das Schema steht unter `text.format`** und heißt `json_schema`.
	form, steht := gesehen["text"].(map[string]any)
	if !steht {
		t.Fatalf("die Anfrage ist %+v", gesehen)
	}

	format, steht := form["format"].(map[string]any)
	if !steht || format["type"] != "json_schema" || format["strict"] != true {
		t.Errorf("die Form ist %+v", form)
	}

	// **`store` steht ausdrücklich auf `false`.**
	//
	// Die Vorgabe der Schnittstelle ist `true` — dann bliebe jede
	// Kartenanfrage beim Anbieter liegen. Das ist der teuerste stille
	// Unterschied dieses Werkstücks.
	if gesehen["store"] != false {
		t.Errorf("`store` ist %v — die Anfrage bliebe beim Anbieter liegen",
			gesehen["store"])
	}

	// **Die Anweisung steht in `instructions`** und nicht als Nachricht: die
	// Dokumentation sagt, sie habe Vorrang vor dem, was in `input` steht.
	if gesehen["instructions"] != "Sei knapp." {
		t.Errorf("die Anweisung ist %v", gesehen["instructions"])
	}
}

// **Vier Ausgänge, und drei davon sind kein Ergebnis.**
func TestJederAusgangWirdUnterschieden(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		rumpf    string
		erwartet openai.Ausgang
	}{
		"fertig": {
			`{"status":"completed","model":"m","output":[{"type":"message",
			  "content":[{"type":"output_text","text":"{}"}]}]}`,
			openai.AusgangFertig,
		},
		"am Zeichenfenster abgeschnitten": {
			`{"status":"incomplete","model":"m",
			  "incomplete_details":{"reason":"max_output_tokens"},
			  "output":[{"type":"message",
			  "content":[{"type":"output_text","text":"{\"na"}]}]}`,
			openai.AusgangUnvollstaendig,
		},
		"vom Filter angehalten": {
			`{"status":"incomplete","model":"m",
			  "incomplete_details":{"reason":"content_filter"},
			  "output":[{"type":"message","content":[]}]}`,
			openai.AusgangAbgelehnt,
		},
		"abgelehnt": {
			`{"status":"completed","model":"m","output":[{"type":"message",
			  "content":[{"type":"refusal","refusal":"tue ich nicht"}]}]}`,
			openai.AusgangAbgelehnt,
		},
		"unbekannter Stand": {
			`{"status":"etwasneues","model":"m","output":[]}`,
			openai.AusgangUnbekannt,
		},
		"ein fremder Ausgabeblock zählt nicht als Text": {
			`{"status":"completed","model":"m","output":[
			  {"type":"reasoning","content":[{"type":"output_text","text":"gedacht"}]},
			  {"type":"message","content":[{"type":"output_text","text":"{}"}]}]}`,
			openai.AusgangFertig,
		},
		"unbekannter Grund": {
			`{"status":"incomplete","model":"m",
			  "incomplete_details":{"reason":"etwasneues"},"output":[]}`,
			openai.AusgangUnbekannt,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := redeGegen(t, 200, fall.rumpf)

			if antwort.Ausgang != fall.erwartet {
				t.Errorf("der Ausgang ist %q, erwartet %q — %+v",
					antwort.Ausgang, fall.erwartet, antwort)
			}
		})
	}
}

// **Ein Block, der keine Nachricht ist, zählt nicht als Text.**
//
// **Und das ist kein Zierrat**: die Ausgabe trägt neben `message` auch
// `reasoning` und, in A12, `image_generation_call`. Wer alles zusammenfügte,
// schöbe interne Überlegungen in das JSON, das danach gegen ein Schema
// geprüft wird.
func TestEinFremderBlockZaehltNichtAlsText(t *testing.T) {
	t.Parallel()

	antwort := redeGegen(t, 200, `{"status":"completed","model":"m","output":[
	  {"type":"reasoning","content":[{"type":"output_text","text":"gedacht"}]},
	  {"type":"message","content":[{"type":"output_text","text":"{}"}]}]}`)

	if antwort.Inhalt != "{}" {
		t.Errorf("der Inhalt ist %q", antwort.Inhalt)
	}
}

// **Eine Ablehnung trägt ihren Satz mit.**
//
// **Und `status` ist dabei `completed`** — wer nur ihn läse, hielte sie für
// ein Ergebnis.
func TestEineAblehnungTraegtIhrenSatz(t *testing.T) {
	t.Parallel()

	antwort := redeGegen(t, 200, `{"status":"completed","model":"m",
	  "output":[{"type":"message",
	  "content":[{"type":"refusal","refusal":"tue ich nicht"}]}]}`)

	if antwort.Naeheres != "tue ich nicht" {
		t.Errorf("das Nähere ist %q", antwort.Naeheres)
	}
}

// **Ein gescheiterter Lauf ist ein Fehler und keine leere Antwort.**
//
// **Der Status ist dabei 200**: die Anfrage ist angekommen, der Lauf ist
// misslungen. Wer das als Ausgang führte, gäbe einen leeren Vorschlag
// zurück — und die Schemaprüfung spräche danach von fehlenden Feldern.
func TestEinGescheiterterLaufIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"status":"failed","model":"m","output":[],
				  "error":{"code":"server_error",
				  "message":"The model failed to generate a response."}}`))
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	_, fehler := klient.Rede(t.Context(), openai.Frage{
		Alias: "m", Auftrag: "x",
	})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "OpenAI") {
		t.Errorf("die Meldung nennt den Anbieter nicht: %v", fehler)
	}
}

// redeGegen fragt eine Gegenstelle mit festem Rumpf.
func redeGegen(t *testing.T, status int, rumpf string) openai.Antwort {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(status)
			_, _ = schreiben.Write([]byte(rumpf)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	antwort, fehler := klient.Rede(t.Context(), openai.Frage{
		Alias: "m", Auftrag: "x",
	})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	return antwort
}
