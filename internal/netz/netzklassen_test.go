package netz_test

// **Drei Netzfehler, drei Klassen — und jede mit ihrem Rat** — D02
// (#2233).
//
// Bis dahin gingen sie roh hinaus — „… fragen: dial tcp: lookup …: no such
// host", „… timeout awaiting response headers", „die Antwort lesen: i/o
// timeout" — und das Fenster zeigte für alle drei „Es ist ein unerwarteter
// Fehler aufgetreten". Hier fahren sie über den **echten** Transport des
// Klienten (`NeuerKlient`, mit seinen Fristen) gegen eine Gegenstelle, die
// genau so versagt:
//
//   - ein Name, den es nicht gibt (`.invalid`, RFC 6761),
//   - ein Port, an dem niemand annimmt,
//   - ein Dienst, der den Kopf nicht schickt — die **Kopffrist** greift,
//   - eine Antwort, die nach dem Kopf stockt — die **Stillstandsfrist** greift.
//
// **Quantitativ**: die zwei Fristen werden gemessen, nicht nur ihr Ausgang.
// **Ohne Geheimnis**: ein Schlüssel in der Adresse steht in keiner Meldung.

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// geschlossenerPort gibt eine Adresse, an der gerade niemand annimmt.
func geschlossenerPort(t *testing.T) string {
	t.Helper()

	horcher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	adresse := horcher.Addr().String()
	_ = horcher.Close() //nolint:errcheck // er soll gerade nicht mehr da sein

	return adresse
}

func TestDreiNetzfehlerTragenIhreKlasse(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "an")

	// **Getrennte Fristen, wie im Betrieb** (30 s und 60 s): bei gleichen
	// Werten gewinnt die Lesefrist auf der Verbindung den Wettlauf, und der
	// Lauf sah den Fehler der Kopffrist nie — den, der als Abbruch galt.
	fristen := netz.Fristen{Verbindung: time.Second, Kopf: 300 * time.Millisecond, Stillstand: 600 * time.Millisecond}
	klient := netz.NeuerKlient(fristen)
	kenner := netz.Kenner{Name: "test", Annahme: "*/*"}

	stumm := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	t.Cleanup(stumm.Close)

	stockend := httptest.NewServer(http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
		schreiben.Header().Set("Content-Length", "100")
		schreiben.WriteHeader(http.StatusOK)
		schreibe(t, schreiben, []byte("anfang"))
		if fehler := http.NewResponseController(schreiben).Flush(); fehler != nil {
			t.Errorf("den Anfang schicken: %v", fehler)
		}

		time.Sleep(2 * time.Second)
	}))
	t.Cleanup(stockend.Close)

	const geheim = "GEHEIM-2233"

	for _, fall := range []struct {
		name    string
		adresse string
		will    error
		frist   time.Duration
	}{
		{"ein Name, den es nicht gibt", "http://gibtsnicht.invalid/karten?schluessel=" + geheim, domain.ErrNameUnaufloesbar, 0},
		{"ein Port, an dem niemand annimmt", "http://" + geschlossenerPort(t) + "/karten?schluessel=" + geheim, domain.ErrNichtErreichbar, 0},
		{"ein Dienst, der nicht antwortet", stumm.URL + "/karten?schluessel=" + geheim, domain.ErrKeineAntwort, fristen.Kopf},
		{"eine Antwort, die stockt", stockend.URL + "/karten?schluessel=" + geheim, domain.ErrAntwortStockt, fristen.Stillstand},
	} {
		t.Run(fall.name, func(t *testing.T) {
			beginn := time.Now()

			_, fehler := netz.Hole(t.Context(), klient, nil, kenner, fall.adresse, 1<<20)
			dauer := time.Since(beginn)
			t.Logf("gemessen: %s nach %v", fall.name, dauer.Round(time.Millisecond))

			if !errors.Is(fehler, fall.will) || !errors.Is(fehler, domain.ErrGegenstelle) {
				t.Fatalf("der Fehler %v trägt nicht die Klasse %v", fehler, fall.will)
			}

			if strings.Contains(fehler.Error(), geheim) {
				t.Errorf("die Meldung trägt das Geheimnis aus der Adresse: %v", fehler)
			}

			// **Eine Frist der Gegenstelle ist kein Abbruch** — `net/http`
			// lässt seine Kopffrist auf `context.DeadlineExceeded` mit Ja
			// antworten, und jede Übersetzung fragt den Abbruch zuerst: das
			// Fenster blendete den Fehler dann aus.
			if errors.Is(fehler, context.DeadlineExceeded) || errors.Is(fehler, context.Canceled) {
				t.Errorf("der Fehler %v gilt als Abbruch", fehler)
			}

			if fall.frist > 0 && (dauer < fall.frist || dauer > fall.frist+time.Second) {
				t.Errorf("nach %v, erwartet die Frist von %v", dauer, fall.frist)
			}
		})
	}
}

