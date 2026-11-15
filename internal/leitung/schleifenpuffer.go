package leitung

// **Eine Schleife mit Puffer** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum `net.Pipe` nicht genügt
//
// **Weil es keinen Puffer hat.** Ein Schreibvorgang darauf kehrt erst zurück,
// wenn die Gegenseite liest — und das hat kein Transport dieser Welt. Ein
// Socket hat einen Kernpuffer, tailcat hat einen, selbst eine serielle
// Schnittstelle hat einen.
//
// **Und dieser Unterschied ist kein Detail, er ist eine Verklemmung.** Sobald
// beide Enden **gleichzeitig** etwas zu sagen haben — der Gast meldet seine
// Rückkehr, der Wirt antwortet mit Bild und Strom —, schreiben beide und
// keiner liest. Gemessen am 2026-09-16: beide Schleifen standen in
// `Sitzungsgriff.Leere`, und der ganze Lauf lief in seine Frist.
//
// **Was die Schleife zeigen soll, ist der Stapel darüber** — Handschlag,
// Hülle, Sitzung, Partie. Eine Eigenschaft, die kein echter Transport hat,
// zeigt davon nichts; sie erfindet nur einen Fehler, den es sonst nirgends
// gibt.
//
// # Was sie weiterhin nicht zeigt
//
// **Keine Paketgrenzen, keine Firewall, kein Einschlafen, keinen
// Netzwechsel** — und auch keinen **vollen** Puffer: wer mehr als
// [Schleifenpuffer] Nachrichten schickt, ohne dass jemand liest, wartet hier
// genauso wie an einem echten Socket. Das ist die Grenze dieser Nachbildung,
// und sie steht hier statt in einer Erwartung.

