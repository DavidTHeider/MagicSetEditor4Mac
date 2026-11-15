package service_test

// **Die Vorprüfung nennt, wer den Bestand sonst offen hält** — M02
// (#2346),
// Befund A-11. Ein Programm, das nach dem Tausch mit dem Code von vorher
// weiterläuft, ist ein Hindernis — und der Satz dazu nennt es.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type halterprobe struct {
	halter []string
	fehler error
}

func (h halterprobe) Halter(context.Context) ([]string, error) { return h.halter, h.fehler }

func TestDieVorpruefungNenntDieHalter(t *testing.T) {
	t.Parallel()

	befund, err := vollstaendig().MitHaltern(halterprobe{halter: []string{"mse-rest (4711)", "sqlite3 (815)"}}).
		Pruefe(t.Context(), passendesManifest(), 100<<20)
	if err != nil {
		t.Fatal(err)
	}

	gefunden := map[string]bool{}

	for _, hindernis := range befund.Hindernisse {
		if hindernis.Art == domain.HindernisHalter {
			gefunden[hindernis.Stelle] = true
		}
	}

	if !gefunden["mse-rest (4711)"] || !gefunden["sqlite3 (815)"] || befund.Bereit {
		t.Errorf("Befund %+v — erwartet beide Halter und nicht bereit", befund)
	}

	ohne, err := vollstaendig().MitHaltern(halterprobe{}).Pruefe(t.Context(), passendesManifest(), 100<<20)
	if err != nil || !ohne.Bereit {
		t.Errorf("ohne Halter: %+v, %v", ohne, err)
	}

	if _, fehler := vollstaendig().MitHaltern(halterprobe{fehler: errors.New("lsof fehlt")}).
		Pruefe(t.Context(), passendesManifest(), 100<<20); fehler == nil {
		t.Error("eine Frage, die scheitert, gilt als beantwortet")
	}
}