// **Die Frist des Aufrufers ist keine Gegenstelle** — sie endet als
// Fristablauf, auch wenn sie greift, während der Kopf aussteht, und die
// Adresse verliert trotzdem ihr Geheimnis.
func TestDieFristDesAufrufersBleibtEinAbbruch(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "an")

	stumm := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	t.Cleanup(stumm.Close)

	klient := netz.NeuerKlient(netz.Fristen{Verbindung: time.Second, Kopf: 5 * time.Second, Stillstand: 5 * time.Second})

	ctx, abbrechen := context.WithTimeout(t.Context(), 200*time.Millisecond)
	t.Cleanup(abbrechen)

	_, fehler := netz.Hole(ctx, klient, nil, netz.Kenner{Name: "test", Annahme: "*/*"}, stumm.URL+"/?schluessel=GEHEIM-2233", 1<<20)

	if !errors.Is(fehler, context.DeadlineExceeded) || errors.Is(fehler, domain.ErrGegenstelle) {
		t.Errorf("die Frist des Aufrufers endete als %v", fehler)
	}

	if strings.Contains(fehler.Error(), "GEHEIM-2233") {
		t.Errorf("die Meldung trägt das Geheimnis: %v", fehler)
	}
}

// **Auch ein 429 nennt die Adresse ohne Geheimnis** — es ging mit der
// ganzen Abfrage in die Meldung.
func TestEinGedrosselterRufNenntDieAdresseOhneGeheimnis(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "an")

	gedrosselt := httptest.NewServer(http.HandlerFunc(func(schreiben http.ResponseWriter, _ *http.Request) {
		schreiben.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(gedrosselt.Close)

	_, fehler := netz.Hole(t.Context(), netz.NeuerKlient(netz.UeblicheFristen()), nil,
		netz.Kenner{Name: "test", Annahme: "*/*"}, gedrosselt.URL+"/karten?schluessel=GEHEIM-2233", 1<<20)

	if !errors.Is(fehler, netz.ErrGedrosselt) || strings.Contains(fehler.Error(), "GEHEIM-2233") ||
		!strings.Contains(fehler.Error(), "/karten") {
		t.Errorf("die Meldung: %v", fehler)
	}
}

// **Die Frist, die griff, ist die genannte** — ist die Stillstandsfrist
// kürzer als die des Kopfes, liegt sie als Lesefrist auf der Verbindung und
// greift zuerst. Die Meldung nannte sonst eine Frist, die gar nicht ablief.
func TestKeineAntwortNenntDieFristDieGriff(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "an")

	stumm := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	t.Cleanup(stumm.Close)

	fristen := netz.Fristen{Verbindung: time.Second, Kopf: 3 * time.Second, Stillstand: 300 * time.Millisecond}

	_, fehler := netz.Hole(t.Context(), netz.NeuerKlient(fristen), nil, netz.Kenner{Name: "test", Annahme: "*/*"}, stumm.URL, 1<<20)

	var gegenstelle *domain.Gegenstellenfehler
	if !errors.Is(fehler, domain.ErrKeineAntwort) || !errors.As(fehler, &gegenstelle) || gegenstelle.Frist != fristen.Stillstand {
		t.Errorf("der Fehler %v nennt nicht die Stillstandsfrist von %v", fehler, fristen.Stillstand)
	}
}
