package rest_test

// **Eine neue Verbindung, die schweigt, hält keinen Platz, den jemand
// braucht** — SEC-13 und SEC-14
// (#2142,
// #2143).
//
// Ein ruhender Pool gab seinen Platz schon her (QA-08); eine Verbindung, die
// nach dem Annehmen **gar nichts** sagte, hielt ihn bis zur Lesefrist. Wer
// alle Plätze so belegte und nachfüllte, ließ jeden rechtmäßigen Aufrufer je
// bis zur Lesefrist warten — keine endlose Sperre, aber eine, die ein lokaler
// Prozess ohne Marke beliebig oft auslösen konnte.

import (
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// dienstMitVorfrist fährt einen Server mit `hoechstens` Plätzen, der Lesefrist
// des Dienstes in Prüfgröße und der Vorfrist `vorfrist`.
func dienstMitVorfrist(t *testing.T, hoechstens int, vorfrist time.Duration) string {
	t.Helper()

	roh, fehler := rest.Lauschen(t.Context(), 0)
	if fehler != nil {
		t.Fatal(fehler)
	}

	lauscher := rest.HoechstensGleichzeitig(roh, hoechstens).MitVorfrist(vorfrist)

	server := &http.Server{
		Handler: http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
			if anfrage.URL.Path == "/lang" {
				time.Sleep(600 * time.Millisecond)
			}

			_, _ = io.WriteString(schreiben, "ja") //nolint:errcheck // der Klient prüft den Rumpf
		}),
		ConnState:         lauscher.Zustand,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() { _ = server.Serve(lauscher) }() //nolint:errcheck // endet mit dem Test

	t.Cleanup(func() { schliesse(t, server) })

	return lauscher.Addr().String()
}

func TestSchweigendeNeueVerbindungenHaltenNiemandenBisZurLesefrist(t *testing.T) {
	t.Parallel()

	const hoechstens = 4

	adresse := dienstMitVorfrist(t, hoechstens, 200*time.Millisecond)

	for range hoechstens {
		schweiger, fehler := net.Dial("tcp", adresse) //nolint:noctx // ein Schweiger braucht keinen Kontext
		if fehler != nil {
			t.Fatal(fehler)
		}

		t.Cleanup(func() { _ = schweiger.Close() }) //nolint:errcheck // nur fürs Aufräumen — womöglich schon zu
	}

	anfang := time.Now()
	holeMit(t, &http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Second}, "http://"+adresse+"/")
	dauer := time.Since(anfang)

	t.Logf("der rechtmäßige Aufrufer wartete %v hinter %d Schweigern", dauer, hoechstens)

	// **Die Lesefrist ist 3 s** — ohne Vorfrist wartete er so lange.
	if dauer > 1500*time.Millisecond {
		t.Errorf("der rechtmäßige Aufrufer wartete %v — bis zur Lesefrist", dauer)
	}
}

// **Eine Antwort, die arbeitet, bleibt** — auch wenn jemand wartet und die
// Vorfrist längst um ist.
func TestEineArbeitendeAntwortBleibtUnberuehrt(t *testing.T) {
	t.Parallel()

	adresse := dienstMitVorfrist(t, 1, 100*time.Millisecond)
	klient := &http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Second}

	lang := make(chan struct{})

	go func() {
		defer close(lang)

		holeMit(t, klient, "http://"+adresse+"/lang")
	}()

	time.Sleep(150 * time.Millisecond) // **die lange Antwort arbeitet schon**

	holeMit(t, &http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Second}, "http://"+adresse+"/")
	<-lang
}

// **Ohne Vorfrist bleibt es beim Warten** — wie vor SEC-14: eine neue
// Verbindung weicht nicht, und der Wartende kommt erst, wenn sie geht.
func TestOhneVorfristWeichtKeineNeueVerbindung(t *testing.T) {
	t.Parallel()

	lauscher, adresse := gedraengeMit(t, 1)
	lauscher.MitVorfrist(0)

	erster, fehler := net.Dial("tcp", adresse) //nolint:noctx // ein Schweiger braucht keinen Kontext
	if fehler != nil {
		t.Fatal(fehler)
	}
	t.Cleanup(func() { _ = erster.Close() }) //nolint:errcheck // nur fürs Aufräumen — womöglich schon zu

	angenommen := nimmAn(t, lauscher)
	lauscher.Zustand(angenommen, http.StateNew)

	zweiter, fehler := net.Dial("tcp", adresse) //nolint:noctx // dito
	if fehler != nil {
		t.Fatal(fehler)
	}
	t.Cleanup(func() { _ = zweiter.Close() }) //nolint:errcheck // nur fürs Aufräumen — womöglich schon zu

	wartet := imHintergrund(lauscher)

	select {
	case <-wartet:
		t.Fatal("die neue Verbindung wich ohne Vorfrist")
	case <-time.After(300 * time.Millisecond):
	}

	_ = angenommen.Close() //nolint:errcheck // der Platz wird frei

	select {
	case <-wartet:
	case <-time.After(2 * time.Second):
		t.Fatal("der Wartende kam nicht, als der Platz frei wurde")
	}
}
