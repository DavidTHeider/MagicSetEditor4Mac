package mtgapi_test

// **Die übrigen Wege des Klienten** — Werkstück 3 von Phase 4.8.
//
// **Eine eigene Datei**, damit `klient_test.go` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, was der Klient **tut**,
// hier, was er bei Fehlern und Sonderfällen **sagt**.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Eine Abfrage wird kodiert.**
//
// Ein Kartenname mit Leerzeichen oder Apostroph gehört kodiert; wer es
// vergisst, bekommt eine Antwort auf eine andere Frage.
func TestEineAbfrageWirdKodiert(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ wert, erwartet string }{
		{"Lightning Bolt", "name=Lightning+Bolt"},
		{"Gaea's Cradle", "name=Gaea%27s+Cradle"},
		{"Ach! Hans, Run!", "name=Ach%21+Hans%2C+Run%21"},
	} {
		if gebaut := mtgapi.Abfrage("name", fall.wert); gebaut != fall.erwartet {
			t.Errorf("%q wurde zu %q, erwartet %q", fall.wert, gebaut, fall.erwartet)
		}
	}
}

// **Die Drossel lässt sich setzen** — sonst wartete jeder Lauf eine Sekunde.
func TestDieDrosselLaesstSichSetzen(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiber, []byte(`{"cards":[]}`))
		}))
	defer server.Close()

	var geschlafen time.Duration

	klient := mtgapi.NeuerKlient(server.URL, netz.KlientUeber(server.Client())).
		MitDrossel(netz.NeueDrossel(time.Hour, time.Now,
			func(wie time.Duration) { geschlafen += wie }))

	// **Zweimal fragen** — beim ersten Mal wartet keine Drossel.
	for range 2 {
		if _, _, err := klient.Karten(t.Context(), "name=X"); err != nil {
			t.Fatalf("fragen: %v", err)
		}
	}

	if geschlafen == 0 {
		t.Error("die gesetzte Drossel hat nicht gewartet")
	}
}

// **Ein unerwarteter Statuscode wird genannt.**
//
// Ein 500 ist weder „nichts gefunden" noch „gedrosselt". Wer ihn als eines
// von beidem meldete, schickte jemanden auf die falsche Suche (§8).
func TestEinUnerwarteterStatusWirdGenannt(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.WriteHeader(http.StatusInternalServerError)
	})

	_, _, err := klient.Karten(t.Context(), "name=X")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("die Meldung %q nennt den Code nicht", err)
	}
}

// **Ein 429 gilt auch als Drosselung** — und zwar schon im Transport.
//
// Der Dienst nennt 403; ein Vermittler davor kann 429 schicken. Beides heißt
// „warte", und beides darf nicht wiederholt werden.
//
// **Der Klient hat dafür keinen eigenen Zweig**, und das ist Absicht:
// `netz.Hole` beantwortet ihn und **sperrt dabei die Drossel**. Ein Zweig
// daneben wäre unerreichbar und machte die schwächere Zusage (§3, §10).
func TestEinVierhundertneunundzwanzigIstAuchEineDrossel(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.WriteHeader(http.StatusTooManyRequests)
	})

	if _, _, err := klient.Karten(t.Context(), "name=X"); !errors.Is(
		err, domain.ErrGedrosselt) {
		t.Errorf("Fehler ist %v, erwartet ErrGedrosselt", err)
	}
}

// **Eine lange Ablehnung wird gekürzt.**
//
// Eine Fehlermeldung mit acht Megabyte darin ist keine.
func TestEineLangeAblehnungWirdGekuerzt(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.WriteHeader(http.StatusForbidden)
		schreibe(t, schreiber, []byte(strings.Repeat("x", 5000)))
	})

	_, _, err := klient.Karten(t.Context(), "name=X")
	if err == nil {
		t.Fatal("angenommen, erwartet ein Fehler")
	}

	if len(err.Error()) > 400 {
		t.Errorf("die Meldung ist %d Zeichen lang", len(err.Error()))
	}

	if !strings.Contains(err.Error(), "…") {
		t.Error("die gekürzte Meldung sagt nicht, dass gekürzt wurde")
	}
}

// **Eine unlesbare Antwort ist ein Fehler.**
func TestEineUnlesbareAntwortIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreibe(t, schreiber, []byte("kein json"))
	})

	if _, _, err := klient.Karten(t.Context(), "name=X"); !errors.Is(
		err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Ein halber Stand ist keiner.**
//
// „5.000 Grenze, und wie viele übrig sind, weiß ich nicht" hilft niemandem.
func TestEinHalberStandGiltAlsUnbekannt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, kopf, wert string }{
		{"nur die Grenze", "Ratelimit-Limit", "5000"},
		{"nur der Rest", "Ratelimit-Remaining", "17"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
				schreiber.Header().Set(fall.kopf, fall.wert)
				schreibe(t, schreiber, []byte(`{"cards":[]}`))
			})

			_, stand, err := klient.Karten(t.Context(), "name=X")
			if err != nil {
				t.Fatalf("fragen: %v", err)
			}

			if stand.Bekannt {
				t.Errorf("ein halber Stand gilt als bekannt: %+v", stand)
			}
		})
	}
}

// **Und eine unsinnige Zahl im Kopf ebenso.**
func TestEineUnsinnigeZahlGiltAlsUnbekannt(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.Header().Set("Ratelimit-Limit", "viele")
		schreiber.Header().Set("Ratelimit-Remaining", "17")
		schreibe(t, schreiber, []byte(`{"cards":[]}`))
	})

	_, stand, err := klient.Karten(t.Context(), "name=X")
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if stand.Bekannt {
		t.Errorf("»viele« gilt als Zahl: %+v", stand)
	}
}

// **Die Drosselmeldung sagt es auch ohne Stand.**
func TestDieDrosselmeldungSagtEsAuchOhneStand(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, func(schreiber http.ResponseWriter, _ *http.Request) {
		schreiber.WriteHeader(http.StatusForbidden)
		schreibe(t, schreiber, []byte("Rate Limit Exceeded"))
	})

	_, _, err := klient.Karten(t.Context(), "name=X")

	if !strings.Contains(err.Error(), "keinen Stand") {
		t.Errorf("die Meldung %q sagt nicht, dass der Stand fehlt", err)
	}
}

// **Ein Netzfehler kommt als solcher an.**
func TestEinNetzfehlerKommtAn(t *testing.T) {
	t.Parallel()

	klient := mtgapi.NeuerKlient("http://127.0.0.1:1", netz.NeuerKlient(
		netz.Fristen{Verbindung: time.Second, Kopf: time.Second, Stillstand: time.Second}))

	if _, _, err := klient.Karten(t.Context(), "name=X"); err == nil {
		t.Error("gegen eine tote Adresse kam kein Fehler")
	}
}
