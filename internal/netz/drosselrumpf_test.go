package netz_test

// **Ein 429 bringt seinen Rumpf mit** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Warum das gebraucht wird
//
// **Weil zwei verschiedene Dinge denselben Status tragen.** Anthropic
// antwortet mit 429 sowohl auf eine Drossel als auch auf die Sperre am
// Monatsdeckel — beide mit `error.type: "rate_limit_error"`. Der Unterschied
// steht **im Rumpf**: `error.details.error_code ==
// "enforced_spend_limit_reached"` heißt „geht nicht vorüber"
// ([0099](../../docs/entscheidungen/0099-was-anthropic-belegt.md)).
//
// **Bis zum 2026-09-11 kam bei einem 429 `Rumpf: nil` zurück.** Der Fehler
// blieb, die Auskunft war fort — und ein Aufrufer, der den Unterschied
// braucht, konnte ihn nicht sehen.
//
// # Was sich dabei **nicht** ändert
//
// **Das 429 bleibt ein Fehler**, die Drossel wird weiter gesperrt, und
// niemand bekommt eine Antwort, die wie ein Erfolg aussieht. Was dazukommt,
// sind die Bytes.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestEinGedrosselterRufBringtSeinenRumpfMit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Retry-After", "7")
			schreiben.WriteHeader(http.StatusTooManyRequests)
			_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
				`{"error":{"details":{"error_code":"enforced_spend_limit_reached"}}}`))
		}))
	t.Cleanup(server.Close)

	antwort, fehler := netz.Hole(t.Context(), netz.KlientUeber(server.Client()),
		nil, netz.Kenner{Name: "Probe", Annahme: "*/*"}, server.URL, 1<<20)

	// **Der Fehler bleibt** — das ist die Zusage, die nicht wackelt.
	if !errors.Is(fehler, netz.ErrGedrosselt) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if antwort.Status != http.StatusTooManyRequests {
		t.Errorf("der Status ist %d", antwort.Status)
	}

	// **Und der Rumpf ist da.**
	if !strings.Contains(string(antwort.Rumpf), "enforced_spend_limit_reached") {
		t.Errorf("der Rumpf ist %q", string(antwort.Rumpf))
	}

	// **Die Kopfzeilen ebenso** — sie waren schon vorher da, und ein Lauf,
	// der sie nicht prüfte, ließe offen, ob der Umbau sie mitgenommen hat.
	if antwort.Kopf.Get("Retry-After") != "7" {
		t.Errorf("die Kopfzeilen sind %v", antwort.Kopf)
	}
}

// **Und ein 429 ohne Rumpf bleibt ein 429.**
//
// **Die Gegenprobe zur Zeile darüber**: ein Lesen, das scheitert, darf den
// Fehler nicht verschlucken und nicht ersetzen.
func TestEinGedrosselterRufOhneRumpfBleibtEiner(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusTooManyRequests)
		}))
	t.Cleanup(server.Close)

	_, fehler := netz.Hole(t.Context(), netz.KlientUeber(server.Client()),
		nil, netz.Kenner{Name: "Probe", Annahme: "*/*"}, server.URL, 1<<20)

	if !errors.Is(fehler, netz.ErrGedrosselt) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}
