package xai_test

// **Die Wege, die ein glücklicher Lauf nicht nimmt** — Werkstück A10 von
// Phase 5.3
// (#1449).
//
// # Warum sie eine eigene Datei haben
//
// **Weil sie sonst hinter dem Erfolgsweg verschwinden.** Ein fehlender
// Schlüssel, eine Adresse, die keine ist, ein Schema, das sich nicht
// schreiben lässt: jeder dieser Fälle ist eine **Auskunft**, und keiner
// entsteht dabei, dass man den guten Fall noch einmal prüft (§10).

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// **Ohne Adresse gilt die Vorgabe — und dieser Lauf geht nicht hinaus.**
//
// **Ein Fuhrwerk statt des Netzes** (§3): der Lauf schreibt die gerufene
// Adresse auf und antwortet selbst. Wer ihn gegen `api.x.ai` fahren ließe,
// bekäme ein Ergebnis, weil es wirklich funktioniert hat.
func TestOhneAdresseGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{antwort: `{"models":[]}`}

	klient := xai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		xai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	// **`language-models` und nicht `models`** — der zweite Weg gibt es auch
	// und nennt die Modalitäten nicht.
	if fuhrwerk.adresse != "https://api.x.ai/v1/language-models" {
		t.Errorf("gerufen wurde %q", fuhrwerk.adresse)
	}

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

	klient := xai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		xai.Zugang{})

	if _, _, fehler := klient.Modelle(t.Context()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if gerufen != 0 {
		t.Errorf("%d Anfragen sind trotzdem hinausgegangen", gerufen)
	}
}

// **Und der Zugang sagt, was fehlt.**
func TestDerZugangSagtWasFehlt(t *testing.T) {
	t.Parallel()

	if fehlt := (xai.Zugang{}).Fehlt(); !strings.Contains(fehlt, "xAI") {
		t.Errorf("es fehlt %q", fehlt)
	}

	if fehlt := (xai.Zugang{Schluessel: "x"}).Fehlt(); fehlt != "" {
		t.Errorf("es fehlt %q, obwohl der Schlüssel dasteht", fehlt)
	}
}

// **Eine Adresse, die keine ist, wird vor dem Ruf bemerkt.**
func TestEineAdresseDieKeineIstWirdVorDemRufBemerkt(t *testing.T) {
	t.Parallel()

	klient := xai.NeuerKlient("://kaputt", netz.KlientUeber(http.DefaultClient),
		xai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ein Schema, das sich nicht schreiben lässt, kostet kein Geld.**
//
// **Der Fall ist nicht erfunden**: das Schema kommt aus `service` und damit
// aus einer Kartenvorlage. Wer es hinausschickte, ohne es schreiben zu
// können, bekäme einen halben Rumpf.
func TestEinUnschreibbaresSchemaGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	fuhrwerk := &mitschrift{}

	klient := xai.NeuerKlient("",
		netz.KlientUeber(&http.Client{Transport: fuhrwerk}),
		xai.Zugang{Schluessel: "geheim"})

	_, fehler := klient.Rede(t.Context(), xai.Frage{
		Alias: "m", Auftrag: "x",
		Schema: map[string]any{"type": make(chan int)},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if fuhrwerk.adresse != "" {
		t.Errorf("gerufen wurde trotzdem %q", fuhrwerk.adresse)
	}
}

// **Ein Dienst, der gar nicht antwortet, wird als Fehler weitergereicht.**
func TestEinStummerDienstWirdWeitergereicht(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			// **Die Verbindung fällt mitten im Satz** — das ist der Fall,
			// den ein Status nicht beschreibt.
			entführer, geht := schreiben.(http.Hijacker)
			if !geht {
				t.Error("die Gegenstelle lässt sich nicht unterbrechen")

				return
			}

			verbindung, _, fehler := entführer.Hijack()
			if fehler == nil {
				_ = verbindung.Close() //nolint:errcheck // sie stirbt hier
			}
		}))
	t.Cleanup(server.Close)

	klient := xai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		xai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler == nil {
		t.Fatal("der abgebrochene Ruf ist als Erfolg durchgegangen")
	}
}

