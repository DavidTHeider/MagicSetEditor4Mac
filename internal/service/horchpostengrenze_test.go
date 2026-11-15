package service_test

// **Wer schweigt, hält den Posten nicht fest** — Befund Z3 des Zweitblicks
// über Phase 6.7 (NET-Teil): vor der Ratengrenze der Einlösung lag
// Hintergrundarbeit ohne Obergrenze und ein Handschlag ohne Frist.

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schweiger ist eine Verbindung, die nichts sagt, bis ihr Kontext endet.
type schweiger struct{ geschlossen atomic.Bool }

func (*schweiger) Kanal() string             { return "schweiger" }
func (*schweiger) Gegenstelle() peer.Kennung { return "" }

func (*schweiger) Sende(context.Context, []byte) error { return nil }

func (*schweiger) Empfange(ctx context.Context) ([]byte, error) {
	<-ctx.Done()

	return nil, ctx.Err()
}

func (s *schweiger) Schliesse() error {
	s.geschlossen.Store(true)

	return nil
}

func TestSchweigendeGegenueberHaltenDenPostenNichtFest(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	var (
		schloss  sync.Mutex
		gemeldet []error
	)

	posten := service.NeuerHorchposten(paar.dort, func(fehler error) {
		schloss.Lock()
		defer schloss.Unlock()

		gemeldet = append(gemeldet, fehler)
	}).MitHandschlagfrist(100 * time.Millisecond)

	schweigende := make([]*schweiger, service.HoechstensHandschlaege+1)
	folge := make([]func() (service.Verbindung, error), 0, len(schweigende))

	for stelle := range schweigende {
		schweigende[stelle] = &schweiger{}
		folge = append(folge, func() (service.Verbindung, error) { return schweigende[stelle], nil })
	}

	anfang := time.Now()
	posten.Bediene(t.Context(), &horchdoppel{folge: folge})

	// **Die Frist beendet jeden Handschlag** — sonst kehrte `Bediene` nie
	// zurück, denn es wartet auf alle.
	if dauer := time.Since(anfang); dauer > 5*time.Second {
		t.Errorf("der Posten stand %s", dauer)
	}

	// **Und wer über der Grenze kam, wurde gleich geschlossen und gemeldet.**
	zuViele := 0

	for _, fehler := range gemeldet {
		if errors.Is(fehler, service.ErrZuVieleHandschlaege) {
			zuViele++
		}
	}

	if zuViele != 1 || !errors.Is(service.ErrZuVieleHandschlaege, domain.ErrGedrosselt) {
		t.Errorf("über der Grenze gemeldet: %d von %v", zuViele, gemeldet)
	}

	for stelle, eine := range schweigende {
		if !eine.geschlossen.Load() {
			t.Errorf("Verbindung %d blieb offen", stelle)
		}
	}
}

// **Eine Paarung erreicht die Aufnahme** — und belegt ihren Platz bis dahin.
func TestEinePaarungErreichtDieAufnahme(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)

	var aufgenommen atomic.Value

	posten := service.NeuerHorchposten(paar.dort, nil).
		MitAufnahme(func(_ context.Context, griff service.Angenommene) {
			aufgenommen.Store(griff.Stand().Zweck)
		})

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		posten.Bediene(t.Context(), &horchdoppel{folge: []func() (service.Verbindung, error){
			func() (service.Verbindung, error) { return paar.dortigeLeitung, nil },
		}})
	}()

	if _, fehler := paar.hier.Oeffne(t.Context(), paar.dortige.Kennung(), domain.ZweckPaarung); fehler != nil {
		t.Fatalf("Oeffne: %v", fehler)
	}

	<-fertig

	if zweck, _ := aufgenommen.Load().(domain.Zweck); zweck != domain.ZweckPaarung { //nolint:errcheck // nur gelesen
		t.Errorf("aufgenommen: %v", aufgenommen.Load())
	}
}
