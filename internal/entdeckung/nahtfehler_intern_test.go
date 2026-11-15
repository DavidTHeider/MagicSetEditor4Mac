package entdeckung

// **Die vier Wege, die dieses Betriebssystem nicht geht** — Werkstück N2 von
// Phase 5.4
// (#1455).
//
// # Warum es diese Läufe gibt
//
// **Weil die Rufe einen Fehler zurückgeben und ihn hier niemand auslöst.**
// `Addrs()` gibt für eine erfundene Schnittstelle eine **leere Liste** und
// keinen Fehler, `NewMDNSService` prüft nur, was vorher geprüft ist,
// `NewServer` bindet auf diesem Rechner, und `Shutdown` ist ausdrücklich
// mehrfach aufrufbar.
//
// **Ein `_ =` daneben wäre die stille Absenkung, die §3 verbietet.** Also
// werden sie betreten — über die Naht, die dafür da ist.

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/mdns"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// eineKennung gibt eine echte Peerkennung.
func eineKennung(t *testing.T) peer.Kennung {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return ich.Kennung()
}

// **Eine Adressliste, die klagt, macht keine Ankündigung.**
func TestEineKlagendeAdresslisteMachtKeineAnkuendigung(t *testing.T) {
	vorher := adresslisteVon
	adresslisteVon = func(*net.Interface) ([]net.Addr, error) {
		return nil, errors.New("das Betriebssystem schweigt")
	}

	t.Cleanup(func() { adresslisteVon = vorher })

	if _, fehler := Melde(Angebot{
		Kennung: eineKennung(t), Port: 41234,
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Ein Dienst, der sich nicht zusammenstellen lässt, ebenso.**
func TestEinDienstDerSichNichtBauenLaesstMachtKeineAnkuendigung(t *testing.T) {
	vorher := dienstBauen
	dienstBauen = func(
		string, string, string, string, int, []net.IP, []string,
	) (*mdns.MDNSService, error) {
		return nil, errors.New("die Zone geht nicht")
	}

	t.Cleanup(func() { dienstBauen = vorher })

	if _, fehler := Melde(Angebot{
		Kennung: eineKennung(t), Port: 41234,
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und ein Server, der nicht aufmacht, auch.**
//
// **Das ist der Weg, den eine abgelehnte Netzwerkberechtigung nimmt** — macOS
// gibt die Steckdose dann nicht her, und das Issue verlangt dafür einen
// „begrenzten verständlichen Fehler" statt eines Spinners.
func TestEinServerDerNichtAufmachtMachtKeineAnkuendigung(t *testing.T) {
	vorher := serverBauen
	serverBauen = func(*mdns.Config) (*mdns.Server, error) {
		return nil, errors.New("die Steckdose bleibt zu")
	}

	t.Cleanup(func() { serverBauen = vorher })

	if _, fehler := Melde(Angebot{
		Kennung: eineKennung(t), Port: 41234,
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Eine Ankündigung wird gedeutet wie eine Suche** — S13.NAT-03 (#1706):
// „Ankündigung und Suche besitzen gleich verständliche Fehleraktionen".
// Hier kam der rohe Text der Steckdose ins Fenster.
func TestEineVerwehrteAnkuendigungSagtWasZuTunIst(t *testing.T) {
	vorher := serverBauen
	serverBauen = func(*mdns.Config) (*mdns.Server, error) {
		return nil, &net.OpError{
			Op: "listen", Net: "udp4",
			Err: &os.SyscallError{Syscall: "setsockopt", Err: syscall.EHOSTUNREACH},
		}
	}

	t.Cleanup(func() { serverBauen = vorher })

	if _, fehler := Melde(Angebot{
		Kennung: eineKennung(t), Port: 41234,
	}); !errors.Is(fehler, domain.ErrOhneNetzerlaubnis) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Eine Ankündigung, die beim Zurücknehmen klagt, sagt es.**
func TestEineAnkuendigungDieBeimZuruecknehmenKlagtSagtEs(t *testing.T) {
	t.Parallel()

	anmeldung := &Anmeldung{dienste: []niederlegbar{klagenderDienst{}}}

	if fehler := anmeldung.Beende(); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// klagenderDienst legt sich nicht nieder.
type klagenderDienst struct{}

func (klagenderDienst) Shutdown() error {
	return errors.New("er bleibt stehen")
}

// **Eine Abfrage, die klagt, gibt kein halbes Ergebnis.**
//
// **Und das ist die Zusage**: wer eine Liste bekommt, hat gefragt und eine
// Antwort erhalten. Ein Ergebnis neben einem Fehler ließe jeden Aufrufer
// entscheiden, welchem von beiden er glaubt.
func TestEineKlagendeAbfrageGibtKeinHalbesErgebnis(t *testing.T) {
	vorher := fragen
	fragen = func(context.Context, *mdns.QueryParam) error {
		return errors.New("die Gruppe nimmt niemanden auf")
	}

	t.Cleanup(func() { fragen = vorher })

	ergebnis, fehler := Suche(t.Context(), Frage{Frist: time.Millisecond})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if len(ergebnis.Funde) != 0 || ergebnis.Verworfen != 0 {
		t.Fatalf("es kam trotzdem %+v", ergebnis)
	}
}

// **Und eine Abfrage, die antwortet, gibt die Funde.**
//
// # Warum dieser Lauf dazugehört
//
// **Weil die Zusage darüber nur die Hälfte ist** (§10). „Kein halbes
// Ergebnis" sagt, was bei einem Fehler **nicht** herauskommt; dass bei einer
// geglückten Abfrage die Antworten wirklich durchgereicht werden, sagte kein
// Lauf über diese Naht.
//
// **Gemerkt hat es die Abdeckung**, und zwar erst, als die Läufe über das
// echte Netz ein Tor bekamen: bis zum 2026-09-20 betraten sie `Suche`s
// Erfolgsweg nebenbei mit. Ein Lauf, dessen Aussage von der Erlaubnis dieses
// Rechners abhängt, ist kein Beleg für eine Naht — er ist es an einem Tag und
// am nächsten nicht.
func TestEineAntwortendeAbfrageGibtIhreFunde(t *testing.T) {
	kennung := eineKennung(t)

	vorher := fragen
	fragen = func(_ context.Context, param *mdns.QueryParam) error {
		param.Entries <- &mdns.ServiceEntry{
			Name:       "probe._mse._tcp.local.",
			Host:       "probe.local.",
			AddrV4:     net.IPv4(127, 0, 0, 1),
			Port:       41234,
			InfoFields: []string{"kennung=" + string(kennung), "name=Davids Rechner"},
		}

		return nil
	}

	t.Cleanup(func() { fragen = vorher })

	ergebnis, fehler := Suche(t.Context(), Frage{Frist: time.Millisecond})
	if fehler != nil {
		t.Fatalf("Suche: %v", fehler)
	}

	if len(ergebnis.Funde) != 1 {
		t.Fatalf("es kam %+v", ergebnis)
	}

	if ergebnis.Funde[0].Kennung != kennung {
		t.Errorf("die Kennung ist %q", ergebnis.Funde[0].Kennung)
	}

	if ergebnis.Funde[0].Name != "Davids Rechner" {
		t.Errorf("der Name ist %q", ergebnis.Funde[0].Name)
	}
}
