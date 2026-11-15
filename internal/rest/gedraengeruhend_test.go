package rest_test

// **Eine ruhende Verbindung gibt ihren Platz her** — QA-08
// (#1739).
//
// # Der Befund
//
// Der Lastlauf vom 2026-10-02 (Profil `ruhend`): ein Klient mit sechzehn
// Verbindungen im Pool fragte je einmal und ließ sie offen, wie jeder Pool
// mit `keepAlive`. Danach nahm `mse-rest` **keine** Verbindung mehr an — ein
// zweiter Klient wartete fünfzehn Sekunden und gab auf. Die Lesefrist greift
// nicht, denn ein Kopf kam ja; was greift, ist der Leerlauf von zwei Minuten.
// So lange hielt **ein** Klient alle anderen draußen.
//
// # Was gilt
//
// **Ist jeder Platz belegt, hält eine ruhende Verbindung keinen.** Sie wird
// geschlossen, sobald sie ruht, oder — wenn sie schon ruhte — sobald ein
// Platz gebraucht wird, die am längsten ruhende zuerst. Ein Klient merkt es
// nur daran, dass er neu verbinden muss. **Eine arbeitende Verbindung bleibt**:
// ihr Platz ist der, den die Grenze schützt.

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// TestRuhendeVerbindungenSperrenNiemandenAus ist der Befund am echten Server.
func TestRuhendeVerbindungenSperrenNiemandenAus(t *testing.T) {
	t.Parallel()

	roh, fehler := rest.Lauschen(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	const hoechstens = 2

	lauscher := rest.HoechstensGleichzeitig(roh, hoechstens)

	// **Zwei Rufe des Pools treffen sich im Behandler** — damit es wirklich
	// zwei Verbindungen sind und nicht eine, die zweimal fragt.
	var tor sync.WaitGroup

	tor.Add(hoechstens)

	server := &http.Server{
		Handler: http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.URL.Path == "/pool" {
				tor.Done()
				tor.Wait()
			}

			_, _ = io.WriteString(schreiben, "ja") //nolint:errcheck // der Klient prüft den Rumpf
		}),
		ConnState:         lauscher.Zustand,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	go func() { _ = server.Serve(lauscher) }() //nolint:errcheck // endet mit dem Test

	t.Cleanup(func() { schliesse(t, server) })

	adresse := "http://" + lauscher.Addr().String()
	pool := &http.Client{Transport: &http.Transport{MaxConnsPerHost: hoechstens, MaxIdleConnsPerHost: hoechstens}}

	var fertig sync.WaitGroup

	for range hoechstens {
		fertig.Go(func() { holeMit(t, pool, adresse+"/pool") })
	}

	fertig.Wait()

	// **Ein zweiter Klient kommt durch** — und nicht erst nach dem Leerlauf.
	fremd := &http.Client{Transport: &http.Transport{}, Timeout: 3 * time.Second}
	holeMit(t, fremd, adresse+"/")
}

