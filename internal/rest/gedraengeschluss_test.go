package rest_test

// **Wer auf einen Platz wartet, geht mit dem Schluss** — SEC-14
// (#2143),
// Zweitblick über E, Befund 2.
//
// `Close` kam vom eingebetteten Lauscher und schloss nur ihn. Ein schon
// angenommener Aufrufer, der auf einen Platz wartete, wartete weiter — und
// wurde der Platz frei, gab `Accept` ihn nach dem Schluss noch heraus.

import (
	"errors"
	"net"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func TestDerSchlussBeendetDasWartenAufEinenPlatz(t *testing.T) {
	t.Parallel()

	lauscher, adresse := gedraengeMit(t, 1)
	lauscher.MitVorfrist(0)

	erster := waehle(t, adresse)
	t.Cleanup(func() { schliesse(t, erster) })

	belegt := nimmAn(t, lauscher)

	zweiter := waehle(t, adresse)
	t.Cleanup(func() { _ = zweiter.Close() }) //nolint:errcheck // vielleicht schon zu

	type angenommen struct {
		verbindung net.Conn
		fehler     error
	}

	ergebnis := make(chan angenommen, 1)

	go func() {
		verbindung, fehler := lauscher.Accept()
		ergebnis <- angenommen{verbindung, fehler}
	}()

	// **Erst, wenn er wirklich wartet** — dann schließen.
	time.Sleep(100 * time.Millisecond)

	if err := lauscher.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case danach := <-ergebnis:
		if danach.fehler == nil || !errors.Is(danach.fehler, net.ErrClosed) {
			t.Errorf("nach dem Schluss: %v, %v", danach.verbindung, danach.fehler)
		}
	case <-time.After(time.Second):
		t.Error("Accept wartet nach dem Schluss weiter auf einen Platz")
		schliesse(t, belegt)
		<-ergebnis
	}

	istZu(t, zweiter)
}

// immerda ist ein Lauscher, der auch nach seinem Schluss noch annimmt — so
// wie einer, der die Verbindung schon hatte, als `Close` kam.
type immerda struct{ net.Listener }

func (immerda) Accept() (net.Conn, error) {
	eine, andere := net.Pipe()
	_ = andere.Close() //nolint:errcheck // die Gegenseite braucht niemand

	return eine, nil
}

func (immerda) Close() error { return nil }

// **Auch ein freier Platz gibt nach dem Schluss nichts heraus** (Nachlauf
// über E): `Close` und eine Verbindung, die schon angenommen war, laufen
// nebeneinander — der Schluss gewinnt.
func TestNachDemSchlussGibtEinFreierPlatzNichtsHeraus(t *testing.T) {
	t.Parallel()

	lauscher := rest.HoechstensGleichzeitig(immerda{}, 4)

	if err := lauscher.Close(); err != nil {
		t.Fatal(err)
	}

	for versuch := range 20 {
		if verbindung, fehler := lauscher.Accept(); !errors.Is(fehler, net.ErrClosed) {
			t.Fatalf("Versuch %d nach dem Schluss: %v, %v", versuch+1, verbindung, fehler)
		}
	}
}

