package ollama_test

// **Der Vektorraum eines Modells** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Wogegen hier geprüft wird
//
// **„Gleich langer Vektor bedeutet nicht gleicher Vektorraum."** Ein Raum
// nennt Modell, Digest, Laufzeit und den gemessenen Vertrag; fehlt einer
// davon, ist er keiner — und dann gibt es einen Befund und kein halbes
// Ergebnis.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// **Ein vermessenes Modell nennt seinen vollständigen Raum.**
func TestEinVermessenesModellNenntSeinenRaum(t *testing.T) {
	t.Parallel()

	raum, befunde, fehler := klientFuer(gegenstelle(t)).Raum(
		t.Context(), "einbetter:300m")
	if fehler != nil || len(befunde) > 0 {
		t.Fatalf("Raum: %v, %+v", fehler, befunde)
	}

	if raum.Modell != "einbetter:300m" ||
		raum.Modellrevision != "aaaa396dbd5992bbe3f5cdb947e8bbc0ee413d7c17e2beaae69f5d569cf98aaaa" ||
		raum.Artefakthash != raum.Modellrevision {
		t.Errorf("der Raum ist %+v", raum)
	}

	if raum.Laufzeit != "ollama 0.12.3" {
		t.Errorf("die Laufzeit ist %q", raum.Laufzeit)
	}

	// **Das Abstandsmaß folgt dem Vertrag** und ist keine Konvention.
	if raum.Vertrag.Normalisiert != (raum.Abstandsmass == domain.AbstandCosinus) {
		t.Errorf("%v bei Normalisierung %v",
			raum.Abstandsmass, raum.Vertrag.Normalisiert)
	}

	// **Und der Einbetter rechnet in genau diesem Raum.**
	if !klientFuer(gegenstelle(t)).Einbetter(raum).Raum().PasstZu(raum) {
		t.Error("der Einbetter kennt seinen Raum nicht wieder")
	}
}

