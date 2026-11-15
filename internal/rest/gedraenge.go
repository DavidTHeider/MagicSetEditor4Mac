package rest

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Vorfrist ist, wie lange eine neue Verbindung schweigen darf, bevor sie
// ihren Platz für einen Wartenden hergibt — SEC-14
// (#2143).
//
// **Gemessen am 2026-10-07** (`vorabschweiger_test.go`, SEC-13): hinter so
// vielen Schweigern, wie es Plätze gibt, wartete ein rechtmäßiger Aufrufer
// bis zur Lesefrist — beim Dienst zehn Sekunden, und wer nachfüllte, ließ ihn
// jedes Mal so lange warten. **Zwei Sekunden**, weil ein Aufrufer seinen
// Kopf im nächsten Augenblick nach dem Verbinden schickt, auch hinter TLS;
// und weil der Preis nur eine Verbindung trifft, die so lange **nichts**
// gesagt hat, **während** jemand wartet.
const Vorfrist = 2 * time.Second

// HoechstensGleichzeitig gibt einen Lauscher, der nicht mehr als `grenze`
// Verbindungen zugleich annimmt.
//
// # Warum die Drossel dafür nicht reicht
//
// Sie zählt Anfragen, und eine Anfrage entsteht erst, wenn ein Behandler
// läuft. Eine **halb geschriebene** Anfrage erreicht sie nie — ein lokaler
// Prozess konnte damit beliebig viele Verbindungen aufmachen und je zehn
// Sekunden halten. Auf macOS ist die Vorgabe für offene Dateien 256 je
// Prozess; danach nimmt der Dienst überhaupt nichts mehr an (Zweitblick über
// Phase U).
//
// **Die Frist begrenzt, wie lange eine Verbindung hängt, nicht wie viele.**
//
// # Was das nicht tut
//
// **Es weist nicht ab, es wartet.** Eine überzählige Verbindung wird
// angenommen und wartet, bis ein Platz frei wird; alle weiteren bleiben im
// Rückstau des Betriebssystems — genauso, wie `golang.org/x/net/netutil` es
// macht. Ein Abweisen bräuchte ein Annehmen und sofortiges Schließen, und das
// kostet mehr, als es hilft: der Aufrufer sähe eine geschlossene Verbindung
// statt einer langsamen.
//
// **Es begrenzt keine Arbeit, sondern Verbindungen.** Wer mit einer Verbindung
// nacheinander teure Aufträge schickt, wird von der Drossel gezählt, nicht
// hier.
//
// # Und eine ruhende Verbindung hält keinen Platz, den jemand braucht
//
// **Gemessen am 2026-10-02** (QA-08, #1739, Profil `ruhend` des Lastlaufs):
// ein Klient mit sechzehn Verbindungen im Pool fragte je einmal und ließ sie
// offen — wie jeder Pool mit `keepAlive`. Danach nahm der Dienst **keine**
// Verbindung mehr an, bis der Leerlauf von zwei Minuten sie schloss; die
// Lesefrist griff nicht, denn ein Kopf kam ja.
//
// **Seitdem gilt: wartet eine Verbindung auf einen Platz, geht eine ruhende**
// — die am längsten ruhende, oder die nächste, die zu ruhen beginnt. **Und
// nur dann**: die erste Fassung schloss bei vollem Haus jede Verbindung, die
// auch nur zwischen zwei Rufen ruhte, und der zweite Lastlauf sah Tausende
// abgerissene Rufe eines Klienten, der genau so viele Verbindungen hielt, wie
// es Plätze gibt. Deshalb wird **erst angenommen** und dann nach dem Platz
// gefragt — so ist sicher, dass wirklich jemand wartet. Dafür muss der Server
// [Gedraenge.Zustand] als `ConnState` bekommen; ohne ihn bleibt es beim
// Warten.

func HoechstensGleichzeitig(darunter net.Listener, grenze int) *Gedraenge {
	return &Gedraenge{
		Listener: darunter,
		plaetze:  make(chan struct{}, grenze),
		ruhend:   make(map[*platzhalter]uint64),
		neu:      make(map[*platzhalter]time.Time),
		vorfrist: Vorfrist,
		zu:       make(chan struct{}),
	}
}

