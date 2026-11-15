package service_test

// **Zwei Schreiber an derselben Partie warten aufeinander** — gefunden mit
// zwei Anwendungen in P9b
// (#2377):
// das Fenster des Wirts und die Schleife eines Gastes schreiben in **einem**
// Prozess. Lasen beide dieselbe Fassung, scheiterte der zweite an der Prüfung
// des Speichers — und dieser Konflikt beendete die Sitzung des Gastes.

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// haltendeProbe hält den ersten Lesenden an, bis er weiter darf.
type haltendeProbe struct {
	*partieprobe

	schloss sync.Mutex
	gelesen int
	halt    chan struct{}
	weiter  chan struct{}
}

func (h *haltendeProbe) Partie(ctx context.Context, kennung string) (domain.Partie, error) {
	h.schloss.Lock()
	h.gelesen++
	erster := h.gelesen == 1
	partie, fehler := h.partieprobe.Partie(ctx, kennung)
	h.schloss.Unlock()

	if erster {
		close(h.halt)
		<-h.weiter
	}

	return partie, fehler
}

func (h *haltendeProbe) Anwenden(
	ctx context.Context, erwartet int64, befehlID string,
	zustand domain.Partie, ereignisse []domain.Ereignis,
) error {
	h.schloss.Lock()
	defer h.schloss.Unlock()

	return h.partieprobe.Anwenden(ctx, erwartet, befehlID, zustand, ereignisse)
}

func (h *haltendeProbe) Befehlsergebnis(
	ctx context.Context, partie, kennung string,
) ([]domain.Ereignis, bool, error) {
	h.schloss.Lock()
	defer h.schloss.Unlock()

	return h.partieprobe.Befehlsergebnis(ctx, partie, kennung)
}

func (h *haltendeProbe) LetzteKette(ctx context.Context, partie string) (kette string, folge int64, fehler error) {
	h.schloss.Lock()
	defer h.schloss.Unlock()

	return h.partieprobe.LetzteKette(ctx, partie)
}

func TestZweiBefehleAnDerselbenPartieWartenAufeinander(t *testing.T) {
	t.Parallel()

	probe := &haltendeProbe{
		partieprobe: &partieprobe{partien: map[string]domain.Partie{}},
		halt:        make(chan struct{}), weiter: make(chan struct{}),
	}
	leitung := service.NeuePartieleitung(probe, service.NeuerSandkasten(service.NeuerZufall()))

	if _, fehler := leitung.Anlegen(context.Background(), einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	befehl := func(kennung string) domain.Spielbefehl {
		return domain.Spielbefehl{
			PartieID: "p1", Kennung: kennung, Absender: "sitz-1",
			Art: "objekt.tappen", Nutzlast: map[string]any{"objekt": "o1"},
		}
	}

	erster := make(chan error, 1)
	go func() {
		_, fehler := leitung.Befehl(context.Background(), befehl("b1"))
		erster <- fehler
	}()

	<-probe.halt

	zweiter := make(chan error, 1)
	go func() {
		enttappen := befehl("b2")
		enttappen.Art = "objekt.enttappen"
		_, fehler := leitung.Befehl(context.Background(), enttappen)
		zweiter <- fehler
	}()

	// **Der zweite darf den ersten nicht überholen** — kurz Zeit geben, es
	// zu versuchen, dann den ersten weiterlassen.
	time.Sleep(20 * time.Millisecond)
	close(probe.weiter)

	for name, kanal := range map[string]chan error{"der erste": erster, "der zweite": zweiter} {
		if fehler := <-kanal; fehler != nil {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}
