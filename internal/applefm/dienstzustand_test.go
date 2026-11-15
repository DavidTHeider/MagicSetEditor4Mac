package applefm_test

// **Abschalten und Absturz** — Z2 und Z3 des Zweitblicks über Phase 6.10
// (#1671).
//
// Z2: „für diese Sitzung aus" setzte einen Schalter in der Anwendung und
// beendete danach den Dienst; ein Vorschlag, der den Anbieter vorher gewählt
// hatte, startete `fm serve` danach neu. Jetzt kennt der **Dienst** den
// Zustand, unter derselben Sperre wie den Start.
//
// Z3: ein abgestürztes `fm serve` blieb bis zum Beenden der Anwendung der
// gemerkte Klient — jede Frage ging an einen toten Socket.

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
)

// zaehlprozess ist ein Probeprozess, der zählt, wie oft er beendet wurde.
type zaehlprozess struct {
	server  *http.Server
	beendet *atomic.Int32
}

func (p zaehlprozess) Beende() error {
	p.beendet.Add(1)

	return p.server.Close()
}

// zaehldienst baut einen Dienst mit einem Starter, der zählt — und vor dem
// Start auf `vorher` wartet, wenn es gesetzt ist.
type zaehldienst struct {
	dienst    *applefm.Dienst
	gestartet atomic.Int32
	beendet   atomic.Int32

	sperre  sync.Mutex
	letzter *http.Server
}

func neuerZaehldienst(t *testing.T, vorher func()) *zaehldienst {
	t.Helper()

	ordner, fehler := os.MkdirTemp("", "fm")
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	t.Cleanup(func() { _ = os.RemoveAll(ordner) }) //nolint:errcheck // nur zum Aufräumen

	z := &zaehldienst{}
	z.dienst = applefm.NeuerDienst(ordner, func(ctx context.Context, argumente ...string) (applefm.Prozess, error) {
		if vorher != nil {
			vorher()
		}

		z.gestartet.Add(1)

		horcher, err := (&net.ListenConfig{}).Listen(ctx, "unix", argumente[len(argumente)-1])
		if err != nil {
			return nil, err
		}

		// **Ein Server und nicht nur ein Horcher**: ein Absturz schließt auch
		// die offenen Verbindungen — sonst antwortete die gehaltene
		// Keep-Alive-Verbindung weiter, und der Lauf prüfte nichts.
		server := &http.Server{ //nolint:gosec // Probe ohne Fristen
			Handler: http.HandlerFunc(func(antwort http.ResponseWriter, _ *http.Request) {
				_, _ = antwort.Write([]byte(`{"models":[{"available":true,"name":"system"}]}`)) //nolint:errcheck // Probe
			}),
		}

		go func() { _ = server.Serve(horcher) }() //nolint:errcheck // endet mit Close

		z.sperre.Lock()
		z.letzter = server
		z.sperre.Unlock()

		return zaehlprozess{server: server, beendet: &z.beendet}, nil
	})

	t.Cleanup(func() { _ = z.dienst.Beende() }) //nolint:errcheck // nur zum Aufräumen

	return z
}

// **Abgeschaltet startet nichts** — und angeschaltet wieder.
func TestEinAbgeschalteterDienstStartetNicht(t *testing.T) {
	t.Parallel()

	z := neuerZaehldienst(t, nil)

	if fehler := z.dienst.Abschalten(); fehler != nil {
		t.Fatalf("abschalten: %v", fehler)
	}

	if _, fehler := z.dienst.Klient(t.Context()); !errors.Is(fehler, applefm.ErrAbgeschaltet) {
		t.Errorf("abgeschaltet: %v", fehler)
	}

	if !z.dienst.Abgeschaltet() {
		t.Error("der Dienst sagt nicht, dass er aus ist")
	}

	if z.gestartet.Load() != 0 {
		t.Errorf("gestartet: %d", z.gestartet.Load())
	}

	z.dienst.Anschalten()

	if z.dienst.Abgeschaltet() {
		t.Error("der Dienst sagt, er sei aus")
	}

	if _, fehler := z.dienst.Klient(t.Context()); fehler != nil || z.gestartet.Load() != 1 {
		t.Errorf("angeschaltet: %v, gestartet %d", fehler, z.gestartet.Load())
	}
}

