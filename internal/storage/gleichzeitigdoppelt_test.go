package storage_test

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Zwei zugleich, derselbe Name: einer gewinnt, der andere hört „gibt es
// schon"** — ENG-05.AK-03 und AK-04
// (#1730).
//
// Zwei Verbindungen auf dieselbe Datei — so, wie Fenster und Kommandozeile
// nebeneinander stehen (0045). Beide legen im selben Augenblick ein Deck
// gleichen Namens an. **Keiner bekommt „belegt" und keiner „unerwarteter
// Fehler"**: der zweite wartet, bis der erste fertig ist (`busy_timeout`,
// `_txlock=immediate`), trifft dann auf die Eindeutigkeit und bekommt den
// doppelten Namen am Namensfeld — dieselbe Antwort wie hintereinander.
func TestZweiZugleichMitDemselbenNamenGebenEinenDoppelten(t *testing.T) {
	t.Parallel()

	ctx := probekontext(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	verbindungen := make([]*storage.Deckablage, 2)

	for i := range verbindungen {
		db, fehler := storage.Oeffnen(ctx, ort)
		if fehler != nil {
			t.Fatalf("öffnen: %v", fehler)
		}

		t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // am Ende eines Laufs

		if fehler := storage.Migriere(ctx, db, nil); fehler != nil {
			t.Fatalf("migrieren: %v", fehler)
		}

		verbindungen[i] = storage.NeueDeckablage(db)
	}

	var (
		start      sync.WaitGroup
		fertig     sync.WaitGroup
		ergebnisse = make([]error, len(verbindungen))
	)

	start.Add(1)

	for i, ablage := range verbindungen {
		fertig.Go(func() {
			start.Wait()

			_, ergebnisse[i] = ablage.Anlegen(ctx, domain.Deck{Name: "Gleich", Spiel: "magic"})
		})
	}

	start.Done()
	fertig.Wait()

	gelungen, doppelt := 0, 0

	for _, fehler := range ergebnisse {
		var befund *domain.FieldError

		switch {
		case fehler == nil:
			gelungen++
		case errors.Is(fehler, domain.ErrDoppelt) && errors.As(fehler, &befund) && befund.Field == "name":
			doppelt++
		default:
			t.Errorf("unerwartet: %v", fehler)
		}
	}

	if gelungen != 1 || doppelt != 1 {
		t.Errorf("%d gelungen, %d doppelt — erwartet je einer", gelungen, doppelt)
	}
}
