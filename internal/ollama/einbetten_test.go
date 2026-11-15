package ollama_test

// **Die Einbettung** — Werkstück A7 von Phase 5.3
// (#1446),
// Ergänzung vom 2026-09-10.
//
// Die Gestalt von `/api/embed` ist am 2026-09-11 unter
// <https://docs.ollama.com/api/embed> nachgelesen; die Beispiele darunter sind
// **selbst geschrieben** (§3).

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// derProbevertrag ist der Vertrag, gegen den die Läufe hier prüfen.
func derProbevertrag() domain.KIEinbettungsvertrag {
	return domain.KIEinbettungsvertrag{
		Dimension: 4, Pooling: domain.KIPoolingMittel,
		Normalisiert: true, Zerteiler: "bert",
	}
}

// **Ein Stapel kommt als Vektoren zurück.**
func TestEinStapelKommtAlsVektorenZurueck(t *testing.T) {
	t.Parallel()

	einbetter := klientFuer(einbettende(t, nil)).
		Einbetter(domain.Vektorraum{Modell: "einbetter:300m", Vertrag: derProbevertrag()})

	vektoren, fehler := einbetter.Einbetten(t.Context(),
		[]string{"Blitzschlag", "Gegenzauber"})
	if fehler != nil {
		t.Fatalf("Einbetten: %v", fehler)
	}

	if len(vektoren) != 2 || len(vektoren[0]) != 4 {
		t.Fatalf("die Vektoren sind %v", vektoren)
	}
}

// **Lange Eingaben werden abgelehnt, nicht gekürzt.**
//
// # Warum das der Kern ist
//
// **Weil Ollama sie sonst still kürzt.** `truncate` steht auf `true`, wenn man
// nichts sagt — „If true, truncate inputs that exceed the context window. If
// false, returns an error" (**nachgelesen am 2026-09-11**). Ein gekürzter Text
// gibt einen Vektor, der zu einem **anderen** Text gehört; nichts scheitert,
// und der Index zeigt woandershin. Das Issue verlangt wörtlich: „lange
// Eingaben explizit aufteilen/ablehnen statt unbemerkt kürzen."
func TestLangeEingabenWerdenAbgelehntStattGekuerzt(t *testing.T) {
	t.Parallel()

	gesehen := make(chan []byte, 1)

	server := einbettende(t, gesehen)

	if _, fehler := klientFuer(server).
		Einbetter(domain.Vektorraum{Modell: "einbetter:300m", Vertrag: derProbevertrag()}).
		Einbetten(t.Context(), []string{"Blitzschlag", "Gegenzauber"}); fehler != nil {
		t.Fatalf("Einbetten: %v", fehler)
	}

	var gestellt map[string]any

	if fehler := json.Unmarshal(<-gesehen, &gestellt); fehler != nil {
		t.Fatalf("den Rumpf lesen: %v", fehler)
	}

	if gestellt["truncate"] != false {
		t.Errorf("»truncate« ist %v — Ollama kürzt dann still", gestellt["truncate"])
	}
}

// **Ein zu großer Stapel wird abgewiesen, bevor er hinausgeht.**
func TestEinZuGrosserStapelWirdVorherAbgewiesen(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	zuviele := make([]string, ollama.HoechstensJeStapel+1)
	for i := range zuviele {
		zuviele[i] = fmt.Sprintf("Karte %d", i)
	}

	if _, fehler := klientFuer(server).
		Einbetter(domain.Vektorraum{Modell: "einbetter:300m", Vertrag: derProbevertrag()}).
		Einbetten(t.Context(), zuviele); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if gefragt {
		t.Error("der zu große Stapel ging trotzdem hinaus")
	}
}

// **Ein leerer Stapel fragt gar nicht erst.**
func TestEinLeererStapelFragtNicht(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	vektoren, fehler := klientFuer(server).
		Einbetter(domain.Vektorraum{Modell: "einbetter:300m", Vertrag: derProbevertrag()}).
		Einbetten(t.Context(), nil)
	if fehler != nil {
		t.Fatalf("Einbetten: %v", fehler)
	}

	if len(vektoren) != 0 || gefragt {
		t.Errorf("%d Vektoren, gefragt: %v", len(vektoren), gefragt)
	}
}

