package app

// **Der Kanal zum Unified Log nach dem Nachlauf über Phase 3** — 0.9.0
// (#2324),
// Befunde 5 und 6 des Nachlaufs (`docs/audits/0.9.0/zweitblicke/3-nachlauf.md`).

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// **Auf einen Prozess wartet genau einer** — Befund 5: `exec.Cmd.Wait`
// darf nur einmal laufen, und der Schluss nach seiner Frist rief es neben
// dem Arbeiter ein zweites Mal.
func TestAufEinenLoggerWartetGenauEiner(t *testing.T) { //nolint:paralleltest // die Frist ist eine Paketvariable
	vorher := systemkanalfrist
	systemkanalfrist = 100 * time.Millisecond

	t.Cleanup(func() { systemkanalfrist = vorher })

	angekommen := make(chan struct{}, 1)
	ende := make(chan struct{})

	var (
		gewartet    atomic.Int32
		abgebrochen atomic.Bool
	)

	kanal := neuerSystemkanal(func() (laufenderProzess, error) {
		return laufenderProzess{
			ein:    nimmtAlles{angekommen: angekommen},
			warten: func() error { gewartet.Add(1); <-ende; return nil },
			abbrechen: func() error {
				if abgebrochen.CompareAndSwap(false, true) {
					close(ende)
				}

				return nil
			},
		}, nil
	})

	if _, err := kanal.Write([]byte("eine Zeile\n")); err != nil {
		t.Fatal(err)
	}

	<-angekommen

	if err := kanal.Close(); err == nil {
		t.Fatal("der Schluss nannte die abgelaufene Frist nicht")
	}

	<-kanal.fertig

	// **Der Abholer des Schlusses läuft danach** — er hat einen Augenblick,
	// um ein zweites Mal zu warten; ob der Riegel greift, zeigt
	// `TestEinmalLaesstNurEinenWarten` ohne Uhr.
	time.Sleep(100 * time.Millisecond)

	if n := gewartet.Load(); n != 1 {
		t.Errorf("%d-mal gewartet — Wait darf nur einmal laufen", n)
	}
}

// **Der Riegel selbst**: acht zugleich, einer wartet, alle hören dasselbe.
func TestEinmalLaesstNurEinenWarten(t *testing.T) {
	t.Parallel()

	var aufgerufen atomic.Int32

	gescheitert := errors.New("beendet")
	warten := einmal(func() error {
		aufgerufen.Add(1)
		time.Sleep(10 * time.Millisecond)

		return gescheitert
	})

	var gruppe sync.WaitGroup

	for range 8 {
		gruppe.Go(func() {
			if err := warten(); !errors.Is(err, gescheitert) {
				t.Errorf("gehört: %v", err)
			}
		})
	}

	gruppe.Wait()

	if n := aufgerufen.Load(); n != 1 {
		t.Errorf("%d-mal gewartet", n)
	}
}

// scheitertEinmal nimmt jede Zeile — bis auf die erste, die gibt sie mit
// einem Fehler zurück, wie ein `logger`, der eben gestorben ist.
type scheitertEinmal struct {
	versuche   *atomic.Int32
	angekommen chan struct{}
}

func (s scheitertEinmal) Write(zeile []byte) (int, error) {
	if s.versuche.Add(1) == 1 {
		return 0, errors.New("broken pipe")
	}

	select {
	case s.angekommen <- struct{}{}:
	default:
	}

	return len(zeile), nil
}

func (scheitertEinmal) Close() error { return nil }

// **Eine verlorene Zeile wird genannt** — Befund 6: der Fehler eines
// gestorbenen Prozesses stand nur bis zur nächsten geglückten Zeile in
// `grund`, und der Schluss sagte nichts. Jetzt zählt sie mit, und der Schluss
// nennt die Zahl.
func TestEineVerloreneZeileNenntDerSchluss(t *testing.T) {
	t.Parallel()

	var versuche atomic.Int32

	angekommen := make(chan struct{}, 1)

	kanal := neuerSystemkanal(func() (laufenderProzess, error) {
		return laufenderProzess{
			ein:       scheitertEinmal{versuche: &versuche, angekommen: angekommen},
			warten:    func() error { return nil },
			abbrechen: func() error { return nil },
		}, nil
	})

	for _, zeile := range []string{"geht verloren\n", "kommt an\n"} {
		if _, err := kanal.Write([]byte(zeile)); err != nil && !strings.Contains(err.Error(), "Unified Log") {
			t.Fatal(err)
		}
	}

	<-angekommen

	err := kanal.Close()
	if err == nil || !strings.Contains(err.Error(), "1 Zeile") {
		t.Errorf("der Schluss sagte %v — eine Zeile ging verloren", err)
	}
}
