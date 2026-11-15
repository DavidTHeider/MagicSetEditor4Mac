package scrydex_test

// **Der Kontostand bei Scrydex** — Werkstück A5 von Phase 5.3
// (#1444).
//
// Die Gestalt ist am 2026-09-11 nachgelesen und steht in
// [0096](../../docs/entscheidungen/0096-welche-optionalen-datenquellen-zugaenglich-sind.md);
// das Beispiel darunter ist **selbst geschrieben** (§3).

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

// **Der Kontostand kommt an — und der Überzugssatz bleibt Text.**
//
// **Als Text und nicht als `float64`**: es ist ein Geldbetrag je Kredit, und
// `0.0004` ist in `float64` nicht `0.0004`. Wer damit rechnete, rechnete
// falsch — dieselbe Überlegung wie bei jedem Preis
// ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)).
func TestDerKontostandKommtAn(t *testing.T) {
	t.Parallel()

	server := kontengegenstelle(t, "kontostand.json")

	stand, fehler := klientFuerKonto(server).Kontostand(t.Context())
	if fehler != nil {
		t.Fatalf("Kontostand: %v", fehler)
	}

	if stand.Gesamt != 100000 || stand.Rest != 84213 || stand.Verbraucht != 15787 {
		t.Errorf("der Stand ist %+v", stand)
	}

	if stand.Ueberzugssatz != "0.0004" {
		t.Errorf("der Überzugssatz ist %q", stand.Ueberzugssatz)
	}
}

// **Der Weg steht fest und wird nicht geraten.**
func TestDerKontostandFragtDenDokumentiertenWeg(t *testing.T) {
	t.Parallel()

	gesehen := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			gesehen <- gestellt.URL.Path
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuerKonto(server).Kontostand(t.Context()); fehler != nil {
		t.Fatalf("Kontostand: %v", fehler)
	}

	if weg := <-gesehen; weg != scrydex.Kontoweg {
		t.Errorf("gefragt wurde %q, erwartet %q", weg, scrydex.Kontoweg)
	}
}

// **Ohne Zugang wird nicht gefragt** — dieselbe Zusage wie bei der Suche.
func TestOhneZugangKeinKontostand(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := scrydex.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()), scrydex.Zugang{})

	if _, fehler := klient.Kontostand(t.Context()); fehler == nil {
		t.Fatal("ohne Zugang kam ein Stand")
	}

	if gefragt {
		t.Error("der Server wurde ohne Zugang gefragt")
	}
}

// **Ein 401 ist ein Fehler und kein leerer Stand.**
func TestEinAbgelehnterKontostandIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusUnauthorized)

			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"error":{"message":"Invalid credentials","code":"INVALID_CREDENTIALS"}}`))
		}))
	t.Cleanup(server.Close)

	_, fehler := klientFuerKonto(server).Kontostand(t.Context())
	if fehler == nil {
		t.Fatal("ein 401 kam als Stand zurück")
	}

	// **Die Meldung des Anbieters reist mit** — „Invalid credentials" sagt,
	// was zu tun ist; „401" sagt es nicht.
	if !strings.Contains(fehler.Error(), "Invalid credentials") {
		t.Errorf("der Fehler ist %v", fehler)
	}
}

// **Und eine kaputte Antwort ebenso.**
func TestEinKaputterKontostandIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte("{nicht json")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuerKonto(server).Kontostand(t.Context()); fehler == nil {
		t.Fatal("eine kaputte Antwort kam als Stand zurück")
	}
}

// kontengegenstelle antwortet mit einer Datei aus `testdaten`.
func kontengegenstelle(t *testing.T, name string) *httptest.Server {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join("testdaten", name)) //nolint:gosec // ein Name aus diesem Test
	if fehler != nil {
		t.Fatalf("%s lesen: %v", name, fehler)
	}

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write(roh) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// klientFuerKonto baut einen Klienten mit Zugang und ohne Wartezeit.
func klientFuerKonto(server *httptest.Server) *scrydex.Klient {
	return scrydex.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()),
		scrydex.Zugang{Schluessel: "schluessel", Team: "team"}).
		MitDrossel(netz.NeueDrossel(0, time.Now, func(time.Duration) {}))
}
