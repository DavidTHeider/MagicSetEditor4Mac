package ollama_test

// **Die Unterhaltung** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// **Eine Antwort kommt als Text zurück.**
func TestEineAntwortKommtAlsTextZurueck(t *testing.T) {
	t.Parallel()

	antwort, fehler := klientFuer(gegenstelle(t)).Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", Auftrag: "Nenne die Karte.",
	})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if antwort.Inhalt != `{"name": "Blitzschlag"}` {
		t.Errorf("der Inhalt ist %q", antwort.Inhalt)
	}

	// **Der Grund des Abschlusses reist mit** — „stop" ist etwas anderes als
	// „length", und nur das zweite heißt: abgeschnitten.
	if antwort.Grund != "stop" || !antwort.Fertig {
		t.Errorf("die Antwort ist %+v", antwort)
	}
}

// **Ohne Strom** — und das ist kein Detail.
//
// **Nachgelesen am 2026-09-11**: `/api/chat` strömt, wenn man nichts sagt.
// Ein strömender Ruf, dessen Teile niemand zusammensetzt, gäbe die **erste**
// Silbe als ganze Antwort aus.
func TestDieAnfrageBittetUmKeinenStrom(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", Auftrag: "x",
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	var gestellt map[string]any

	if fehler := json.Unmarshal(<-gesehen.rumpf, &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v", fehler)
	}

	if gestellt["stream"] != false {
		t.Errorf("»stream« ist %v", gestellt["stream"])
	}

	if gestellt["model"] != "gemma4:8b" {
		t.Errorf("»model« ist %v", gestellt["model"])
	}
}

// **Die Frage bittet ums Antworten, nicht ums Denken** — gemessen in AI-02
// (#1743)
// am 2026-10-04: ohne `think` schaltet Ollama 0.35.1 das Denken bei
// `qwen3:0.6b` ein — über die 32 Fragen des Testsatzes im Median 267 Token
// und 1,8 s vor jeder Antwort, und eine Frage nach einem Kartennamen hatte
// nach drei Minuten und über 7.400 Token Denktext noch keinen Inhalt. Mit
// `false` 14 Token und 0,16 s. `gemma3:1b`, das nicht denken kann, nimmt das
// Feld an.
//
// **Mit und ohne Schema**: dieses Paket liest `message.thinking` gar nicht —
// Denktext wäre in jedem Fall bezahlt und weggeworfen.
func TestDieAnfrageBittetUmKeinDenken(t *testing.T) {
	t.Parallel()

	for _, schema := range []map[string]any{nil, {"type": "object"}} {
		gesehen := gesehenerRumpf(t)

		if _, fehler := klientFuer(gesehen.server).Rede(t.Context(), ollama.Frage{
			Alias: "qwen3:0.6b", Auftrag: "x", Schema: schema,
		}); fehler != nil {
			t.Fatalf("Rede: %v", fehler)
		}

		var gestellt map[string]any

		if fehler := json.Unmarshal(<-gesehen.rumpf, &gestellt); fehler != nil {
			t.Fatalf("den Rumpf lesen: %v", fehler)
		}

		if denken, da := gestellt["think"]; !da || denken != false {
			t.Errorf("mit Schema %v: »think« ist %v (gesetzt: %v)", schema, denken, da)
		}
	}
}

// **Ein Schema geht als `format` hinaus.**
//
// **Nachgelesen am 2026-09-11**: der Parameter heißt `format` und nimmt ein
// JSON-Schema-Objekt. Er ist eine Eigenschaft der **Laufzeit** — Ollama hält
// die Form, statt darum zu bitten.
func TestEinSchemaGehtAlsFormatHinaus(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{"name": map[string]any{"type": "string"}},
		"required":   []string{"name"},
	}

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", Auftrag: "x", Schema: schema,
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	var gestellt map[string]any

	if fehler := json.Unmarshal(<-gesehen.rumpf, &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v", fehler)
	}

	form, gibts := gestellt["format"].(map[string]any)
	if !gibts || form["type"] != "object" {
		t.Errorf("»format« ist %v", gestellt["format"])
	}

	// **Und die Temperatur wird gesenkt** — die Dokumentation empfiehlt es
	// ausdrücklich für verlässliche Form: „Lower the temperature (e.g., set
	// it to `0`)".
	optionen, hatOptionen := gestellt["options"].(map[string]any)
	if !hatOptionen || optionen["temperature"] != float64(0) {
		t.Errorf("»options« ist %v", gestellt["options"])
	}
}

