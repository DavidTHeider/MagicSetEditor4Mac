package aktualisierung_test

// **Gemeldet wird das Ende, das wirklich gilt** — Nachlauf über Phase 0,
// Befund 2 (#2333).
//
// # Warum
//
// Ein 429 sperrt schon in [netz.Hole] — ohne `Retry-After` dreißig
// Sekunden —, und die Deutung von GitHub sperrt danach ein zweites Mal.
// **Die längere gilt** ([netz.Drossel.Sperre]); gemeldet wurde bis hierher
// trotzdem die eigene. Wer um 14:02 wieder fragt, weil die Meldung das
// sagte, steht vor einer Sperre bis 14:03.
//
// # Was sie nicht prüfen
//
// Die Absprache über Prozessgrenzen — sie hält `internal/netz` selbst
// (`prozessgrenze_test.go`); hier zählt, dass die Sperre das ganze Werk
// erreicht und nicht nur die eine Drossel.

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestGemeldetWirdDasWirksameEndeDerSperre(t *testing.T) {
	t.Parallel()

	// **Die Rücksetzung liegt vor der Sperre des Transports** — genau der
	// Fall, in dem beide auseinanderliefen.
	dienst := kopfstelle(t, http.StatusTooManyRequests, map[string]string{
		"X-Ratelimit-Remaining": "0",
		"X-Ratelimit-Reset":     strconv.FormatInt(time.Now().Add(10*time.Second).Unix(), 10),
	})

	_, err := probeklient(t, dienst.URL).NeuesteFreigabe(t.Context())

	frueheste := time.Now().Add(netz.SperreOhneAngabe - 5*time.Second)
	if bis := ende(t, err); bis.Before(frueheste) {
		t.Errorf("gemeldet %v, gesperrt ist aber mindestens bis %v", bis, frueheste)
	}
}

func TestDieGrenzeVonGitHubSperrtDasGanzeWerk(t *testing.T) {
	t.Parallel()

	dienst := kopfstelle(t, http.StatusForbidden, map[string]string{
		"X-Ratelimit-Remaining": "0",
		"X-Ratelimit-Reset":     strconv.FormatInt(time.Now().Add(37*time.Minute).Unix(), 10),
	})

	werk := netz.NeuesDrosselwerk(time.Now, func(time.Duration) {})
	nachbar := werk.Fuer("github-beigaben", time.Second)

	klient := aktualisierung.NeuerKlient(dienst.URL,
		netz.NeuerKlient(netz.UeblicheFristen()), werk, "")
	if _, err := klient.NeuesteFreigabe(t.Context()); err == nil {
		t.Fatal("eine erschöpfte Grenze wurde als Freigabe gelesen")
	}

	if !nachbar.Gesperrt() {
		t.Error("die Grenze von GitHub sperrt nur die eine Drossel, nicht das Werk")
	}
}
