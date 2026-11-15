// Package leitung trägt Nachrichten über einen Strom.
//
// **Ein Blatt** (§2): es kennt `internal/domain` und `internal/peer` und sonst
// nichts aus diesem Projekt. Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum es diese Schicht gibt
//
// **Weil ein Strom keine Nachrichten kennt.** `service.Verbindung` verlangt
// `Sende([]byte)` — eine Nachricht. TCP liefert Bytes in irgendeiner
// Aufteilung: wer drei Hüllen hineinschreibt, bekommt am anderen Ende drei
// Bytefolgen, die mit den Hüllengrenzen nichts zu tun haben. Eine Hülle in
// zwei Stücken ist kein JSON mehr, und zwei Hüllen in einem Stück auch nicht.
//
// # Warum sie über `net.Conn` geht und nicht über ein Transport
//
// **Weil drei Wege denselben Strom liefern**: `net.Pipe` für zwei Anwendungen
// im selben Prozess, ein gewöhnlicher TCP-Wählvorgang für den Weg „Adresse von
// Hand", und `tailcat` für die Interneteinladung
// ([0112](../../docs/entscheidungen/0112-tailcat-statt-eines-eigenen-dienstes.md)).
// Eine Rahmung je Transport wären drei Rahmungen, von denen zwei nie jemand
// prüft.
//
// # Was hier **nicht** entschieden wird
//
// **Wer am anderen Ende ist.** [Leitung.Gegenueber] wird von außen gesetzt;
// belegt wird die Identität vom Handschlag aus N1, nicht von der Leitung. Eine
// Leitung, die eine Kennung behauptete, wäre die Stelle, an der sich jemand
// eine aussucht.
package leitung

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Rahmenlaenge ist das Gewicht des Längenpräfixes in Byte.
//
// **Vier, und damit passt jede erlaubte Länge hinein.** [peer.Hoechstlaenge]
// ist ein Mebibyte; vier Bytes tragen das Viertausendfache. Die Grenze
// entsteht also nicht aus dem Rahmen, sondern aus der Prüfung — und die steht
// dort, wo sie steht, weil sie sonst niemand sähe.
const Rahmenlaenge = 4

// rahmenfrist ist, wie lange der Rumpf eines angekündigten Rahmens brauchen
// darf — D05 (#2236).
//
// **Eine Sitzung darf ruhen, ein angekündigter Rahmen nicht.** Die langen
// Sitzungen empfangen ohne Frist im Kontext; ein Gegenüber, das einen Kopf
// über ein Mebibyte schickte und den Rumpf tröpfelte, hielt Leser und Puffer
// bis zum Ende der Sitzung. **Eine Minute**, weil ein Mebibyte über ein
// langsames Relais einige Sekunden braucht und keine Minute.
var rahmenfrist = time.Minute

// Leitung ist eine Verbindung, die Nachrichten trägt.
//
// **Sie erfüllt `service.Verbindung`, ohne das Paket zu kennen.** Go prüft
// Schnittstellen über die Gestalt; ein Import dorthin wäre eine Kante, die
// dieses Blatt nicht braucht (§2).
type Leitung struct {
	// schreiben hält einen Sender zur Zeit.
	//
	// # Was es hält und was heute ohnehin hält
	//
	// **Ein Rahmen geht als *ein* `Write` hinaus**, und darauf beruht die
	// Trennung zweier Nachrichten. `net.Pipe` und `*net.TCPConn` serialisieren
	// einen einzelnen `Write` von sich aus — die Gegenprobe zeigt es: dieses
	// Schloss zu entfernen macht **keinen** Lauf rot.
	//
	// **Es steht trotzdem da, und der Grund ist die Schnittstelle.**
	// `net.Conn` sagt nichts über gleichzeitige Schreibvorgänge zu; ein
	// eigener Strom — ein Doppel, ein Adapter aus tailcat — muss es nicht
	// tun. Und wer hier später zwei `Write` daraus macht, hätte ohne das
	// Schloss einen Fehler, den kein Lauf zeigt.
	//
	// **Was es damit ist**: eine Vorsorge, keine tragende Zusage. Der
	// Unterschied gehört hierher, weil ihn sonst niemand mehr kennt (§10).
	schreiben sync.Mutex

	// lesen hält einen Empfänger zur Zeit — aus demselben Grund.
	//
	// **Hier trägt es mehr**: ein Empfang sind **zwei** `Read` (Rahmen, dann
	// Rumpf). Zwei gleichzeitige Empfänger läsen den Rahmen des einen und den
	// Rumpf des anderen.
	lesen sync.Mutex

	stand sync.Mutex
	zu    bool

	strom net.Conn
	kanal string

	// gegenstelle ist die Kennung aus dem Zeugnis der Gegenseite — leer über
	// einen Strom ohne TLS (SEC-16).
	gegenstelle peer.Kennung
}

