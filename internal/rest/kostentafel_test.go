package rest_test

// **Der Dienst bucht nach der Kostentafel** — Folgebefund F1 des Lastlaufs
// QA-08 (#2003).
//
// Bis zum 2026-10-02 stand hier eine eigene, kurze Liste: drei teure Wege,
// der Rest nach der Methode. Eine Karte zeichnen kostete 2, ein Bogen 2 —
// über MCP 8. Jetzt fragt der Dienst [domain.KostenklasseFuer] mit dem Muster
// des Weges, den die Anfrage trifft.

import (
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// buchprobe merkt jede Buchung — und weist alles ab, wenn sie leer ist.
type buchprobe struct {
	leer bool

	mu      sync.Mutex
	gebucht []domain.Kostenklasse
}

func (b *buchprobe) Abbuchen(_ string, klasse domain.Kostenklasse, _ bool) (bool, time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.gebucht = append(b.gebucht, klasse)

	if b.leer {
		return false, time.Second
	}

	return true, 0
}

func (b *buchprobe) buchungen() []domain.Kostenklasse {
	b.mu.Lock()
	defer b.mu.Unlock()

	return append([]domain.Kostenklasse(nil), b.gebucht...)
}

// gebuchtFuer ruft einen Weg und gibt, was dabei gebucht wurde.
func gebuchtFuer(t *testing.T, methode, pfad string) []domain.Kostenklasse {
	t.Helper()

	probe := &buchprobe{}
	dienst := neuerDienst(t).MitHaushalt(probe, neueSperrprobe(festeUhr))
	ruf(t, dienst, endpunkt{methode: methode, weg: pfad, rumpf: "{}"}, gueltig)

	return probe.buchungen()
}

func TestJederWegDerTafelBuchtSeineKlasse(t *testing.T) {
	t.Parallel()

	for _, zeile := range domain.Kostentafel() {
		for _, weg := range zeile.Wege[domain.FlaecheREST] {
			methode, muster, _ := strings.Cut(weg, " ")
			gebucht := gebuchtFuer(t, methode, strings.ReplaceAll(muster, "{}", "1"))

			// **Ein Abbruch fragt das Guthaben gar nicht erst** — was nichts
			// kostet, hat nichts zu buchen.
			if zeile.Klasse == domain.KostenAbbruch {
				if len(gebucht) != 0 {
					t.Errorf("%s buchte %v", weg, gebucht)
				}

				continue
			}

			if len(gebucht) != 1 || gebucht[0] != zeile.Klasse {
				t.Errorf("%s (%s) buchte %v, die Tafel sagt %s", weg, zeile.Faehigkeit, gebucht, zeile.Klasse)
			}
		}
	}
}

// **Was nicht in der Tafel steht, kostet nach seiner Methode** — die Art,
// die der Dienst kennt.
func TestEinWegAusserhalbDerTafelBuchtNachSeinerArt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		methode, pfad string
		will          domain.Kostenklasse
	}{
		{http.MethodGet, "/sets", domain.KostenLesen},
		{http.MethodPost, "/sets/neu", domain.KostenSchreiben},
		{http.MethodGet, "/mtgjson/spiele", domain.KostenLesen},
		{http.MethodGet, "/gibtesnicht", domain.KostenLesen},

		// **Die Tafel nennt jede Rechnung des Symboleditors mit ihrem Wert**
		// (#2020) — eine, die sie nicht nennt, kostet nach ihrer Methode.
		{http.MethodPost, "/symbol/gibtesnicht", domain.KostenSchreiben},
	}

	for _, fall := range faelle {
		if gebucht := gebuchtFuer(t, fall.methode, fall.pfad); len(gebucht) != 1 || gebucht[0] != fall.will {
			t.Errorf("%s %s buchte %v, erwartet %s", fall.methode, fall.pfad, gebucht, fall.will)
		}
	}

	// **Auch ohne Anfrage**: mit Abfrage, und ein Pfad, der keiner ist,
	// kostet nach seiner Methode.
	for _, fall := range []struct {
		methode, pfad string
		will          domain.Kostenklasse
	}{
		{http.MethodPost, "/karten/7/bild", domain.KostenRendern},
		{http.MethodGet, "/sammlung/ausgeben?format=csv", domain.KostenAusfuhr},
		{http.MethodPost, "kein Pfad", domain.KostenSchreiben},
	} {
		if klasse := rest.KostenklasseFuer(fall.methode, fall.pfad); klasse != fall.will {
			t.Errorf("KostenklasseFuer(%s %q): %s, erwartet %s", fall.methode, fall.pfad, klasse, fall.will)
		}
	}
}

// **Ein Abbruch wird nie gedrosselt** — auch nicht bei leerem Guthaben. Er ist
// die eine Anfrage, die Last nimmt; im Lastlauf bekam er ein 429, und der
// Auftrag lief zu Ende.
func TestEinAbbruchWirdNieGedrosselt(t *testing.T) {
	t.Parallel()

	probe := &buchprobe{leer: true}
	dienst := neuerDienst(t).MitHaushalt(probe, neueSperrprobe(festeUhr))

	// **Erst der Beleg, dass das Guthaben wirklich leer ist** — sonst sagte
	// das Nein darunter nichts.
	if antwort := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/sets"}, gueltig); antwort.Code != http.StatusTooManyRequests {
		t.Fatalf("das leere Guthaben ließ durch: %d", antwort.Code)
	}

	for _, pfad := range []string{"/auftraege/unbekannt", "/index/aufbau"} {
		if antwort := ruf(t, dienst, endpunkt{methode: http.MethodDelete, weg: pfad}, gueltig); antwort.Code == http.StatusTooManyRequests {
			t.Errorf("DELETE %s wurde gedrosselt: %s", pfad, antwort.Body.String())
		}
	}
}
