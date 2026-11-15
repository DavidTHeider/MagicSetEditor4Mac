package leitung

// **Unter welcher Adresse jemand anderes ankommt** — der Befund aus der
// Sichtprüfung vom 2026-09-14.
//
// # Warum das nicht die Bindeadresse ist
//
// **Weil ein Lauscher sagt, wo er gebunden hat, und nicht, wo jemand
// ankommt.** Gemessen am 2026-09-14 auf diesem Rechner:
//
//	:41234          -> [::]:41234
//	127.0.0.1:41234 -> 127.0.0.1:41234
//
// Das erste ist keine Adresse, zu der sich wählen ließe; das zweite zeigt auf
// den eigenen Rechner. Beides stand im Fenster unter „Erreichbar unter …",
// und `192.168.178.51:41234` — die einzige brauchbare — stand nirgends.
//
// # Warum ein Name dazugehört
//
// **Weil eine IP-Adresse geliehen ist.** Sie kommt in den meisten Netzen über
// DHCP, und eine Lease endet: wer die Adresse weitergibt, gibt eine
// Momentaufnahme weiter. Ein Name, den das Netz auflöst, überlebt den
// Wechsel — er zeigt nach dem neuen Lease auf die neue Adresse.
//
// **Und er wird geprüft statt geraten.** Gemessen am selben Tag:
//
//	os.Hostname()                               = "MacBook-Air-von-Ada.fritz.box"
//	LookupHost("MacBook-Air-von-Ada.fritz.box") = [192.168.178.51 2003:… fdc0:…]
//	LookupHost("…fritz.box.local")              = []  (i/o timeout)
//
// Der Router vergibt hier einen auflösbaren Namen; ein blind angehängtes
// `.local` ergibt Unsinn. Ein Name zählt deshalb nur, wenn er auf eine
// Adresse **dieses** Rechners auflöst — sonst schickte er jemanden zu einem
// Fremden.
//
// # Was dauerhaft trägt, steht noch aus
//
// **Weder Adresse noch Name sind Zusagen.** Beide gelten in dem Netz, in dem
// sie erhoben wurden. Die Adresse, die überall gilt, ist die Peerkennung
// selbst — und die trägt erst N3
// ([0112](../../docs/entscheidungen/0112-tailcat-statt-eines-eigenen-dienstes.md)).
//
// # Warum hier und nicht in `internal/entdeckung`
//
// **Weil es die Frage dieses Pakets ist.** Dort geht es darum, **wer** da
// ist; hier darum, **wie** Bytes von dort hierher kommen. Der Weg „Adresse
// von Hand" braucht diese Antwort und kennt mDNS gar nicht.

