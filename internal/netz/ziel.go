package netz

// **Wohin ein Klient wählen darf** — SEC-P4
// (#2000),
// Befund P4 in `docs/PENTEST.md`.
//
// # Der Befund
//
// **Der Klient wählte jede Adresse, die man ihm nannte** — ohne Prüfung auf
// die Schleife, das eigene Netz oder `169.254.0.0/16`, und ohne eigene
// Politik für Weiterleitungen. Wer eine Adresse setzen konnte — ein Token
// `schreiben` beim Paketbezug —, ließ die Anwendung damit einen Dienst
// fragen, den nur dieser Rechner erreicht.
//
// # Zwei Bereiche, und die Grundsperre gilt beiden
//
// **Die Grundsperre** ist, was kein Dienst je braucht: `0.0.0.0` und `::` —
// eine Verbindung dorthin landet auf macOS bei diesem Rechner, gemessen am
// 2026-10-02 —, die verbindungslokalen Bereiche, in denen Metadatendienste
// stehen, Gruppen- und reservierte Adressen. Sie gilt **jedem** Klienten,
// auch denen mit fester Adresse, und auch der Leitung zum Proxy.
//
// **[ZielOeffentlich] sperrt dazu die Schleife und das eigene Netz** — für
// eine Adresse, die ein Rufer nennt. Die Schleife öffnet allein der benannte
// Schalter [SchleifeUmgebung]; das eigene Netz öffnet nichts.
//
// # Warum an drei Stellen
//
//  1. **Vor der Anfrage** ([Klient.darfFragen]): eine Adresse in Ziffern und
//     das Schema. Das ist die einzige Prüfung, die auch hinter einem Proxy
//     greift — was drüben aufgelöst wird, sieht dieser Rechner nicht.
//  2. **Nach der Auflösung** (`net.Dialer.Control`): die Adresse, mit der
//     wirklich verbunden wird. Ein Name, der auf die Schleife zeigt, kommt
//     erst hier zum Vorschein — und hier, **bevor** ein Paket hinausgeht.
//  3. **Bei jeder Weiterleitung** (`http.Client.CheckRedirect`): derselbe
//     Maßstab wie vor der ersten Anfrage. Ohne ihn führte eine öffentliche
//     Adresse mit einem `302` hinter die Sperre.
//
// # Was sie nicht deckt
//
//   - **Einen Namen, den ein Proxy auflöst.** Die Leitung zum Proxy steht
//     unter der Grundsperre und darf im eigenen Netz liegen — dort steht ein
//     Firmenproxy. Wohin er weiterreicht, entscheidet er.
//   - **Die Leitung zu einem Gegenüber** (`internal/leitung`): sie wählt,
//     was eine Einladung nennt, und das eigene Netz ist dort der Zweck. Sie
//     steht hinter `externSenden` statt hinter einer Bereichssperre.
//   - **6to4 und Teredo**: sie betten eine IPv4-Adresse ein, die ein fremdes
//     Relais erreicht — nicht dieser Rechner. Eingebettet geprüft wird nur
//     NAT64 (`64:ff9b::/96`), weil das Relais dort im eigenen Netz steht.

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Zielbereich sagt, wohin ein Klient wählen darf.
type Zielbereich int

const (
	// ZielUeberall ist jede Adresse außer der Grundsperre — für Dienste mit
	// fester Adresse, und für die Schleife, auf der ein Lauf sie nachstellt.
	ZielUeberall Zielbereich = iota

	// ZielOeffentlich ist nur eine öffentlich gerichtete Adresse — für eine
	// Adresse, die ein Rufer nennt.
	ZielOeffentlich
)

// SchleifeUmgebung ist der benannte Schalter für die Schleife.
//
// **`an` öffnet die Schleife für [ZielOeffentlich] — und nur sie.** Das
// eigene Netz und die Grundsperre bleiben zu. Gebraucht wird er von Läufen,
// deren Gegenstelle ein `httptest`-Server auf `127.0.0.1` ist; im Betrieb
// setzt ihn niemand.
//
// **Eine Umgebungsangabe und keine Einstellung**: eine Einstellung setzt,
// wer schreiben darf — also genau der Rufer, gegen den die Sperre steht. Die
// Umgebung setzt, wer den Prozess startet.
const SchleifeUmgebung = "MSE_NETZ_SCHLEIFE"

