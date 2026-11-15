package flaechen_test

// **Ruhende Verbindungen sperren niemanden aus** — QA-08
// (#1739).
//
// # Warum am gebauten Dienst
//
// **Weil die Zusage an zwei Stellen hängt**: der Lauscher in `internal/rest`
// verdrängt eine ruhende Verbindung nur, wenn der Server ihm sagt, dass sie
// ruht — und das tut `ConnState` in `cmd/mse-rest`. Ein Lauf des Lauschers
// allein bliebe grün, wenn die Wurzel die Zeile nicht hätte. Gemessen hat
// den Befund der Lastlauf (Profil `ruhend`): sechzehn ruhende Verbindungen
// eines Pools, und ein zweiter Klient wartete, bis der Leerlauf von zwei
// Minuten sie schloss.

import (
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestRuhendeVerbindungenSperrenAmDienstNiemandenAus(t *testing.T) {
	an := starteDienst(t, mitEigenerSaat(t))

	// **Sechzehn — so viele, wie der Dienst annimmt** (`cmd/mse-rest`).
	const plaetze = 16

	pool := &http.Client{Transport: &http.Transport{MaxConnsPerHost: plaetze, MaxIdleConnsPerHost: plaetze}}
	t.Cleanup(pool.CloseIdleConnections)

	var fertig sync.WaitGroup

	for range plaetze {
		fertig.Go(func() { rufMit(t, pool, an, http.StatusOK) })
	}

	fertig.Wait()

	// **Ein zweiter Klient kommt durch**, und zwar gleich.
	fremd := &http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Second}
	t.Cleanup(fremd.CloseIdleConnections)

	rufMit(t, fremd, an, http.StatusOK)
}

// rufMit fragt die Grenzen über einen eigenen Klienten.
func rufMit(t *testing.T, klient *http.Client, an dienst, erwartet int) {
	t.Helper()

	anfrage, fehler := http.NewRequestWithContext(t.Context(), http.MethodGet, an.adresse+"/grenzen", http.NoBody)
	if fehler != nil {
		t.Fatal(fehler)
	}

	anfrage.Header.Set("Authorization", "Bearer "+an.marke)

	antwort, fehler := klient.Do(anfrage)
	if fehler != nil {
		t.Errorf("GET /grenzen: %v", fehler)

		return
	}

	defer func() { _ = antwort.Body.Close() }() //nolint:errcheck // gelesen

	_, _ = io.Copy(io.Discard, antwort.Body) //nolint:errcheck // der Status zählt

	if antwort.StatusCode != erwartet {
		t.Errorf("GET /grenzen: %d", antwort.StatusCode)
	}
}
