package schriften_test

// **Auch der Ersatz wird nebenläufig gefragt** — gefunden in der Nacht zum
// 2026-09-25, als die WebKit-Probe `play-beginn` zwei von sechs Malen den
// Server verlor: „fatal error: concurrent map writes" in `Suche.Ersatz`.
//
// Der Anlass ist derselbe wie beim Test daneben
// (`TestDieSucheHaeltNebenlaeufigenFragenStand`): **eine Starthand legt
// sieben Karten auf einmal auf den Tisch**, und ihre Bilder werden
// gleichzeitig gezeichnet. `Suche.Suche` bekam damals ein Schloss;
// `Ersatz` und `Fundort` lesen und schreiben dieselbe Art Zwischenspeicher
// und bekamen keines. Die Klasse war nicht behoben, nur der Fall (§10).
//
// Mit `-race` sind diese Läufe die Wächter darüber.

import (
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

func TestDerErsatzHaeltNebenlaeufigenFragenStand(t *testing.T) {
	suche := schriften.Neue([]string{verzeichnisMitSchrift(t, "MPlantin.ttf")})

	var warten sync.WaitGroup

	for stelle := range 16 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			// **Verschiedene Zeichen**, damit jeder Lauf in den Speicher
			// schreibt und nicht nur liest.
			suche.Ersatz(rune(0x2190 + stelle))
		}()
	}

	warten.Wait()
}

func TestDerFundortHaeltNebenlaeufigenFragenStand(t *testing.T) {
	suche := schriften.Neue([]string{verzeichnisMitSchrift(t, "MPlantin.ttf")})

	var warten sync.WaitGroup

	for range 16 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			if _, err := suche.Fundort("MPlantin", false); err != nil {
				t.Errorf("Fundort: %v", err)
			}
		}()
	}

	warten.Wait()
}
