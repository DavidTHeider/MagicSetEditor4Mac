package fernleitung

// **Die Naht zu tailcat** — Werkstück N3 von Phase 5.4
// (#1456).
//
// # Warum sie eine eigene Datei ist
//
// **Weil sie ohne ein DERP-Relais im Internet nicht auszuführen ist.** Was
// hier steht, ist die dünnste mögliche Schicht: sie baut die Verbünde von
// tailcat und reicht seine Methoden durch. Jede fachliche Entscheidung — die
// Fassung der Einladung, die Schlange, der Lebenszyklus — steht in
// `fernleitung.go` und wird gemessen.
//
// Der Eintrag in `.covignore` nennt genau das. Dieselbe Gestalt wie
// `internal/app/fensternaht.go`: jede Naht ist eine Veränderliche, **damit**
// alles darüber ohne sie läuft.

import (
	"context"
	"net"

	"github.com/tailscale/tailcat"
)

// Die Nähte zu tailcat — **als Veränderliche, damit ein Lauf ohne Internet
// auskommt** (§3). Ein Lauf, der wirklich wählte, bräuchte ein DERP-Relais
// und zwei Rechner; was hier geprüft wird, ist der Weg davor und danach.
var (
	serverBauen = func(annehmen func(net.Conn)) fernserver {
		return &tailcatserver{unten: &tailcat.Server{
			Logf: tailcatprotokoll,
			OnTCP: func(port uint16) func(net.Conn) {
				if port != Port {
					return nil
				}

				return annehmen
			},
		}}
	}

	klientBauen = func(adresse string) fernklient {
		return &tailcatklient{unten: &tailcat.Client{
			Server: tailcat.Addr(adresse),
			Logf:   tailcatprotokoll,
		}}
	}
)

type tailcatserver struct{ unten *tailcat.Server }

func (t *tailcatserver) Start() error { return ohneZugangsdatenFehler(t.unten.Start()) }

func (t *tailcatserver) Einladungsadresse() string {
	return string(t.unten.TailcatAddr())
}

func (t *tailcatserver) Close() error { return t.unten.Close() }

// Wege liest den Stand der Gegenstellen — `CurAddr` für den direkten Weg,
// `Relay` für die DERP-Region.
func (t *tailcatserver) Wege() []Fernweg {
	stand := t.unten.Status()
	if stand == nil {
		return nil
	}

	wege := make([]Fernweg, 0, len(stand.Peer))
	for _, gegenstelle := range stand.Peer {
		wege = append(wege, Fernweg{Direkt: gegenstelle.CurAddr, Relais: gegenstelle.Relay})
	}

	return wege
}

type tailcatklient struct{ unten *tailcat.Client }

func (t *tailcatklient) DialTCPPort(
	ctx context.Context, port uint16,
) (net.Conn, error) {
	verbindung, fehler := t.unten.DialTCPPort(ctx, port)

	return verbindung, ohneZugangsdatenFehler(fehler)
}

func (t *tailcatklient) Close() error { return t.unten.Close() }
