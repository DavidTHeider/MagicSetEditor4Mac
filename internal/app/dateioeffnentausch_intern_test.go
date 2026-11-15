package app

// **Ein Doppelklick im Finder und ein Tausch des Bestands gehen nacheinander**
// — Nachlauf über Phase G, zweiter Teil, Befund 3
// (#2292).
//
// # Was schiefging
//
// Der Finder-Weg liest in Go ein und meldet danach nur die **Kennung**. Ein
// Tausch wartet laufende Aufrufe nicht ab: las der Doppelklick noch in den
// alten Bestand, kam seine Kennung nach dem Neuanfang an, und das Fenster
// schlug sie im neuen auf — ein anderes Set oder keines.
//
// **Nicht parallel**: die Fälle halten die Sperre des Tauschs.

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// finderprobe ist eine laufende Anwendung, deren Meldungen ein Kanal hört.
func finderprobe(t *testing.T) (a *App, gemeldet chan string) {
	t.Helper()

	a = gestartet(t)
	gemeldet = make(chan string, 4)

	a.MitMeldung(func(_ context.Context, ereignis string, _ ...any) {
		switch ereignis {
		case EreignisDateiGeoeffnet:
			gemeldet <- "geöffnet"
		case EreignisDateiFehler:
			gemeldet <- "Fehler"
		}
	})
	a.Bereit(t.Context())

	return a, gemeldet
}

func TestEinDoppelklickWartetEinenTauschAb(t *testing.T) {
	a, gemeldet := finderprobe(t)

	// **Ein Tausch läuft** — er hält die Sperre vom Schließen bis zum
	// letzten Öffnen ([App.imTausch]).
	a.lauf.wechsel.Lock()

	gehalten := true
	t.Cleanup(func() {
		if gehalten {
			a.lauf.wechsel.Unlock()
		}
	})

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		a.DateiOeffnen(filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	}()

	select {
	case was := <-gemeldet:
		t.Fatalf("während des Tauschs %s gemeldet — der Doppelklick las in den alten Bestand", was)
	case <-time.After(300 * time.Millisecond):
	}

	gehalten = false
	a.lauf.wechsel.Unlock()

	// **Und danach öffnet er** — im neuen Bestand; ohne diesen Beleg hieße
	// das Schweigen oben nur, dass niemand zuhört.
	select {
	case was := <-gemeldet:
		if was != "geöffnet" {
			t.Errorf("nach dem Tausch %s gemeldet, erwartet war das geöffnete Set", was)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("nach dem Tausch kam keine Meldung")
	}

	<-fertig
}

// **Nach dem Beginn des Schlusses öffnet ein Doppelklick nichts mehr** — der
// Bestand geht gleich zu, und ein Fenster, das etwas zeigen könnte, auch.
func TestEinDoppelklickNachDemSchlussOeffnetNichts(t *testing.T) {
	a, gemeldet := finderprobe(t)

	a.lauf.tor.Store(torEndet)
	t.Cleanup(func() { a.lauf.tor.Store(torOffen) })

	a.DateiOeffnen(filepath.Join("..", "..", "testdaten", "probe.mse-set"))

	select {
	case was := <-gemeldet:
		t.Errorf("nach dem Schluss %s gemeldet", was)
	default:
	}
}