// **Die Drossel wartet vor der Anfrage.**
//
// **Gemessen und nicht behauptet**: die Wartezeit wird eingesetzt und
// aufgeschrieben. Ohne diese Zusicherung wäre `MitDrossel` eine Zeile, die
// gesetzt und nie gelesen wird (§10).
func TestDieDrosselWartetVorDerAnfrage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte(`{"models":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	var gewartet time.Duration

	drossel := netz.NeueDrossel(xai.Abstand, time.Now,
		func(dauer time.Duration) { gewartet += dauer })

	klient := xai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		xai.Zugang{Schluessel: "geheim"}).MitDrossel(drossel)

	for range 2 {
		if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
			t.Fatalf("Modelle: %v", fehler)
		}
	}

	if gewartet <= 0 {
		t.Error("zwischen zwei Anfragen wurde nicht gewartet")
	}
}

// **Und der Abstand der Vorgabe ist gesetzt.**
//
// **Vorsichtig und nicht geraten**: xAI nennt keine Zahl (§16, Regel 4).
func TestDerAbstandDerVorgabeIstGesetzt(t *testing.T) {
	t.Parallel()

	if xai.Abstand != time.Second {
		t.Errorf("der Abstand ist %v", xai.Abstand)
	}
}

// **Ein abgebrochener Ruf endet als Fehler und nicht als leere Antwort.**
//
// **Der Fall ist der Normalfall und nicht der Ausnahmefall**: eine
// Kartenvorschau wird abgebrochen, während jemand tippt (§7), und ein Ruf an
// ein Denkmodell dauert. Wer den Abbruch als „das Modell hat nichts gesagt"
// läse, schriebe ein leeres Feld.
//
// **Die Gegenstelle wartet eine feste Zeit und nicht auf den Abbruch.** Ein
// Handler, der am Kontext hängt, hält `server.Close` fest, wenn der Ruf ihn
// nicht erreicht — und ein Lauf, der hängt, hält die ganze Stufe an. Das ist
// am 2026-09-12 einmal passiert.
func TestEinAbgebrochenerRufIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			time.Sleep(300 * time.Millisecond)

			_, _ = schreiben.Write([]byte(`{"models":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := xai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		xai.Zugang{Schluessel: "geheim"})

	ctx, abbrechen := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer abbrechen()

	_, fehler := klient.Rede(ctx, xai.Frage{Alias: "m", Auftrag: "x"})
	if fehler == nil {
		t.Fatal("der abgebrochene Ruf ist als Erfolg durchgegangen")
	}

	if !errors.Is(fehler, context.DeadlineExceeded) {
		t.Errorf("der Fehler nennt den Abbruch nicht: %v", fehler)
	}
}

// **Eine Antwort über der Grenze wird abgewiesen, nicht gelesen.**
//
// # Warum das mehr ist als Sparsamkeit
//
// **Weil die Gegenstelle bestimmt, wie viel sie schickt.** Ohne Grenze
// entscheidet ein fremder Dienst, wie viel Arbeitsspeicher dieses Programm
// belegt — und eine Antwort, die nicht aufhört, hält das Fenster an.
func TestEineZuGrosseAntwortWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			// **Ein Stück über der Grenze** — 4 MiB plus Rahmen.
			fuellung := strings.Repeat("x", 4<<20)

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"models":[],"fuellung":"` + fuellung + `"}`))
		}))
	t.Cleanup(server.Close)

	klient := xai.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		xai.Zugang{Schluessel: "geheim"})

	if _, _, fehler := klient.Modelle(t.Context()); fehler == nil {
		t.Fatal("die zu große Antwort ist durchgegangen")
	}
}