// **Unnormalisierte Vektoren werden mit dem Skalarprodukt verglichen.**
//
// **Nicht mit dem Cosinus**, und das ist keine Feinheit: der Cosinus
// dividiert die Länge heraus, und bei unnormalisierten Vektoren trägt sie
// Bedeutung. Dieselben Zahlen ergeben zwei verschiedene Trefferlisten.
func TestUnnormalisierteVektorenNehmenDasSkalarprodukt(t *testing.T) {
	t.Parallel()

	// **Ein Modell, dessen Vektoren nicht auf Länge eins liegen** — gemessen
	// und nicht behauptet: [ollama.Klient.Vermessen] rechnet nach.
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			switch gestellt.URL.Path {
			case ollama.Fassungsweg:
				_, _ = schreiben.Write([]byte(`{"version":"0.12.3"}`)) //nolint:errcheck // die Gegenstelle
			case ollama.Modellweg:
				schreibeDatei(t, schreiben, "tags.json")
			case ollama.Schauweg:
				schreibeDatei(t, schreiben, "show-einbetter.json")
			default:
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"embeddings":[[3,4,0,0]]}`))
			}
		}))
	t.Cleanup(server.Close)

	raum, befunde, fehler := ollama.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client())).Raum(t.Context(), "einbetter:300m")
	if fehler != nil || len(befunde) > 0 {
		t.Fatalf("Raum: %v, %+v", fehler, befunde)
	}

	if raum.Vertrag.Normalisiert {
		t.Fatalf("der Vertrag behauptet Normalisierung: %+v", raum.Vertrag)
	}

	if raum.Abstandsmass != domain.AbstandPunkt {
		t.Errorf("das Abstandsmaß ist %q", raum.Abstandsmass)
	}
}

// **Ein Modell ohne Digest gibt keinen Raum.**
//
// **Sondern einen Befund.** Ohne den Abdruck der Gewichte ist nicht zu sagen,
// was gerechnet hat — und ein Index, dessen Manifest das offenlässt, lässt
// sich später mit nichts vergleichen.
func TestEinModellOhneDigestGibtKeinenRaum(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(gegenstelle(t)).Raum(
		t.Context(), "ohne-digest:1b")
	if fehler != nil {
		t.Fatalf("Raum: %v", fehler)
	}

	if len(befunde) == 0 {
		t.Fatal("ohne Digest kam kein Befund")
	}
}

// **Ein Modell, das die Laufzeit gar nicht führt, gibt keinen Raum.**
func TestEinUnbekanntesModellGibtKeinenRaum(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(gegenstelle(t)).Raum(
		t.Context(), "einbetter:300m-anders")
	if fehler != nil {
		t.Fatalf("Raum: %v", fehler)
	}

	if len(befunde) != 1 ||
		!strings.Contains(befunde[0].Meldung, "installierten Modelle") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Ein Modell ohne Einbettungsfähigkeit gibt keinen Raum.**
func TestEinModellOhneEinbettungGibtKeinenRaum(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := klientFuer(gegenstelle(t)).Raum(t.Context(), "gemma4:8b")
	if fehler != nil {
		t.Fatalf("Raum: %v", fehler)
	}

	if len(befunde) == 0 {
		t.Fatal("ein Textmodell kam als Einbetter durch")
	}
}

// **Antwortet die Laufzeit nicht, gibt es keinen Raum und keinen Befund.**
//
// **Der Unterschied gehört dem Aufrufer** (§10): ein Befund sagt „dieses
// Modell taugt nicht", ein Fehler sagt „es kam gar nichts".
func TestOhneLaufzeitGibtEsKeinenRaum(t *testing.T) {
	t.Parallel()

	for name, antwort := range map[string]func(http.ResponseWriter, *http.Request){
		"keine Fassung": func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Fassungsweg {
				schreiben.WriteHeader(http.StatusInternalServerError)

				return
			}

			antwortenWieGewohnt(t, schreiben, gestellt)
		},
		"keine Modellliste": func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Modellweg {
				schreiben.WriteHeader(http.StatusInternalServerError)

				return
			}

			antwortenWieGewohnt(t, schreiben, gestellt)
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(antwort))
			t.Cleanup(server.Close)

			if _, _, fehler := ollama.NeuerKlient(server.URL,
				netz.KlientUeber(server.Client())).Raum(t.Context(), "einbetter:300m"); fehler == nil {
				t.Fatal("die schweigende Laufzeit kam durch")
			}
		})
	}
}

// antwortenWieGewohnt gibt die üblichen Antworten.
func antwortenWieGewohnt(
	t *testing.T, schreiben http.ResponseWriter, gestellt *http.Request,
) {
	t.Helper()

	switch gestellt.URL.Path {
	case ollama.Fassungsweg:
		_, _ = schreiben.Write([]byte(`{"version":"0.12.3"}`)) //nolint:errcheck // die Gegenstelle
	case ollama.Modellweg:
		schreibeDatei(t, schreiben, "tags.json")
	case ollama.Schauweg:
		schreibeDatei(t, schreiben, "show-einbetter.json")
	default:
		rumpf := make([]byte, 16384)
		gelesen, _ := gestellt.Body.Read(rumpf) //nolint:errcheck // die Gegenstelle

		schreibeVektoren(t, schreiben, string(rumpf[:gelesen]))
	}
}

// **Ein Name ohne Marke meint `latest`.**
//
// # Woher die Regel kommt
//
// **Gemessen und nicht überlegt** (2026-09-12, Ollama 0.33.3): `/api/show`
// und `/api/embed` nehmen `all-minilm` entgegen, `/api/tags` führt dasselbe
// Modell als `all-minilm:latest`. Der erste Messlauf bekam dafür einen
// Befund — für ein Modell, das gerade gerechnet hatte.
func TestEinNameOhneMarkeMeintLatest(t *testing.T) {
	t.Parallel()

	// **Eine eigene Modellliste**, weil die gemeinsame Vorrichtung gegen eine
	// feste Aufzählung geprüft wird; ein Eintrag mehr darin machte einen
	// anderen Lauf rot.
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Modellweg {
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"models":[{"name":"mitlatest:latest",` +
						`"digest":"cccc396dbd5992bbe3f5cdb947e8bbc0ee413d7c17e2beaae69f5d569cf98ccc",` +
						`"details":{"family":"bert"}}]}`))

				return
			}

			antwortenWieGewohnt(t, schreiben, gestellt)
		}))
	t.Cleanup(server.Close)

	klient := ollama.NeuerKlient(server.URL, netz.KlientUeber(server.Client()))

	raum, befunde, fehler := klient.Raum(t.Context(), "mitlatest")
	if fehler != nil || len(befunde) > 0 {
		t.Fatalf("Raum: %v, %+v", fehler, befunde)
	}

	if raum.Modell != "mitlatest" {
		t.Errorf("der Raum nennt %q", raum.Modell)
	}

	// **Und eine andere Marke ist ein anderes Modell.** `gemma3:270m` und
	// `gemma3:latest` sind nicht dasselbe; wer die Marke wegließe, verglich
	// Gewichte, die er nie gesehen hat.
	_, befunde, fehler = klient.Raum(t.Context(), "mitlatest:alt")
	if fehler != nil || len(befunde) == 0 {
		t.Error("eine fremde Marke kam durch")
	}
}