import (
	"context"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Namensfrist ist, wie lange auf die Auflösung eines Namens gewartet wird.
//
// **Zwei Sekunden für alle Anwärter zusammen**, nicht je einer. Ein Name, der
// nicht auflöst, läuft in eine Zeitüberschreitung — gemessen am 2026-09-14
// drei Sekunden für `…fritz.box.local` —, und wer je Anwärter wartete,
// hielte den Knopfdruck doppelt so lange auf.
const Namensfrist = 2 * time.Second

// Die Nähte zu diesem Rechner — **als Veränderliche, damit ein Lauf ohne ihn
// auskommt** (§3). Sonst prüfte jeder Test, was zufällig in diesem Netz
// steht, und wäre auf einer anderen Maschine rot.
var (
	schnittstellenadressen = net.InterfaceAddrs
	rechnername            = os.Hostname
	aufloesen              = net.DefaultResolver.LookupHost
)

// Erreichbarkeit ist, was ein Gegenüber eintippen kann.
//
// **Beide Listen sind leer und nie `nil`.** Eine nil-Scheibe wird zu `null`,
// und das Fenster bekäme keine leere Liste, sondern gar keine — genau die
// Klasse, vor der §3 warnt und die `leerelistenlaufzeit_test.go` hält.
type Erreichbarkeit struct {
	// Namen sind die geprüften — **sie überleben einen Leasewechsel**.
	Namen []string

	// Adressen sind die IP-Adressen dieses Rechners mit dem Port.
	Adressen []string
}

// Leer sagt, ob von außen niemand ankommt.
func (e Erreichbarkeit) Leer() bool {
	return len(e.Namen) == 0 && len(e.Adressen) == 0
}

// ErreichbarUnter sagt, unter welchen Adressen ein Horchposten erreichbar ist.
//
// Hinein geht, was der Lauscher meldet; heraus kommt, was ein Gegenüber
// eintippen kann. Drei Fälle, und der mittlere ist der, den es zu sagen gilt:
//
//   - **Ein bestimmter Wirt** — genau er, denn dort und nirgends sonst wird
//     gehorcht.
//   - **Die Rückschleife** — **nichts**. Wer dort horcht, ist von außen nicht
//     erreichbar, und eine leere Antwort ist die ehrliche Auskunft. Eine
//     Adresse zu nennen, unter der niemand ankommt, wäre die teurere Hälfte.
//   - **Jede Schnittstelle** (`::`, `0.0.0.0` oder leer) — die Adressen
//     dieses Rechners, ohne die, die niemandem nützen, dazu seine geprüften
//     Namen.
//
// # Was ausgelassen wird und warum
//
// **Die Rückschleife**, weil sie auf den Fragenden selbst zeigt.
// **Link-Local** (`fe80::…`), weil eine solche Adresse ohne Zonenangabe
// (`%en0`) nicht wählbar ist — und die Zone des einen Rechners sagt über die
// des anderen nichts. **Und alles, was keine IP ist**: `net.InterfaceAddrs`
// gibt `net.Addr`, und darin steckt nicht immer eine.
func ErreichbarUnter(
	ctx context.Context, horchadresse string,
) (Erreichbarkeit, error) {
	wirt, port, fehler := domain.ZerlegeAdresse(horchadresse)
	if fehler != nil {
		return leer(), fehler
	}

	if wirt != "" {
		gebunden := net.ParseIP(wirt)

		// **Ein Name ist kein Grund, die Liste zu ziehen.** Wer einen
		// Rechnernamen gebunden hat, hat ihn gebunden — er geht so hinaus,
		// wie er hereinkam.
		if gebunden == nil {
			return Erreichbarkeit{Namen: []string{mitPort(wirt, port)}}, nil
		}

		if gebunden.IsLoopback() {
			return Erreichbarkeit{}, nil
		}

		if !gebunden.IsUnspecified() {
			return Erreichbarkeit{
				Adressen: []string{mitPort(wirt, port)},
			}, nil
		}
	}

	eigene, fehler := eigeneAdressen()
	if fehler != nil {
		return leer(), fehler
	}

	return Erreichbarkeit{
		Namen:    geprueftenNamen(ctx, eigene, port),
		Adressen: mitPortAlle(eigene, port),
	}, nil
}

// eigeneAdressen zählt die Adressen dieses Rechners auf, die etwas taugen.
// leer ist die Antwort „von außen erreicht dich niemand".
func leer() Erreichbarkeit {
	return Erreichbarkeit{Namen: []string{}, Adressen: []string{}}
}

func eigeneAdressen() ([]net.IP, error) {
	liste, fehler := schnittstellenadressen()
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Adressen dieses Rechners lesen")
	}

	var brauchbar []net.IP

	for _, eintrag := range liste {
		netzadresse, taugt := eintrag.(*net.IPNet)
		if !taugt {
			continue
		}

		adresse := netzadresse.IP
		if adresse.IsLoopback() || adresse.IsLinkLocalUnicast() ||
			adresse.IsUnspecified() {
			continue
		}

		brauchbar = append(brauchbar, adresse)
	}

	ordne(brauchbar)

	return brauchbar, nil
}

