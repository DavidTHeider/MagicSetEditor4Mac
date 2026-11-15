package app_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Ausfuhr bekommt die Stile des Sets** — QA-04-F3
// (#2039).
// Ohne die Quelle stand `set.stylesheet` leer da, und die Lackey-Vorlage nahm
// ihren Zweig ohne Stil. Hier steht die **Naht**: welches Set und welcher Name
// bei der Stilwahl ankommen. Die Suche im Paketverzeichnis selbst hält
// `internal/service/stilpaketwahl_test.go` über die Saat.
func TestDieAusfuhrBekommtDieStileDesSets(t *testing.T) {
	gesehen := map[string]string{}

	probe := &vorlagenprobe{
		vorlagen: []msepaket.Exportvorlage{probevorlage()},
		ausgabe:  "fertig",
		waehrend: func(mittel service.Ausfuhrmittel) {
			for _, paket := range []string{"", "kunst", "gibt-es-nicht"} {
				if stil, da := mittel.Stile(paket); da {
					gesehen[paket] = stil.VollerName
				}
			}
		},
	}

	anwendung := mitVorlagen(t, probe)

	wahl := &ausfuhrstilwahl{}
	anwendung.MitStilwahl(wahl)

	antwort := anwendung.VorlageAusfuehren(t.Context(), "2",
		"probe.mse-export-template", filepath.Join(t.TempDir(), "a.txt"), nil, nil)
	if antwort.Fehler != nil {
		t.Fatalf("%v", antwort.Fehler)
	}

	if gesehen[""] != "der Stil des Sets" || gesehen["kunst"] != "kunst" {
		t.Errorf("gesehen: %v", gesehen)
	}

	if _, da := gesehen["gibt-es-nicht"]; da {
		t.Error("ein Stil, den es nicht gibt, wurde gefunden")
	}

	for _, set := range wahl.sets {
		if set != 2 {
			t.Errorf("gefragt wurde Set %d, ausgeführt Set 2", set)
		}
	}
}

// ausfuhrstilwahl kennt den Stil des Sets und `kunst`, und schreibt mit,
// nach welchem Set gefragt wird.
type ausfuhrstilwahl struct {
	stilwaehler

	sets []int64
}

func (w *ausfuhrstilwahl) Stilpaket(_ context.Context, setID int64, paket string) (msepaket.Stil, error) {
	w.sets = append(w.sets, setID)

	switch paket {
	case "":
		return msepaket.Stil{VollerName: "der Stil des Sets"}, nil
	case "kunst":
		return msepaket.Stil{VollerName: "kunst"}, nil
	}

	return msepaket.Stil{}, domain.ErrNotFound
}
