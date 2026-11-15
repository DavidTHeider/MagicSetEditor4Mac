package entdeckung

// **Ein gescheiterter Ruf sagt, was zu tun ist** — der Befund aus der
// Sichtprüfung vom 2026-09-14.
//
// # Was gesehen wurde
//
// „Wer in der Nähe ist" antwortete mit „Es ist ein unerwarteter Fehler
// aufgetreten", und der Grund stand nur im Protokoll:
//
//	write udp4 0.0.0.0:52316->224.0.0.251:5353: sendto: no route to host
//
// `224.0.0.251` ist die Multicastadresse von mDNS. macOS gibt für sie
// `EHOSTUNREACH`, solange die Anwendung die Erlaubnis **Lokales Netzwerk**
// nicht hat — und das ist mit zwei Klicks erledigt, wenn man es weiß.

import (
	"errors"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineVerwehrteSucheSagtWasZuTunIst(t *testing.T) {
	t.Parallel()

	for name, ausgang := range map[string]error{
		// **In der Hülle, in der er wirklich ankommt.** Ein blankes
		// `syscall.EHOSTUNREACH` prüfte eine Gestalt, die das Netzpaket nie
		// liefert — `errors.Is` muss durch beide Schichten greifen.
		"keine Route zum Wirt": &net.OpError{
			Op: "write", Net: "udp4",
			Err: &os.SyscallError{Syscall: "sendto", Err: syscall.EHOSTUNREACH},
		},
		"nicht erlaubt": &net.OpError{
			Op: "write", Net: "udp4",
			Err: &os.SyscallError{Syscall: "sendto", Err: syscall.EPERM},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gedeutet := deuteSuchfehler(ausgang)

			if !errors.Is(gedeutet, domain.ErrOhneNetzerlaubnis) {
				t.Errorf("erwartet war ErrOhneNetzerlaubnis, kam: %v", gedeutet)
			}

			// **Und der Grund bleibt erhalten** — er gehört ins Protokoll,
			// auch wenn der Satz ihn nicht zeigt (§8).
			if !errors.Is(gedeutet, ausgang) {
				t.Errorf("der Grund ist verloren: %v", gedeutet)
			}
		})
	}
}

// **Keine Route ins Netz ist keine fehlende Erlaubnis** — NET-02.NAT-03 und
// S14.NAT-02 (#1706).
// `ENETUNREACH` heißt, dass es gar keinen Weg gibt — etwa ohne verbundene
// Schnittstelle. Hier stand es bei den dreien, und der Satz dazu schickte in
// die Systemeinstellungen, wo nichts zu tun war.
func TestKeineRouteInsNetzIstEinEigenerBefund(t *testing.T) {
	t.Parallel()

	ausgang := &net.OpError{
		Op: "write", Net: "udp4",
		Err: &os.SyscallError{Syscall: "sendto", Err: syscall.ENETUNREACH},
	}

	gedeutet := deuteSuchfehler(ausgang)

	if !errors.Is(gedeutet, domain.ErrOhneNetzroute) || errors.Is(gedeutet, domain.ErrOhneNetzerlaubnis) {
		t.Errorf("gedeutet als %v", gedeutet)
	}

	if !errors.Is(gedeutet, ausgang) {
		t.Errorf("der Grund ist verloren: %v", gedeutet)
	}

	// **Und er führt trotzdem auf IPv4 zurück** — keine Route für IPv6 ist
	// genau der Fall des Rückfalls.
	if !ohneRoute(gedeutet) {
		t.Error("ohneRoute erkennt ihn nicht")
	}
}

// **Und alles andere bleibt, was es war.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne sie sagte der Lauf nur, dass
// irgendetwas zu `ErrOhneNetzerlaubnis` wird.
func TestEinAndererFehlerBleibtEineValidierung(t *testing.T) {
	t.Parallel()

	gedeutet := deuteSuchfehler(errors.New("etwas anderes"))

	if errors.Is(gedeutet, domain.ErrOhneNetzerlaubnis) {
		t.Errorf("das ist keine fehlende Erlaubnis: %v", gedeutet)
	}

	if !errors.Is(gedeutet, domain.ErrValidation) {
		t.Errorf("erwartet war ErrValidation, kam: %v", gedeutet)
	}
}