// ordne stellt die brauchbarste Adresse nach vorn.
//
// # Warum überhaupt geordnet wird
//
// **Weil die Reihenfolge des Betriebssystems keine ist.** Die Sichtprüfung
// vom 2026-09-14 zeigte sieben Zeilen, und `192.168.178.51` — die eine, die
// ein Mensch eintippt — stand an **vierter** Stelle. Sechs der sieben waren
// IPv6, fünf davon temporäre Adressen aus den Privacy Extensions.
//
// # Die Ordnung, und warum sie so und nicht anders ist
//
//  1. **IPv4** — das ist, was ein Mensch abschreibt und was in jedem Netz
//     ankommt, auch in einem ohne IPv6.
//  2. **Eindeutig lokal** (`fc00::/7`) — im eigenen Netz routbar und
//     **stabil**: sie kommt nicht von einem Anbieter und wechselt nicht mit
//     einer Zuteilung.
//  3. **Alles Übrige** — global gültige IPv6. Sie trägt weiter, wechselt aber
//     mit dem Präfix des Anschlusses, und macOS bildet daraus täglich neue
//     temporäre Adressen.
//
// **Gekürzt wird nicht.** Was dasteht, ist wahr, und eine Liste, die etwas
// verschweigt, wäre die teurere Hälfte.
//
// **Und temporär wird nicht von dauerhaft getrennt.** `net.InterfaceAddrs`
// gibt keine Flagge dafür her; wer es unterscheiden wollte, müsste `ifconfig`
// lesen — plattformeigen, teuer und für eine Reihenfolge zu viel. Der Satz
// unter der Liste sagt stattdessen, dass eine geliehene Adresse wechseln
// kann.
//
// **Stabil sortiert**, damit zwei Läufe auf demselben Rechner dieselbe
// Reihenfolge geben: eine Liste, die bei jedem Zeichnen springt, ist nicht
// abzuschreiben.
func ordne(adressen []net.IP) {
	sort.SliceStable(adressen, func(a, b int) bool {
		return rang(adressen[a]) < rang(adressen[b])
	})
}

// rang sagt, wie weit vorn eine Adresse steht — kleiner ist weiter vorn.
func rang(adresse net.IP) int {
	switch {
	case adresse.To4() != nil:
		return 0
	case adresse[0]&0xfe == 0xfc:
		return 1
	default:
		return 2
	}
}

// geprueftenNamen gibt die Namen, die auf diesen Rechner zeigen.
//
// **Ein Befund kostet hier nur den Namen.** Wer keinen Namen auflösen kann,
// bekommt die Adressen — und nicht eine Fehlermeldung statt beider.
func geprueftenNamen(ctx context.Context, eigene []net.IP, port int) []string {
	geprueft := []string{}

	name, fehler := rechnername()
	if fehler != nil || name == "" {
		return geprueft
	}

	anwaerter := []string{name}

	// **`.local` nur an einen unqualifizierten Namen.** An
	// `…fritz.box` gehängt ergibt es einen Namen, den niemand kennt —
	// gemessen am 2026-09-14: drei Sekunden bis zur Zeitüberschreitung.
	if !strings.Contains(name, ".") {
		anwaerter = append(anwaerter, name+".local")
	}

	frist, ablauf := context.WithTimeout(ctx, Namensfrist)
	defer ablauf()

	for _, anwaerter := range anwaerter {
		if zeigtHierher(frist, anwaerter, eigene) {
			geprueft = append(geprueft, mitPort(anwaerter, port))
		}
	}

	return geprueft
}

// zeigtHierher sagt, ob ein Name auf eine Adresse dieses Rechners auflöst.
func zeigtHierher(ctx context.Context, name string, eigene []net.IP) bool {
	aufgeloest, fehler := aufloesen(ctx, name)
	if fehler != nil {
		return false
	}

	for _, text := range aufgeloest {
		adresse := net.ParseIP(text)
		if adresse == nil {
			continue
		}

		for _, meine := range eigene {
			if adresse.Equal(meine) {
				return true
			}
		}
	}

	return false
}

// mitPort setzt Wirt und Port zusammen — eckige Klammern inklusive.
func mitPort(wirt string, port int) string {
	return net.JoinHostPort(wirt, strconv.Itoa(port))
}

// mitPortAlle tut dasselbe für eine Liste.
func mitPortAlle(adressen []net.IP, port int) []string {
	liste := make([]string, 0, len(adressen))

	for _, adresse := range adressen {
		liste = append(liste, mitPort(adresse.String(), port))
	}

	return liste
}
