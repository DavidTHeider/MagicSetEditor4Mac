package anthropic_test

// **Die Unterhaltung in fester Form** — Werkstück A8 von Phase 5.3
// (#1447).
//
// Die Gestalt ist am 2026-09-11 unter
// <https://platform.claude.com/docs/en/build-with-claude/structured-outputs>
// und <https://platform.claude.com/docs/en/api/messages> nachgelesen; die
// Beispiele darunter sind **selbst geschrieben** (§3).

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Antwort kommt als Text zurück — und gilt als fertig.**
func TestEineAntwortKommtAlsTextZurueck(t *testing.T) {
	t.Parallel()

	antwort, fehler := klientFuer(gegenstelle(t)).Rede(t.Context(),
		anthropic.Frage{Alias: "claude-opus-5", Auftrag: "Nenne die Karte."})
	if fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	if antwort.Inhalt != `{"name": "Blitzschlag"}` {
		t.Errorf("der Inhalt ist %q", antwort.Inhalt)
	}

	if antwort.Ausgang != anthropic.AusgangFertig {
		t.Errorf("der Ausgang ist %q", antwort.Ausgang)
	}

	// **Das antwortende Modell steht dabei** — nicht dasselbe wie der Alias
	// der Frage; wer es nicht vergleicht, merkt einen stillen Tausch nicht.
	if antwort.Modell != "claude-opus-5" {
		t.Errorf("das Modell ist %q", antwort.Modell)
	}

	// **Und der Verbrauch** — das Issue verlangt „Usage-Metadaten an 5.2
	// weitergeben".
	if antwort.Verbrauch.Ausgabe != 503 {
		t.Errorf("der Verbrauch ist %+v", antwort.Verbrauch)
	}

	// **Die Eingabe zählt den Zwischenspeicher mit**: `input_tokens` sind
	// allein die **nach** der letzten Marke.
	if antwort.Verbrauch.Eingabezeichen() != 2095+1024 {
		t.Errorf("die Eingabezeichen sind %d", antwort.Verbrauch.Eingabezeichen())
	}
}

// **Das Schema geht als `output_config.format` hinaus.**
//
// # Warum nicht als `output_format`
//
// **Weil das der alte Weg ist.** Nachgelesen am 2026-09-11: „The
// `output_format` parameter has moved to `output_config.format`, and beta
// headers are no longer required." Der alte wird übergangsweise noch
// angenommen — wer ihn heute neu schriebe, schriebe eine Übergangsfassung.
func TestDasSchemaGehtAlsOutputConfigHinaus(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(),
		anthropic.Frage{
			Alias: "claude-opus-5", Auftrag: "x",
			Schema: map[string]any{"type": "object"},
		}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	var gestellt map[string]any

	if fehler := json.Unmarshal(<-gesehen.rumpf, &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v", fehler)
	}

	aufbau, gibts := gestellt["output_config"].(map[string]any)
	if !gibts {
		t.Fatalf("»output_config« fehlt: %v", gestellt)
	}

	form, hatForm := aufbau["format"].(map[string]any)
	if !hatForm || form["type"] != "json_schema" {
		t.Fatalf("»format« ist %v", aufbau["format"])
	}

	if form["schema"] == nil {
		t.Errorf("das Schema fehlt: %v", form)
	}

	// **Und der alte Weg geht nicht daneben hinaus.**
	if _, alt := gestellt["output_format"]; alt {
		t.Errorf("»output_format« steht daneben: %v", gestellt["output_format"])
	}
}

// **Ohne Schema geht kein `output_config` hinaus.**
func TestOhneSchemaGehtKeinOutputConfigHinaus(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(),
		anthropic.Frage{Alias: "claude-opus-5", Auftrag: "x"}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	var gestellt map[string]any

	if fehler := json.Unmarshal(<-gesehen.rumpf, &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v", fehler)
	}

	if _, gibts := gestellt["output_config"]; gibts {
		t.Errorf("»output_config« steht in der Anfrage: %v", gestellt["output_config"])
	}

	// **`max_tokens` steht immer da** — es ist Pflicht in der Messages-API.
	if gestellt["max_tokens"] == nil {
		t.Error("»max_tokens« fehlt")
	}

	if gestellt["stream"] != nil {
		t.Errorf("»stream« steht in der Anfrage: %v", gestellt["stream"])
	}
}

// **Die Systemanweisung geht als eigenes Feld hinaus.**
//
// **Und nicht als Nachricht mit `role: "system"`** — die Messages-API kennt
// nur `user` und `assistant`; eine dritte Rolle wäre ein 400.
func TestDieSystemanweisungGehtAlsEigenesFeld(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(),
		anthropic.Frage{
			Alias: "claude-opus-5", System: "Du bist knapp.", Auftrag: "x",
		}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	rumpf := string(<-gesehen.rumpf)

	if !strings.Contains(rumpf, `"system":"Du bist knapp."`) {
		t.Errorf("der Rumpf ist %s", rumpf)
	}

	if strings.Contains(rumpf, `"role":"system"`) {
		t.Errorf("die Anweisung ging als Rolle hinaus: %s", rumpf)
	}
}

// **Ein Bild geht als Block mit Medientyp hinaus.**
func TestEinBildGehtAlsBlockMitMedientyp(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(),
		anthropic.Frage{
			Alias: "claude-opus-5", Auftrag: "Was steht darauf?",
			Bilder: []anthropic.Bild{{
				Medientyp: "image/png", Daten: []byte{0x89, 0x50, 0x4e, 0x47},
			}},
		}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	var gestellt struct {
		Nachrichten []struct {
			Inhalt []struct {
				Art    string `json:"type"`
				Quelle struct {
					Art       string `json:"type"`
					Medientyp string `json:"media_type"`
					Daten     string `json:"data"`
				} `json:"source"`
			} `json:"content"`
		} `json:"messages"`
	}

	if fehler := json.Unmarshal(<-gesehen.rumpf, &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v", fehler)
	}

	if len(gestellt.Nachrichten) != 1 || len(gestellt.Nachrichten[0].Inhalt) != 2 {
		t.Fatalf("die Nachrichten sind %+v", gestellt.Nachrichten)
	}

	// **Das Bild steht vor dem Text** — die Dokumentation empfiehlt es, und
	// die Reihenfolge ist eine Aussage über die Frage.
	bild := gestellt.Nachrichten[0].Inhalt[0]

	if bild.Art != "image" || bild.Quelle.Art != "base64" {
		t.Fatalf("der erste Block ist %+v", bild)
	}

	if bild.Quelle.Medientyp != "image/png" || bild.Quelle.Daten != "iVBORw==" {
		t.Errorf("die Quelle ist %+v", bild.Quelle)
	}
}

// **Ein Bild ohne bekannten Medientyp geht gar nicht erst hinaus.**
//
// **Nachgelesen am 2026-09-11**: vier Typen, und sonst keiner. Ein fünfter
// käme als 400 zurück, und die Meldung spräche von der Anfrage.
func TestEinBildOhneBekanntenMedientypGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	_, fehler := klientFuer(server).Rede(t.Context(), anthropic.Frage{
		Alias: "claude-opus-5", Auftrag: "x",
		Bilder: []anthropic.Bild{{Medientyp: "image/tiff", Daten: []byte{1}}},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if gefragt {
		t.Error("das unbekannte Bild ging hinaus")
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

			schreibeDatei(t, schreiben, "antwort.json")
		}))

	t.Cleanup(spion.server.Close)

	return spion
}
