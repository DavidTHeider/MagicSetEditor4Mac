package scrydex_test

// **Der Klient für Scrydex** — Werkstück 5 von Phase 4.8
// (#1321).
//
// # Was hier geprüft wird und was nicht
//
// **Geprüft**: dass **beide** Kopfzeilen gesetzt werden, die Gestalt der
// Anfrage und die Deutung der Antwort — gegen einen eigenen Testserver.
// `testdaten/ohnezugang.json` ist **echt** (abgerufen am 2026-09-02, ohne
// Kopfzeilen).
//
// **Nicht geprüft**: dass der echte Dienst Kartendaten so schickt. Es gibt
// hier kein Konto — Vermerk 0056, Entscheidung 4. Die Oberfläche sagt es dort,
// wo man ihn einschaltet.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

// beispiel liest eine Datei aus `testdaten/`.
func beispiel(t *testing.T, name string) []byte {
	t.Helper()

	//nolint:gosec // der Name kommt aus dem Lauf, nicht von außen
	roh, err := os.ReadFile(filepath.Join("testdaten", name))
	if err != nil {
		t.Fatalf("%s lesen: %v", name, err)
	}

	return roh
}

// schreibe antwortet und meldet einen Schreibfehler.
func schreibe(t *testing.T, schreiber http.ResponseWriter, rumpf []byte) {
	t.Helper()

	if _, err := schreiber.Write(rumpf); err != nil {
		t.Errorf("antworten: %v", err)
	}
}

// klientGegen baut einen Klienten gegen einen Testserver.
func klientGegen(
	t *testing.T, zugang scrydex.Zugang, bediener http.HandlerFunc,
) *scrydex.Klient {
	t.Helper()

	server := httptest.NewServer(bediener)
	t.Cleanup(server.Close)

	return scrydex.NeuerKlient(server.URL, netz.KlientUeber(server.Client()), zugang)
}

// vollerZugang ist ein Zugang mit beiden Angaben.
func vollerZugang() scrydex.Zugang {
	return scrydex.Zugang{Schluessel: "geheim", Team: "mannschaft"}
}

// **Beide Kopfzeilen werden gesetzt.**
//
// **Nachgelesen am 2026-09-02** unter
// `https://scrydex.com/docs/getting-started/authentication`: `X-Api-Key`
// **und** `X-Team-ID`. Ein Dienst, der nur eine bekäme, wäre eingerichtet und
// antwortete trotzdem nicht.
func TestBeideKopfzeilenWerdenGesetzt(t *testing.T) {
	t.Parallel()

	var gesehen http.Header

	klient := klientGegen(t, vollerZugang(),
		func(schreiber http.ResponseWriter, gestellt *http.Request) {
			gesehen = gestellt.Header.Clone()

			schreibe(t, schreiber, beispiel(t, "karten.json"))
		})

	if _, err := klient.Karten(t.Context(), "Blitzschlag"); err != nil {
		t.Fatalf("fragen: %v", err)
	}

	for kopf, erwartet := range map[string]string{
		"X-Api-Key": "geheim",
		"X-Team-Id": "mannschaft",
	} {
		if gelesen := gesehen.Get(kopf); gelesen != erwartet {
			t.Errorf("%s ist %q, erwartet %q", kopf, gelesen, erwartet)
		}
	}
}

// **Ohne beide Angaben wird gar nicht erst gefragt.**
//
// # Die Doku sagt etwas anderes als der Dienst
//
// `https://scrydex.com/docs/getting-started/authentication` sagt am
// 2026-09-02: „Requests made without authentication headers will proceed, but
// with heavily reduced rate limits."
//
// **Gemessen am selben Tag** gegen `api.scrydex.com`:
//
//	HTTP/2 401
//	{"error":{"message":"Invalid credentials","code":"INVALID_CREDENTIALS"}}
//
// Es geht **nicht** weiter. Eine Anfrage ohne Zugang ist damit eine Anfrage,
// deren Antwort feststeht (§16, Regel 4).
func TestOhneZugangWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name   string
		zugang scrydex.Zugang
	}{
		{"gar nichts", scrydex.Zugang{}},
		{"nur der Schlüssel", scrydex.Zugang{Schluessel: "geheim"}},
		{"nur das Team", scrydex.Zugang{Team: "mannschaft"}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			var anfragen int

			klient := klientGegen(t, fall.zugang,
				func(schreiber http.ResponseWriter, _ *http.Request) {
					anfragen++

					schreibe(t, schreiber, beispiel(t, "karten.json"))
				})

			_, err := klient.Karten(t.Context(), "Blitzschlag")

			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
			}

			if anfragen != 0 {
				t.Errorf("%d Anfragen ohne Zugang, erwartet keine", anfragen)
			}
		})
	}
}