// Gedraenge hält die Zahl der offenen Verbindungen.
type Gedraenge struct {
	net.Listener

	plaetze chan struct{}

	// ruhend sind die Verbindungen, die gerade auf die nächste Anfrage
	// warten, mit der Folge, in der sie zu ruhen begannen; wartend ist die
	// Zahl der angenommenen Verbindungen, die auf einen Platz warten.
	sperre  sync.Mutex
	ruhend  map[*platzhalter]uint64
	folge   uint64
	wartend int

	// neu sind die angenommenen Verbindungen, die noch keine Anfrage
	// geschickt haben, mit dem Augenblick ihrer Annahme (SEC-14).
	neu map[*platzhalter]time.Time

	// vorfrist ist, wie lange eine neue Verbindung schweigen darf, bevor sie
	// ihren Platz für einen Wartenden hergibt (SEC-14) — null heißt nie.
	vorfrist time.Duration

	// endet ist gesetzt, sobald [Gedraenge.StummeSchliessen] lief: wer danach
	// schweigt oder zu ruhen beginnt, geht sofort.
	endet bool

	// zu schließt mit [Gedraenge.Close] und beendet das Warten auf einen
	// Platz (Zweitblick über E, Befund 2).
	zu      chan struct{}
	schluss sync.Once
}

// Close schließt den Lauscher — **und** jeden, der auf einen Platz wartet.
//
// **Vom eingebetteten Lauscher geerbt schloss es nur ihn**: ein schon
// angenommener Aufrufer wartete weiter, und wurde ein Platz frei, gab
// `Accept` ihn nach dem Schluss noch heraus.
func (g *Gedraenge) Close() error {
	g.schluss.Do(func() { close(g.zu) })

	return g.Listener.Close() //nolint:wrapcheck // der Aufrufer erwartet den Fehler des Lauschers
}

// StummeSchliessen schließt beim Dienstende jede Verbindung, die gerade nichts
// sagt — die noch keine Anfrage geschickt hat und die ruhenden.
//
// **Weil `http.Server.Shutdown` auf eine neue Verbindung fünf Sekunden
// wartet**: es hält `StateNew` erst danach für ruhend (`net/http/server.go`,
// Zeilen 3310–3314, Go 1.27.1, Issue 22682). Ein Klient, der eine Verbindung
// vorwählt und nicht braucht, hielt damit die ganze Frist des Dienstendes
// fest. Eine Verbindung, die arbeitet, bleibt — ihre Antwort soll noch
// hinausgehen.
//
// **Und gilt fort, nicht nur für den Augenblick** — angestrichen vom
// Zweitblick über Phase H: Go trägt eine Verbindung in seine Liste ein,
// bevor es `ConnState` ruft (`net/http/server.go`, Zeile 1865); wessen
// Meldung erst danach kam, entging einer bloßen Momentaufnahme. Danach geht
// in [Gedraenge.Zustand], wer schweigt oder zu ruhen beginnt.
func (g *Gedraenge) StummeSchliessen() {
	g.sperre.Lock()
	g.endet = true
	stumme := make([]*platzhalter, 0, len(g.neu)+len(g.ruhend))
	for halter := range g.neu {
		stumme = append(stumme, halter)
	}

	for halter := range g.ruhend {
		stumme = append(stumme, halter)
	}
	g.sperre.Unlock()

	for _, halter := range stumme {
		_ = halter.Close() //nolint:errcheck // wer nichts sagt, hat nichts zu verlieren
	}
}

// MitVorfrist setzt die Vorfrist.
func (g *Gedraenge) MitVorfrist(vorfrist time.Duration) *Gedraenge {
	g.vorfrist = vorfrist

	return g
}

// Zustand folgt `http.Server.ConnState`: eine Verbindung, die zu ruhen
// beginnt, während jemand auf einen Platz wartet, wird geschlossen; sonst
// wird sie gemerkt. Eine, die arbeitet, bleibt.
func (g *Gedraenge) Zustand(verbindung net.Conn, zustand http.ConnState) {
	halter, unsere := verbindung.(*platzhalter)
	if !unsere {
		return
	}

	g.sperre.Lock()

	// **Eine neue Verbindung wird mit ihrer Zeit gemerkt** (SEC-14) — bis
	// sie etwas sagt.
	if zustand == http.StateNew {
		g.neu[halter] = time.Now()
	} else {
		delete(g.neu, halter)
	}

	if g.endet && (zustand == http.StateNew || zustand == http.StateIdle) {
		delete(g.neu, halter)
		delete(g.ruhend, halter)
		g.sperre.Unlock()
		_ = halter.Close() //nolint:errcheck // wer beim Dienstende nichts sagt, hat nichts zu verlieren

		return
	}

	if zustand != http.StateIdle {
		delete(g.ruhend, halter)
		g.sperre.Unlock()

		return
	}

	if g.wartend > 0 {
		g.sperre.Unlock()
		_ = halter.Close() //nolint:errcheck // eine ruhende Verbindung hat nichts mehr zu sagen

		return
	}

	g.folge++
	g.ruhend[halter] = g.folge
	g.sperre.Unlock()
}