// ErrZielGesperrt meldet ein Ziel in einem gesperrten Bereich.
//
// **Eine Eingabe, die nicht taugt** ([domain.ErrValidation]): die Adresse
// hat jemand genannt, und an ihr ist etwas zu ändern.
var ErrZielGesperrt = fmt.Errorf("%w: dieses Ziel ist gesperrt", domain.ErrValidation)

// hoechstensWeiterleitungen ist die Grenze, die Go selbst zieht.
//
// **Übernommen und nicht neu gewählt**: ein eigenes `CheckRedirect` ersetzt
// die Vorgabe ganz, und ohne diese Zeile folgte der Klient einem Kreis bis
// zur Frist.
const hoechstensWeiterleitungen = 10

// NeuerKlientNach baut einen Klienten für einen Zielbereich.
func NeuerKlientNach(fristen Fristen, quelle Proxyquelle, bereich Zielbereich) *Klient {
	return neuerKlient(fristen, quelle, bereich)
}

// pruefeZiel prüft eine Adresse vor der Anfrage — Schema und Ziffern.
//
// **Ein Name geht hier durch.** Was er bedeutet, sagt erst die Auflösung,
// und die prüft [kontrolleFuer].
func pruefeZiel(ziel *url.URL, bereich Zielbereich) error {
	if ziel.Scheme != "http" && ziel.Scheme != "https" {
		return fmt.Errorf("%w — %q ist kein http und kein https", ErrZielGesperrt, ziel.Scheme)
	}

	adresse, istZiffer := ipAusName(ziel.Hostname())
	if !istZiffer {
		return nil
	}

	return gesperrtFuer(adresse, bereich)
}

// ipAusName liest einen Rechnernamen als Adresse — ohne Zone.
func ipAusName(name string) (netip.Addr, bool) {
	adresse, fehler := netip.ParseAddr(name)
	if fehler != nil {
		return netip.Addr{}, false
	}

	return adresse.WithZone(""), true
}

// gesperrtFuer gibt den Fehler, wenn die Adresse gesperrt ist.
func gesperrtFuer(adresse netip.Addr, bereich Zielbereich) error {
	if grund := sperrgrund(adresse, bereich, schleifeOffen()); grund != "" {
		return fmt.Errorf("%w — %s %s", ErrZielGesperrt, adresse, grund)
	}

	return nil
}

// schleifeOffen liest den benannten Schalter — **je Ruf**, damit ein Lauf
// ihn mit `t.Setenv` setzen kann.
func schleifeOffen() bool { return os.Getenv(SchleifeUmgebung) == "an" }

// kontrolleFuer prüft die Adresse, mit der wirklich verbunden wird.
//
// **Nach der Auflösung und vor dem Verbinden**: `net.Dialer.Control` läuft,
// wenn das Socket steht und noch nichts gesendet ist. Eine Adresse, die
// sich nicht lesen lässt, wird abgewiesen — eine Sperre, die bei einem
// Rätsel öffnet, wäre keine.
func kontrolleFuer(bereich Zielbereich) func(string, string, syscall.RawConn) error {
	return func(_, ziel string, _ syscall.RawConn) error {
		gelesen, fehler := netip.ParseAddrPort(ziel)
		if fehler != nil {
			return fmt.Errorf("%w — %q ist keine Adresse", ErrZielGesperrt, ziel)
		}

		return gesperrtFuer(gelesen.Addr().WithZone(""), bereich)
	}
}

// weiterleitungFuer prüft jede Weiterleitung wie die erste Anfrage.
func weiterleitungFuer(bereich Zielbereich) func(*http.Request, []*http.Request) error {
	return func(anfrage *http.Request, vorher []*http.Request) error {
		if len(vorher) >= hoechstensWeiterleitungen {
			return fmt.Errorf("nach %d Weiterleitungen angehalten", hoechstensWeiterleitungen)
		}

		return pruefeZiel(anfrage.URL, bereich)
	}
}

