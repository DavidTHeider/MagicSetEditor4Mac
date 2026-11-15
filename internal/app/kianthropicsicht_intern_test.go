package app

// **Die Naht schickt ein Bild** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Was hier gemessen wird
//
// **Dass das Bild wirklich hinausgeht — und als das, was es ist.** Anthropic
// nimmt ein Bild als `base64` mit **seinem Medientyp** entgegen; ein Weg, der
// die Bytes fallen ließe, bekäme trotzdem eine Antwort, weil das Modell dann
// eben nur den Text liest. Die Zusage „das Modell sieht die Karte" wäre eine
// Behauptung über etwas, das nie über die Leitung ging.
//
// **Paketintern und gegen einen `httptest.Server`** — dieselbe Begründung wie
// bei [TestDieAnthropicnahtFragtUndUebersetzt] nebenan.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// **Das Bild geht mit — samt Medientyp.**
func TestDieAnthropicnahtSchicktDasBild(t *testing.T) {
	t.Parallel()

	db := anthropicAn(t)

	var gestellt map[string]any

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if strings.HasSuffix(anfrage.URL.Path, "/models") {
				_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			roh, _ := io.ReadAll(anfrage.Body) //nolint:errcheck // die Gegenstelle
			_ = json.Unmarshal(roh, &gestellt) //nolint:errcheck // die Gegenstelle

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"model":"claude-probe","stop_reason":"end_turn",
				  "content":[{"type":"text","text":"{\"beobachtungen\": []}"}],
				  "usage":{"input_tokens":11,"output_tokens":18}}`))
		}))

	t.Cleanup(server.Close)

	quelle := mitSchluessel(db).anthropicnaht(server.URL)

	sicht, kann := any(quelle).(KISichtquelle)
	if !kann {
		t.Fatalf("%T liest keine Bilder — die Naht erfüllt KISichtquelle nicht", quelle)
	}

	antwort, fehler := sicht.Sehen(riegelkontext(t), KISichtwunsch{
		Anweisung: KIErkennungsanweisung, Alias: "claude-probe",
		Auftrag: "Lies diese Karte.",
		Bild:    []byte{0x89, 'P', 'N', 'G'}, Medientyp: "image/png",
		Schema: map[string]any{"type": "object"},
	})
	if fehler != nil {
		t.Fatalf("Sehen: %v", fehler)
	}

	if antwort.Modell != "claude-probe" {
		t.Errorf("die Antwort ist %+v", antwort)
	}

	// **Und der Rumpf trug das Bild** — nicht nur den Text.
	roh, fehler := json.Marshal(gestellt)
	if fehler != nil {
		t.Fatalf("den gestellten Rumpf lesen: %v", fehler)
	}

	for _, teil := range []string{`"image"`, `"image/png"`, `"base64"`} {
		if !strings.Contains(string(roh), teil) {
			t.Errorf("der Rumpf nennt %s nicht:\n%s", teil, roh)
		}
	}
}