// **Beim Dienstende geht, wer nichts sagt — und nur der** — gefunden am
// 2026-10-07: `http.Server.Shutdown` hält eine Verbindung ohne Anfrage
// (`StateNew`) erst nach fünf Sekunden für ruhend und wartete die ganze
// Frist auf sie. Eine Verbindung, die arbeitet, bleibt offen.
func TestDieStummenSchliessenNurWerNichtsSagt(t *testing.T) {
	t.Parallel()

	lauscher, adresse := gedraengeMit(t, 3)

	// **Auch eine ruhende** — zwischen zwei Anfragen sagt sie ebenso nichts.
	ruhend := waehle(t, adresse)
	t.Cleanup(func() { _ = ruhend.Close() }) //nolint:errcheck // vielleicht schon zu

	ruhendDort := nimmAn(t, lauscher)
	lauscher.Zustand(ruhendDort, http.StateNew)
	lauscher.Zustand(ruhendDort, http.StateActive)
	lauscher.Zustand(ruhendDort, http.StateIdle)

	stumm := waehle(t, adresse)
	t.Cleanup(func() { _ = stumm.Close() }) //nolint:errcheck // vielleicht schon zu

	stummDort := nimmAn(t, lauscher)
	lauscher.Zustand(stummDort, http.StateNew)

	arbeitend := waehle(t, adresse)
	t.Cleanup(func() { schliesse(t, arbeitend) })

	arbeitendDort := nimmAn(t, lauscher)
	lauscher.Zustand(arbeitendDort, http.StateNew)
	lauscher.Zustand(arbeitendDort, http.StateActive)

	// **Die Serverseiten bleiben bis zum Schluss erreichbar** — sonst räumt
	// sie im vollen Paketlauf der Garbage Collector ab, und der Finalizer
	// schließt sie: die arbeitende stand dann geschlossen da, und eine stumme
	// wäre es auch ohne diese Methode gewesen (gemessen: dreimal von drei).
	defer runtime.KeepAlive([]net.Conn{ruhendDort, stummDort, arbeitendDort})

	lauscher.StummeSchliessen()

	for name, verbindung := range map[string]net.Conn{"stumme": stumm, "ruhende": ruhend} {
		_ = verbindung.SetReadDeadline(time.Now().Add(time.Second)) //nolint:errcheck // eine Frist zum Lesen
		if _, fehler := verbindung.Read(make([]byte, 1)); fehler == nil || errors.Is(fehler, os.ErrDeadlineExceeded) {
			t.Errorf("die %s Verbindung blieb offen: %v", name, fehler)
		}
	}

	_ = arbeitend.SetReadDeadline(time.Now().Add(200 * time.Millisecond)) //nolint:errcheck // dito
	if _, fehler := arbeitend.Read(make([]byte, 1)); !errors.Is(fehler, os.ErrDeadlineExceeded) {
		t.Errorf("die arbeitende Verbindung wurde geschlossen: %v", fehler)
	}
}

// **Auch wer sich erst nach dem Schluss meldet, geht** — angestrichen vom
// Zweitblick über Phase H (#2086): Go trägt eine Verbindung in seine eigene
// Liste ein, **bevor** es `ConnState` ruft (`net/http/server.go`, Zeile 1865,
// Go 1.27.1). Kam ihre Meldung erst nach [rest.Gedraenge.StummeSchliessen],
// sah die einmalige Momentaufnahme sie nicht, und `Shutdown` wartete wieder
// die ganze Frist.
func TestWerSichNachDemSchlussMeldetGehtAuch(t *testing.T) {
	t.Parallel()

	lauscher, adresse := gedraengeMit(t, 3)

	spaet := waehle(t, adresse)
	t.Cleanup(func() { _ = spaet.Close() }) //nolint:errcheck // vielleicht schon zu

	spaetDort := nimmAn(t, lauscher)

	arbeitend := waehle(t, adresse)
	t.Cleanup(func() { _ = arbeitend.Close() }) //nolint:errcheck // dito

	arbeitendDort := nimmAn(t, lauscher)

	defer runtime.KeepAlive([]net.Conn{spaetDort, arbeitendDort})

	lauscher.Zustand(arbeitendDort, http.StateNew)
	lauscher.Zustand(arbeitendDort, http.StateActive)

	lauscher.StummeSchliessen()

	lauscher.Zustand(spaetDort, http.StateNew)

	_ = spaet.SetReadDeadline(time.Now().Add(time.Second)) //nolint:errcheck // eine Frist zum Lesen
	if _, fehler := spaet.Read(make([]byte, 1)); fehler == nil || errors.Is(fehler, os.ErrDeadlineExceeded) {
		t.Errorf("die späte stumme Verbindung blieb offen: %v", fehler)
	}

	// **Eine, die beim Schluss arbeitet, antwortet zu Ende** — und geht,
	// sobald sie danach ruht.
	_ = arbeitend.SetReadDeadline(time.Now().Add(200 * time.Millisecond)) //nolint:errcheck // dito
	if _, fehler := arbeitend.Read(make([]byte, 1)); !errors.Is(fehler, os.ErrDeadlineExceeded) {
		t.Errorf("die arbeitende Verbindung wurde geschlossen: %v", fehler)
	}

	lauscher.Zustand(arbeitendDort, http.StateIdle)

	_ = arbeitend.SetReadDeadline(time.Now().Add(time.Second)) //nolint:errcheck // dito
	if _, fehler := arbeitend.Read(make([]byte, 1)); fehler == nil || errors.Is(fehler, os.ErrDeadlineExceeded) {
		t.Errorf("die nach dem Schluss ruhende Verbindung blieb offen: %v", fehler)
	}
}
