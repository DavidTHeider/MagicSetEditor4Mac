package applefm_test

// **Ein Stop wartet nicht auf einen Start, der hängt** — BK-02.AK-02
// (#2110).
//
// `Klient` hält die Sperre des Dienstes, während es auf `/health` wartet — bis
// zu zehn Sekunden. `Beende` nahm dieselbe Sperre und wartete mit: wer die
// Anwendung schloss, während das Gerätemodell anlief, sah ein Fenster, das
// nicht zuging.

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
)

// stummerProzess läuft und antwortet nie — ein Start, der hängt.
type stummerProzess struct{ beendet atomic.Bool }

func (p *stummerProzess) Beende() error {
	p.beendet.Store(true)

	return nil
}

func TestEinStopBeendetEinenHaengendenStartInDerFrist(t *testing.T) {
	t.Parallel()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	prozess := &stummerProzess{}
	gestartet := make(chan struct{})

	dienst := applefm.NeuerDienst(ordner, func(context.Context, ...string) (applefm.Prozess, error) {
		close(gestartet)

		return prozess, nil
	})

	startfehler := make(chan error, 1)

	go func() {
		_, fehler := dienst.Klient(context.Background())
		startfehler <- fehler
	}()

	<-gestartet

	angefangen := time.Now()

	if fehler := dienst.Beende(); fehler != nil {
		t.Errorf("Beende: %v", fehler)
	}

	if dauer := time.Since(angefangen); dauer > time.Second {
		t.Errorf("der Stop wartete %v auf den Start", dauer)
	}

	select {
	case fehler := <-startfehler:
		if fehler == nil {
			t.Error("der abgebrochene Start meldete Erfolg")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("der Start kehrte nach dem Stop nicht zurück")
	}

	// **Und kein eigener Kindprozess bleibt** — der Start, der abbrach, hat
	// ihn beendet.
	if !prozess.beendet.Load() {
		t.Error("der hängende Prozess blieb stehen")
	}

	// **Ein zweiter Stop ist keiner.**
	if fehler := dienst.Beende(); fehler != nil {
		t.Errorf("zweites Beende: %v", fehler)
	}
}
