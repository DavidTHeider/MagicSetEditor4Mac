package entdeckung

// **IPv4 trägt, wenn IPv6 keine Route hat** — NET-02.NAT-01
// (#1706).
//
// # Der Befund
//
// **Zwei Anwendungen im selben WLAN fanden einander nicht**, und der
// Ausweg war, die IPv4-Schnittstelle von Hand zu nennen
// (`MSE_ENTDECKUNGSSCHNITTSTELLE`). Warum das half, steht in der
// Bibliothek — `hashicorp/mdns@v1.0.7`, `client.go`, `sendQuery`:
//
//	if c.ipv4UnicastConn != nil {
//	    _, err = c.ipv4UnicastConn.WriteToUDP(buf, ipv4Addr)
//	    if err != nil { return err }
//	}
//	if c.ipv6UnicastConn != nil {
//	    _, err = c.ipv6UnicastConn.WriteToUDP(buf, ipv6Addr)
//	    if err != nil { return err }
//	}
//
// Geschickt wird **erst über IPv4, dann über IPv6** — und jeder
// Schreibfehler bricht die ganze Abfrage ab. Gibt macOS für `ff02::fb` ein
// `EHOSTUNREACH`, weil für IPv6-Multicast keine Route steht, scheitert die
// Suche, **obwohl der IPv4-Ruf schon hinausgegangen war**.
//
// **Beim Ankündigen gibt es das nicht**: `server.go` verwirft den Fehler des
// IPv6-Listeners (`ipv6List, _ := net.ListenMulticastUDP(…)`) und scheitert
// nur, wenn **beide** fehlen. Die Ankündigung war tolerant, die Suche nicht.
//
// # Warum ein zweiter Ruf und kein `DisableIPv6` von vornherein
//
// **Weil IPv6 nicht das Problem ist, sondern seine fehlende Route.** Wer ihn
// immer abschaltete, fände ein Gegenüber nicht mehr, das **nur** über IPv6
// erreichbar ist — und das wäre eine Zusage weniger für einen Fehler, den es
// auf vielen Netzen gar nicht gibt.
//
// # Was das nicht belegt
//
// **Dass zwei Rechner einander im WLAN finden.** Hier steht ein eingespeister
// Schreibfehler, kein Netz. `docs/HAERTUNG.md` sagt es für den ganzen
// mDNS-Weg: der Lauf geht über `lo0`, und „Multicast über die Rückschleife
// ist nicht Multicast über ein WLAN".

import (
	"context"
	"errors"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/hashicorp/mdns"
)

// unerreichbar ist der Fehler, den macOS für ein Multicastziel ohne Route gibt —
// in der Hülle, in der er wirklich ankommt.
func unerreichbar(netzart string) error {
	return &net.OpError{
		Op: "write", Net: netzart,
		Err: &os.SyscallError{Syscall: "sendto", Err: syscall.EHOSTUNREACH},
	}
}

// mitRufen hängt eine Antwortfolge an die Naht und sammelt die Rufe.
//
// # Warum mit Schloss
//
// **Weil die Suche nebenläufig fragt** — seit NET-02.NAT-01 ein Ruf je
// Schnittstelle (#1706). Ohne Schloss schreiben zwei Goroutinen in dieselbe
// Liste, und **der Race-Detektor hat es gefunden**: `make ci` war rot, wo
// `make vorpruefung` grün war, denn der Riegel fährt ohne `-race`
// (gemessen: 48-fache Laufzeit, [0094](../../docs/entscheidungen/0094-der-detektor-misst-nicht-wazero.md)).
//
// **Es ist die Vorrichtung, die rannte, nicht der Code**: `ueberAlle`
// schreibt je Goroutine in **eine eigene Stelle** der Ergebnisliste. Aber
// eine Vorrichtung, die rennt, macht einen Lauf unbrauchbar, und das ist
// derselbe Befund (§3).
// ohneSchnittstellen lässt die Wahl beim Betriebssystem — **unabhängig von
// den Schnittstellen dieses Rechners** (C06,
// #2231).
// Die Fälle hier zählen die Rufe **eines** Wegs; lief daneben eine VM, kam
// ihre Brücke als zweite taugliche Schnittstelle dazu, und jede Zahl
// verdoppelte sich (gemessen am 2026-10-09 mit `bridge100`).
func ohneSchnittstellen(t *testing.T) {
	t.Helper()

	vorher := tauglicheLesen
	tauglicheLesen = func() ([]net.Interface, error) { return nil, nil }

	t.Cleanup(func() { tauglicheLesen = vorher })
}

