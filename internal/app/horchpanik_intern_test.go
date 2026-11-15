package app

// **Ein Horchposten, der zusammenbricht, sagt es** — BK-05a
// (#2111).
//
// Horchen und Fernhorchen bedienen ihren Lauscher in einer eigenen Goroutine.
// Bricht sie zusammen — im Lauscher selbst, also in fremdem Code —, endete
// bisher der Prozess. Jetzt endet nur dieser Posten, und sein Stand sagt es,
// statt weiter „läuft" zu zeigen. Ein neuer Start gelingt danach.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zerbrechenderLauscher bricht beim ersten `Nimm` zusammen.
type zerbrechenderLauscher struct{ probelauscher }

func (*zerbrechenderLauscher) Nimm(context.Context) (service.Verbindung, error) {
	panic("der Lauscher bricht zusammen")
}

// zerbrechenderFernhorcher bricht beim ersten `Nimm` zusammen.
type zerbrechenderFernhorcher struct{ probefernhorcher }

func (*zerbrechenderFernhorcher) Nimm(context.Context) (service.Verbindung, error) {
	panic("der Fernhorcher bricht zusammen")
}

// bisNichtMehr wartet, bis `laeuft` falsch sagt — höchstens fünf Sekunden.
func bisNichtMehr(laeuft func() bool) bool {
	frist := time.Now().Add(5 * time.Second)

	for laeuft() {
		if time.Now().After(frist) {
			return false
		}

		time.Sleep(10 * time.Millisecond)
	}

	return true
}

func TestEinZusammenbrechenderHorchpostenEndetNurSelbst(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineHorchende(t, &zerbrechenderLauscher{}, nil)

	if antwort := anwendung.HorchStarten(ctx, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	if !bisNichtMehr(func() bool { return anwendung.HorchStand(ctx).Stand.Laeuft }) {
		t.Fatalf("nach dem Zusammenbruch steht noch: %+v", anwendung.HorchStand(ctx).Stand)
	}

	// **Und ein neuer Start gelingt** — über einen Lauscher, der hält.
	anwendung.MitHorcher(func(context.Context, string, peer.Identitaet) (service.Hereinkommend, error) {
		return &probelauscher{}, nil
	})

	if antwort := anwendung.HorchStarten(ctx, "127.0.0.1:0"); antwort.Fehler != nil || !antwort.Stand.Laeuft {
		t.Errorf("der zweite Start: %+v", antwort)
	}

	anwendung.HorchBeenden(ctx)
}

func TestEinZusammenbrechenderFernhorcherEndetNurSelbst(t *testing.T) {
	t.Parallel()

	anwendung, ctx := eineFernhorchende(t, &zerbrechenderFernhorcher{
		probefernhorcher: probefernhorcher{einladung: "mse4mac/einladung/1/probe"},
	}, nil)

	if antwort := anwendung.FernHorchStarten(ctx); antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	if !bisNichtMehr(func() bool { return anwendung.FernHorchStand(ctx).Stand.Laeuft }) {
		t.Fatalf("nach dem Zusammenbruch steht noch: %+v", anwendung.FernHorchStand(ctx).Stand)
	}

	if stand := anwendung.FernHorchStand(ctx).Stand; stand.Einladung != "" {
		t.Errorf("die Einladung steht noch da: %q", stand.Einladung)
	}
}

// **Eine Ansage, deren Leser zusammenbricht, ist ein Fehler** — nicht das
// Ende des Prozesses, der den Dienst gerade starten wollte.
func TestEinZusammenbrechenderAnsageleserIstEinFehler(t *testing.T) {
	t.Parallel()

	if _, fehler := leseAnsage(t.Context(), zerbrechenderLeser{}, time.Second); !errors.Is(fehler, domain.ErrZusammengebrochen) {
		t.Fatalf("die Ansage gab %v", fehler)
	}
}

// zerbrechenderLeser bricht beim Lesen zusammen.
type zerbrechenderLeser struct{}

func (zerbrechenderLeser) Read([]byte) (int, error) {
	panic("der Leser bricht zusammen")
}

// **Ein verspäteter Zusammenbruch räumt nicht den neuen Lauf** — beendet und
// neu gestartet heißt: der Stand gehört einem anderen `fertig`.
func TestEinVerspaeteterZusammenbruchRaeumtNichtDenNeuenLauf(t *testing.T) {
	t.Parallel()

	horchend, horchkontext := eineHorchende(t, &probelauscher{}, nil)
	if antwort := horchend.HorchStarten(horchkontext, "127.0.0.1:0"); antwort.Fehler != nil {
		t.Fatalf("HorchStarten: %+v", antwort.Fehler)
	}

	t.Cleanup(func() { horchend.HorchBeenden(horchkontext) })

	horchend.horchenZusammengebrochen(make(chan struct{}), domain.ErrZusammengebrochen)

	if !horchend.HorchStand(horchkontext).Stand.Laeuft {
		t.Error("ein fremder Zusammenbruch hat den laufenden Horchposten geräumt")
	}

	fern, fernkontext := eineFernhorchende(t, &probefernhorcher{einladung: "mse4mac/einladung/1/probe"}, nil)
	if antwort := fern.FernHorchStarten(fernkontext); antwort.Fehler != nil {
		t.Fatalf("FernHorchStarten: %+v", antwort.Fehler)
	}

	t.Cleanup(func() { fern.FernHorchBeenden(fernkontext) })

	fern.fernhorchenZusammengebrochen(make(chan struct{}), domain.ErrZusammengebrochen)

	if !fern.FernHorchStand(fernkontext).Stand.Laeuft {
		t.Error("ein fremder Zusammenbruch hat das laufende Fernhorchen geräumt")
	}
}
