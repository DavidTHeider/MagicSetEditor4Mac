package ollama_test

// **Die Fehlerwege der Vermessung** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// schauUndEinbettung baut eine Gegenstelle aus zwei festen Antworten.
func schauUndEinbettung(t *testing.T, schau, einbettung string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Schauweg {
				_, _ = schreiben.Write([]byte(schau)) //nolint:errcheck // die Gegenstelle

				return
			}

			_, _ = schreiben.Write([]byte(einbettung)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// derEinbetterschau ist eine Schau mit vollständigen Angaben.
const derEinbetterschau = `{"capabilities":["embedding"],"license":"MIT",
	"model_info":{"bert.pooling_type":1,"tokenizer.ggml.model":"bert"}}`

// **Ohne Zerteiler gibt es keinen Vertrag.**
func TestOhneZerteilerGibtEsKeinenVertrag(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(schauUndEinbettung(t,
		`{"capabilities":["embedding"],"model_info":{"bert.pooling_type":1}}`,
		`{"embeddings":[[1,0,0,0]]}`)).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) != 1 ||
		!strings.Contains(befunde[0].Meldung, "tokenizer.ggml.model") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Und ein Zerteiler, der kein Name ist, ebenso.**
func TestEinZerteilerOhneNamenGibtKeinenVertrag(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(schauUndEinbettung(t,
		`{"capabilities":["embedding"],
		  "model_info":{"bert.pooling_type":1,"tokenizer.ggml.model":7}}`,
		`{"embeddings":[[1,0,0,0]]}`)).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) != 1 ||
		!strings.Contains(befunde[0].Meldung, "kein Name") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Eine Zusammenfassung, die keine Zahl ist, zählt nicht als eine.**
//
// **Der Schlüssel steht da und der Wert taugt nicht** — das ist etwas anderes
// als ein fehlender Schlüssel und endet doch am selben Punkt: ohne
// Zusammenfassung kein Vertrag.
func TestEineZusammenfassungOhneZahlZaehltNicht(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(schauUndEinbettung(t,
		`{"capabilities":["embedding"],
		  "model_info":{"bert.pooling_type":"mittel",
		                "tokenizer.ggml.model":"bert"}}`,
		`{"embeddings":[[1,0,0,0]]}`)).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) != 1 ||
		!strings.Contains(befunde[0].Meldung, "pooling_type") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Zwei Vektoren auf einen Text sind ein Befund.**
func TestZweiVektorenAufEinenTextSindEinBefund(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(schauUndEinbettung(t, derEinbetterschau,
		`{"embeddings":[[1,0],[0,1]]}`)).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "2 Vektoren") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Ein leerer Vektor ergibt keinen gültigen Vertrag.**
//
// **Die Dimension wird gemessen** — und eine gemessene Null ist keine
// Dimension. Der Vertrag prüft sich selbst, bevor er herausgeht.
func TestEinLeererVektorErgibtKeinenVertrag(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(schauUndEinbettung(t, derEinbetterschau,
		`{"embeddings":[[]]}`)).Vermessen(t.Context(), "einbetter:300m")
	if fehler != nil {
		t.Fatalf("Vermessen: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "Dimension") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Eine Laufzeit, die beim Einbetten ausfällt, ist ein Fehler.**
//
// **Und kein Befund**: es kam gar nichts, und daraus folgt nichts über das
// Modell (§8).
func TestEinAusfallBeimEinbettenIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Schauweg {
				_, _ = schreiben.Write([]byte(derEinbetterschau)) //nolint:errcheck // die Gegenstelle

				return
			}

			schreiben.WriteHeader(http.StatusServiceUnavailable)
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).
		Vermessen(t.Context(), "einbetter:300m"); fehler == nil {
		t.Fatal("der Ausfall kam ohne Fehler durch")
	}
}