func mitRufen(t *testing.T, antworte func(int, *mdns.QueryParam) error) *[]mdns.QueryParam {
	t.Helper()

	var (
		schloss sync.Mutex
		rufe    = &[]mdns.QueryParam{}
	)

	vorher := fragen

	fragen = func(_ context.Context, param *mdns.QueryParam) error {
		schloss.Lock()
		*rufe = append(*rufe, *param)
		zahl := len(*rufe)
		schloss.Unlock()

		return antworte(zahl, param)
	}

	t.Cleanup(func() { fragen = vorher })

	return rufe
}

func TestEineSucheFaelltAufIPv4Zurueck(t *testing.T) {
	ohneSchnittstellen(t)

	rufe := mitRufen(t, func(_ int, param *mdns.QueryParam) error {
		if param.DisableIPv6 {
			return nil
		}

		return unerreichbar("udp6")
	})

	if _, fehler := Suche(context.Background(), Frage{}); fehler != nil {
		t.Fatalf("die Suche scheiterte, statt auf IPv4 zurückzufallen: %v",
			fehler)
	}

	if len(*rufe) != 2 {
		t.Fatalf("es gab %d Rufe, erwartet 2", len(*rufe))
	}

	if (*rufe)[0].DisableIPv6 {
		t.Error("der erste Ruf schaltete IPv6 ab — er soll es versuchen")
	}

	if !(*rufe)[1].DisableIPv6 {
		t.Error("der zweite Ruf schaltete IPv6 nicht ab")
	}

	// **Und IPv4 bleibt an.** Ein Rückfall, der beides abschaltete, ließe
	// `newClient` mit „Must enable at least one of IPv4 and IPv6" scheitern.
	if (*rufe)[1].DisableIPv4 {
		t.Error("der zweite Ruf schaltete auch IPv4 ab")
	}
}

// **Scheitert auch IPv4, gilt die Deutung von vorher.**
//
// **Der Unterschied ist die Erlaubnis**: `EHOSTUNREACH` über **beide** Wege
// heißt auf macOS, dass diese Anwendung nicht ins lokale Netz darf — und das
// ist etwas anderes als „IPv6 hat keine Route".
func TestScheitertAuchIPv4BleibtEsBeiDerErlaubnis(t *testing.T) {
	ohneSchnittstellen(t)

	rufe := mitRufen(t, func(_ int, _ *mdns.QueryParam) error {
		return unerreichbar("udp4")
	})

	_, fehler := Suche(context.Background(), Frage{})

	if !errors.Is(fehler, domain.ErrOhneNetzerlaubnis) {
		t.Errorf("der Fehler war %v, erwartet die fehlende Netzerlaubnis",
			fehler)
	}

	if len(*rufe) != 2 {
		t.Errorf("es gab %d Rufe, erwartet 2", len(*rufe))
	}
}

// **Der Steuerlauf** (§10): ein Ruf, der gelingt, wird nicht wiederholt.
//
// Ohne ihn sagte der Lauf darüber nicht, ob der zweite Ruf am Fehler hängt —
// er könnte auch immer stattfinden, und dann fragte jede Suche zweimal.
func TestEineGelungeneSucheFragtEinmal(t *testing.T) {
	ohneSchnittstellen(t)

	rufe := mitRufen(t, func(_ int, _ *mdns.QueryParam) error { return nil })

	if _, fehler := Suche(context.Background(), Frage{}); fehler != nil {
		t.Fatalf("die Suche scheiterte: %v", fehler)
	}

	if len(*rufe) != 1 {
		t.Errorf("es gab %d Rufe, erwartet 1", len(*rufe))
	}
}

// **Und ein fachlicher Fehler wird nicht wiederholt.**
//
// Ein Rückfall auf jeden Fehler wäre eine Wiederholung ohne Grund — und bei
// einem Fehler, der nichts mit der Route zu tun hat, dauerte jede Suche
// doppelt so lange.
func TestEinFremderFehlerWirdNichtWiederholt(t *testing.T) {
	ohneSchnittstellen(t)

	rufe := mitRufen(t, func(_ int, _ *mdns.QueryParam) error {
		return errors.New("etwas ganz anderes")
	})

	if _, fehler := Suche(context.Background(), Frage{}); fehler == nil {
		t.Fatal("die Suche gelang, erwartet war ein Fehler")
	}

	if len(*rufe) != 1 {
		t.Errorf("es gab %d Rufe, erwartet 1", len(*rufe))
	}
}