// TestUnterVollerLastBrichtKeineVerbindungAb hält die andere Seite: ein Pool
// mit genau so vielen Verbindungen, wie es Plätze gibt, fragt ohne Pause —
// und keiner seiner Rufe scheitert. **Gefunden vom zweiten Lastlauf am
// 2026-10-02**: die erste Fassung schloss jede Verbindung, die bei vollem
// Haus auch nur zwischen zwei Rufen ruhte, und ein Klient mit sechzehn
// Verbindungen sah Tausende abgerissene Rufe. Gefragt wird mit `POST`, denn
// einen `GET` wiederholte Go nach einem Abriss still.
func TestUnterVollerLastBrichtKeineVerbindungAb(t *testing.T) {
	t.Parallel()

	roh, fehler := rest.Lauschen(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	const hoechstens = 2

	lauscher := rest.HoechstensGleichzeitig(roh, hoechstens)
	server := &http.Server{
		Handler: http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
			_, _ = io.Copy(io.Discard, anfrage.Body) //nolint:errcheck // gelesen ist gelesen
			_, _ = io.WriteString(schreiben, "ja")   //nolint:errcheck // der Klient prüft den Rumpf
		}),
		ConnState:         lauscher.Zustand,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	go func() { _ = server.Serve(lauscher) }() //nolint:errcheck // endet mit dem Test

	t.Cleanup(func() { schliesse(t, server) })

	adresse := "http://" + lauscher.Addr().String() + "/"
	pool := &http.Client{Transport: &http.Transport{MaxConnsPerHost: hoechstens, MaxIdleConnsPerHost: hoechstens}}

	var (
		fertig      sync.WaitGroup
		sperre      sync.Mutex
		gescheitert []string
	)

	for range hoechstens {
		fertig.Go(func() {
			for range 300 {
				anfrage, err := http.NewRequestWithContext(t.Context(), http.MethodPost, adresse, strings.NewReader("x"))
				if err != nil {
					t.Error(err)

					return
				}

				antwort, err := pool.Do(anfrage)
				if err != nil {
					sperre.Lock()
					gescheitert = append(gescheitert, err.Error())
					sperre.Unlock()

					continue
				}

				_, _ = io.Copy(io.Discard, antwort.Body) //nolint:errcheck // gelesen
				_ = antwort.Body.Close()                 //nolint:errcheck // dito
			}
		})
	}

	fertig.Wait()

	if len(gescheitert) > 0 {
		t.Errorf("%d von %d Rufen abgerissen, etwa: %s", len(gescheitert), hoechstens*300, gescheitert[0])
	}
}

// TestEineRuhendeVerbindungGehtErstWennJemandWartet hält den ersten Weg:
// eine ruhende Verbindung bleibt, bis eine neue einen Platz braucht — dann
// geht sie, und die neue kommt durch.
func TestEineRuhendeVerbindungGehtErstWennJemandWartet(t *testing.T) {
	t.Parallel()

	lauscher, adresse := gedraengeMit(t, 1)

	klient := waehle(t, adresse)
	t.Cleanup(func() { _ = klient.Close() }) //nolint:errcheck // vielleicht schon zu

	serverseite := nimmAn(t, lauscher)

	// **Arbeitend bleibt sie, ruhend auch — solange niemand wartet.** Und
	// ein unbekannter Gegenstand ändert nichts.
	lauscher.Zustand(serverseite, http.StateActive)
	lauscher.Zustand(&net.TCPConn{}, http.StateIdle)
	lauscher.Zustand(serverseite, http.StateIdle)
	istOffen(t, klient)

	neu := waehle(t, adresse)
	t.Cleanup(func() { schliesse(t, neu) })

	neueSeite := nimmAn(t, lauscher)
	t.Cleanup(func() { schliesse(t, neueSeite) })

	istZu(t, klient)
}

// TestEinWartenderNimmtDieNaechsteRuhende hält den zweiten Weg: wartet eine
// Verbindung schon, geht die nächste, die zu ruhen beginnt.
func TestEinWartenderNimmtDieNaechsteRuhende(t *testing.T) {
	t.Parallel()

	lauscher, adresse := gedraengeMit(t, 1)

	klient := waehle(t, adresse)
	t.Cleanup(func() { _ = klient.Close() }) //nolint:errcheck // vielleicht schon zu

	serverseite := nimmAn(t, lauscher)
	lauscher.Zustand(serverseite, http.StateActive)

	neu := waehle(t, adresse)
	t.Cleanup(func() { schliesse(t, neu) })

	angenommen := imHintergrund(lauscher)

	// **Erst warten lassen, dann ruhen** — sonst nähme der Wartende die
	// Ruhende über den ersten Weg, und dieser Fall prüfte ihn noch einmal.
	select {
	case <-angenommen:
		t.Fatal("die neue Verbindung kam ohne Platz durch")
	case <-time.After(200 * time.Millisecond):
	}

	lauscher.Zustand(serverseite, http.StateIdle)
	istZu(t, klient)

	select {
	case neueSeite := <-angenommen:
		t.Cleanup(func() { schliesse(t, neueSeite) })
	case <-time.After(2 * time.Second):
		t.Error("die wartende Verbindung bekam keinen Platz")
	}
}

