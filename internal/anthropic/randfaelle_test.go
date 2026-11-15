package anthropic_test

// **Die Wege, die ein glücklicher Lauf nicht nimmt** — Werkstück A8 von
// Phase 5.3
// (#1447).
//
// # Warum sie eine eigene Datei haben
//
// **Weil sie sonst hinter dem Erfolgsweg verschwinden.** Eine fehlende
// Kopfzeile, ein unlesbares Datum, eine zweite Seite, ein Rumpf, den niemand
// verpacken kann: jeder dieser Fälle ist eine **Auskunft**, und keiner
// entsteht dabei, dass man den guten Fall noch einmal prüft (§10).

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Die Positivliste ist die des Vertrags.**
//
// **Was er nicht hält**: dass diese zwei Wege gerufen werden. Das prüft der
// Spion in `positivliste_test.go`; hier steht nur, dass die Aufzählung nicht
// still wächst.
func TestDiePositivlisteNenntZweiWege(t *testing.T) {
	t.Parallel()

	wege := anthropic.Wege()
	if len(wege) != 2 {
		t.Fatalf("die Positivliste führt %v", wege)
	}

	for _, weg := range wege {
		if !strings.HasPrefix(weg, "/v1/") {
			t.Errorf("%q sieht nicht wie ein Weg dieser Schnittstelle aus", weg)
		}
	}
}

// **Ohne Adresse gilt die Vorgabe.**
//
// # Und dieser Lauf geht nicht hinaus
//
// **Der erste Anlauf tat es**, und er war grün: er baute den Klienten ohne
// Adresse, rief `Modelle` und las die Adresse aus dem Fehlergrund — den
// `api.anthropic.com` geschickt hatte, mit einem echten 401 auf einen
// erfundenen Schlüssel. Genau die Gestalt, vor der §3 warnt: „ein Test, der
// nach draußen greift, ohne dass jemand es wollte".
//
// Hier hängt deshalb ein **Fuhrwerk** unter dem Klienten, das die Adresse
// aufschreibt und selbst antwortet. Es erreicht kein Netz, und es belegt
// mehr als vorher: die volle Adresse samt Weg und Fragenteil.
func TestOhneAdresseGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{}

	klient := anthropic.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		anthropic.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if !strings.HasPrefix(fuhrwerk.adresse, "https://api.anthropic.com/v1/models?") {
		t.Errorf("gerufen wurde %q", fuhrwerk.adresse)
	}
}

// mitschrift schreibt die gerufene Adresse auf und antwortet selbst.
type mitschrift struct{ adresse string }

func (m *mitschrift) RoundTrip(anfrage *http.Request) (*http.Response, error) {
	m.adresse = anfrage.URL.String()

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"data":[]}`)),
		Request:    anfrage,
	}, nil
}

// **Die Grenzen der letzten Antwort stehen am Klienten.**
func TestDieGrenzenDerLetztenAntwortStehenBereit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			kopf := schreiben.Header()
			kopf.Set("anthropic-ratelimit-requests-limit", "1000")
			kopf.Set("anthropic-ratelimit-requests-remaining", "999")
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if grenzen := klient.Grenzen(); !grenzen.Anfragen.Genannt ||
		grenzen.Anfragen.Grenze != 1000 || grenzen.Anfragen.Rest != 999 {
		t.Errorf("die Grenzen sind %+v", grenzen)
	}
}

// **Eine Kopfzeile, die keine Zahl ist, ist keine Angabe.**
//
// **Und das ist nicht dasselbe wie „null übrig"**: der Unterschied zwischen
// »gleich gesperrt« und »weiß nicht« ist die ganze Aussage.
func TestEineUnlesbareKopfzeileIstKeineAngabe(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			kopf := schreiben.Header()
			// **Beide unlesbar**, denn `Genannt` ist ein Oder: eine
			// lesbare Restzahl neben einer unlesbaren Grenze ist sehr wohl
			// eine Angabe.
			kopf.Set("anthropic-ratelimit-requests-limit", "viele")
			kopf.Set("anthropic-ratelimit-requests-remaining", "")
			kopf.Set("retry-after", "bald")
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	grenzen := klient.Grenzen()
	if grenzen.Anfragen.Genannt {
		t.Errorf("»viele« gilt als Zahl: %+v", grenzen.Anfragen)
	}

	if grenzen.Warten != 0 {
		t.Errorf("»bald« gilt als Frist: %v", grenzen.Warten)
	}
}

// **Eine zweite Seite wird gemeldet und nicht geholt.**
//
// Eine halbe Liste, die wie eine ganze aussieht, ist teurer als eine
// Meldung — und dieser Lauf ist der Beleg, dass die Meldung kommt.
func TestEineZweiteSeiteWirdGemeldet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte(`{"data":[],` + //nolint:errcheck // die Gegenstelle
				`"has_more":true,"last_id":"claude-letztes"}`))
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	_, befunde, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "claude-letztes") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	if befunde[0].Art != domain.ArtUmgebung {
		t.Errorf("der Befund ist %q", befunde[0].Art)
	}
}

// **Ein unlesbares Datum wird leer und nicht erfunden.**
//
// Ohne Datum fällt das Modell an [domain.KIGeprueft.Validate] durch — und
// genau das soll es: ein erfundener Tag sähe aus wie ein gemessener.
func TestEinUnlesbaresDatumWirdLeer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte(`{"data":[{ ` + //nolint:errcheck // die Gegenstelle
				`"id":"claude-probe","display_name":"Probe",` +
				`"created_at":"neulich",` +
				`"capabilities":{"structured_outputs":{"supported":true}}}]}`))
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	modelle, befunde, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(modelle) != 0 {
		t.Fatalf("ein Modell ohne Datum steht im Katalog: %+v", modelle)
	}

	if len(befunde) == 0 {
		t.Fatal("das Modell fällt aus und niemand sagt es")
	}
}

// **Eine Frage ohne Modell und eine ohne Auftrag werden abgewiesen.**
func TestEineUnbrauchbareFrageGehtGarNichtHinaus(t *testing.T) {
	t.Parallel()

	gerufen := 0

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gerufen++

			_, _ = schreiben.Write([]byte(`{}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	for name, frage := range map[string]anthropic.Frage{
		"ohne Modell":   {Auftrag: "schreib etwas"},
		"ohne Auftrag":  {Alias: "claude-probe"},
		"leerer Alias":  {Alias: "   ", Auftrag: "schreib etwas"},
		"leerer Auftra": {Alias: "claude-probe", Auftrag: "  "},
	} {
		t.Run(name, func(t *testing.T) {
			if _, fehler := klient.Rede(t.Context(), frage); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("der Fehler ist %v", fehler)
			}
		})
	}

	// **Die Gegenprobe zur Zusage**: abgewiesen heißt, dass nichts hinausging.
	if gerufen != 0 {
		t.Errorf("%d Anfragen sind trotzdem hinausgegangen", gerufen)
	}
}

