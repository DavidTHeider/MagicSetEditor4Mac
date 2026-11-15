package domain

// **Wohin ein Dienst bindet** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Warum das ein Wertetyp ist und kein Paar aus Zeichenkette und Zahl
//
// **Weil dieselben vier Fragen sonst an jeder Fläche neu beantwortet würden**:
// was ein leeres Feld heißt, wie die Adresse in einen Lauscher geht, ob sie
// den Rechner verlässt und was überhaupt eine ist. Vier Antworten an drei
// Stellen sind zwölf Gelegenheiten, auseinanderzulaufen — und bei der
// dritten Frage hängt an der Abweichung, ob jemand ohne TLS ins Netz hört.
//
// # Warum `net/netip` und nicht `net`
//
// **Weil `netip.Addr` ein Wert ist und kein Zugang.** `internal/domain` darf
// das Netz nicht erreichen (`internal/schichten/netz_test.go`), und `net`
// steht dort auf der Liste, weil `net.Listen` nichts weiter braucht.
// `net/netip` kann **nichts** öffnen — dieselbe Lage wie bei `net/url`, das
// der Wächter aus demselben Grund draußen lässt. Der Unterschied ist echt und
// keine Umgehung.
//
// **Weil die eine Frage, die daran hängt, nicht aus dem Augenschein folgt**:
// *verlässt diese Bindung diesen Rechner?* Von ihr hängt ab, ob TLS und ein
// Zugangstoken Pflicht sind
// ([0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)) —
// und eine Antwort, die an drei Stellen neu erraten würde, liefe an einer
// davon auseinander.
//
// # Und warum „auf der Schleife" mehr ist als `127.0.0.1`
//
// **Weil `::1` auch eine ist, und `localhost` einer sein kann.** Bis N6 nahm
// `internal/rest` genau eine Zeichenkette an und wies jede andere ab — das
// war richtig, solange es **nur** die Schleife geben sollte. Sobald es eine
// Wahl gibt, muss die Frage beantwortet werden statt umgangen.

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

// Schleifenadresse ist die Vorgabe jeder Bindung.
//
// **Nicht `localhost`.** Der Name wird aufgelöst und kann auf einem fremd
// eingerichteten Rechner auf etwas anderes zeigen; eine Adresse kann das
// nicht. Der Satz stand schon in `internal/rest` und gilt weiter.
const Schleifenadresse = "127.0.0.1"

// JedeSchnittstelle ist die Bindung, die den Rechner am weitesten öffnet.
//
// **Sie steht hier als Konstante, damit man sie suchen kann.** Wer wissen
// will, wo dieses Programm `0.0.0.0` überhaupt zulässt, findet die Stellen
// über diesen Namen und nicht über eine Zeichenkette in einem Formular.
const JedeSchnittstelle = "0.0.0.0"

// Bindung ist eine gewünschte Adresse samt Port.
type Bindung struct {
	// Adresse ist die Schnittstelle — **leer heißt die Schleife**.
	Adresse string

	// Port ist der gewünschte Port — **0 heißt: einer vom Betriebssystem**.
	//
	// Welcher es wurde, sagt erst der offene Lauscher. Eine Bindung ist ein
	// **Wunsch**; was daraus wird, steht danach am Dienst.
	Port int
}

// Wirkadresse gibt die Adresse, die wirklich gemeint ist.
//
// **Leer heißt die Schleife**, und diese Deutung steht an **einer** Stelle —
// sonst hieße ein leeres Feld an einer Fläche „die Schleife" und an einer
// anderen „jede Schnittstelle", und das ist der Unterschied zwischen einem
// Dienst auf diesem Rechner und einem im Netz.
func (b Bindung) Wirkadresse() string {
	if strings.TrimSpace(b.Adresse) == "" {
		return Schleifenadresse
	}

	return b.Adresse
}

// Adresstext gibt die Bindung in der Gestalt, die ein Lauscher erwartet.
//
// **Mit Klammern um eine IPv6-Adresse** — ohne sie läse jeder Lauscher den
// Doppelpunkt vor dem Port als Teil der Adresse.
func (b Bindung) Adresstext() string {
	adresse := b.Wirkadresse()
	if strings.Contains(adresse, ":") {
		return "[" + adresse + "]:" + strconv.Itoa(b.Port)
	}

	return adresse + ":" + strconv.Itoa(b.Port)
}

// VerlaesstDenRechner sagt, ob jemand von außen anklopfen kann.
//
// **Die Frage, an der TLS und Token hängen** (0115). Sie wird **an der
// Adresse** beantwortet und nicht am Wunsch: `net.IP.IsLoopback` kennt den
// ganzen Bereich `127.0.0.0/8` und `::1`, und wer nur `127.0.0.1` mit einer
// Zeichenkette verglich, hielte `127.0.0.2` für eine Öffnung nach außen.
//
// **Und was keine Adresse ist, verlässt den Rechner.** Ein Name wird
// aufgelöst, und wohin, weiß dieses Paket nicht — `fail closed` ist hier die
// einzige vertretbare Antwort.
func (b Bindung) VerlaesstDenRechner() bool {
	adresse, fehler := netip.ParseAddr(b.Wirkadresse())

	return fehler != nil || !adresse.IsLoopback()
}

// PruefeBindung weist ab, was keine Bindung sein kann.
//
// **Ein Name wird abgewiesen und nicht aufgelöst.** `localhost` sähe aus wie
// die Schleife und ist es auf einem fremd eingerichteten Rechner nicht; alles
// andere wäre eine Adresse, die dieses Programm zur Laufzeit nachschlägt, um
// eine **Sicherheitsfrage** zu beantworten. Wer eine Schnittstelle meint,
// nennt ihre Adresse.
func PruefeBindung(b Bindung) error {
	if b.Port < 0 || b.Port > 65535 {
		return fmt.Errorf("%w: %d ist kein Port", ErrValidation, b.Port)
	}

	if strings.TrimSpace(b.Adresse) == "" {
		return nil
	}

	if _, fehler := netip.ParseAddr(b.Adresse); fehler != nil {
		return fmt.Errorf(
			"%w: %q ist keine Adresse — ein Name wird aufgelöst, und wohin, "+
				"entscheidet dieser Rechner erst zur Laufzeit",
			ErrValidation, b.Adresse)
	}

	return nil
}

// Dienstart sagt, welcher der beiden Dienste gemeint ist.
//
// **Zwei und nicht beliebig viele.** Der Vermerk 0115 entscheidet über REST
// und MCP; ein dritter wäre eine neue Entscheidung und keine neue Zeile.
type Dienstart string

const (
	// DienstREST ist der Dienst aus `internal/rest`.
	DienstREST Dienstart = "rest"

	// DienstMCP ist der Werkzeugserver aus `internal/mcp`.
	DienstMCP Dienstart = "mcp"
)

// Dienstarten gibt beide.
func Dienstarten() []Dienstart { return []Dienstart{DienstREST, DienstMCP} }