// Ueber baut eine Leitung über einen Strom.
//
// **Nur der Kanal kommt von außen, und keine Kennung.** Eine angenommene
// Verbindung hat kein Gegenüber, bis der Handschlag eines nennt — und die
// hereingereichte Kennung einer gewählten wäre eine Behauptung, die niemand
// prüft. Die Leitung trägt Bytes.
func Ueber(strom net.Conn, kanal string) *Leitung {
	return &Leitung{strom: strom, kanal: kanal}
}

// Kanal ist die Bindung dieser einen Verbindung.
func (l *Leitung) Kanal() string { return l.kanal }

// Gegenstelle ist die Kennung, deren Zeugnis die Gegenseite im TLS-Handschlag
// vorgezeigt hat — **geprüft**, nicht behauptet: der Schlüssel im Zeugnis ist
// die Kennung (SEC-16, #2145).
//
// **Leer über einen Strom ohne TLS** — eine Röhre im selben Prozess hat kein
// Zeugnis, und dort sagt allein der Handschlag, wer da ist.
func (l *Leitung) Gegenstelle() peer.Kennung { return l.gegenstelle }

// Sende schreibt eine Nachricht als einen Rahmen.
func (l *Leitung) Sende(ctx context.Context, nachricht []byte) error {
	if fehler := pruefeNachricht(nachricht); fehler != nil {
		return fehler
	}

	if fehler := l.offen(); fehler != nil {
		return fehler
	}

	aufhoeren := l.fristSetzen(ctx, false)
	defer aufhoeren()

	var kopf [Rahmenlaenge]byte

	// **Die Länge ist geprüft, bevor sie hier steht** — [pruefeNachricht]
	// hält sie zwischen eins und [peer.Hoechstlaenge], und ein Mebibyte
	// passt in 32 Bit.
	//
	// **`#nosec` und nicht `//nolint`**: `gosec` läuft in der Pipeline
	// eigenständig und liest die Anweisung von golangci-lint nicht — der
	// erste Anlauf hatte `//nolint:gosec`, und die Stufe `security` wurde rot.
	binary.BigEndian.PutUint32(kopf[:], uint32(len(nachricht))) //#nosec G115 -- siehe darüber: höchstens ein Mebibyte

	l.schreiben.Lock()
	defer l.schreiben.Unlock()

	if _, err := l.strom.Write(append(kopf[:], nachricht...)); err != nil {
		return domain.MitGrund(domain.ErrConflict, err,
			"eine Nachricht über die Leitung schicken")
	}

	return nil
}

// Empfange liest die nächste Nachricht.
//
// **Ein Ruf, eine Nachricht.** Wer einen Strom läse, bekäme irgendeine
// Aufteilung; hier kommt heraus, was am anderen Ende hineinging.
func (l *Leitung) Empfange(ctx context.Context) ([]byte, error) {
	if fehler := l.offen(); fehler != nil {
		return nil, fehler
	}

	aufhoeren := l.fristSetzen(ctx, true)
	defer aufhoeren()

	l.lesen.Lock()
	defer l.lesen.Unlock()

	var kopf [Rahmenlaenge]byte

	if _, err := io.ReadFull(l.strom, kopf[:]); err != nil {
		return nil, domain.MitGrund(domain.ErrConflict, err,
			"den Rahmen einer Nachricht lesen")
	}

	laenge := binary.BigEndian.Uint32(kopf[:])

	// **Geprüft wird, bevor zugeteilt wird, und das ist die ganze Zusage.**
	//
	// Der Rahmen ist eine **Behauptung** des Gegenübers. Wer ihr glaubt und
	// zuteilt, hat einem Fremden erlaubt, den Arbeitsspeicher dieses Rechners
	// zu bestimmen — mit vier Bytes.
	if laenge == 0 || laenge > peer.Hoechstlaenge {
		return nil, fmt.Errorf(
			"%w: der Rahmen behauptet %d Byte, erlaubt sind 1 bis %d",
			domain.ErrValidation, laenge, peer.Hoechstlaenge)
	}

	nachricht := make([]byte, laenge)
	l.rumpffristSetzen(ctx)

	if _, err := io.ReadFull(l.strom, nachricht); err != nil {
		return nil, domain.MitGrund(domain.ErrConflict, err,
			"eine Nachricht von der Leitung lesen")
	}

	return nachricht, nil
}

// Schliesse legt die Leitung nieder.
//
// **Zweimal schließen ist einmal** — das Ziel ist erreicht, und wer einen
// Fehler bekäme, fügte ihn seinem Aufräumweg hinzu.
func (l *Leitung) Schliesse() error {
	l.stand.Lock()

	if l.zu {
		l.stand.Unlock()

		return nil
	}

	l.zu = true

	l.stand.Unlock()

	if err := l.strom.Close(); err != nil {
		return domain.MitGrund(domain.ErrConflict, err,
			"die Leitung niederlegen")
	}

	return nil
}