import (
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Schleifenpuffer ist, wie viele Stücke unterwegs sein dürfen.
//
// **Vierundsechzig, und die Zahl ist eine Setzung**: sie ist groß genug für
// jeden Wortwechsel, den dieser Stapel führt — eine Rückkehr samt Bild und
// nachgereichtem Strom —, und klein genug, dass ein Lauf, der wirklich
// klemmt, stehenbleibt statt unbegrenzt zu wachsen.
const Schleifenpuffer = 64

// schleifenende ist ein [net.Conn] über zwei Kanäle.
type schleifenende struct {
	hinaus chan<- []byte
	herein <-chan []byte

	// rest ist, was vom letzten Stück noch nicht abgeholt wurde.
	//
	// **Ein Strom hat keine Grenzen**, und ein Leser darf weniger holen, als
	// geschrieben wurde — wer den Überhang wegwürfe, verlöre Bytes mitten in
	// einem Rahmen.
	rest []byte

	// zu sagt, dass **dieses** Ende niedergelegt ist, drueben, dass es das
	// andere ist.
	//
	// **Zwei Signale und nicht eines** (#2504). Hier stand ein gemeinsames,
	// und wer las, nachdem drüben geschlossen war, bekam `net.ErrClosed` —
	// auch für Stücke, die vor dem Schließen hinausgegangen waren. Ein Socket
	// liefert sie und meldet **danach** das Ende; ein Handschlag, dessen
	// letzter Schritt kurz vor dem Schließen hinausging, kam sonst unter Last
	// nicht an.
	//
	// **Und je Ende ein eigener Riegel**, denn jedes schließt nur sein
	// eigenes Signal. Der erste Anlauf teilte den Kanal und gab jedem Ende
	// ein eigenes `sync.Once` — wer beide Enden niederlegte, schloss denselben
	// Kanal zweimal.
	zu      chan struct{}
	drueben <-chan struct{}
	einmal  sync.Once

	// fristen hütet die Zeitpunkte und ihre zwei Wecksignale.
	//
	// **Die Signale sind der Punkt.** `internal/leitung` bricht ein
	// **laufendes** Warten ab, indem es eine Frist in die **Vergangenheit**
	// setzt — so kennt `net.Conn` das. Wer die Frist nur beim Eintritt läse,
	// bemerkte das nie, und ein abgebrochenes Empfangen hinge weiter.
	fristen                     sync.Mutex
	lesefrist, schreibfrist     time.Time
	lesewechsel, schreibwechsel chan struct{}

	// gewartet zählt, wie oft ein Lesen oder Schreiben ins Warten ging —
	// **hinter** der Prüfung in [schleifenende.vorbei]. Ein Lauf, der ein
	// **laufendes** Warten abbrechen will, wartet darauf statt auf eine Frist:
	// zwanzig Millisekunden reichten unter voller Last nicht, und der Abbruch
	// ging dann durch `vorbei` statt durch das `select` (gemessen am
	// 2026-09-27 in `make ci`, `Write` bei 86,7 %).
	gewartet atomic.Int64
}

// Read holt das nächste Stück — oder den Rest des vorigen.
func (e *schleifenende) Read(ziel []byte) (int, error) {
	for len(e.rest) == 0 {
		stueck, fehler := e.hole()
		if fehler != nil {
			return 0, fehler
		}

		e.rest = stueck
	}

	genommen := copy(ziel, e.rest)
	e.rest = e.rest[genommen:]

	return genommen, nil
}

// hole wartet auf das nächste Stück — und auf eine Frist, die sich ändert.
func (e *schleifenende) hole() ([]byte, error) {
	for {
		if fehler := e.vorbei(true); fehler != nil {
			return nil, fehler
		}

		if stueck, ende, gibt := e.bereit(); gibt {
			return stueck, ende
		}

		wann, wechsel := e.frist(true)

		ablauf, aufhoeren := e.wecker(wann)

		e.gewartet.Add(1)

		select {
		case stueck := <-e.herein:
			aufhoeren()

			return stueck, nil

		case <-e.zu:
			aufhoeren()

			return nil, net.ErrClosed

		case <-e.drueben:
			// **Drüben ist niedergelegt** — noch einmal von vorn: was bis
			// dahin unterwegs war, holt [schleifenende.bereit].
			aufhoeren()

		case <-ablauf:
			aufhoeren()

			return nil, os.ErrDeadlineExceeded

		case <-wechsel:
			// **Die Frist hat sich geändert** — noch einmal von vorn, mit
			// der neuen. Steht sie in der Vergangenheit, schlägt der Wecker
			// sofort.
			aufhoeren()
		}
	}
}

// Write legt ein Stück in den Puffer.
//
// **Mit einer eigenen Kopie.** Der Aufrufer darf seinen Puffer danach
// wiederverwenden; wer die Scheibe weiterreichte, schickte irgendwann etwas
// anderes, als er geschrieben hat.
func (e *schleifenende) Write(quelle []byte) (int, error) {
	stueck := make([]byte, len(quelle))
	copy(stueck, quelle)

	for {
		if fehler := e.vorbei(false); fehler != nil {
			return 0, fehler
		}

		wann, wechsel := e.frist(false)

		ablauf, aufhoeren := e.wecker(wann)

		e.gewartet.Add(1)

		select {
		case e.hinaus <- stueck:
			aufhoeren()

			return len(quelle), nil

		case <-e.zu:
			aufhoeren()

			return 0, net.ErrClosed

		case <-e.drueben:
			aufhoeren()

			return 0, net.ErrClosed

		case <-ablauf:
			aufhoeren()

			return 0, os.ErrDeadlineExceeded

		case <-wechsel:
			aufhoeren()
		}
	}
}

// vorbei sagt, ob es gar nicht erst losgeht — zu, oder Frist verstrichen.
//
// # Warum das vor dem `select` steht und nicht darin
//
// **Weil `select` unter mehreren bereiten Zweigen würfelt.** Steht im Puffer
// noch Platz **und** ist die Verbindung zu, wäre beides bereit — und die
// Hälfte der Läufe schriebe fröhlich in einen Kanal, den niemand mehr liest.
// Dasselbe gilt für eine Frist, die schon verstrichen ist: der Wecker wäre
// bereit, der Puffer auch, und es entschiede der Zufall.
//
// **Gemessen am 2026-09-16**: der Lauf über den dritten Schritt des
// Handschlags und der über die abgelaufene Frist waren mal grün, mal rot —
// ohne dass sich etwas geändert hätte. Genau so verhält sich `net.Pipe` auch:
// es prüft beides, bevor es wartet.
//
// **Drüben zu hält nur das Schreiben an.** Wer liest, bekommt erst, was noch
// unterwegs ist ([schleifenende.bereit]).
func (e *schleifenende) vorbei(lesend bool) error {
	select {
	case <-e.zu:
		return net.ErrClosed
	default:
	}

	if !lesend {
		select {
		case <-e.drueben:
			return net.ErrClosed
		default:
		}
	}

	if wann, _ := e.frist(lesend); !wann.IsZero() && !time.Now().Before(wann) {
		return os.ErrDeadlineExceeded
	}

	return nil
}

// bereit gibt, was ohne Warten geht: ein Stück aus dem Puffer, sonst das
// Ende, wenn drüben niedergelegt ist (#2504).
//
// **In dieser Reihenfolge und vor dem Warten**, wie ein Socket: was gesendet
// ist, kommt an, und danach `io.EOF`. Im `select` würfelte es, wenn beides
// bereit ist.
func (e *schleifenende) bereit() (stueck []byte, ende error, gibt bool) {
	select {
	case stueck := <-e.herein:
		return stueck, nil, true
	default:
	}

	select {
	case <-e.drueben:
		return nil, io.EOF, true
	default:
		return nil, nil, false
	}
}

// Close legt dieses Ende nieder — und zwar genau einmal.
func (e *schleifenende) Close() error {
	e.einmal.Do(func() { close(e.zu) })

	return nil
}

// LocalAddr und RemoteAddr geben dieselbe gedachte Adresse.
//
// **Sie steht da, weil [net.Conn] sie verlangt**, und sie sagt die Wahrheit:
// hier ist kein Rechner, hier ist ein Prozess.
func (e *schleifenende) LocalAddr() net.Addr  { return schleifenadresse{} }
func (e *schleifenende) RemoteAddr() net.Addr { return schleifenadresse{} }

// Die drei Fristen gelten wirklich.
//
// **Und das ist keine Höflichkeit gegenüber [net.Conn].** `internal/leitung`
// setzt sie — für die Frist eines Handschlags und für den Abbruch eines
// wartenden Empfangens. Eine Schleife, die sie stillschweigend überginge,
// zeigte genau die Schicht nicht, für die es sie gibt.
func (e *schleifenende) SetDeadline(wann time.Time) error {
	e.fristen.Lock()
	defer e.fristen.Unlock()

	e.lesefrist, e.schreibfrist = wann, wann
	e.lesewechsel = geweckt(e.lesewechsel)
	e.schreibwechsel = geweckt(e.schreibwechsel)

	return nil
}

func (e *schleifenende) SetReadDeadline(wann time.Time) error {
	e.fristen.Lock()
	defer e.fristen.Unlock()

	e.lesefrist = wann
	e.lesewechsel = geweckt(e.lesewechsel)

	return nil
}

func (e *schleifenende) SetWriteDeadline(wann time.Time) error {
	e.fristen.Lock()
	defer e.fristen.Unlock()

	e.schreibfrist = wann
	e.schreibwechsel = geweckt(e.schreibwechsel)

	return nil
}

// geweckt schlägt das alte Signal und gibt ein frisches.
func geweckt(bisher chan struct{}) chan struct{} {
	close(bisher)

	return make(chan struct{})
}

// frist gibt die gesetzte Frist und das Signal für ihre Änderung.
func (e *schleifenende) frist(lesend bool) (
	wann time.Time, wechsel <-chan struct{},
) {
	e.fristen.Lock()
	defer e.fristen.Unlock()

	if lesend {
		return e.lesefrist, e.lesewechsel
	}

	return e.schreibfrist, e.schreibwechsel
}

// wecker gibt einen Kanal, der zur Frist schlägt — und nie, wenn keine steht.
//
// **Ein `nil`-Kanal blockiert für immer**, und genau das ist im `select` der
// Fall „ohne Frist": er wird nie gewählt, und die anderen Zweige entscheiden.
func (e *schleifenende) wecker(wann time.Time) (
	ablauf <-chan time.Time, aufhoeren func(),
) {
	if wann.IsZero() {
		return nil, func() {}
	}

	uhr := time.NewTimer(time.Until(wann))

	return uhr.C, func() { uhr.Stop() }
}

// schleifenadresse ist die Adresse, die es nicht gibt.
type schleifenadresse struct{}

func (schleifenadresse) Network() string { return "schleife" }
func (schleifenadresse) String() string  { return "im selben Prozess" }

// gepuffertesPaar gibt zwei Enden, die einander hören.
func gepuffertesPaar() (hiesiges, dortiges net.Conn) {
	hin := make(chan []byte, Schleifenpuffer)
	her := make(chan []byte, Schleifenpuffer)

	// **Je Ende ein Signal, und jedes hört das des anderen** (#2504): wer
	// eines niederlegt, legt die Verbindung nieder — und was unterwegs war,
	// kommt drüben noch an, genau wie bei einem Socket.
	einer := &schleifenende{
		hinaus: hin, herein: her, zu: make(chan struct{}),
		lesewechsel: make(chan struct{}), schreibwechsel: make(chan struct{}),
	}
	anderer := &schleifenende{
		hinaus: her, herein: hin, zu: make(chan struct{}),
		lesewechsel: make(chan struct{}), schreibwechsel: make(chan struct{}),
	}
	einer.drueben, anderer.drueben = anderer.zu, einer.zu

	return einer, anderer
}