// **Zu wenige Vektoren sind ein Fehler und keine kurze Liste.**
//
// **Sonst verschöbe sich die Zuordnung**: der zweite Vektor gehörte zum
// ersten Text, und niemand sähe es.
func TestZuWenigeVektorenSindEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"model":"einbetter:300m","embeddings":[[0.5,0.5,0.5,0.5]]}`))
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuer(server).
		Einbetter(domain.Vektorraum{Modell: "einbetter:300m", Vertrag: derProbevertrag()}).
		Einbetten(t.Context(), []string{"eins", "zwei"}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Eine falsche Dimension ebenso.**
func TestEineFalscheDimensionIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"model":"einbetter:300m","embeddings":[[0.5,0.5]]}`))
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuer(server).
		Einbetter(domain.Vektorraum{Modell: "einbetter:300m", Vertrag: derProbevertrag()}).
		Einbetten(t.Context(), []string{"eins"}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Die Anweisung des Vertrags geht mit hinaus.**
func TestDieAnweisungDesVertragsGehtMit(t *testing.T) {
	t.Parallel()

	gesehen := make(chan []byte, 1)

	server := einbettende(t, gesehen)

	vertrag := derProbevertrag()
	vertrag.Frageanweisung = "Frage: "

	if _, fehler := klientFuer(server).
		Einbetter(domain.Vektorraum{Modell: "einbetter:300m", Vertrag: vertrag}).
		Einbetten(t.Context(), []string{vertrag.FuerFrage("Blitzschlag")}); fehler != nil {
		t.Fatalf("Einbetten: %v", fehler)
	}

	if rumpf := string(<-gesehen); !strings.Contains(rumpf, "Frage: Blitzschlag") {
		t.Errorf("der Rumpf ist %s", rumpf)
	}
}

// einbettende baut eine Gegenstelle für `/api/show` und `/api/embed`.
//
// **Der Rumpf wird durchgereicht**, wo ein Kanal übergeben ist.
func einbettende(t *testing.T, gesehen chan []byte) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			roh := make([]byte, 16384)
			gelesen, _ := gestellt.Body.Read(roh) //nolint:errcheck // die Gegenstelle

			if gestellt.URL.Path == ollama.Schauweg {
				schreibeDatei(t, schreiben, "show-einbetter.json")

				return
			}

			if gesehen != nil {
				gesehen <- roh[:gelesen]
			}

			schreibeVektoren(t, schreiben, string(roh[:gelesen]))
		}))

	t.Cleanup(server.Close)

	return server
}

// schreibeVektoren gibt je Eingabe einen normalisierten Vektor.
func schreibeVektoren(t *testing.T, schreiben http.ResponseWriter, rumpf string) {
	t.Helper()

	var gestellt struct {
		Eingaben []string `json:"input"`
	}

	if fehler := json.Unmarshal([]byte(rumpf), &gestellt); fehler != nil {
		t.Fatalf("die Eingabe lesen: %v", fehler)
	}

	vektoren := make([][]float32, 0, len(gestellt.Eingaben))

	for _, text := range gestellt.Eingaben {
		// **Verschiedene Texte geben verschiedene Vektoren**, und derselbe
		// Text immer denselben — sonst prüfte der Vertragslauf aus A6 nichts.
		quer := float32(0)
		for _, zeichen := range text {
			quer += float32(zeichen)
		}

		betrag := float32(math.Sqrt(float64(quer*quer + 1)))
		vektoren = append(vektoren, []float32{quer / betrag, 1 / betrag, 0, 0})
	}

	antwort, fehler := json.Marshal(map[string]any{
		"model": "einbetter:300m", "embeddings": vektoren,
	})
	if fehler != nil {
		t.Fatalf("die Antwort bauen: %v", fehler)
	}

	_, _ = schreiben.Write(antwort) //nolint:errcheck // die Gegenstelle
}
