package netz_test

import (
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Jede Anfrage nennt sich** (#1002).
//
// Nachgelesen am 2026-08-23 unter `https://scryfall.com/docs/api`:
// „All HTTP requests to api.scryfall.com **must** include a User-Agent header",
// und „The Accept header **must** be present". Beides ist Pflicht, nicht
// Höflichkeit — eine Anfrage ohne sie wird abgewiesen.
func TestJedeAnfrageNenntSich(t *testing.T) {
	t.Parallel()

	var gesehen http.Header

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gesehen = anfrage.Header.Clone()
			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(gegenstelle.Close)

	kenner := netz.Kenner{Name: "MagicSetEditor4Mac/0.2.0", Annahme: "application/json"}

	if _, fehler := netz.Hole(t.Context(), netz.KlientUeber(gegenstelle.Client()), nil,
		kenner, gegenstelle.URL, 1<<20); fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	if name := gesehen.Get("User-Agent"); name != kenner.Name {
		t.Errorf("User-Agent %q, erwartet war %q", name, kenner.Name)
	}

	if annahme := gesehen.Get("Accept"); annahme != kenner.Annahme {
		t.Errorf("Accept %q, erwartet war %q", annahme, kenner.Annahme)
	}
}

// **Ein 429 wird gemeldet, nicht wiederholt.**
//
// „It is not acceptable to ignore HTTP 429 responses" und „Continuing to
// overload the API after this point may result in a temporary or permanent
// ban" — beides nachgelesen am 2026-08-23.
func TestEin429WirdGemeldetUndNichtWiederholt(t *testing.T) {
	t.Parallel()

	rufe := 0

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			rufe++
			schreiben.Header().Set("Retry-After", "30")
			schreiben.WriteHeader(http.StatusTooManyRequests)
		}))
	t.Cleanup(gegenstelle.Close)

	antwort, fehler := netz.Hole(t.Context(), netz.KlientUeber(gegenstelle.Client()), nil,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gegenstelle.URL, 1<<20)

	if !errors.Is(fehler, netz.ErrGedrosselt) {
		t.Errorf("Fehler %v ist kein %v", fehler, netz.ErrGedrosselt)
	}

	if rufe != 1 {
		t.Errorf("die Anfrage wurde %dmal gestellt — wiederholt wird nicht", rufe)
	}

	// **Der Kopf reist mit**, damit der Aufrufer den `Retry-After` lesen kann.
	if antwort.Kopf.Get("Retry-After") != "30" {
		t.Errorf("der Retry-After fehlt in der Antwort: %v", antwort.Kopf)
	}

	// **Und die Meldung sagt, wie lange.** Ein „zu viele Anfragen" ohne Dauer
	// lässt den Aufrufer raten.
	if !strings.Contains(fehler.Error(), "30s") {
		t.Errorf("die Meldung nennt die Wartezeit nicht: %v", fehler)
	}
}

// **`Retry-After` wird in beiden Formen gelesen.**
//
// # Warum beide
//
// RFC 9110 lässt Sekunden **und** einen Zeitpunkt zu. Wer nur die eine liest,
// hält die Kopfzeile bei der anderen für nicht vorhanden — und wartet gar
// nicht. Das ist genau das Übergehen, das die Bedingungen verbieten.
func TestDerRetryAfterWirdInBeidenFormenGelesen(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	faelle := map[string]struct {
		wert     string
		erwartet time.Duration
		gefunden bool
	}{
		"in Sekunden":                        {"30", 30 * time.Second, true},
		"als Zeitpunkt":                      {jetzt.Add(time.Minute).UTC().Format(http.TimeFormat), time.Minute, true},
		"ein Zeitpunkt in der Vergangenheit": {jetzt.Add(-time.Minute).UTC().Format(http.TimeFormat), 0, true},
		"eine Null":                          {"0", 0, true},
		"etwas Negatives":                    {"-5", 0, true},
		"gar nichts":                         {"", 0, false},
		"Unsinn":                             {"bald", 0, false},
		// **Mehr, als eine Dauer fasst, ist das meiste und nicht wenig**
		// (J05, #2269) — umgelaufen wäre es eine negative Wartezeit.
		"über die Dauer hinaus": {"9223372037", math.MaxInt64, true},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			kopf := http.Header{}
			if fall.wert != "" {
				kopf.Set("Retry-After", fall.wert)
			}

			dauer, gefunden := netz.WarteZeit(kopf, jetzt)
			if gefunden != fall.gefunden {
				t.Fatalf("gefunden = %v, erwartet war %v", gefunden, fall.gefunden)
			}

			if dauer != fall.erwartet {
				t.Errorf("Dauer %v, erwartet war %v", dauer, fall.erwartet)
			}
		})
	}
}

