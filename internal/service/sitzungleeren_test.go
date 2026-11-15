package service_test

// **Zwei, die zugleich leeren** — gefunden am 2026-09-24, als ein geteilter
// Racelauf in einer fremden Goroutine mit „slice bounds out of range [1:0]"
// abstürzte.
//
// `Leere` las die erste wartende Nachricht, gab das Schloss frei, schickte sie
// und nahm sie erst danach aus der Schlange. **Zwei Rufe zugleich** — die
// Netzschleife einer Partie (`partienetz.go`) und eine Handlung wie
// `App.PartieWuenschen` leeren denselben Griff — lasen dieselbe Nachricht,
// schickten sie **zweimal**, und der zweite schnitt aus einer leeren Schlange.

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestZweiZugleichLeerendeSchickenJedeNachrichtEinmal(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	hier, dort := paar.eineSitzung(t, domain.ZweckPartie)

	if fehler := hier.Sende(t.Context(), []byte("eins"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	// **Zwei Leerende, und die Leitung hat keinen Puffer**: der erste hängt im
	// Schicken, bis die Gegenseite liest — lange genug, dass der zweite
	// dieselbe Schlange sieht.
	var (
		warten sync.WaitGroup
		pannen = make(chan string, 2)
	)

	for range 2 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			// **Eine Panik wird gemeldet und nicht geworfen** — sonst endete
			// der ganze Lauf, und der Bericht nennte nicht diesen Fall.
			defer func() {
				if grund := recover(); grund != nil {
					pannen <- fmt.Sprint(grund)
				}
			}()

			_ = hier.Leere(t.Context()) //nolint:errcheck // geprüft wird, was ankommt
		}()
	}

	time.Sleep(50 * time.Millisecond)

	erste, fehler := dort.Empfange(t.Context())
	if fehler != nil || string(erste) != "eins" {
		t.Fatalf("zuerst kam %q (%v)", erste, fehler)
	}

	// **Danach darf nichts mehr kommen** — außer dem, was eigens geschickt
	// wird. Eine zweite „eins" wäre die doppelte Zustellung.
	zweite := make(chan []byte, 1)

	go func() {
		ctx, abbrechen := context.WithTimeout(t.Context(), 2*time.Second)
		defer abbrechen()

		nachricht, _ := dort.Empfange(ctx) //nolint:errcheck // leer heißt: nichts kam
		zweite <- nachricht
	}()

	if nochmal := <-zweite; nochmal != nil {
		t.Errorf("dieselbe Nachricht kam ein zweites Mal: %q", nochmal)
	}

	warten.Wait()
	close(pannen)

	for grund := range pannen {
		t.Errorf("ein Leerender stürzte ab: %s", grund)
	}
}
