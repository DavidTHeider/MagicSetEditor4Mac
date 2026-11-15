package service_test

// **Zwei Vorschauläufe teilen eine Verwendung** —
// #1325.
//
// # Der Befund
//
// `go test -race ./internal/...` meldete am 2026-09-03 zwei bis drei
// Datenrennen je Lauf, alle in `Schluesselwortverwendung.Vermerke` — ein
// `append` auf einen Slice, den zwei Goroutinen zugleich anfassen.
//
// **Er war nicht neu und nicht durch eine Änderung entstanden**: derselbe Lauf
// über den Stand davor meldete ihn ebenso, sobald genug parallel lief. Ein
// grüner Lauf davor war Glück, kein Beleg.
//
// # Warum das im Betrieb zählt
//
// Eine Kartenvorschau wird **abgebrochen**, nicht abgewartet (§7): wer tippt,
// stößt mehrere Läufe an, und der vorige rechnet noch, während der nächste
// beginnt. Beide setzen Schlüsselwörter ein, und beide vermerken.
//
// # Was dieser Lauf prüft
//
// Dass gleichzeitiges Vermerken und Lesen kein Rennen ist. **Er beweist nichts
// ohne `-race`** — ohne den Melder liefe er auch mit dem Fehler grün; das
// steht hier, damit niemand mehr hineinliest.

import (
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestZweiLaeufeVermerkenOhneRennen(t *testing.T) {
	t.Parallel()

	verwendung := &service.Schluesselwortverwendung{}

	var warten sync.WaitGroup

	for i := range 8 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			verwendung.BeiFeld("rule text")
			verwendung.Vermerke("rule text", "Fliegend")

			if i%2 == 0 {
				verwendung.Raeume("rule text")
			}
		}()
	}

	warten.Wait()
}