// **Ohne Schema geht kein `format` hinaus.**
//
// **Die Gegenprobe**: ein leeres `format` ist etwas anderes als keines, und
// wer es immer schickte, zwänge jede Antwort in eine Form, die niemand
// verlangt hat.
func TestOhneSchemaGehtKeinFormatHinaus(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", Auftrag: "x",
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	var gestellt map[string]any

	if fehler := json.Unmarshal(<-gesehen.rumpf, &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v", fehler)
	}

	if _, gibts := gestellt["format"]; gibts {
		t.Errorf("»format« steht in der Anfrage: %v", gestellt["format"])
	}
}

// **Ein Bild geht als Base64 in die Nachricht.**
func TestEinBildGehtAlsBase64InDieNachricht(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", Auftrag: "Was steht darauf?",
		Bilder: [][]byte{{0x89, 0x50, 0x4e, 0x47}},
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	var gestellt struct {
		Nachrichten []struct {
			Bilder []string `json:"images"`
		} `json:"messages"`
	}

	if fehler := json.Unmarshal(<-gesehen.rumpf, &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v", fehler)
	}

	if len(gestellt.Nachrichten) != 1 || len(gestellt.Nachrichten[0].Bilder) != 1 {
		t.Fatalf("die Nachrichten sind %+v", gestellt.Nachrichten)
	}

	if gestellt.Nachrichten[0].Bilder[0] != "iVBORw==" {
		t.Errorf("das Bild ist %q", gestellt.Nachrichten[0].Bilder[0])
	}
}

// **Ohne Alias wird nicht gefragt.**
func TestOhneAliasWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuer(server).Rede(t.Context(), ollama.Frage{
		Auftrag: "x",
	}); fehler == nil {
		t.Fatal("ohne Alias kam eine Antwort")
	}

	if gefragt {
		t.Error("der Server wurde ohne Alias gefragt")
	}
}

// **Und ohne Auftrag ebenso.**
func TestOhneAuftragWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuer(server).Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", Auftrag: "   ",
	}); fehler == nil {
		t.Fatal("ohne Auftrag kam eine Antwort")
	}

	if gefragt {
		t.Error("der Server wurde ohne Auftrag gefragt")
	}
}

// **Eine abgeschnittene Antwort sagt es.**
//
// Das Issue nennt sie unter denen, die **keine Kartenänderung** erzeugen
// dürfen. Wer `done_reason` überginge, nähme eine halbe Antwort für eine ganze.
func TestEineAbgeschnitteneAntwortSagtEs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"model":"gemma4:8b","message":{"role":"assistant",
				  "content":"{\"name\": \"Blitz"},"done":true,
				  "done_reason":"length"}`))
		}))
	t.Cleanup(server.Close)

	antwort, fehler := klientFuer(server).Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", Auftrag: "x",
	})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if !antwort.Abgeschnitten() {
		t.Errorf("die Antwort gilt nicht als abgeschnitten: %+v", antwort)
	}
}

// **Ein Fehler der Laufzeit bleibt einer.**
func TestEinFehlerDerLaufzeitBleibtEiner(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusNotFound)
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuer(server).Rede(t.Context(), ollama.Frage{
		Alias: "gibtesnicht", Auftrag: "x",
	}); fehler == nil {
		t.Fatal("ein 404 kam als Antwort zurück")
	}
}

// rumpfspion hält den Server und den gesehenen Rumpf.
type rumpfspion struct {
	server *httptest.Server
	rumpf  chan []byte
}

// gesehenerRumpf baut eine Gegenstelle, die ihren Rumpf durchreicht.
func gesehenerRumpf(t *testing.T) rumpfspion {
	t.Helper()

	spion := rumpfspion{rumpf: make(chan []byte, 1)}

	spion.server = httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			roh := make([]byte, 8192)
			gelesen, _ := gestellt.Body.Read(roh) //nolint:errcheck // die Gegenstelle

			spion.rumpf <- roh[:gelesen]

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"message":{"role":"assistant","content":"x"},"done":true,
				  "done_reason":"stop"}`))
		}))

	t.Cleanup(spion.server.Close)

	return spion
}
