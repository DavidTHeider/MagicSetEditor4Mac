package service_test

// **Der Wirt legt die Reihenfolge fest** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// Das Issue verlangt es als Erstes: „G1-Host verarbeitet gleichzeitig
// eintreffende Commands in **einer** Reihenfolge."
//
// **Zwei Spieler, die im selben Augenblick ziehen, ziehen nacheinander** —
// und welcher zuerst, entscheidet der Wirt. Ohne diese Stelle hätten zwei
// Teilnehmer zwei verschiedene Geschichten derselben Partie.

import (
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDieErsteFolgeIstEins(t *testing.T) {
	t.Parallel()

	reihe := service.NeueReihenfolge(1)

	epoche, folge := reihe.Naechste()
	if epoche != 1 || folge != 1 {
		t.Errorf("Epoche %d, Folge %d — erwartet 1 und 1", epoche, folge)
	}
}

func TestJedeFolgeKommtNurEinmal(t *testing.T) {
	t.Parallel()

	reihe := service.NeueReihenfolge(1)
	gesehen := map[uint64]bool{}

	for i := 0; i < 100; i++ {
		_, folge := reihe.Naechste()
		if gesehen[folge] {
			t.Fatalf("Folge %d kam zweimal", folge)
		}

		gesehen[folge] = true
	}
}

// **Und auch dann nicht, wenn hundert Goroutinen zugleich fragen.**
//
// Das ist die Zusage, für die es diese Stelle gibt: „gleichzeitig eintreffende
// Commands in einer Reihenfolge". Der Race-Detektor läuft über dieses Paket
// (AGENTS.md §4).
func TestHundertGleichzeitigeBekommenHundertFolgen(t *testing.T) {
	t.Parallel()

	reihe := service.NeueReihenfolge(1)

	var (
		wand    sync.WaitGroup
		schloss sync.Mutex
		gesehen = map[uint64]bool{}
	)

	for i := 0; i < 100; i++ {
		wand.Add(1)

		go func() {
			defer wand.Done()

			_, folge := reihe.Naechste()

			schloss.Lock()
			defer schloss.Unlock()

			gesehen[folge] = true
		}()
	}

	wand.Wait()

	if len(gesehen) != 100 {
		t.Errorf("%d verschiedene Folgen, erwartet 100", len(gesehen))
	}
}

// **Eine neue Epoche fängt von vorn an.**
//
// Sie entsteht, wenn der Wirt neu anfängt — nach einem Neustart etwa. Die
// Folge beginnt wieder bei eins, und die Epoche trennt die beiden Läufe.
func TestEineNeueEpocheFaengtBeiEinsAn(t *testing.T) {
	t.Parallel()

	reihe := service.NeueReihenfolge(1)
	reihe.Naechste()
	reihe.Naechste()

	reihe.NeueEpoche()

	epoche, folge := reihe.Naechste()
	if epoche != 2 || folge != 1 {
		t.Errorf("Epoche %d, Folge %d — erwartet 2 und 1", epoche, folge)
	}
}

// **Der Stand lässt sich ablesen, ohne ihn weiterzuzählen.**
//
// Ein Gast, der nach dem Stand fragt, soll ihn nicht verändern — sonst
// entstünde eine Lücke allein durchs Hinsehen.
func TestDerStandLaesstSichAblesen(t *testing.T) {
	t.Parallel()

	reihe := service.NeueReihenfolge(3)
	reihe.Naechste()
	reihe.Naechste()

	epoche, folge := reihe.Stand()
	if epoche != 3 || folge != 2 {
		t.Errorf("Epoche %d, Folge %d — erwartet 3 und 2", epoche, folge)
	}

	nochmal, wieder := reihe.Stand()
	if nochmal != epoche || wieder != folge {
		t.Error("das Ablesen hat den Stand verändert")
	}
}

// **Eine Reihenfolge fängt in der Epoche an, die ihr gegeben wird** — nach
// einem Neustart ist das die gespeicherte, um eins erhöht.
func TestEineReihenfolgeFaengtInIhrerEpocheAn(t *testing.T) {
	t.Parallel()

	epoche, folge := service.NeueReihenfolge(7).Stand()
	if epoche != 7 || folge != 0 {
		t.Errorf("Epoche %d, Folge %d — erwartet 7 und 0", epoche, folge)
	}
}