// **Und die Meldung sagt, welche Angabe fehlt.**
//
// „Zugang fehlt" hilft nicht, wenn man den Schlüssel längst eingetragen hat
// und nur die Team-Kennung vergaß (§8).
func TestDieMeldungNenntDieFehlendeAngabe(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, scrydex.Zugang{Schluessel: "geheim"},
		func(http.ResponseWriter, *http.Request) {})

	_, err := klient.Karten(t.Context(), "X")

	if !strings.Contains(err.Error(), "Team") {
		t.Errorf("die Meldung %q nennt die fehlende Angabe nicht", err)
	}
}

// **Ein 401 ist ein Zugangsfehler und kein Ausfall.**
func TestEinVierhunderteinsIstEinZugangsfehler(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, vollerZugang(),
		func(schreiber http.ResponseWriter, _ *http.Request) {
			schreiber.WriteHeader(http.StatusUnauthorized)

			schreibe(t, schreiber, beispiel(t, "ohnezugang.json"))
		})

	_, err := klient.Karten(t.Context(), "X")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if !strings.Contains(err.Error(), "Invalid credentials") {
		t.Errorf("die Meldung %q gibt den Dienst nicht wieder", err)
	}
}

// **Eine Karte kommt durch.**
func TestEineKarteKommtDurch(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, vollerZugang(),
		func(schreiber http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiber, beispiel(t, "karten.json"))
		})

	karten, err := klient.Karten(t.Context(), "Blitzschlag")
	if err != nil {
		t.Fatalf("fragen: %v", err)
	}

	if len(karten) != 1 {
		t.Fatalf("%d Karten, erwartet 1", len(karten))
	}

	eine := karten[0]

	for _, fall := range []struct{ feld, gelesen, erwartet string }{
		{"Kennung", eine.Kennung, "prb-141"},
		{"Name", eine.Name, "Blitzschlag"},
		{"Manakosten", eine.Manakosten, "{R}"},
		{"Typzeile", eine.Typzeile, "Instant"},
		{"Seltenheit", eine.Seltenheit, "common"},
		{"Regeltext", eine.Regeltext, "Blitzschlag fügt einem Ziel 3 Schadenspunkte zu."},
		{"Kuenstler", eine.Kuenstler, "Selbst gezeichnet"},

		// **Der Setcode steckt in einem Unterverbund.** Scrydex nennt ihn
		// `expansion.code`, nicht `set` — wer flach läse, bekäme nichts.
		{"Setcode", eine.Setcode, "PRB"},
		{"Setname", eine.Setname, "Probeset"},
	} {
		if fall.gelesen != fall.erwartet {
			t.Errorf("%s ist %q, erwartet %q", fall.feld, fall.gelesen, fall.erwartet)
		}
	}
}

// **Die Blätterung wird gesetzt, und die Seitengröße ist die erlaubte.**
//
// **Nachgelesen am 2026-09-02**: „The highest allowable value is `100`".
func TestDieBlaetterungWirdGesetzt(t *testing.T) {
	t.Parallel()

	var gesehen string

	klient := klientGegen(t, vollerZugang(),
		func(schreiber http.ResponseWriter, gestellt *http.Request) {
			gesehen = gestellt.URL.RawQuery

			schreibe(t, schreiber, beispiel(t, "karten.json"))
		})

	if _, err := klient.Karten(t.Context(), "Blitz Schlag"); err != nil {
		t.Fatalf("fragen: %v", err)
	}

	for _, teil := range []string{"page_size=100", "q=", "Blitz"} {
		if !strings.Contains(gesehen, teil) {
			t.Errorf("die Abfrage %q enthält %q nicht", gesehen, teil)
		}
	}
}

// **Ein 429 kommt als Drosselung an.**
//
// **Nachgelesen am 2026-09-02**: „There is a rate limit of 100 requests per
// second … you will receive a `429 Too Many Requests` response."
func TestEinVierhundertneunundzwanzigIstEineDrossel(t *testing.T) {
	t.Parallel()

	klient := klientGegen(t, vollerZugang(),
		func(schreiber http.ResponseWriter, _ *http.Request) {
			schreiber.WriteHeader(http.StatusTooManyRequests)
		})

	if _, err := klient.Karten(t.Context(), "X"); !errors.Is(
		err, domain.ErrGedrosselt) {
		t.Errorf("Fehler ist %v, erwartet ErrGedrosselt", err)
	}
}