// **Ein Abschalten während eines Starts beendet, was gestartet wurde** — die
// Reihenfolge aus Z2: der Start läuft schon, dann kommt „aus".
func TestAbschaltenWaehrendEinesStarts(t *testing.T) {
	t.Parallel()

	drin, weiter := make(chan struct{}), make(chan struct{})
	z := neuerZaehldienst(t, func() {
		close(drin)
		<-weiter
	})

	start := make(chan error, 1)
	go func() {
		_, fehler := z.dienst.Klient(context.Background())
		start <- fehler
	}()

	<-drin

	ab := make(chan error, 1)
	go func() { ab <- z.dienst.Abschalten() }()

	close(weiter)

	// **Seit BK-02 bricht „aus" den Start ab**, statt auf ihn zu warten
	// (#2110): wer gerade startete, bekommt den Abbruch — und nicht einen
	// Klienten, dessen Prozess gleich danach beendet wird. Bis dahin lief
	// der Start zu Ende, und `Abschalten` wartete bis zu zehn Sekunden.
	if fehler := <-start; !errors.Is(fehler, context.Canceled) {
		t.Fatalf("start: %v", fehler)
	}

	if fehler := <-ab; fehler != nil {
		t.Fatalf("abschalten: %v", fehler)
	}

	if z.beendet.Load() != 1 {
		t.Errorf("beendet: %d — ein Prozess lief nach dem Abschalten weiter", z.beendet.Load())
	}

	if _, fehler := z.dienst.Klient(t.Context()); !errors.Is(fehler, applefm.ErrAbgeschaltet) {
		t.Errorf("danach: %v", fehler)
	}
}

// **Ein abgestürzter Dienst wird beim nächsten Ruf neu gestartet** — und der
// alte aufgeräumt.
func TestEinToterDienstWirdNeuGestartet(t *testing.T) {
	t.Parallel()

	z := neuerZaehldienst(t, nil)

	if _, fehler := z.dienst.Klient(t.Context()); fehler != nil {
		t.Fatalf("erster Start: %v", fehler)
	}

	// **Der Absturz**: Socket und Verbindungen schließen, ohne dass der
	// Dienst es weiß.
	z.sperre.Lock()
	_ = z.letzter.Close() //nolint:errcheck // der Absturz selbst
	z.sperre.Unlock()

	klient, fehler := z.dienst.Klient(t.Context())
	if fehler != nil || z.gestartet.Load() != 2 || z.beendet.Load() != 1 {
		t.Fatalf("neu: %v, gestartet %d, beendet %d", fehler, z.gestartet.Load(), z.beendet.Load())
	}

	if fehler := klient.Gesund(t.Context()); fehler != nil {
		t.Errorf("gesund: %v", fehler)
	}
}

// **Ein abgebrochener Auftrag beendet keinen gesunden Dienst** — N3 des
// Nachlaufs über Phase 6.10: die Gesundheitsfrage scheiterte am Kontext des
// einen Rufs, und der Dienst, den ein zweiter Ruf gerade nutzte, wurde
// beendet.
func TestEinAbgebrochenerRufBeendetKeinenGesundenDienst(t *testing.T) {
	t.Parallel()

	z := neuerZaehldienst(t, nil)

	if _, fehler := z.dienst.Klient(t.Context()); fehler != nil {
		t.Fatalf("erster Start: %v", fehler)
	}

	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, fehler := z.dienst.Klient(abgebrochen); !errors.Is(fehler, context.Canceled) {
		t.Errorf("abgebrochen: %v", fehler)
	}

	if _, fehler := z.dienst.Klient(t.Context()); fehler != nil || z.gestartet.Load() != 1 || z.beendet.Load() != 0 {
		t.Errorf("danach: %v, gestartet %d, beendet %d", fehler, z.gestartet.Load(), z.beendet.Load())
	}
}
