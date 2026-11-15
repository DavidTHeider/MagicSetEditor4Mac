package service_test

// **Der Bogen meldet seinen Stand und lässt sich anhalten** — FMT-03.AK-02
// und AK-04 (#1692).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerBogenZaehltSeineKarten(t *testing.T) {
	var erledigt []int

	ctx := service.MitFortschritt(t.Context(), func(stand domain.Fortschritt) {
		if stand.Gesamt != 3 || stand.Einheit != domain.EinheitKarte {
			t.Errorf("Stand: %+v", stand)
		}

		erledigt = append(erledigt, stand.Erledigt)
	})

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	if _, fehler := bogenleger(t, 3).Lege(ctx, 7, ziel, service.StandardBogen(),
		domain.Kartenauswahl{}); fehler != nil {
		t.Fatalf("legen: %v", fehler)
	}

	// **Von null bis zur Gesamtzahl, und die letzte vor dem Schreiben** — wer
	// dann abbricht, bekommt keine Datei.
	if !slices.Equal(erledigt, []int{0, 1, 2, 3}) {
		t.Errorf("erledigt: %v", erledigt)
	}
}

// **Ein Abbruch mitten im Bogen hinterlässt keine Datei** (AK-04) — der
// Bogen entsteht im Speicher und wird erst am Ende geschrieben. **Auch nach
// der letzten Karte nicht**: die letzte Frage steht vor dem Schreiben.
func TestEinAbgebrochenerBogenHinterlaesstNichts(t *testing.T) {
	for _, fall := range []struct {
		name string
		bei  int
	}{{"nach der ersten Karte", 1}, {"nach der letzten Karte", 3}} {
		t.Run(fall.name, func(t *testing.T) {
			ctx, abbrechen := context.WithCancel(t.Context())
			defer abbrechen()

			ctx = service.MitFortschritt(ctx, func(stand domain.Fortschritt) {
				if stand.Erledigt == fall.bei {
					abbrechen()
				}
			})

			ziel := filepath.Join(t.TempDir(), "bogen.pdf")

			_, fehler := bogenleger(t, 3).Lege(ctx, 7, ziel, service.StandardBogen(), domain.Kartenauswahl{})
			if !errors.Is(fehler, context.Canceled) {
				t.Fatalf("Fehler: %v", fehler)
			}

			if _, statfehler := os.Stat(ziel); !errors.Is(statfehler, os.ErrNotExist) {
				t.Errorf("nach dem Abbruch liegt etwas am Ziel: %v", statfehler)
			}
		})
	}
}
