package netz

// **Protokoll und Fristen des Transports** — Werkstück A18 von Phase 5.3
// (#1533),
// die Kopffrist mit A17
// (#1532).
//
// # Warum eine eigene Datei
//
// **Weil `transporteigenschaften_intern_test.go` über die Zeilengrenze aus §2
// ging.** Der Schnitt folgt der Frage: dort steht, **wohin** eine Anfrage
// geht — der Proxy —, hier, **wie** sie geht und **wie lange** sie darf.

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// **Gegen einen Server, der HTTP/2 spricht, spricht der Klient HTTP/2.**
//
// # Der Befund
//
// **Er sprach HTTP/1.1.** Go schaltet HTTP/2 bei einem eigenen `DialContext`
// nur mit `ForceAttemptHTTP2` zu, und das stand nicht da.
func TestGegenEinenHTTP2ServerSprichtDerKlientHTTP2(t *testing.T) {
	t.Parallel()

	server := httptest.NewUnstartedServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			schreiben.Header().Set("X-Fassung", anfrage.Proto)
			schreiben.WriteHeader(http.StatusOK)
		}))
	server.EnableHTTP2 = true
	server.StartTLS()

	defer server.Close()

	klient := NeuerKlient(UeblicheFristen())

	// **Das Zertifikat des Laufservers, nicht ein abgeschaltetes
	// Prüfverhalten**: `InsecureSkipVerify` machte den Lauf grün und die
	// Zusage wertlos.
	derKlientTransport(t, klient).TLSClientConfig = serverzertifikat(t, server)

	antwort, fehler := hole(t, klient, server.URL)
	if fehler != nil {
		t.Fatalf("die Anfrage ging nicht durch: %v", fehler)
	}

	defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

	if antwort.Proto != "HTTP/2.0" {
		t.Fatalf("der Klient sprach %q mit einem Server, der h2 kann",
			antwort.Proto)
	}

	// **Und die Gegenstelle hat es auch so gesehen** — eine Fassung, die nur
	// der Klient meldet, wäre seine Behauptung über sich selbst.
	if gesehen := antwort.Header.Get("X-Fassung"); gesehen != "HTTP/2.0" {
		t.Errorf("der Server sah %q", gesehen)
	}
}

// **Und gegen einen, der nur HTTP/1.1 kann, bleibt es bei HTTP/1.1.**
//
// **Die Gegenprobe** (§10): ein Klient, der h2 erzwingt und einen alten
// Server dann nicht mehr erreicht, wäre die teurere Hälfte des Befunds.
func TestGegenEinenAltenServerBleibtEsBeiHTTP11(t *testing.T) {
	t.Parallel()

	server := httptest.NewUnstartedServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusOK)
		}))
	server.StartTLS()

	defer server.Close()

	klient := NeuerKlient(UeblicheFristen())
	derKlientTransport(t, klient).TLSClientConfig = serverzertifikat(t, server)

	antwort, fehler := hole(t, klient, server.URL)
	if fehler != nil {
		t.Fatalf("der alte Server war nicht mehr erreichbar: %v", fehler)
	}

	defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

	if antwort.Proto != "HTTP/1.1" {
		t.Fatalf("gegen einen Server ohne h2 kam %q heraus", antwort.Proto)
	}
}

// serverzertifikat gibt die TLS-Einstellung, die den Laufserver kennt.
func serverzertifikat(t *testing.T, server *httptest.Server) *tls.Config {
	t.Helper()

	transport, ist := server.Client().Transport.(*http.Transport)
	if !ist {
		t.Fatalf("der Server trägt %T", server.Client().Transport)
	}

	return transport.TLSClientConfig
}

// **Die Stillstandsfrist greift auch unter HTTP/2.**
//
// # Warum das eigens gemessen wird
//
// **Weil h2 eine Verbindung teilt.** Die Frist hängt an der **Verbindung**
// (`fristverbindung`, #916) und nicht am Antwortrumpf; unter HTTP/1.1 trägt
// eine Verbindung eine Anfrage, unter h2 mehrere Ströme. Das Issue verlangt
// die Messung als eigene Zeile — eine Frist, die dort nicht mehr griffe, wäre
// der unbeobachtete Rumpf zurück, für den es Befund E-37 gab.
//
// **Gemessen und nicht überlegt**: die Gegenstelle schickt ihren Kopf und
// schweigt dann. Wer den Rumpf liest, muss abgewiesen werden.
func TestDieStillstandsfristGreiftAuchUnterHTTP2(t *testing.T) {
	t.Parallel()

	stille := make(chan struct{})

	server := httptest.NewUnstartedServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.Proto != "HTTP/2.0" {
				t.Errorf("die Gegenstelle sah %q", anfrage.Proto)
			}

			// **Kopf ja, Rumpf nein** — genau die Lage, die
			// `ResponseHeaderTimeout` nicht mehr erfasst.
			schreiben.WriteHeader(http.StatusOK)

			if spueler, kann := schreiben.(http.Flusher); kann {
				spueler.Flush()
			}

			<-stille
		}))
	server.EnableHTTP2 = true
	server.StartTLS()

	defer func() { close(stille); server.Close() }()

	// **Eine kurze Frist**, damit der Lauf nicht eine Minute steht.
	klient := NeuerKlient(Fristen{
		Verbindung: 5 * time.Second,
		Kopf:       5 * time.Second,
		Stillstand: 300 * time.Millisecond,
	})
	derKlientTransport(t, klient).TLSClientConfig = serverzertifikat(t, server)

	antwort, fehler := hole(t, klient, server.URL)
	if fehler != nil {
		t.Fatalf("die Anfrage ging nicht durch: %v", fehler)
	}

	defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

	if antwort.Proto != "HTTP/2.0" {
		t.Fatalf("der Lauf prüft die Frist unter %q und nicht unter h2",
			antwort.Proto)
	}

	begonnen := time.Now()

	if _, lesefehler := io.ReadAll(antwort.Body); lesefehler == nil {
		t.Fatal("der schweigende Rumpf kam durch — die Frist greift unter " +
			"h2 nicht mehr")
	}

	// **Und sie greift in ihrer Zeit** — eine, die erst nach einer Minute
	// zuschlägt, wäre keine kurze Frist, sondern eine andere.
	if gedauert := time.Since(begonnen); gedauert > 5*time.Second {
		t.Errorf("die Frist griff erst nach %s", gedauert)
	}
}

