package openai_test

// **Die Wege, die ein glücklicher Lauf nicht nimmt** — Werkstück A9 von
// Phase 5.3
// (#1448).
//
// # Warum sie eine eigene Datei haben
//
// **Weil sie sonst hinter dem Erfolgsweg verschwinden.** Eine fehlende
// Kopfzeile, ein Bild in einem Format, das niemand liest, eine Frage ohne
// Modell: jeder dieser Fälle ist eine **Auskunft**, und keiner entsteht
// dabei, dass man den guten Fall noch einmal prüft (§10).

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

// **Ohne Adresse gilt die Vorgabe — und dieser Lauf geht nicht hinaus.**
//
// **Ein Fuhrwerk statt des Netzes** (§3): der Lauf schreibt die gerufene
// Adresse auf und antwortet selbst. Wer ihn gegen `api.openai.com` fahren
// ließe, bekäme ein Ergebnis, weil es wirklich funktioniert hat.
func TestOhneAdresseGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{antwort: `{"data":[]}`}

	klient := openai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		openai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if fuhrwerk.adresse != "https://api.openai.com/v1/models" {
		t.Errorf("gerufen wurde %q", fuhrwerk.adresse)
	}

	// **Und die Aufschrift ist `Authorization: Bearer`** — nicht
	// `x-api-key` wie bei Anthropic.
	if fuhrwerk.aufschrift != "Bearer geheim" {
		t.Errorf("die Aufschrift ist %q", fuhrwerk.aufschrift)
	}
}

// mitschrift schreibt Adresse, Aufschrift und Rumpf auf und antwortet selbst.
type mitschrift struct {
	adresse    string
	aufschrift string
	rumpf      string
	antwort    string
}

func (m *mitschrift) RoundTrip(anfrage *http.Request) (*http.Response, error) {
	m.adresse = anfrage.URL.String()
	m.aufschrift = anfrage.Header.Get("Authorization")

	if anfrage.Body != nil {
		roh, _ := io.ReadAll(anfrage.Body) //nolint:errcheck // ein Puffer
		m.rumpf = string(roh)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(m.antwort)),
		Request:    anfrage,
	}, nil
}

// **Eine Frage ohne Schlüssel geht gar nicht erst hinaus.**
func TestOhneSchluesselGehtNichtsHinaus(t *testing.T) {
	t.Parallel()

	gerufen := 0

	server := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) { gerufen++ }))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{})

	if _, _, fehler := klient.Modelle(t.Context()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if gerufen != 0 {
		t.Errorf("%d Anfragen sind trotzdem hinausgegangen", gerufen)
	}
}

