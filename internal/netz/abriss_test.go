package netz_test

// **Eine Gegenstelle, die mittendrin auflegt** — #2487.
//
// Gefunden mit der Hüllenprobe `p10handel09` (P10, #2378): schloss die
// Gegenstelle eine angenommene Verbindung, bevor der Kopf kam, ordnete
// `deuteTransport` den Fehler nicht ein, und am Fenster stand „Es ist ein
// unerwarteter Fehler aufgetreten" — etwa unter „Preise holen". Ein
// abgewiesener Port und ein unauflösbarer Name sagten dagegen, was zu tun
// ist. **Hier fahren drei Abrisse über den echten Transport**: ein Schließen
// (`EOF`), ein Zurücksetzen (`ECONNRESET`) und eines mitten im Rumpf.
//
// **Keine Wiederholung durch diese Klasse** (§16, Satz 4): gezählt wird, wie
// oft die Gegenstelle gefragt wurde.

import (
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// auflegend gibt eine Gegenstelle, die jede Verbindung annimmt, die Anfrage
// liest und dann tut, was `auflegen` sagt.
func auflegend(t *testing.T, auflegen func(*net.TCPConn)) (string, *atomic.Int32) {
	t.Helper()

	horcher, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("horchen: %v", fehler)
	}

	t.Cleanup(func() { _ = horcher.Close() }) //nolint:errcheck // Aufräumen

	var gefragt atomic.Int32

	go func() {
		for {
			strom, fehler := horcher.Accept()
			if fehler != nil {
				return
			}

			gefragt.Add(1)

			// **Erst die Anfrage lesen** — sonst sähe der Klient ein
			// Schreiben ins Leere statt einer Antwort, die nicht kommt.
			_ = strom.SetReadDeadline(time.Now().Add(time.Second)) //nolint:errcheck // nur eine Frist
			puffer := make([]byte, 4096)
			_, _ = strom.Read(puffer) //nolint:errcheck // gelesen wird, um aufzulegen

			if tcp, ist := strom.(*net.TCPConn); ist {
				auflegen(tcp)
			}
		}
	}()

	return "http://" + horcher.Addr().String(), &gefragt
}

func TestEineGegenstelleDieAuflegtTraegtIhreKlasse(t *testing.T) {
	t.Setenv(netz.SchleifeUmgebung, "an")

	klient := netz.NeuerKlient(netz.Fristen{Verbindung: time.Second, Kopf: 3 * time.Second, Stillstand: 3 * time.Second})
	kenner := netz.Kenner{Name: "test", Annahme: "*/*"}

	const geheim = "GEHEIM-2487"

	for _, fall := range []struct {
		name     string
		auflegen func(*net.TCPConn)
	}{
		{"geschlossen, bevor der Kopf kam", func(strom *net.TCPConn) {
			_ = strom.Close() //nolint:errcheck // das ist der Fall
		}},
		{"zurückgesetzt, bevor der Kopf kam", func(strom *net.TCPConn) {
			_ = strom.SetLinger(0) //nolint:errcheck // ein RST statt eines FIN
			_ = strom.Close()      //nolint:errcheck // das ist der Fall
		}},
		{"geschlossen mitten im Rumpf", func(strom *net.TCPConn) {
			_, _ = strom.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\nanfang")) //nolint:errcheck // der Anfang
			_ = strom.Close()                                                                  //nolint:errcheck // das ist der Fall
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			adresse, gefragt := auflegend(t, fall.auflegen)

			_, fehler := netz.Hole(t.Context(), klient, nil, kenner, adresse+"/karten?schluessel="+geheim, 1<<20)

			if !errors.Is(fehler, domain.ErrVerbindungAbgerissen) || !errors.Is(fehler, domain.ErrGegenstelle) {
				t.Fatalf("der Fehler %v trägt nicht die Klasse %v", fehler, domain.ErrVerbindungAbgerissen)
			}

			var gegenstelle *domain.Gegenstellenfehler
			if !errors.As(fehler, &gegenstelle) || !strings.HasPrefix(adresse, "http://"+gegenstelle.Wirt) {
				t.Errorf("der Fehler %v nennt nicht den Wirt aus %s", fehler, adresse)
			}

			if strings.Contains(fehler.Error(), geheim) {
				t.Errorf("die Meldung trägt das Geheimnis aus der Adresse: %v", fehler)
			}

			if zahl := gefragt.Load(); zahl != 1 {
				t.Errorf("die Gegenstelle wurde %d-mal gefragt, erwartet einmal", zahl)
			}
		})
	}
}