// **Eine zu große Antwort wird gemeldet, nicht gekappt.**
//
// Sonst sähe ein zu großer Rumpf aus wie ein vollständiger, der an einer
// unglücklichen Stelle endet (#898).
func TestEineZuGrosseAntwortWirdGemeldet(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiben, []byte(strings.Repeat("x", 100)))
		}))
	t.Cleanup(gegenstelle.Close)

	if _, fehler := netz.Hole(t.Context(), netz.KlientUeber(gegenstelle.Client()), nil,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gegenstelle.URL, 10); fehler == nil {
		t.Error("eine zu große Antwort kam durch")
	}
}

// **Und eine, die passt, kommt an.**
//
// Ohne diesen Fall wäre auch eine Grenze grün, die jede Antwort abwiese.
func TestEinePassendeAntwortKommtAn(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiben, []byte(`{"gut":true}`))
		}))
	t.Cleanup(gegenstelle.Close)

	antwort, fehler := netz.Hole(t.Context(), netz.KlientUeber(gegenstelle.Client()), nil,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gegenstelle.URL, 1<<20)
	if fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	if antwort.Status != http.StatusOK {
		t.Errorf("Status %d", antwort.Status)
	}

	if string(antwort.Rumpf) != `{"gut":true}` {
		t.Errorf("Rumpf %q", antwort.Rumpf)
	}
}

// **Die Drossel wird eingehalten, bevor die Anfrage rausgeht.**
//
// Gedrosselt wird **unter** der genannten Zahl (§16, Regel 4) — nicht als
// Antwort auf ein 429.
func TestDieDrosselGreiftVorDerAnfrage(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(gegenstelle.Close)

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	gewartet := 0

	drossel := netz.NeueDrossel(500*time.Millisecond,
		func() time.Time { return jetzt },
		func(dauer time.Duration) { gewartet++; jetzt = jetzt.Add(dauer) })

	kenner := netz.Kenner{Name: "test", Annahme: "*/*"}

	for range 3 {
		if _, fehler := netz.Hole(t.Context(), netz.KlientUeber(gegenstelle.Client()), drossel,
			kenner, gegenstelle.URL, 1<<20); fehler != nil {
			t.Fatalf("holen: %v", fehler)
		}
	}

	// Die erste geht sofort, die beiden anderen warten.
	if gewartet != 2 {
		t.Errorf("es wurde %dmal gewartet, erwartet waren zwei", gewartet)
	}
}

// **Eine Adresse, die keine ist, wird gemeldet, bevor etwas hinausgeht.**
func TestEineUnbrauchbareAdresseWirdGemeldet(t *testing.T) {
	t.Parallel()

	if _, fehler := netz.Hole(t.Context(), netz.KlientUeber(http.DefaultClient), nil,
		netz.Kenner{Name: "test", Annahme: "*/*"},
		"http://[::1]:namedport/", 1<<20); fehler == nil {
		t.Error("eine unbrauchbare Adresse kam durch")
	}
}

// **Ein Abbruch trifft schon die Drossel.**
//
// Wer abbricht, während gewartet wird, soll nicht erst noch eine Anfrage
// hinausgehen sehen.
func TestEinAbbruchTrifftSchonDieDrossel(t *testing.T) {
	t.Parallel()

	gerufen := 0

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gerufen++
			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(gegenstelle.Close)

	drossel := netz.NeueDrossel(time.Minute, time.Now, nil)

	// Die erste füllt die Drossel.
	if _, fehler := netz.Hole(t.Context(), netz.KlientUeber(gegenstelle.Client()), drossel,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gegenstelle.URL, 1<<20); fehler != nil {
		t.Fatalf("die erste Anfrage: %v", fehler)
	}

	abgebrochen, abbrechen := contextMitAbbruch(t)
	abbrechen()

	if _, fehler := netz.Hole(abgebrochen, netz.KlientUeber(gegenstelle.Client()), drossel,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gegenstelle.URL, 1<<20); fehler == nil {
		t.Error("die zweite Anfrage ging trotz Abbruch hinaus")
	}

	if gerufen != 1 {
		t.Errorf("die Gegenstelle wurde %dmal gefragt, erwartet war einmal", gerufen)
	}
}

