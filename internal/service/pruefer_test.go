package service_test

// **Der Riegel vor jedem Weg, der Dateien schreibt** — Werkstück O1 von
// Phase 5.2 (#1425).
//
// **`query_only` ersetzt keine Autorisierung** (0078): die Platte fragt den
// Bestand nicht. Jeder Dienst mit Dateiwirkung fragt deshalb einen Prüfer,
// bevor er etwas anlegt — und ein Nein hinterlässt nichts.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// neinsager ist ein Prüfer, der jede Wirkung verweigert — und mitschreibt,
// wonach gefragt wurde.
type neinsager struct{ gefragt []domain.Wirkung }

func (n *neinsager) Pruefe(_ context.Context, wirkung domain.Wirkung) error {
	n.gefragt = append(n.gefragt, wirkung)

	return fmt.Errorf("%w: hier wird nur gelesen", domain.ErrNurLesend)
}

// jasager erlaubt alles — die Gegenprobe, dass der Riegel nicht mehr sperrt,
// als er soll.
type jasager struct{}

func (jasager) Pruefe(context.Context, domain.Wirkung) error { return nil }

func TestEinPrueferDerNeinSagtHaeltJedesSchreibenAuf(t *testing.T) {
	t.Parallel()

	ziel := t.TempDir()
	ctx := context.Background()

	faelle := map[string]func(pruefer service.Pruefer) (string, error){
		"Set sichern": func(pruefer service.Pruefer) (string, error) {
			pfad := filepath.Join(ziel, "kopie.mse-set")
			_, err := service.NeuerSetschreiber(&archivleser{}, &ortsschreiber{}).
				MitPruefer(pruefer).Schreibe(ctx, 7, pfad)

			return pfad, err
		},
		"MTG JSON": func(pruefer service.Pruefer) (string, error) {
			pfad := filepath.Join(ziel, "set.json")
			_, err := service.NeuerExporteur(probenbestand()).MitPruefer(pruefer).
				MTGJSON(ctx, 7, "magic", pfad)

			return pfad, err
		},
		"Kartenbild": func(pruefer service.Pruefer) (string, error) {
			_, err := bildausfuhr(t, planDouble{}).MitPruefer(pruefer).
				Schreibe(ctx, 1, ziel, "karte.png", service.Bildformat("png"), 0, "", render.GueteMittel)

			return filepath.Join(ziel, "karte.png"), err
		},
		"Druckbogen": func(pruefer service.Pruefer) (string, error) {
			pfad := filepath.Join(ziel, "bogen.pdf")
			_, err := service.NeuerBogenleger(
				kartenDouble{leserDouble: karten("Feuerball")}, planDouble{}, render.NeuerRasterer(nil)).
				MitPruefer(pruefer).Lege(ctx, 7, pfad, service.StandardBogen(), domain.Kartenauswahl{})

			return pfad, err
		},
	}

	for name, lauf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pruefer := &neinsager{}

			pfad, err := lauf(pruefer)
			if !errors.Is(err, domain.ErrNurLesend) {
				t.Fatalf("%s: %v, erwartet ErrNurLesend", name, err)
			}

			if _, statfehler := os.Stat(pfad); !errors.Is(statfehler, fs.ErrNotExist) {
				t.Errorf("%s: %s liegt trotzdem da (%v)", name, pfad, statfehler)
			}

			// **Gefragt wird nach dem Schreiben** — und genau einmal.
			if len(pruefer.gefragt) != 1 || pruefer.gefragt[0] != domain.WirkungSchreiben {
				t.Errorf("%s: gefragt nach %v, erwartet [lokalSchreiben]", name, pruefer.gefragt)
			}
		})
	}
}

// **Und ein Ja lässt den Weg, wie er war** — die Gegenprobe am Exporteur,
// dessen Ausgabe ohne Rasterer entsteht.
func TestEinPrueferDerJaSagtSchreibt(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "set.json")

	ausfuhr, err := service.NeuerExporteur(probenbestand()).MitPruefer(jasager{}).
		MTGJSON(context.Background(), 7, "magic", pfad)
	if err != nil || ausfuhr.Karten != 1 {
		t.Fatalf("mit Ja: %+v, %v", ausfuhr, err)
	}

	if _, statfehler := os.Stat(pfad); statfehler != nil {
		t.Errorf("die Datei fehlt: %v", statfehler)
	}
}