// offen sagt, ob noch gearbeitet werden darf.
func (l *Leitung) offen() error {
	l.stand.Lock()
	defer l.stand.Unlock()

	if l.zu {
		return fmt.Errorf(
			"%w: diese Leitung ist geschlossen", domain.ErrConflict)
	}

	return nil
}

// fristSetzen überträgt Frist und Abbruch des Kontexts auf eine Richtung.
//
// # Warum je Richtung und nicht für beide
//
// **Weil `SetDeadline` beide setzt.** Ein Senden und ein Empfangen laufen
// nebeneinander; wer eine gemeinsame Frist setzte, bräche dem anderen die
// seine — und wer sie danach zurücknähme, höbe dem anderen seine auf.
// `SetWriteDeadline` und `SetReadDeadline` sind zwei Uhren, und jede gehört
// dem, der sie gesetzt hat.
//
// # Warum Frist **und** Abbruch
//
// **Weil ein Abbruch keine Frist ist.** `ctx.Deadline()` deckt „bis dahin und
// nicht länger"; wer mittendrin abbricht, hat keine — und ein `Read`, das auf
// ein Gegenüber wartet, das nichts mehr schickt, endete nie. `context.AfterFunc`
// setzt dann eine Frist in die Vergangenheit, und der Strom gibt auf.
//
// **Der Rückgabewert nimmt beides zurück**: er hält den Wächter an und räumt
// die Frist wieder weg. Ohne das Zweite bliebe sie für den nächsten Ruf
// stehen, und der scheiterte an einer Uhr, die ihn nichts angeht.
func (l *Leitung) fristSetzen(ctx context.Context, lesend bool) func() {
	setzen := l.strom.SetWriteDeadline
	if lesend {
		setzen = l.strom.SetReadDeadline
	}

	if frist, gibts := ctx.Deadline(); gibts {
		_ = setzen(frist) //nolint:errcheck // ein Strom ohne Fristen bleibt einer mit Abbruch
	}

	anhalten := context.AfterFunc(ctx, func() {
		// **Eine Frist in der Vergangenheit** bricht ab, was gerade läuft —
		// das ist der Weg, den `net.Conn` dafür kennt.
		_ = setzen(time.Unix(1, 0)) //nolint:errcheck // dito
	})

	return func() {
		anhalten()

		_ = setzen(time.Time{}) //nolint:errcheck // dito
	}
}

// rumpffristSetzen begrenzt den Rumpf eines angekündigten Rahmens auf die
// frühere von Kontextfrist und [rahmenfrist].
//
// **Ein Abbruch geht vor**: kam er zwischen Kopf und Rumpf, hat `fristSetzen`
// die Frist schon in die Vergangenheit gelegt, und eine neue überschriebe
// ihn. Deshalb wird danach noch einmal gefragt.
func (l *Leitung) rumpffristSetzen(ctx context.Context) {
	frist := time.Now().Add(rahmenfrist)
	if ende, gibts := ctx.Deadline(); gibts && ende.Before(frist) {
		frist = ende
	}

	_ = l.strom.SetReadDeadline(frist) //nolint:errcheck // ein Strom ohne Fristen bleibt einer mit Abbruch

	if ctx.Err() != nil {
		_ = l.strom.SetReadDeadline(time.Unix(1, 0)) //nolint:errcheck // dito
	}
}

// pruefeNachricht hält eine ausgehende Nachricht gegen die Grenzen.
func pruefeNachricht(nachricht []byte) error {
	if len(nachricht) == 0 {
		return fmt.Errorf(
			"%w: eine leere Nachricht ist keine", domain.ErrValidation)
	}

	if len(nachricht) > peer.Hoechstlaenge {
		return fmt.Errorf(
			"%w: die Nachricht wiegt %d Byte, erlaubt sind %d",
			domain.ErrValidation, len(nachricht), peer.Hoechstlaenge)
	}

	return nil
}

// Schleife gibt zwei Leitungen, die einander erreichen — ohne Netz.
//
// # Wofür sie da ist
//
// **Damit zwei Anwendungen im selben Prozess miteinander sprechen können.**
// Der ganze Stapel läuft darüber — Handschlag, Hülle, Kontakte, Sitzung und
// später der Objektaustausch —, und zwar **ohne ein einziges Paket**, ohne
// Berechtigung, ohne zweiten Rechner und ohne Flackern.
//
// **Sie ist kein Ersatz für zwei Macs** und behauptet das nicht: was sie nicht
// zeigt, sind die Berechtigung von macOS, die Firewall, das Einschlafen und
// der Netzwechsel. Sie zeigt jede Schicht darüber.
func Schleife(kanal string) (hiesige, dortige *Leitung) {
	// **Mit Puffer und nicht über `net.Pipe`** — der Grund steht in
	// `schleifenpuffer.go`: ein Strom ohne Puffer verklemmt, sobald beide
	// Enden zugleich etwas zu sagen haben, und das hat kein echter Transport.
	einer, anderer := gepuffertesPaar()

	return Ueber(einer, kanal), Ueber(anderer, kanal)
}
