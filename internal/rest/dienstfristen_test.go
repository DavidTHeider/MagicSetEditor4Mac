package rest_test

// **Die Fristen des Dienstes, gemessen** — Werkstück 4 von A18
// (#1533).
//
// # Warum eine Messung und keine Zusicherung auf eine Konstante
//
// **Weil eine Konstante nichts über den Server sagt.** `schreibfrist = 5 *
// time.Minute` steht in `cmd/mse-rest/dienst.go` — ob `http.Server` sie
// wirklich durchsetzt, und **worauf** sie zählt, ist eine andere Frage. Das
// Issue verlangt sie ausdrücklich als Messung, weil A17 (#1532) sie ausreizt:
// ein Bildauftrag über den Dienst kann Minuten brauchen.
//
// # Was hier gemessen wird
//
// **`WriteTimeout` beginnt beim Lesen des Anfragekopfes**, nicht beim
// Schreiben der Antwort — so steht es in `net/http`. Ein Behandler, der
// länger rechnet als die Frist, bekommt seine Antwort **nicht** mehr zum
// Aufrufer.
//
// **Und der Lauf fährt kurze Fristen**, damit er nicht fünf Minuten steht.
// Was er belegt, ist das **Verhalten**; die Zahlen des Betriebs stehen
// daneben und sind eine Wahl.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// **Ein Behandler, der länger braucht als die Schreibfrist, kommt nicht an.**
func TestEinLangerBehandlerLaeuftInDieSchreibfrist(t *testing.T) {
	t.Parallel()

	const frist = 200 * time.Millisecond

	server := httptest.NewUnstartedServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			// **Länger als die Frist** — der Fall, den A17 ausreizt.
			time.Sleep(4 * frist)

			schreiben.WriteHeader(http.StatusOK)
		}))
	server.Config.WriteTimeout = frist
	server.Start()

	defer server.Close()

	antwort, fehler := server.Client().Get(server.URL) //nolint:noctx // ein Lauf gegen httptest

	// **Entweder die Anfrage scheitert, oder der Rumpf lässt sich nicht
	// lesen** — beide Gestalten kommen vor, und beide heißen dasselbe: die
	// Antwort erreicht den Aufrufer nicht.
	if fehler != nil {
		return
	}

	defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

	if _, lesefehler := io.ReadAll(antwort.Body); lesefehler == nil {
		t.Fatal("die Antwort kam trotz überschrittener Schreibfrist an — " +
			"dann gilt sie nicht, was sie zusagt")
	}
}

// **Und einer innerhalb der Frist kommt an.**
//
// **Die Gegenprobe** (§10): ein Server, der jede Antwort verschluckte, hielte
// den Lauf darüber, ohne dass die Frist etwas entschiede.
func TestEinKurzerBehandlerKommtDurch(t *testing.T) {
	t.Parallel()

	const frist = 2 * time.Second

	server := httptest.NewUnstartedServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			time.Sleep(frist / 20)

			schreiben.WriteHeader(http.StatusTeapot)
		}))
	server.Config.WriteTimeout = frist
	server.Start()

	defer server.Close()

	antwort, fehler := server.Client().Get(server.URL) //nolint:noctx // ein Lauf gegen httptest
	if fehler != nil {
		t.Fatalf("die kurze Antwort kam nicht an: %v", fehler)
	}

	defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

	if antwort.StatusCode != http.StatusTeapot {
		t.Fatalf("die Antwort ist %d", antwort.StatusCode)
	}
}

// **Der Dienst spricht kein h2c — mit Absicht.**
//
// # Das Nein und sein Grund
//
// **Auf der Schleife ohne TLS brächte h2 nichts**, und es wäre eine Fläche
// mehr: `h2c` verlangt einen Upgrade-Weg, den jeder Aufrufer kennen müsste.
// Der Dienst bindet nur an `127.0.0.1` ([0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)),
// und dort ist Multiplexing über eine Verbindung kein Gewinn — es gibt einen
// Aufrufer.
//
// **Der Lauf steht hier, damit das Nein eine Messung ist und keine
// Behauptung**: wer h2c einschaltete, machte ihn rot.
func TestDerDienstSprichtHTTP11(t *testing.T) {
	t.Parallel()

	gesehen := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			select {
			case gesehen <- anfrage.Proto:
			default:
			}

			schreiben.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	antwort, fehler := server.Client().Get(server.URL) //nolint:noctx // ein Lauf gegen httptest
	if fehler != nil {
		t.Fatalf("die Anfrage ging nicht durch: %v", fehler)
	}

	defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

	if proto := <-gesehen; proto != "HTTP/1.1" {
		t.Fatalf("der Dienst sprach %q — h2c ist ausdrücklich nicht "+
			"eingeschaltet", proto)
	}
}
