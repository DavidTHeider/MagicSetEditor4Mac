// Package fernleitung verbindet zwei Rechner über fremde Netze — über tailcat.
//
// **Ein Blatt** (§2), mit **einer** benannten Ausnahme: es kennt
// `internal/leitung`. Der Grund ist derselbe wie bei `svg` zu `render` — es
// ist der **zweite Verbraucher** desselben Weges: tailcat liefert einen
// `net.Conn`, und alles darüber — TLS 1.3, das Festnageln auf die erwartete
// Kennung, die Kanalbindung, die Rahmung — ist dasselbe wie im eigenen Netz
// und soll es bleiben.
//
// Werkstück N3 von Phase 5.4
// (#1456),
// entschieden in [0114](../../docs/entscheidungen/0114-tailcat-gepinnt.md).
//
// # Was tailcat leistet
//
// **WireGuard zur Verschlüsselung, magicsock zur NAT-Durchquerung, DERP als
// Rendezvous und Rückfall — ohne Tailscale-Konto und ohne Kontrollebene.**
// Die Einladung trägt die Schlüssel des Servers und seine DERP-Region und
// wird „exchanged out of band, however you want".
//
// # Was es nicht zusagt
//
// **Nichts.** „the Go API, the CLI flags and output, and the wire format may
// all change." Deshalb ist es auf `v0.7.0` gepinnt, und deshalb trägt die
// Einladung eine Fassung ([domain.NeueEinladung]) — der Bruch wird **vor**
// dem Wählen erkannt statt als Zeitüberschreitung.
//
// # Und was hier bewusst nicht geschieht
//
// **Es horcht nichts von selbst.** Dieselbe Linie wie bei N2: wer [Horche]
// nicht ruft, ist nicht erreichbar — auch nach einem Neustart nicht.
package fernleitung

import (
	"context"
	"net"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Port ist die Portnummer im Tunnel.
//
// **Eine feste, und das ist keine Einschränkung**: der Tunnel gehört diesen
// beiden Rechnern allein, und was darin an welchem Port liegt, entscheiden
// sie. Eine einstellbare Zahl wäre eine Frage ohne Antwort.
const Port = 4123

// Wahlfrist ist, wie lange das Wählen über ein fremdes Netz dauern darf.
//
// **Länger als im eigenen Netz** ([leitung.Wahlfrist] gibt zwanzig Sekunden):
// über DERP kommt erst ein Rendezvous zustande, dann versucht magicsock den
// direkten Weg, und erst wenn beides scheitert, steht das Ergebnis fest.
const Wahlfrist = 60 * time.Second

// fernserver ist, was dieses Paket von einem tailcat-Server braucht.
type fernserver interface {
	Start() error
	Einladungsadresse() string
	Close() error

	// Wege sagt je Gegenstelle, wie die Verbindung geht (NET-02.AK-03).
	Wege() []Fernweg
}

// Fernweg ist, wie eine Verbindung über fremde Netze geht — NET-02.AK-03
// (#1706).
//
// **Gesehen wird er nur an dieser Seite**, der, die eine Einladung
// ausgegeben hat: tailcat gibt den Stand seiner Gegenstellen am Server her
// (`ipnstate.PeerStatus`, `CurAddr` und `Relay`, nachgelesen in
// `tailscale.com/ipn/ipnstate` der gepinnten Fassung), am Klienten nicht.
type Fernweg struct {
	// Direkt ist die Adresse, über die WireGuard die Gegenstelle erreicht —
	// leer, solange es keinen direkten Weg gibt.
	Direkt string

	// Relais ist die DERP-Region, über die es geht, solange kein direkter
	// Weg steht.
	Relais string
}

// fernklient ist, was es von einem tailcat-Klienten braucht.
type fernklient interface {
	DialTCPPort(ctx context.Context, port uint16) (net.Conn, error)
	Close() error
}

// Waehle verbindet sich über eine Einladung.
//
// **Die Einladung wird zerlegt, bevor irgendetwas ins Netz geht** — eine aus
// einer neueren Fassung endet hier und nicht in einer Zeitüberschreitung.
func Waehle(
	ctx context.Context, einladung domain.Einladung,
	ich peer.Identitaet, erwartet peer.Kennung,
) (*leitung.Leitung, error) {
	adresse, fehler := domain.ZerlegeEinladung(einladung)
	if fehler != nil {
		return nil, fehler
	}

	ctx, abbrechen := context.WithTimeout(ctx, Wahlfrist)
	defer abbrechen()

	klient := klientBauen(adresse)

	strom, fehler := klient.DialTCPPort(ctx, Port)
	if fehler != nil {
		_ = klient.Close() //nolint:errcheck // der Weg endet hier ohnehin

		return nil, domain.MitGrund(domain.ErrPeerNichtErreichbar, fehler,
			"die Gegenstelle ist über das fremde Netz nicht erreichbar")
	}

	gesichert, fehler := leitung.UeberStrom(ctx, strom, ich, erwartet)
	if fehler != nil {
		_ = klient.Close() //nolint:errcheck // dito

		return nil, fehler
	}

	return gesichert, nil
}
