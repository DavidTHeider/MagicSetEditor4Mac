package flaechen_test

// **Eine Netzvorrichtung zwischen zwei Instanzen** — QA-10.AK-02
// (#1741).
//
// # Was sie ist
//
// **Ein TCP-Vermittler auf der Schleifenadresse**: er nimmt an, wählt das
// Ziel und reicht die Bytes in beide Richtungen durch — ohne sie zu lesen.
// TLS geht Ende zu Ende durch ihn hindurch; er ist damit genau das, was ein
// Relais im Netz ist: ein Weg, dem man nicht trauen muss.
//
// Zwischen Annahme und Durchreichen sitzen die Schalter: **gesperrt** (der
// Weg ist fort), **verzögert** (jede Richtung wartet vor jedem Stück), und
// **einmal gekappt** (die nächste Verbindung endet, bevor ein Byte durchgeht).
//
// # Was sie nicht ist
//
// **Ein fremdes Netz.** Ein Relais über DERP, ein Router mit NAT oder eine
// Firewall verhalten sich anders als ein Prozess auf derselben Maschine; das
// bleibt die Abnahme von Hand aus QA-10.AK-04.

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// vermittler ist die Vorrichtung.
type vermittler struct {
	horcher net.Listener
	ziel    string

	// ctx endet mit dem Test — und mit ihm jedes Wählen, das noch läuft.
	ctx context.Context

	schloss      sync.Mutex
	gesperrt     bool
	verzoegerung time.Duration
	kappen       int
	offen        []net.Conn

	// durchgereicht zählt die Verbindungen, die bis zum Ziel kamen — der
	// Beleg, dass ein Weg wirklich über den Vermittler lief.
	durchgereicht atomic.Int64
}

// neuerVermittler horcht auf einem freien Anschluss und reicht zum Ziel.
func neuerVermittler(t *testing.T, ziel string) *vermittler {
	t.Helper()

	horcher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("der Vermittler horcht nicht: %v", fehler)
	}

	v := &vermittler{horcher: horcher, ziel: ziel, ctx: t.Context()}

	go v.annehmen()

	t.Cleanup(v.schliessen)

	return v
}

// adresse ist, was eine Instanz als Adresse des Gegenübers einträgt.
func (v *vermittler) adresse() string { return v.horcher.Addr().String() }

// sperren schaltet den Weg ab oder wieder an.
func (v *vermittler) sperren(gesperrt bool) {
	v.schloss.Lock()
	defer v.schloss.Unlock()

	v.gesperrt = gesperrt
}

// verzoegern setzt die Wartezeit je Stück und Richtung.
func (v *vermittler) verzoegern(dauer time.Duration) {
	v.schloss.Lock()
	defer v.schloss.Unlock()

	v.verzoegerung = dauer
}

// kappeDieNaechsten lässt so viele Verbindungen enden, bevor ein Byte
// durchgeht — der Verlust auf dem Weg.
func (v *vermittler) kappeDieNaechsten(anzahl int) {
	v.schloss.Lock()
	defer v.schloss.Unlock()

	v.kappen = anzahl
}

// umlenken setzt ein neues Ziel — für einen Horcher, der neu gestartet ist.
func (v *vermittler) umlenken(ziel string) {
	v.schloss.Lock()
	defer v.schloss.Unlock()

	v.ziel = ziel
}

func (v *vermittler) annehmen() {
	for {
		verbindung, fehler := v.horcher.Accept()
		if fehler != nil {
			return
		}

		go v.vermitteln(verbindung)
	}
}

// vermitteln entscheidet über eine Verbindung und reicht sie durch.
func (v *vermittler) vermitteln(herein net.Conn) {
	v.schloss.Lock()
	gesperrt, verzoegerung, ziel := v.gesperrt, v.verzoegerung, v.ziel
	gekappt := v.kappen > 0

	if gekappt {
		v.kappen--
	}

	v.offen = append(v.offen, herein)
	v.schloss.Unlock()

	if gesperrt || gekappt {
		_ = herein.Close() //nolint:errcheck // der Weg ist absichtlich fort

		return
	}

	hinaus, fehler := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(v.ctx, "tcp", ziel)
	if fehler != nil {
		_ = herein.Close() //nolint:errcheck // das Ziel ist fort; dasselbe für den Anrufer

		return
	}

	v.durchgereicht.Add(1)

	v.schloss.Lock()
	v.offen = append(v.offen, hinaus)
	v.schloss.Unlock()

	go reiche(hinaus, herein, verzoegerung)
	reiche(herein, hinaus, verzoegerung)
}

// reiche kopiert eine Richtung, Stück für Stück, mit Wartezeit davor.
func reiche(nach io.WriteCloser, von io.Reader, verzoegerung time.Duration) {
	defer func() { _ = nach.Close() }() //nolint:errcheck // das Ende einer Richtung

	puffer := make([]byte, 32*1024)

	for {
		gelesen, fehler := von.Read(puffer)
		if gelesen > 0 {
			time.Sleep(verzoegerung)

			if _, schreibfehler := nach.Write(puffer[:gelesen]); schreibfehler != nil {
				return
			}
		}

		if fehler != nil {
			return
		}
	}
}

// schliessen räumt Horcher und jede offene Verbindung ab.
func (v *vermittler) schliessen() {
	_ = v.horcher.Close() //nolint:errcheck // beim Abräumen

	v.schloss.Lock()
	defer v.schloss.Unlock()

	for _, verbindung := range v.offen {
		_ = verbindung.Close() //nolint:errcheck // dito
	}
}