// **Ein 429 ohne `Retry-After` sagt trotzdem, was zu tun ist.**
//
// Ohne diesen Fall bliebe der Zweig ungeprüft, der die Meldung ohne Zahl baut
// — und eine Meldung, die dann leer bliebe, wäre schlimmer als eine ohne
// Angabe.
func TestEin429OhneRetryAfterSagtTrotzdemWasZuTunIst(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusTooManyRequests)
		}))
	t.Cleanup(gegenstelle.Close)

	_, fehler := netz.Hole(t.Context(), netz.KlientUeber(gegenstelle.Client()), nil,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gegenstelle.URL, 1<<20)

	if !errors.Is(fehler, netz.ErrGedrosselt) {
		t.Fatalf("Fehler %v ist kein %v", fehler, netz.ErrGedrosselt)
	}

	if !strings.Contains(fehler.Error(), "später") {
		t.Errorf("die Meldung sagt nicht, was zu tun ist: %v", fehler)
	}
}

// **Ein POST mit JSON-Rumpf trägt Methode, Typ und Inhalt** —
// #1348.
//
// # Warum das hier steht und nicht beim Aufrufer
//
// `POST /cards/collection` fragt bis zu 75 Karten auf einmal; die Kennungen
// stehen im Rumpf, weil sie in keine Abfrage passen. Alles, was [netz.Hole]
// zusagt — drosseln vor der Grenze, ein 429 als solches, den Rumpf begrenzen
// —, gilt dafür genauso, und deshalb ist es derselbe Weg.
//
// **Und deshalb steht die Methode hier fest**: `internal/scryfall` darf
// `net/http` nicht kennen (`internal/schichten/netz_test.go`), also darf es
// die Methode auch nicht nennen.
func TestEinPostTraegtSeinenRumpfUndSeinenTyp(t *testing.T) {
	t.Parallel()

	var (
		methode    string
		inhaltstyp string
		gesehen    []byte
	)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			methode = anfrage.Method
			inhaltstyp = anfrage.Header.Get("Content-Type")

			roh, fehler := io.ReadAll(anfrage.Body)
			if fehler != nil {
				t.Errorf("den Rumpf lesen: %v", fehler)
			}

			gesehen = roh

			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(gegenstelle.Close)

	kenner := netz.Kenner{Name: "MagicSetEditor4Mac/0.3.0", Annahme: "application/json"}

	if _, fehler := netz.SendeJSON(t.Context(), netz.KlientUeber(gegenstelle.Client()),
		nil, kenner, gegenstelle.URL, []byte(`{"identifiers":[]}`), 1<<20); fehler != nil {
		t.Fatalf("senden: %v", fehler)
	}

	if methode != http.MethodPost {
		t.Errorf("Methode %q, erwartet war POST", methode)
	}

	if !strings.HasPrefix(inhaltstyp, "application/json") {
		t.Errorf("Content-Type %q", inhaltstyp)
	}

	if string(gesehen) != `{"identifiers":[]}` {
		t.Errorf("der Rumpf kam als %q an", gesehen)
	}
}

// **Ein GET trägt keinen Rumpf und keinen Typ.**
//
// Die Gegenprobe zur Zeile darüber: ohne sie wäre auch ein Weg grün, der
// jedem Aufruf einen `Content-Type` anhängte.
func TestEinGetTraegtKeinenRumpfUndKeinenTyp(t *testing.T) {
	t.Parallel()

	var (
		inhaltstyp string
		laenge     int
	)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			inhaltstyp = anfrage.Header.Get("Content-Type")

			roh, _ := io.ReadAll(anfrage.Body) //nolint:errcheck // ein leerer Rumpf
			laenge = len(roh)

			schreiben.WriteHeader(http.StatusOK)
		}))
	t.Cleanup(gegenstelle.Close)

	if _, fehler := netz.Hole(t.Context(), netz.KlientUeber(gegenstelle.Client()), nil,
		netz.Kenner{Name: "x", Annahme: "application/json"},
		gegenstelle.URL, 1<<20); fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	if inhaltstyp != "" {
		t.Errorf("Content-Type %q — ein GET hat keinen", inhaltstyp)
	}

	if laenge != 0 {
		t.Errorf("%d Bytes im Rumpf eines GET", laenge)
	}
}