// **Eine unbrauchbare Frage wird abgewiesen, bevor sie Geld kostet.**
func TestEineUnbrauchbareFrageGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	gerufen := 0

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gerufen++

			_, _ = schreiben.Write([]byte(`{}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	for name, frage := range map[string]openai.Frage{
		"ohne Modell":  {Auftrag: "x"},
		"ohne Auftrag": {Alias: "m"},
		"leerer Alias": {Alias: "  ", Auftrag: "x"},
		"fremdes Bildformat": {
			Alias: "m", Auftrag: "x",
			Bilder: []openai.Bild{{Medientyp: "image/bmp", Daten: "AAAA"}},
		},
		"Bild ohne Daten": {
			Alias: "m", Auftrag: "x",
			Bilder: []openai.Bild{{Medientyp: "image/png"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, fehler := klient.Rede(t.Context(), frage); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Errorf("der Fehler ist %v", fehler)
			}
		})
	}

	if gerufen != 0 {
		t.Errorf("%d Anfragen sind trotzdem hinausgegangen", gerufen)
	}
}

// **Ein Bild reist vor dem Auftrag mit.**
func TestEinBildReistVorDemAuftrag(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{antwort: `{"status":"completed","model":"m",
	  "output":[{"type":"message","content":[{"type":"output_text","text":"{}"}]}]}`}

	klient := openai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		openai.Zugang{Schluessel: "geheim"})

	if _, fehler := klient.Rede(t.Context(), openai.Frage{
		Alias: "m", Auftrag: "beschreibe",
		Bilder: []openai.Bild{{Medientyp: "image/png", Daten: "AAAA"}},
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	bild := strings.Index(fuhrwerk.rumpf, "input_image")
	text := strings.Index(fuhrwerk.rumpf, "input_text")

	if bild < 0 || text < 0 || bild > text {
		t.Errorf("die Reihenfolge ist bild=%d text=%d: %s",
			bild, text, fuhrwerk.rumpf)
	}

	// **Als Datenadresse und nicht als Netzadresse**: sonst holte der
	// Anbieter selbst etwas, und wohin, entschiede ein Kartenfeld.
	if !strings.Contains(fuhrwerk.rumpf, "data:image/png;base64,AAAA") {
		t.Errorf("das Bild reist anders mit: %s", fuhrwerk.rumpf)
	}
}

// **Die Grenzen der letzten Antwort stehen am Klienten.**
func TestDieGrenzenDerLetztenAntwortStehenBereit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			kopf := schreiben.Header()
			kopf.Set("x-ratelimit-limit-requests", "500")
			kopf.Set("x-ratelimit-remaining-requests", "499")
			kopf.Set("x-ratelimit-reset-requests", "6m0s")
			kopf.Set("retry-after", "12")
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	grenzen := klient.Grenzen()
	if !grenzen.Anfragen.Genannt || grenzen.Anfragen.Grenze != 500 ||
		grenzen.Anfragen.Rest != 499 {
		t.Fatalf("die Grenzen sind %+v", grenzen)
	}

	// **Eine Dauer und kein Zeitpunkt** — das ist der Unterschied zu
	// Anthropic, und er ist leicht zu übersehen.
	if grenzen.Anfragen.Frisch != 6*time.Minute {
		t.Errorf("die Frist ist %v", grenzen.Anfragen.Frisch)
	}

	if grenzen.Warten != 12*time.Second {
		t.Errorf("gewartet werden soll %v", grenzen.Warten)
	}
}

// **Unlesbare Kopfzeilen sind keine Angabe.**
func TestUnlesbareKopfzeilenSindKeineAngabe(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			kopf := schreiben.Header()
			kopf.Set("x-ratelimit-limit-tokens", "viele")
			kopf.Set("x-ratelimit-reset-tokens", "bald")
			kopf.Set("retry-after", "gleich")
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	grenzen := klient.Grenzen()
	if grenzen.Zeichen.Genannt || grenzen.Zeichen.Frisch != 0 ||
		grenzen.Warten != 0 {
		t.Errorf("Unlesbares gilt als Angabe: %+v", grenzen)
	}
}

// **Eine kaputte Adresse wird vor dem Ruf abgewiesen.**
func TestEineKaputteAdresseWirdVorDemRufAbgewiesen(t *testing.T) {
	t.Parallel()

	klient := openai.NeuerKlient("://ohne-schema",
		netz.NeuerKlient(netz.UeblicheFristen()),
		openai.Zugang{Schluessel: "geheim"})

	_, _, fehler := klient.Modelle(t.Context())
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
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

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Transportfehler bleibt einer.**
func TestEinTransportfehlerWirdNichtGedeutet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) {}))
	server.Close()

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler == nil {
		t.Fatal("der Ruf gegen einen geschlossenen Server ist geglückt")
	}
}

// **Ein Schema, das sich nicht verpacken lässt, geht nicht hinaus.**
func TestEinUnverpackbaresSchemaGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	gerufen := 0

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gerufen++

			_, _ = schreiben.Write([]byte(`{}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	if _, fehler := klient.Rede(t.Context(), openai.Frage{
		Alias: "m", Auftrag: "x",
		Schema: map[string]any{"kanal": make(chan int)},
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if gerufen != 0 {
		t.Errorf("%d Anfragen sind trotzdem hinausgegangen", gerufen)
	}
}

// **Eine angekündigte Abschaltung steht im Befund.**
func TestEineAngekuendigteAbschaltungStehtImBefund(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"data":[{"id":"alt","object":"model","created":1,
				  "owned_by":"system","shutdown_date":"2027-01-01"}]}`))
		}))
	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"})

	_, befunde, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "2027-01-01") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Eine Drossel wartet vor der Anfrage.**
func TestDieDrosselWartetVorDerAnfrage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	var gewartet time.Duration

	drossel := netz.NeueDrossel(openai.Abstand, time.Now,
		func(dauer time.Duration) { gewartet += dauer })

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		openai.Zugang{Schluessel: "geheim"}).MitDrossel(drossel)

	for range 2 {
		if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
			t.Fatalf("Modelle: %v", fehler)
		}
	}

	if gewartet <= 0 {
		t.Error("zwischen zwei Anfragen wurde nicht gewartet")
	}
}
