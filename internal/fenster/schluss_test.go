package fenster

// **Der Schluss läuft, auch wenn ein Schritt davor oder darin wirft** —
// Befund 3 des Zweitblicks über Phase 1 von 0.9.0 (F02, #2341). Bis dahin
// stand die Frage nach der Fenstergröße im selben Riegel wie der Schluss: warf
// sie, entfiel der ganze Schluss, und `panik09` belegte nur, dass der Prozess
// weg war — nicht, dass der Bestand geordnet zuging.

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestDieFenstergroesseHaeltDenSchlussNichtAuf(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		frage func() (int, int)
		da    bool
	}{
		"sie antwortet": {func() (int, int) { return 1440, 900 }, true},
		"sie wirft":     {func() (int, int) { panic("Size") }, false},
		// **Und sie hängt** — in Wails' `InvokeSync`, das nach einer inneren
		// Panik nie zurückkehrt (`mainthread.go`, beta.27).
		"sie hängt": {func() (int, int) { select {} }, false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			frist, aufhoeren := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer aufhoeren()

			breite, hoehe, da := fenstergroesse(frist, fall.frage)
			if da != fall.da || (da && (breite != 1440 || hoehe != 900)) {
				t.Errorf("ergibt %d × %d, %v", breite, hoehe, da)
			}
		})
	}
}

func TestDerSchlussLaeuftNachEinerPanikNochEinmal(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		groesse  func() (int, int)
		merken   func(context.Context, int, int)
		versuche int32
	}{
		// **Die Frage wirft**: geschlossen wird trotzdem, einmal und ohne Größe.
		"die Größe wirft": {func() (int, int) { panic("Size") }, func(context.Context, int, int) {}, 1},
		// **Das Merken wirft mitten im Schluss**: der zweite läuft ohne es.
		"das Merken wirft": {
			func() (int, int) { return 1440, 900 },
			func(context.Context, int, int) { panic("Merken") }, 2,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var versuche, geraet atomic.Int32

			schlussFahren(context.Background(), fall.groesse,
				func(ctx context.Context, bestenfalls func(context.Context)) error {
					versuche.Add(1)

					if bestenfalls != nil {
						bestenfalls(ctx)
					}

					return nil
				},
				fall.merken, func() { geraet.Add(1) })

			if versuche.Load() != fall.versuche {
				t.Errorf("der Schluss lief %d-mal, erwartet %d", versuche.Load(), fall.versuche)
			}

			if geraet.Load() != 1 {
				t.Errorf("das Gerätemodell endete %d-mal", geraet.Load())
			}
		})
	}
}

func TestEinGlatterSchlussLaeuftEinmalUndMerktDieGroesse(t *testing.T) {
	t.Parallel()

	var versuche atomic.Int32

	var gemerkt [2]int

	schlussFahren(context.Background(),
		func() (int, int) { return 1200, 800 },
		func(ctx context.Context, bestenfalls func(context.Context)) error {
			versuche.Add(1)
			bestenfalls(ctx)

			return nil
		},
		func(_ context.Context, breite, hoehe int) { gemerkt = [2]int{breite, hoehe} },
		func() {})

	if versuche.Load() != 1 || gemerkt != [2]int{1200, 800} {
		t.Errorf("%d Versuche, gemerkt %v", versuche.Load(), gemerkt)
	}
}
