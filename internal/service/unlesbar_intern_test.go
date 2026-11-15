package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Was an der Datei scheitert, ist eine Ablehnung mit Stelle** — und was
// nicht an ihr scheitert, ordnet `unlesbar` nicht ein (FMT-06.AK-04, #1695).
func TestUnlesbarOrdnetEin(t *testing.T) {
	t.Parallel()

	// **Die Stelle nennt die Datei im Paket, nicht das Paket zweimal** —
	// `msepaket` setzt den Paketnamen davor, und der Satz beginnt schon mit
	// ihm.
	stelle := &msepaket.Zerteilfehler{
		Stelle:  msepaket.Stelle{Datei: "probe.mse-set/set", Zeile: 5, Spalte: 14},
		Meldung: "diese Zeile ist kein gültiges UTF-8",
		Grund:   msepaket.ErrKodierung,
	}

	faelle := []struct {
		name   string
		fehler error
		satz   string // leer: nicht eingeordnet
	}{
		{
			"eine Stelle in der Datei", fmt.Errorf("lesen: %w", stelle),
			"probe.mse-set lässt sich nicht lesen — in »set«, Zeile 5, Spalte 14: diese Zeile ist kein gültiges UTF-8",
		},
		{
			"ein kaputtes Archiv", fmt.Errorf("öffnen: %w", zip.ErrFormat),
			"probe.mse-set ist kein lesbares Archiv — beschädigt oder unvollständig",
		},
		{
			"eine Prüfsumme, die nicht stimmt", zip.ErrChecksum,
			"probe.mse-set ist kein lesbares Archiv — beschädigt oder unvollständig",
		},
		{
			"eine Datei über der Grenze", fmt.Errorf("lesen: %w", msepaket.ErrZuGross),
			"probe.mse-set lässt sich nicht öffnen — eine Datei darin ist größer als 256 MiB",
		},
		{
			"ein Archiv über der Ressourcenregel", fmt.Errorf("öffnen: %w", msepaket.ErrArchivZuGross),
			"probe.mse-set lässt sich nicht öffnen — das Archiv hat mehr als 100000 Einträge oder ein Verzeichnis über 16 MiB",
		},
		{"eine Datei, die es nicht gibt", os.ErrNotExist, ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			eingeordnet := unlesbar("/irgendwo/probe.mse-set", fall.fehler)

			if fall.satz == "" {
				if eingeordnet != nil {
					t.Fatalf("eingeordnet, obwohl es nicht an der Datei liegt: %v", eingeordnet)
				}

				return
			}

			if !errors.Is(eingeordnet, domain.ErrValidation) {
				t.Errorf("keine Validierung: %v", eingeordnet)
			}

			if !errors.Is(eingeordnet, fall.fehler) {
				t.Errorf("die Ursache ist verloren: %v", eingeordnet)
			}

			if satz := domain.SatzVon(eingeordnet); satz != fall.satz {
				t.Errorf("der Satz lautet\n  %q\nerwartet\n  %q", satz, fall.satz)
			}
		})
	}
}
