package rest_test

// **Die Ränder des vereinten Lauschers** — G08
// (#2272).
//
// Was ein echter Lauscher selten oder nie tut, tut hier einer zum Einsetzen:
// eine Verbindung bringen, nachdem der Dienst schon zu ist, und von selbst
// schließen, während jemand wartet. Beides muss den vereinten Lauscher
// sauber zurücklassen — keine Verbindung offen, kein Annehmen, das hängt.

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// einsetzbar ist ein Lauscher, der auf Zuruf liefert.
type einsetzbar struct {
	herein chan net.Conn
	zu     chan struct{}
	einmal atomic.Bool
}

func neuerEinsetzbarer() *einsetzbar {
	return &einsetzbar{herein: make(chan net.Conn), zu: make(chan struct{})}
}

func (e *einsetzbar) Accept() (net.Conn, error) {
	select {
	case verbindung := <-e.herein:
		return verbindung, nil
	case <-e.zu:
		return nil, net.ErrClosed
	}
}

func (e *einsetzbar) Close() error {
	if e.einmal.CompareAndSwap(false, true) {
		close(e.zu)
	}

	return nil
}

func (e *einsetzbar) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1} }

// gezaehlt merkt sich, ob sie geschlossen wurde.
type gezaehlt struct {
	net.Conn
	geschlossen atomic.Bool
}

func (g *gezaehlt) Close() error {
	g.geschlossen.Store(true)

	return nil
}

// **Eine Verbindung, die nach dem Schließen ankommt, wird geschlossen** —
// niemand nimmt sie mehr an, und offen bliebe sie ein Loch im Dienst.
func TestEineVerspaeteteVerbindungWirdGeschlossen(t *testing.T) {
	t.Parallel()

	innen := neuerEinsetzbarer()
	vereint := rest.Vereint(innen)

	// **Der innere bleibt offen**, damit sein Annehmen eine Verbindung bringt
	// statt eines Fehlers: geschlossen wird nur der vereinte.
	innen.einmal.Store(true)

	if fehler := vereint.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	spaet := &gezaehlt{}
	innen.herein <- spaet

	frist := time.Now().Add(5 * time.Second)
	for !spaet.geschlossen.Load() && time.Now().Before(frist) {
		time.Sleep(10 * time.Millisecond)
	}

	if !spaet.geschlossen.Load() {
		t.Error("die verspätete Verbindung blieb offen")
	}
}

// **Schließt ein innerer Lauscher von selbst, erfährt es der Wartende** — und
// das Annehmen dahinter endet, statt sich zu drehen.
func TestEinInnererLauscherSchliesstVonSelbst(t *testing.T) {
	t.Parallel()

	innen := neuerEinsetzbarer()
	vereint := rest.Vereint(innen)

	t.Cleanup(func() { _ = vereint.Close() }) //nolint:errcheck // aufräumen

	ergebnis := make(chan error, 1)

	go func() {
		_, fehler := vereint.Accept()
		ergebnis <- fehler
	}()

	// **Ob der Wartende schon da ist oder erst kommt**: das Annehmen innen
	// reicht den Fehler erst weiter, wenn ihn jemand abnimmt.
	if fehler := innen.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	select {
	case fehler := <-ergebnis:
		if !errors.Is(fehler, net.ErrClosed) {
			t.Errorf("der Wartende bekam %v — erwartet net.ErrClosed", fehler)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("der Wartende erfuhr nichts")
	}
}

// **Eine ungültige Bindung startet nichts** — geprüft vor dem ersten Lauscher.
func TestEineUngueltigeBindungStartetNichts(t *testing.T) {
	t.Parallel()

	if _, _, fehler := rest.LauschenAufAllen(t.Context(),
		[]domain.Bindung{{Adresse: "127.0.0.1", Port: -1}}, nil); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("abgewiesen mit %v — erwartet ErrValidation", fehler)
	}
}

// **Ein gewöhnlicher Lauscher nennt seine eine Adresse.**
func TestEinGewoehnlicherLauscherNenntSeineAdresse(t *testing.T) {
	t.Parallel()

	innen := neuerEinsetzbarer()

	if adressen := rest.Adressen(innen); len(adressen) != 1 || adressen[0] != "127.0.0.1:1" {
		t.Errorf("die Adressen sind %v", adressen)
	}
}
