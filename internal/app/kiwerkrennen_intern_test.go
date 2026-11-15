package app

// **Zwei Bindungen bauen das Auftragswerk zugleich** — Befund 1 des
// Zweitblicks über Phase 5.3.
//
// # Wogegen hier geprüft wird
//
// **Wails bedient jede Bindung in einer eigenen Goroutine**, und sechs von
// ihnen bauen das Werk neu ([App.werkzeugAus]). Bis zu diesem Lauf standen
// Lesen und Schreiben von `a.kiwerk` ohne Schloss nebeneinander.
//
// **Der Race-Detektor ist hier der Prüfer und nicht die Zusicherung.** Der
// Lauf endet ohne Fehler, gleich ob das Schloss steht; rot wird er unter
// `-race`, und `internal/app` läuft dort mit (§4). Wer das Schloss entfernt,
// sieht `WARNING: DATA RACE` auf `App.kiwerk` — die Gegenprobe ist gefahren.

import (
	"sync"
	"testing"
)

func TestZweiRufeBauenDasWerkOhneRennen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&nurEingeschaltete{})
	anwendung.MitKILaufzeit(func(string) KILaufzeitquelle {
		return festeLaufzeit{modelle: einModellFuerDenBauplan()}
	})

	staende, _, standfehler := anwendung.frageAnbieter(riegelkontext(t))
	if standfehler != nil {
		t.Fatalf("frageAnbieter: %v", standfehler)
	}

	// **Beide Goroutinen starten am selben Tor.** Ohne es liefe die zweite
	// regelmäßig nach der ersten, und der Detektor sähe nichts.
	tor := make(chan struct{})

	var warten sync.WaitGroup

	for range 2 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			<-tor

			if _, _, fehler := anwendung.werkzeugAus(
				riegelkontext(t), staende, ollamaanbieter, kideckung{},
			); fehler != nil {
				t.Errorf("werkzeugAus: %v", fehler)
			}
		}()
	}

	close(tor)
	warten.Wait()

	if anwendung.kiwerk == nil {
		t.Fatal("nach zwei Rufen steht kein Werk")
	}
}
