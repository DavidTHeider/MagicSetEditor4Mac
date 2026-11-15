package entdeckung

// **Eine Ankündigung, die nicht hinausgeht, ist keine** — K02
// (#2188).
//
// Gemessen am 2026-10-07 in einem macOS-15.2-Gast mit der gebauten
// Anwendung: verwehrt macOS die Erlaubnis *Lokales Netzwerk*, bindet der
// mDNS-Server seine Sockets trotzdem — die Absage kommt erst beim **Senden**,
// und das tut der Server nur als Antwort, deren Fehler die Bibliothek
// verschluckt. Die Anwendung zeigte „kündigt sich an", und niemand sah sie.

import (
	"bytes"
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sendefehler ist, was macOS bei verwehrter Erlaubnis zurückgibt.
func sendefehler(grund syscall.Errno) error {
	return &net.OpError{Op: "write", Net: "udp4", Err: &os.SyscallError{Syscall: "sendto", Err: grund}}
}

func TestEineAnkuendigungDieNichtHinausgehtSagtWasZuTunIst(t *testing.T) {
	for name, fall := range map[string]struct {
		grund syscall.Errno
		soll  error
	}{
		"verwehrte Erlaubnis": {syscall.EHOSTUNREACH, domain.ErrOhneNetzerlaubnis},
		"verwehrt mit EPERM":  {syscall.EPERM, domain.ErrOhneNetzerlaubnis},
		"kein Weg ins Netz":   {syscall.ENETUNREACH, domain.ErrOhneNetzroute},
	} {
		t.Run(name, func(t *testing.T) {
			vorher := probeSenden
			probeSenden = func(*net.Interface) error { return sendefehler(fall.grund) }

			t.Cleanup(func() { probeSenden = vorher })

			anmeldung, fehler := Melde(Angebot{Kennung: eineKennung(t), Port: 41234})
			if !errors.Is(fehler, fall.soll) {
				t.Fatalf("es kam %v, erwartet %v", fehler, fall.soll)
			}

			if anmeldung != nil {
				t.Errorf("neben dem Fehler stand eine Anmeldung: %+v", anmeldung)
			}
		})
	}
}

func TestEineSchnittstelleDieHinausgehtGenuegt(t *testing.T) {
	t.Parallel()

	erste, zweite := &net.Interface{Index: 1, Name: "eins"}, &net.Interface{Index: 2, Name: "zwei"}

	gefragt := 0
	senden := func(schnitt *net.Interface) error {
		gefragt++
		if schnitt == erste {
			return sendefehler(syscall.ENETUNREACH)
		}

		return nil
	}

	if fehler := hinaus([]*net.Interface{erste, zweite}, senden); fehler != nil {
		t.Fatalf("die zweite ging hinaus, und es kam %v", fehler)
	}

	if gefragt != 2 {
		t.Errorf("%d Proben, erwartet zwei", gefragt)
	}

	if fehler := hinaus([]*net.Interface{erste, erste}, senden); !errors.Is(fehler, domain.ErrOhneNetzroute) {
		t.Errorf("keine ging hinaus, und es kam %v", fehler)
	}
}

func TestDieProbeGehtUeberDieSchleifeHinaus(t *testing.T) {
	t.Parallel()

	schleife, fehler := net.InterfaceByName("lo0")
	if fehler != nil {
		t.Skipf("keine Schleife lo0: %v", fehler)
	}

	if fehler := mdnsProbe(schleife); fehler != nil {
		t.Fatalf("die Probe ging nicht hinaus: %v", fehler)
	}
}

func TestDieProbeIstEineMDNSFrageNachDerDienstart(t *testing.T) {
	t.Parallel()

	paket := probefrage()

	if kopf := []byte{0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0}; !bytes.Equal(paket[:12], kopf) {
		t.Fatalf("der Kopf ist %v, erwartet eine Frage ohne Antworten %v", paket[:12], kopf)
	}

	// **Der Name aus Dienstart und Bereich**, nicht abgeschrieben.
	var name []byte
	for _, etikett := range strings.Split(Dienstart+"."+Bereich, ".") {
		name = append(name, byte(len(etikett))) //nolint:gosec // drei kurze Etiketten aus Konstanten
		name = append(name, etikett...)
	}

	name = append(name, 0)

	if !bytes.Equal(paket[12:12+len(name)], name) {
		t.Fatalf("der Name ist %q, erwartet %q", paket[12:12+len(name)], name)
	}

	if rest := paket[12+len(name):]; !bytes.Equal(rest, []byte{0, 0x0c, 0, 0x01}) {
		t.Errorf("Art und Klasse sind %v, erwartet PTR und IN", rest)
	}
}

func TestOhneSchnittstelleGibtEsNichtsZuPruefen(t *testing.T) {
	t.Parallel()

	if fehler := hinaus(nil, func(*net.Interface) error { return errors.New("nie gerufen") }); fehler != nil {
		t.Errorf("es kam %v", fehler)
	}
}

func TestEineUnbekannteSchnittstelleLaesstDieProbeScheitern(t *testing.T) {
	t.Parallel()

	if fehler := mdnsProbe(&net.Interface{Index: 999999, Name: "keine"}); fehler == nil {
		t.Error("über eine Schnittstelle, die es nicht gibt, ging die Probe hinaus")
	}
}

func TestOhneSteckdoseGehtKeineProbeHinaus(t *testing.T) {
	vorher := probeSteckdose
	probeSteckdose = func(context.Context) (net.PacketConn, error) {
		return nil, errors.New("keine Steckdose")
	}

	t.Cleanup(func() { probeSteckdose = vorher })

	if fehler := mdnsProbe(nil); fehler == nil {
		t.Error("ohne Steckdose ging die Probe hinaus")
	}
}