// TestEinNeuerPlatzVerdraengtDieAmLaengstenRuhende hält die Reihenfolge:
// ruhen zwei, geht die ältere.
func TestEinNeuerPlatzVerdraengtDieAmLaengstenRuhende(t *testing.T) {
	t.Parallel()

	lauscher, adresse := gedraengeMit(t, 2)

	erste := waehle(t, adresse)
	t.Cleanup(func() { _ = erste.Close() }) //nolint:errcheck // vielleicht schon zu

	erstesEnde := nimmAn(t, lauscher)
	lauscher.Zustand(erstesEnde, http.StateIdle)

	zweite := waehle(t, adresse)
	t.Cleanup(func() { schliesse(t, zweite) })

	zweitesEnde := nimmAn(t, lauscher)
	t.Cleanup(func() { schliesse(t, zweitesEnde) })
	lauscher.Zustand(zweitesEnde, http.StateIdle)

	dritte := waehle(t, adresse)
	t.Cleanup(func() { schliesse(t, dritte) })

	drittesEnde := nimmAn(t, lauscher)
	t.Cleanup(func() { schliesse(t, drittesEnde) })

	istZu(t, erste)
	istOffen(t, zweite)
}

// gedraengeMit gibt einen Lauscher mit `hoechstens` Plätzen und seine Adresse.
func gedraengeMit(t *testing.T, hoechstens int) (lauscher *rest.Gedraenge, adresse string) {
	t.Helper()

	roh, fehler := rest.Lauschen(t.Context(), 0)
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	lauscher = rest.HoechstensGleichzeitig(roh, hoechstens)
	t.Cleanup(func() { _ = lauscher.Close() }) //nolint:errcheck // Aufräumen

	return lauscher, lauscher.Addr().String()
}

// nimmAn holt die nächste Verbindung ab — oder bricht nach zwei Sekunden ab.
func nimmAn(t *testing.T, lauscher net.Listener) net.Conn {
	t.Helper()

	select {
	case verbindung := <-imHintergrund(lauscher):
		return verbindung
	case <-time.After(2 * time.Second):
		t.Fatal("keine Verbindung angenommen")

		return nil
	}
}

// imHintergrund nimmt die nächste Verbindung in einer eigenen Goroutine an.
func imHintergrund(lauscher net.Listener) <-chan net.Conn {
	angenommen := make(chan net.Conn, 1)

	go func() {
		if verbindung, fehler := lauscher.Accept(); fehler == nil {
			angenommen <- verbindung
		}
	}()

	return angenommen
}

// istZu sagt, dass die Gegenseite geschlossen hat.
func istZu(t *testing.T, klient net.Conn) {
	t.Helper()

	_ = klient.SetReadDeadline(time.Now().Add(2 * time.Second)) //nolint:errcheck // die Frist ist der Test

	if _, fehler := klient.Read(make([]byte, 1)); !errors.Is(fehler, io.EOF) {
		t.Errorf("die Verbindung ist nicht geschlossen: %v", fehler)
	}
}

// istOffen sagt, dass die Gegenseite nicht geschlossen hat.
func istOffen(t *testing.T, klient net.Conn) {
	t.Helper()

	_ = klient.SetReadDeadline(time.Now().Add(100 * time.Millisecond)) //nolint:errcheck // die Frist ist der Test

	if _, fehler := klient.Read(make([]byte, 1)); errors.Is(fehler, io.EOF) {
		t.Error("die Verbindung wurde geschlossen")
	}
}

// holeMit fragt und erwartet „ja".
func holeMit(t *testing.T, klient *http.Client, adresse string) {
	t.Helper()

	anfrage, fehler := http.NewRequestWithContext(t.Context(), http.MethodGet, adresse, http.NoBody)
	if fehler != nil {
		t.Fatal(fehler)
	}

	antwort, fehler := klient.Do(anfrage)
	if fehler != nil {
		t.Errorf("%s: %v", adresse, fehler)

		return
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen

	if rumpf, _ := io.ReadAll(antwort.Body); string(rumpf) != "ja" { //nolint:errcheck // der Rumpf zählt
		t.Errorf("%s: %q", adresse, rumpf)
	}
}