// hole ruft eine Adresse mit einem Kontext.
//
// **Mit und nicht ohne**: `http.Client.Get` baut die Anfrage ohne, und ein
// Lauf, der hängt, hinge bis zum Zeitablauf der ganzen Stufe.
func hole(t *testing.T, klient *Klient, adresse string) (*http.Response, error) {
	t.Helper()

	anfrage, fehler := http.NewRequestWithContext(t.Context(),
		http.MethodGet, adresse, http.NoBody)
	if fehler != nil {
		t.Fatalf("die Anfrage bauen: %v", fehler)
	}

	return klient.unten.Do(anfrage) //nolint:wrapcheck // der Lauf prüft den Fehler selbst
}

// **Die Kopffrist ist die, an der eine lange Generierung scheitert.**
//
// # Warum sie eigens gemessen wird
//
// **Weil sie bis zur Statuszeile zählt und nicht bis zum letzten Byte.** Ein
// Bildmodell antwortet erst, wenn es fertig ist — es schickt keinen Kopf und
// dann langsam den Rumpf. Wer die Frist zu knapp setzt, bricht den Auftrag
// ab, **während der Anbieter ihn noch abrechnet**.
//
// **Das ist der Grund, warum die KI-Nähte eigene Fristen tragen** (Ollama
// 5 min, OpenAI und xAI 3 min, Anthropic 2 min) statt der üblichen 30
// Sekunden. Die Zahlen stehen in `docs/NETZEIGENSCHAFTEN.md` — hier steht,
// dass die Frist überhaupt greift, und **wie** sie es tut.
func TestDieKopffristGreiftBevorEinKopfKommt(t *testing.T) {
	t.Parallel()

	stille := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			// **Kein Kopf, kein Rumpf** — die Lage eines Modells, das noch
			// rechnet.
			select {
			case <-stille:
			case <-anfrage.Context().Done():
			}

			schreiben.WriteHeader(http.StatusOK)
		}))

	defer func() { close(stille); server.Close() }()

	klient := NeuerKlient(Fristen{
		Verbindung: 5 * time.Second,
		Kopf:       300 * time.Millisecond,
		Stillstand: 30 * time.Second,
	})

	begonnen := time.Now()

	antwort, fehler := hole(t, klient, server.URL)
	if fehler == nil {
		defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

		t.Fatal("der schweigende Anbieter kam durch — dann gibt es keine " +
			"Kopffrist")
	}

	// **Und sie greift in ihrer Zeit.** Eine Frist, die erst nach der
	// Stillstandsfrist zuschlüge, wäre eine andere.
	if gedauert := time.Since(begonnen); gedauert > 5*time.Second {
		t.Errorf("die Kopffrist griff erst nach %s", gedauert)
	}
}

// **Und eine Antwort innerhalb der Frist kommt durch — auch eine langsame.**
//
// **Die Gegenprobe** (§10): eine Frist, die jede Antwort abschnitte, hielte
// den Lauf darüber, und kein Bildauftrag käme je zurück.
func TestEineLangsameAntwortInnerhalbDerFristKommtDurch(t *testing.T) {
	t.Parallel()

	const frist = 3 * time.Second

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			// **Deutlich langsamer als ein gewöhnlicher Ruf** und trotzdem
			// innerhalb der Frist — die Lage eines Modells, das rechnet.
			time.Sleep(frist / 6)

			schreiben.WriteHeader(http.StatusTeapot)
		}))
	defer server.Close()

	antwort, fehler := hole(t, NeuerKlient(Fristen{
		Verbindung: 5 * time.Second, Kopf: frist, Stillstand: frist,
	}), server.URL)
	if fehler != nil {
		t.Fatalf("die langsame Antwort kam nicht durch: %v", fehler)
	}

	defer antwort.Body.Close() //nolint:errcheck // der Lauf endet ohnehin

	if antwort.StatusCode != http.StatusTeapot {
		t.Fatalf("die Antwort ist %d", antwort.StatusCode)
	}
}
