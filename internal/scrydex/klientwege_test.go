package scrydex_test

// **Die übrigen Wege des Scrydex-Klienten** — Werkstück 5 von Phase 4.8.
//
// **Eine eigene Datei**, damit `klient_test.go` unter der Zeilengrenze aus §2
// bleibt.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

// **Die Drossel lässt sich setzen** — sonst wartete jeder Lauf eine Sekunde.
func TestDieDrosselLaesstSichSetzen(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			if _, err := schreiber.Write([]byte(`{"data":[]}`)); err != nil {
				t.Errorf("antworten: %v", err)
			}
		}))
	defer server.Close()

	var geschlafen time.Duration

	klient := scrydex.NeuerKlient(server.URL, netz.KlientUeber(server.Client()),
		scrydex.Zugang{Schluessel: "geheim", Team: "mannschaft"}).
		MitDrossel(netz.NeueDrossel(time.Hour, time.Now,
			func(wie time.Duration) { geschlafen += wie }))

	for range 2 {
		if _, err := klient.Karten(t.Context(), "X"); err != nil {
			t.Fatalf("fragen: %v", err)
		}
	}

	if geschlafen == 0 {
		t.Error("die gesetzte Drossel hat nicht gewartet")
	}
}

// **Eine unlesbare Antwort ist ein Fehler und keine leere Liste.**
func TestEineUnlesbareAntwortIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, vollerZugang(),
		func(schreiber http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiber, []byte("kein json"))
		})

	if _, err := klient.Karten(t.Context(), "X"); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Ein unerwarteter Status ohne Fehlerrumpf wird genannt.**
//
// Der Rumpf steht vor dem Code — aber wenn er nichts sagt, muss der Code es
// tun. Ein Schweigen wäre die schlechteste Antwort (§8).
func TestEinUnerwarteterStatusWirdGenannt(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, vollerZugang(),
		func(schreiber http.ResponseWriter, _ *http.Request) {
			schreiber.WriteHeader(http.StatusInternalServerError)

			schreibe(t, schreiber, []byte(`{"data":[]}`))
		})

	_, err := klient.Karten(t.Context(), "X")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("die Meldung %q nennt den Code nicht", err)
	}
}

// **Eine leere Antwort ist eine leere Liste und kein Fehler.**
func TestEineLeereAntwortIstKeinFehler(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, vollerZugang(),
		func(schreiber http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiber, []byte(`{"status":200,"data":[],"totalCount":0}`))
		})

	karten, err := klient.Karten(t.Context(), "Gibtsnicht")
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if len(karten) != 0 {
		t.Errorf("%d Karten, erwartet keine", len(karten))
	}
}

// **Ein Netzfehler kommt als solcher an.**
func TestEinNetzfehlerKommtAn(t *testing.T) {
	t.Parallel()

	klient := scrydex.NeuerKlient("http://127.0.0.1:1", netz.NeuerKlient(
		netz.Fristen{Verbindung: time.Second, Kopf: time.Second, Stillstand: time.Second}),
		scrydex.Zugang{Schluessel: "geheim", Team: "mannschaft"})

	if _, err := klient.Karten(t.Context(), "X"); err == nil {
		t.Error("gegen eine tote Adresse kam kein Fehler")
	}
}

// **Ein vollständiger Zugang meldet nichts als fehlend.**
//
// Ohne diesen Lauf sagte [scrydex.Zugang.Fehlt] nur, dass es immer etwas zu
// meckern gibt.
func TestEinVollstaendigerZugangIstVollstaendig(t *testing.T) {
	t.Parallel()

	if fehlt := vollerZugang().Fehlt(); fehlt != "" {
		t.Errorf("ein vollständiger Zugang meldet %q als fehlend", fehlt)
	}
}
