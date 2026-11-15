package ollama_test

// **Die Vermessung eines Einbettungsmodells** — Werkstück A7 von Phase 5.3
// (#1446),
// Ergänzung vom 2026-09-10.
//
// Die Läufe hier gehören zu [ollama.Klient.Vermessen]; die Gegenstellen dazu
// stehen in `einbetten_test.go`.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// **Ein vermessenes Modell bekommt seinen Vertrag aus Metadaten und Lauf.**
//
// # Woher jeder Teil kommt
//
// | Teil | Quelle |
// | --- | --- |
// | Zusammenfassung | `model_info` unter einem Schlüssel auf `.pooling_type` |
// | Zerteiler | `model_info["tokenizer.ggml.model"]` |
// | Dimension | **gemessen** an einem echten Lauf |
// | Normalisiert | **gemessen** am Betrag des Vektors |
//
// **Nichts davon ist geraten** — das Issue verlangt „Capabilities und
// Modellfingerprint statt Namensheuristik".
func TestEinVermessenesModellBekommtSeinenVertrag(t *testing.T) {
	t.Parallel()

	vertrag, befunde, fehler := klientFuer(einbettende(t, nil)).
		Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Fatalf("%d Befunde: %+v", len(befunde), befunde)
	}

	if vertrag.Dimension != 4 || vertrag.Pooling != domain.KIPoolingMittel {
		t.Errorf("der Vertrag ist %+v", vertrag)
	}

	if vertrag.Zerteiler != "bert" || !vertrag.Normalisiert {
		t.Errorf("der Vertrag ist %+v", vertrag)
	}

	if fehler := vertrag.Pruefe(); fehler != nil {
		t.Errorf("der gemessene Vertrag ist ungültig: %v", fehler)
	}
}

// **Ohne Zusammenfassung in den Metadaten gibt es keinen Vertrag.**
//
// **Die Gegenprobe zur Messung**: Dimension und Normalisierung sind zu messen,
// die Zusammenfassung nicht. Wer sie riete, bekäme Vektoren, die systematisch
// danebenliegen, **ohne dass etwas scheitert**.
func TestOhneZusammenfassungGibtEsKeinenVertrag(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Schauweg {
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"capabilities":["embedding"],"license":"MIT",
					  "model_info":{"tokenizer.ggml.model":"bert"}}`))

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"embeddings":[[0.5,0.5,0.5,0.5]]}`))
		}))
	t.Cleanup(server.Close)

	_, befunde, fehler := klientFuer(server).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) == 0 {
		t.Fatal("ohne Zusammenfassung kam kein Befund")
	}

	if !strings.Contains(befunde[0].Meldung, "pooling_type") {
		t.Errorf("die Meldung nennt den fehlenden Schlüssel nicht: %q",
			befunde[0].Meldung)
	}
}

// **Ein Modell ohne die Fähigkeit wird gar nicht erst vermessen.**
func TestEinModellOhneEinbettungWirdNichtVermessen(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(gegenstelle(t)).
		Vermessen(t.Context(), "gemma4:8b")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) == 0 {
		t.Fatal("ein Modell ohne die Fähigkeit kam ohne Befund durch")
	}
}

// **Eine unbekannte Zusammenfassung ist ein Befund und keine Annahme.**
func TestEineUnbekannteZusammenfassungIstEinBefund(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Schauweg {
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"capabilities":["embedding"],"license":"MIT",
					  "model_info":{"bert.pooling_type":99,
					                "tokenizer.ggml.model":"bert"}}`))

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"embeddings":[[0.5,0.5,0.5,0.5]]}`))
		}))
	t.Cleanup(server.Close)

	_, befunde, fehler := klientFuer(server).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) == 0 {
		t.Fatal("eine unbekannte Zusammenfassung kam ohne Befund durch")
	}
}

// **Die Zusammenfassung „keine" ist kein Einbettungsmodell.**
//
// **Nachgelesen am 2026-09-11** in `gguf-py/gguf/constants.py` von llama.cpp:
// `PoolingType` zählt `NONE = 0`, `MEAN = 1`, `CLS = 2`, `LAST = 3`, `RANK = 4`.
// Bei `NONE` kommt ein Vektor **je Zeichen** heraus und keiner je Text — und
// A6 kennt dafür keinen Vertrag.
func TestDieZusammenfassungKeineIstKeinVertrag(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Schauweg {
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"capabilities":["embedding"],"license":"MIT",
					  "model_info":{"bert.pooling_type":0,
					                "tokenizer.ggml.model":"bert"}}`))

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"embeddings":[[0.5,0.5,0.5,0.5]]}`))
		}))
	t.Cleanup(server.Close)

	_, befunde, fehler := klientFuer(server).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) == 0 {
		t.Fatal("»keine Zusammenfassung« kam ohne Befund durch")
	}
}

// **Ein unnormalisierter Vektor wird als solcher vermessen.**
//
// **Die Gegenprobe zur Messung der Normalisierung**: ein Messer, der immer
// `true` sagt, wäre in jedem anderen Lauf hier grün.
func TestEinUnnormalisierterVektorWirdSoVermessen(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Schauweg {
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"capabilities":["embedding"],"license":"MIT",
					  "model_info":{"bert.pooling_type":1,
					                "tokenizer.ggml.model":"bert"}}`))

				return
			}

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"embeddings":[[3,4,0,0]]}`))
		}))
	t.Cleanup(server.Close)

	vertrag, _, fehler := klientFuer(server).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if vertrag.Normalisiert {
		t.Error("ein Vektor der Länge 5 gilt als normalisiert")
	}
}