// Accept nimmt eine Verbindung an und gibt sie heraus, sobald sie einen Platz
// hat.
func (g *Gedraenge) Accept() (net.Conn, error) {
	verbindung, fehler := g.Listener.Accept()
	if fehler != nil {
		return nil, fehler //nolint:wrapcheck // der Aufrufer erwartet den Fehler des Lauschers
	}

	if !g.platzBelegen() {
		_ = verbindung.Close() //nolint:errcheck // der Schluss ist die Antwort

		return nil, net.ErrClosed
	}

	// **Der Schluss gewinnt auch gegen einen freien Platz** (Nachlauf über
	// E): eine Verbindung, die schon angenommen war, als `Close` kam, bekam
	// sonst ihren Platz — über jeden der drei Wege, frei, gewartet oder
	// verdrängt. Geprüft wird deshalb hier, hinter allen dreien.
	select {
	case <-g.zu:
		<-g.plaetze
		_ = verbindung.Close() //nolint:errcheck // der Schluss ist die Antwort

		return nil, net.ErrClosed
	default:
	}

	return &platzhalter{Conn: verbindung, frei: func() { <-g.plaetze }}, nil
}

// platzBelegen nimmt einen Platz — und macht einen frei, wenn eine
// Verbindung nur ruht. Ruht keine, wartet es, und die nächste, die zu ruhen
// beginnt, geht ([Gedraenge.Zustand]). Falsch heißt: geschlossen.
func (g *Gedraenge) platzBelegen() bool {
	for {
		select {
		case g.plaetze <- struct{}{}:
			return true
		default:
		}

		g.sperre.Lock()

		weichen := g.aeltesteRuhende()
		if weichen == nil {
			weichen = g.aeltesterSchweiger(time.Now())
		}

		if weichen == nil {
			g.wartend++
		}

		g.sperre.Unlock()

		if weichen != nil {
			_ = weichen.Close() //nolint:errcheck // eine ruhende oder schweigende Verbindung hat nichts zu sagen

			g.plaetze <- struct{}{}

			return true
		}

		switch g.warte() {
		case platzBelegt:
			return true
		case geschlossen:
			return false
		case nochmal:
		}
	}
}

// Wie ein Warten ausgeht.
type wartausgang int

const (
	nochmal wartausgang = iota
	platzBelegt
	geschlossen
)

// warte wartet auf einen freien Platz — höchstens eine Vorfrist lang; danach
// kann eine Verbindung schweigend über ihr stehen und weichen (SEC-14).
//
// **Ohne Vorfrist wartet sie, bis ein Platz frei wird**, wie vor SEC-14.
func (g *Gedraenge) warte() wartausgang {
	var frist <-chan time.Time

	if g.vorfrist > 0 {
		uhr := time.NewTimer(g.vorfrist)
		defer uhr.Stop()

		frist = uhr.C
	}

	ausgang := nochmal

	select {
	case g.plaetze <- struct{}{}:
		ausgang = platzBelegt
	case <-g.zu:
		ausgang = geschlossen
	case <-frist:
	}

	g.sperre.Lock()
	g.wartend--
	g.sperre.Unlock()

	return ausgang
}

// aeltesterSchweiger nimmt die neue Verbindung, die am längsten nichts gesagt
// hat — wenn es länger ist als die Vorfrist. **Nur dann**: ein Aufrufer, der
// eben verbunden hat, schickt seinen Kopf im nächsten Augenblick, und ihn zu
// schließen hieße, einen rechtmäßigen gegen einen anderen zu tauschen.
func (g *Gedraenge) aeltesterSchweiger(jetzt time.Time) *platzhalter {
	if g.vorfrist <= 0 {
		return nil
	}

	var (
		aeltester *platzhalter
		seit      = jetzt
	)

	for halter, angenommen := range g.neu {
		if jetzt.Sub(angenommen) >= g.vorfrist && angenommen.Before(seit) {
			aeltester, seit = halter, angenommen
		}
	}

	delete(g.neu, aeltester)

	return aeltester
}

// platzhalter gibt seinen Platz zurück, wenn die Verbindung endet.
type platzhalter struct {
	net.Conn

	frei   func()
	einmal sync.Once
}

// Close schließt die Verbindung und gibt den Platz frei.
//
// **Über `sync.Once`**: `net/http` schließt eine Verbindung durchaus zweimal,
// und ein zweites Freigeben nähme einem anderen den Platz weg.
func (p *platzhalter) Close() error {
	p.einmal.Do(p.frei)

	return p.Conn.Close() //nolint:wrapcheck // der Aufrufer erwartet den Fehler der Verbindung
}

// aeltesteRuhende nimmt die am längsten ruhende Verbindung aus der Reihe —
// `nil`, wenn keine ruht. Der Aufrufer hält die Sperre.
func (g *Gedraenge) aeltesteRuhende() *platzhalter {
	var (
		aelteste *platzhalter
		kleinste = ^uint64(0)
	)

	for halter, folge := range g.ruhend {
		if folge < kleinste {
			aelteste, kleinste = halter, folge
		}
	}

	delete(g.ruhend, aelteste)

	return aelteste
}
