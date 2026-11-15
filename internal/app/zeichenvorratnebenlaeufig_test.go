package app_test

// **Nebenläufige Rufe über einem behaltenen Werk** —
// #2005.
//
// `mse-rest` bedient Anfragen nebeneinander, und seit die Zeichennaht ihr
// Werk behält, gehen sie alle über **dieselben** Speicher: Stile, Schriften,
// Bilder, Ebenen. Der Lauf hier zeichnet, wertet aus und wechselt den Stil
// zugleich; was er belegt, belegt er **unter `-race`** — so fährt ihn die
// Pipeline. Ohne den Detektor prüft er nur, dass kein Ruf scheitert.

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

func TestNebenlaeufigeRufeTeilenDasWerkOhneWettlauf(t *testing.T) {
	lage := zeichenzuhause(t)
	ort := t.TempDir()

	var (
		gruppe sync.WaitGroup
		sperre sync.Mutex
		fehler []error
		werke  = map[*app.Zeichenwerk]bool{}
	)

	melde := func(err error) {
		sperre.Lock()
		defer sperre.Unlock()

		fehler = append(fehler, err)
	}

	for faden := range 6 {
		gruppe.Go(func() {
			for runde := range 3 {
				werk, err := app.Zeichennaht(lage.db)
				if err != nil {
					melde(err)

					return
				}

				sperre.Lock()
				werke[werk] = true
				sperre.Unlock()

				ziel := filepath.Join(ort, fmt.Sprintf("karte-%d-%d.png", faden, runde))
				if _, fehler := werk.SchreibeBild(probekontext(t), lage.karteID, ziel, render.GueteMittel); fehler != nil {
					melde(fmt.Errorf("zeichnen: %w", fehler))
				}

				konsole, err := app.SkriptnahtUeber(lage.db)
				if err != nil {
					melde(err)

					continue
				}

				if _, err := konsole.Auswerten(probekontext(t), lage.karteID, "card.name"); err != nil {
					melde(fmt.Errorf("auswerten: %w", err))
				}
			}
		})
	}

	// **Und einer wechselt den Stil dazwischen hin und her** — das Vergessen
	// läuft gegen die Leser oben.
	gruppe.Go(func() {
		weg, err := app.StilnahtUeber(lage.db)
		if err != nil {
			melde(err)

			return
		}

		for runde := range 4 {
			ziel := []string{"probe-bild", "probe-schlicht"}[runde%2]
			if err := weg.Wechseln(probekontext(t), lage.setID, ziel); err != nil {
				melde(fmt.Errorf("wechseln: %w", err))
			}
		}
	})

	gruppe.Wait()

	if err := errors.Join(fehler...); err != nil {
		t.Fatal(err)
	}

	if len(werke) != 1 {
		t.Errorf("%d Werke über einem Bestand — erwartet eines", len(werke))
	}
}
