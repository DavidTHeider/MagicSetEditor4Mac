package mtggraphql_test

// **Die übrigen Wege des GraphQL-Klienten** — Werkstück 4 von Phase 4.8.
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
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Die Drossel lässt sich setzen** — sonst wartete jeder Lauf zehn Sekunden.
func TestDieDrosselLaesstSichSetzen(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			if _, err := schreiber.Write([]byte(`{"data":{"cards":[]}}`)); err != nil {
				t.Errorf("antworten: %v", err)
			}
		}))
	defer server.Close()

	var geschlafen time.Duration

	klient := mtggraphql.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()), "geheim").
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

// **Ein Netzfehler kommt als solcher an.**
func TestEinNetzfehlerKommtAn(t *testing.T) {
	t.Parallel()

	klient := mtggraphql.NeuerKlient("http://127.0.0.1:1", netz.NeuerKlient(
		netz.Fristen{Verbindung: time.Second, Kopf: time.Second, Stillstand: time.Second}),
		"geheim")

	if _, err := klient.Karten(t.Context(), "X"); err == nil {
		t.Error("gegen eine tote Adresse kam kein Fehler")
	}
}

// **Eine unlesbare Antwort ist ein Fehler und keine leere Liste.**
func TestEineUnlesbareAntwortIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, []byte("kein json"))
	})

	if _, err := klient.Karten(t.Context(), "X"); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Eine leere Antwort ist eine leere Liste und kein Fehler.**
//
// „Nichts gefunden" ist eine Auskunft; sie als Fehler zu melden hieße, dass
// eine Suche ohne Treffer wie ein Ausfall aussieht (§8).
func TestEineLeereAntwortIstKeinFehler(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, []byte(`{"data":{"cards":[]}}`))
	})

	karten, err := klient.Karten(t.Context(), "Gibtsnicht")
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if len(karten) != 0 {
		t.Errorf("%d Karten, erwartet keine", len(karten))
	}
}

// **Mehrere Fehler werden alle genannt.**
//
// GraphQL darf mehrere schicken. Nur den ersten zu melden hieße, dass die
// übrigen erst beim nächsten Versuch auffallen — einer nach dem anderen.
func TestMehrereFehlerWerdenAlleGenannt(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, []byte(
			`{"errors":[{"message":"erster"},{"message":"zweiter"}]}`))
	})

	_, err := klient.Karten(t.Context(), "X")
	if err == nil {
		t.Fatal("angenommen, erwartet ein Fehler")
	}

	for _, teil := range []string{"erster", "zweiter"} {
		if !contains(err.Error(), teil) {
			t.Errorf("die Meldung %q nennt %q nicht", err, teil)
		}
	}
}

// contains ist `strings.Contains` — hier, damit die Zeile kurz bleibt.
func contains(wo, was string) bool { return strings.Contains(wo, was) }

// **Der allgemeine Abfrageweg meldet, was JSON nicht schreiben kann.**
//
// `Karten` reicht Zeichenketten durch und kann diesen Zweig nicht betreten —
// hier kommen die Variablen von außen.
func TestDerAbfragewegMeldetUnschreibbaresJSON(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, "geheim", func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, []byte(`{"data":{"cards":[]}}`))
	})

	_, err := klient.Abfragen(t.Context(), "{ cards { name } }",
		map[string]any{"kanal": make(chan int)})

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Und auch er fragt ohne Marke nicht.**
func TestDerAbfragewegFragtOhneMarkeNicht(t *testing.T) {
	t.Parallel()

	var anfragen int

	klient := klientGegen(t, "", func(schreiber http.ResponseWriter, _ *http.Request) {
		anfragen++

		schreibe(t, schreiber, []byte(`{"data":{"cards":[]}}`))
	})

	if _, err := klient.Abfragen(t.Context(), "{ cards { name } }", nil); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if anfragen != 0 {
		t.Errorf("%d Anfragen ohne Marke, erwartet keine", anfragen)
	}
}
