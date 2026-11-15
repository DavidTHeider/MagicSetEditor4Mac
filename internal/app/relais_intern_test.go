package app

// **Ein Relais, das sich anhalten und kappen lässt** — für D01
// (#2232).
//
// Es reicht Bytes zwischen Gast und Wirt durch, **unterhalb** von TLS und
// Rahmung: was es tut, sieht für beide Enden aus wie ein Netz, das stockt
// oder abreißt.
//
//   - **anhalten**: nichts geht mehr durch, aber jede Verbindung bleibt offen
//     — ein kurzer Ausfall, den TCP überlebt;
//   - **kappen**: jede offene Verbindung wird geschlossen, und neue werden
//     abgewiesen, bis es wieder **freigegeben** ist — ein langer Ausfall.

import (
	"errors"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type relais struct {
	t       *testing.T
	ziel    string
	horcher net.Listener

	schloss  sync.Mutex
	offen    map[net.Conn]struct{}
	gekappt  bool
	angehalt chan struct{} // geschlossen, solange es durchlässt
}

// neuesRelais horcht auf der Rückschleife und reicht an `ziel` durch.
func neuesRelais(t *testing.T, ziel string) *relais {
	t.Helper()

	horcher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("das Relais horcht nicht: %v", fehler)
	}

	durch := make(chan struct{})
	close(durch)

	r := &relais{t: t, ziel: ziel, horcher: horcher, offen: map[net.Conn]struct{}{}, angehalt: durch}

	go r.annehmen()

	t.Cleanup(func() {
		_ = horcher.Close() //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten
		r.alleSchliessen()
	})

	return r
}

func (r *relais) adresse() string { return r.horcher.Addr().String() }

func (r *relais) annehmen() {
	for {
		herein, fehler := r.horcher.Accept()
		if fehler != nil {
			return
		}

		r.schloss.Lock()
		gekappt := r.gekappt
		r.schloss.Unlock()

		if gekappt {
			_ = herein.Close() //nolint:errcheck // abgewiesen, wie ein Netz ohne Weg

			continue
		}

		hinaus, fehler := (&net.Dialer{}).DialContext(r.t.Context(), "tcp", r.ziel)
		if fehler != nil {
			_ = herein.Close() //nolint:errcheck // der Wirt ist nicht da

			continue
		}

		r.merken(herein, hinaus)

		go r.reichen(herein, hinaus)
		go r.reichen(hinaus, herein)
	}
}

// reichen kopiert in Stücken und wartet vor jedem, solange angehalten ist.
func (r *relais) reichen(von, nach net.Conn) {
	puffer := make([]byte, 32<<10)

	for {
		gelesen, fehler := von.Read(puffer)
		if gelesen > 0 {
			r.schloss.Lock()
			warte := r.angehalt
			r.schloss.Unlock()
			<-warte

			if _, schreibfehler := nach.Write(puffer[:gelesen]); schreibfehler != nil {
				return
			}
		}

		if fehler != nil {
			if !errors.Is(fehler, io.EOF) {
				_ = nach.Close() //nolint:errcheck // die andere Richtung endet mit
			}

			return
		}
	}
}

func (r *relais) merken(verbindungen ...net.Conn) {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	for _, eine := range verbindungen {
		r.offen[eine] = struct{}{}
	}
}

// anhalten lässt nichts mehr durch; [relais.freigeben] hebt es auf.
func (r *relais) anhalten() {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	r.angehalt = make(chan struct{})
}

// kappen schließt jede offene Verbindung und weist neue ab.
func (r *relais) kappen() {
	r.schloss.Lock()
	r.gekappt = true
	r.schloss.Unlock()

	r.alleSchliessen()
}

// freigeben lässt wieder durch und nimmt wieder an.
func (r *relais) freigeben() {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	r.gekappt = false

	select {
	case <-r.angehalt:
	default:
		close(r.angehalt)
	}
}

func (r *relais) alleSchliessen() {
	r.schloss.Lock()
	offen := r.offen
	r.offen = map[net.Conn]struct{}{}
	r.schloss.Unlock()

	for eine := range offen {
		_ = eine.Close() //nolint:errcheck // gekappt ist gekappt
	}
}

// ueberDenWirt stellt über die **Anwendung des Wirts**, wie sein Fenster es
// tut — nur ihr Zug geht über ihren Tischfunk an die Gäste.
//
// **Bis D04 stellten die Läufe über eine zweite Fläche** auf demselben
// Bestand (`ueberDieFlaeche`), am Funk vorbei: der Gast sah einen Zug des
// Wirts erst, wenn sein eigener Wunsch oder seine Rückkehr das Bild nachzog.
// Gefunden, als ein verdeckter Zug des Wirts über den Fernweg nie ankam —
// und über TCP genauso wenig.
func ueberDenWirt(wirt *seite) befehlsweg {
	return func(befehl domain.Spielbefehl) BefehlAntwort {
		return wirt.anwendung.PartieBefehl(wirt.ctx, befehl)
	}
}