// **Eine lesbare Wartefrist wird gelesen.**
//
// **Die Gegenprobe steht oben**: „bald" ist keine. Ohne diesen Lauf prüfte
// nur der Fehlerweg, und „0" sähe dann aus wie „gelesen und null".
func TestEineLesbareWartefristWirdGelesen(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("retry-after", "42")
			schreiben.WriteHeader(http.StatusTooManyRequests)
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"error":{"type":"rate_limit_error","message":"x"}}`))
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if warten := klient.Grenzen().Warten; warten != 42*time.Second {
		t.Errorf("gewartet werden soll %v", warten)
	}
}

// **Eine Adresse, die keine ist, wird vor dem Ruf abgewiesen.**
//
// **Und zwar als Validierung**: der Fehler liegt in der Einstellung, nicht
// beim Dienst — wer eine kaputte Adresse einträgt, soll nicht lesen, dass
// Anthropic abgelehnt hat (§8).
func TestEineKaputteAdresseWirdVorDemRufAbgewiesen(t *testing.T) {
	t.Parallel()

	klient := anthropic.NeuerKlient("://ohne-schema",
		netz.NeuerKlient(netz.UeblicheFristen()),
		anthropic.Zugang{Schluessel: "geheim"})

	_, _, fehler := klient.Modelle(t.Context())
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "/v1/models") {
		t.Errorf("die Meldung nennt den Weg nicht: %v", fehler)
	}
}

// **Ein Transportfehler bleibt ein Transportfehler.**
//
// Er wird **nicht** in einen Dienstfehler übersetzt: es gab keine Antwort,
// über die sich etwas sagen ließe.
func TestEinTransportfehlerWirdNichtGedeutet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) {}))
	server.Close()

	klient := anthropic.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler == nil {
		t.Fatal("der Ruf gegen einen geschlossenen Server ist geglückt")
	}
}

// **Ein Rumpf, der kein JSON ist, ist ein Fehler und keine leere Liste.**
func TestEinUnlesbarerRumpfIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte("kein JSON")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	_, _, fehler := klient.Modelle(t.Context())
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	// **Der Grund nennt den Weg** — sonst weiß niemand, welche der zwei
	// Antworten unlesbar war.
	if !strings.Contains(fehler.Error(), "/v1/models") {
		t.Errorf("die Meldung nennt den Weg nicht: %v", fehler)
	}
}

// **Ein Schema, das sich nicht verpacken lässt, geht nicht hinaus.**
//
// Ein Kanal in einem `map[string]any` ist kein JSON. Die Frage wird
// abgewiesen, **bevor** eine Anfrage entsteht — sonst schickte dieses
// Programm einen halben Rumpf.
func TestEinUnverpackbaresSchemaGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	gerufen := 0

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gerufen++

			_, _ = schreiben.Write([]byte(`{}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	_, fehler := klient.Rede(t.Context(), anthropic.Frage{
		Alias: "claude-probe", Auftrag: "schreib etwas",
		Schema: map[string]any{"kanal": make(chan int)},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if gerufen != 0 {
		t.Errorf("%d Anfragen sind trotzdem hinausgegangen", gerufen)
	}
}