// Die Bereiche, die `netip` nicht selbst benennt.
var (
	// unbestimmt4 ist „dieses Netz" (RFC 791) — `0.0.0.0` allein ist zu eng.
	unbestimmt4 = netip.MustParsePrefix("0.0.0.0/8")

	// reserviert4 ist die frühere Klasse E samt Rundruf (RFC 1112).
	reserviert4 = netip.MustParsePrefix("240.0.0.0/4")

	// nat64 bettet eine IPv4-Adresse ein (RFC 6052).
	nat64 = netip.MustParsePrefix("64:ff9b::/96")

	// eigenesNetz sind die Bereiche, die kein Dienst im Internet trägt und
	// `netip.Addr.IsPrivate` nicht kennt.
	eigenesNetz = []netip.Prefix{
		netip.MustParsePrefix("100.64.0.0/10"), // geteilt (RFC 6598) — auch Tailscale
		netip.MustParsePrefix("192.0.0.0/24"),  // Protokollzuweisungen (RFC 6890)
		netip.MustParsePrefix("198.18.0.0/15"), // Prüfnetze (RFC 2544)
		netip.MustParsePrefix("fec0::/10"),     // standortlokal (RFC 3879)
	}
)

// sperrgrund sagt, warum eine Adresse gesperrt ist — leer heißt: frei.
func sperrgrund(adresse netip.Addr, bereich Zielbereich, schleife bool) string {
	adresse = adresse.Unmap()

	if nat64.Contains(adresse) {
		roh := adresse.As16()
		adresse = netip.AddrFrom4([4]byte(roh[12:]))
	}

	if grund := grundsperre(adresse); grund != "" {
		return grund
	}

	if bereich == ZielUeberall {
		return ""
	}

	switch {
	case adresse.IsLoopback() && !schleife:
		return "liegt auf der Schleife"
	case adresse.IsPrivate() || imEigenenNetz(adresse):
		return "liegt im eigenen Netz"
	}

	return ""
}

// grundsperre ist, was kein Klient je wählt.
func grundsperre(adresse netip.Addr) string {
	switch {
	case adresse.IsUnspecified() || unbestimmt4.Contains(adresse):
		return "ist unbestimmt und meint diesen Rechner"
	case adresse.IsLinkLocalUnicast():
		return "ist verbindungslokal — dort stehen Metadatendienste"
	case adresse.IsMulticast() || reserviert4.Contains(adresse):
		return "ist eine Gruppen- oder reservierte Adresse"
	}

	return ""
}

// imEigenenNetz prüft die Bereiche, die `IsPrivate` nicht kennt.
func imEigenenNetz(adresse netip.Addr) bool {
	for _, bereich := range eigenesNetz {
		if bereich.Contains(adresse) {
			return true
		}
	}

	return false
}

// istProxyleitung sagt, ob eine Verbindung zum Proxy geht und nicht zum Ziel.
//
// **Der Transport wählt beide über denselben Wähler**; was sie unterscheidet,
// ist die Adresse. Verglichen wird mit dem Stand, der **jetzt** gilt — dem,
// aus dem auch die Auswahl je Anfrage rechnet ([proxyAus]).
//
// **Für die Leitung zum Proxy gilt die Grundsperre und nicht der Bereich des
// Klienten**: ein Firmenproxy steht im eigenen Netz, und eine Sperre, die
// ihn träfe, machte jeden Bezug dahinter unmöglich. Dass ein Rufer den Proxy
// nicht auf ein Ziel seiner Wahl stellt, hält der Riegel an den Einstellungen.
func istProxyleitung(ctx context.Context, ziel string, quelle Proxyquelle) bool {
	stand := StandMit(ctx, quelle)

	for _, wert := range []string{stand.HTTP.Wert, stand.HTTPS.Wert} {
		if leitung, gelesen := proxyleitung(wert); gelesen && strings.EqualFold(leitung, ziel) {
			return true
		}
	}

	return false
}

// proxyports sind die Ports, die der Transport ohne Angabe wählt.
var proxyports = map[string]string{"http": "80", "https": "443", "socks5": "1080", "socks5h": "1080"}

// proxyleitung rechnet aus einer Proxyangabe die Adresse, die der Transport
// wählt — **wie `httpproxy`**: ohne Schema heißt `http://`.
func proxyleitung(wert string) (string, bool) {
	gelesen, fehler := url.Parse(strings.TrimSpace(wert))
	if fehler != nil || gelesen.Scheme == "" || gelesen.Host == "" {
		gelesen, fehler = url.Parse("http://" + strings.TrimSpace(wert))
		if fehler != nil {
			return "", false
		}
	}

	port := gelesen.Port()
	if port == "" {
		port = proxyports[gelesen.Scheme]
	}

	if gelesen.Hostname() == "" || port == "" {
		return "", false
	}

	return net.JoinHostPort(strings.ToLower(gelesen.Hostname()), port), true
}
