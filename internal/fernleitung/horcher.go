package fernleitung

// **Die horchende Seite** — sie gibt die Einladung aus.

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Schlangenlaenge ist, wie viele hereinkommende Ströme warten dürfen.
//
// **Klein, und das ist Absicht.** Wer einlädt, lädt einen ein; eine lange
// Schlange wäre eine Einladung an jeden, der die Adresse hat, und der
// Handschlag käme erst danach. Was darüber hinaus anklopft, wird geschlossen.
const Schlangenlaenge = 4

// Horcher nimmt Verbindungen aus fremden Netzen entgegen.
//
// **Er sieht aus wie [leitung.Horcher] und ist es nicht** — er hat keine
// Adresse, sondern eine **Einladung**, und die trägt die Schlüssel. Gemeinsam
// ist beiden, was hinter dem Strom liegt.
type Horcher struct {
	server  fernserver
	ich     peer.Identitaet
	stroeme chan net.Conn

	einmal sync.Once
	zu     chan struct{}

	// schlange ordnet Einreihen und Niederlegen: wer nach dem Niederlegen
	// einreiht, findet `geschlossen` gesetzt und die Schlange geleert.
	schlange    sync.Mutex
	geschlossen bool

	// angebunden zählt die offenen Ströme — SEC-21 (#2162), Grund bei
	// [Horcher.Angebunden].
	angebunden atomic.Int64
}

// Horche macht dieses Profil über fremde Netze erreichbar.
//
// **Nur auf Aufruf**, wie in N2: wer nicht ruft, ist nicht erreichbar.
func Horche(ich peer.Identitaet) (*Horcher, error) {
	horcher := &Horcher{
		ich:     ich,
		stroeme: make(chan net.Conn, Schlangenlaenge),
		zu:      make(chan struct{}),
	}

	horcher.server = serverBauen(horcher.annehmen)

	if fehler := horcher.server.Start(); fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"über fremde Netze erreichbar werden")
	}

	return horcher, nil
}

// annehmen reiht einen hereinkommenden Strom ein.
//
// **Nicht blockierend**: tailcat ruft aus seiner eigenen Goroutine, und ein
// Warten hier hielte den Tunnel auf. Was nicht hineinpasst, wird geschlossen —
// und das ist besser als ein Gegenüber, das ewig wartet.
func (h *Horcher) annehmen(strom net.Conn) {
	h.schlange.Lock()
	defer h.schlange.Unlock()

	if h.geschlossen {
		_ = strom.Close() //nolint:errcheck // der Horcher ist zu

		return
	}

	select {
	case h.stroeme <- strom:
	default:
		_ = strom.Close() //nolint:errcheck // die Schlange ist voll
	}
}

// Einladung ist, was ein Gegenüber zum Wählen braucht.
//
// **Mit Fassung**, und der Grund steht bei [domain.NeueEinladung].
func (h *Horcher) Einladung() domain.Einladung {
	return domain.NeueEinladung(h.server.Einladungsadresse())
}

// Adresse ist, wo dieser Horcher steht — und das **ist** die Einladung.
//
// **Sie erfüllt damit [service.Hereinkommend]**, und das ist keine
// Verrenkung: die Schnittstelle fragt, *wo bin ich erreichbar*, und über ein
// fremdes Netz lautet die Antwort nicht "Wirt und Port", sondern
// "diese Zeichenkette". [Horcher.Einladung] gibt dasselbe getippt zurück.
func (h *Horcher) Adresse() string { return string(h.Einladung()) }

// Wege sagt, wie die Verbindungen gerade gehen — direkt oder über ein
// Relais (NET-02.AK-03, #1706).
func (h *Horcher) Wege() []Fernweg { return h.server.Wege() }

// Angebunden sagt, wie viele Ströme offen sind — von der Übernahme durch
// [Horcher.Nimm] bis zu ihrem Ende, wer auch immer es herbeiführt.
//
// **Gezählt wird, was jemand beobachtet** (Nachlauf über Phase H): ein Ende
// sieht dieser Horcher am Schließen oder an einem Lesen oder Schreiben, das es
// meldet — die Leitung liest ständig, also sieht sie es. Ein Strom, der noch
// in der Schlange wartet, zählt deshalb nicht: stürbe er dort, merkte es
// niemand.
//
// **Gezählt wird hier, weil tailcat Gegenstellen nennt und nicht Ströme.**
// Bis v0.6.0 nannte es gar nichts: `locoBackend.Status` baute den
// `ipnstate.StatusBuilder` ohne `WantPeers`, und [Horcher.Wege] blieb auch bei
// offener Verbindung leer (SEC-21,
// #2162).
// Seit v0.7.0 setzt es `WantPeers` (`tailcat.go`, `locoBackend.Status`,
// nachgelesen am 2026-10-07) — K01
// (#2187).
// **Ein Eintrag dort ist eine Gegenstelle, die magicsock kennt**, kein offener
// Strom: zwischen Anbinden und erstem Weg, und nachdem ein Strom endete, sagen
// beide Zahlen Verschiedenes. **Dass** ein Strom offen ist, weiß nur dieser
// Horcher.
func (h *Horcher) Angebunden() int { return int(h.angebunden.Load()) }

// zaehlstrom zählt beim ersten Ende einmal ab — eine Leitung schließt ihren
// Strom, und ein Aufräumen danach darf die Zahl nicht unter null ziehen.
//
// **Ein Ende ist auch, was darunter geschieht** — angestrichen vom Zweitblick
// über Phase H: tailcat schließt den Strom unter dieser Hülle, wenn es seinen
// Netzstapel niederlegt, und das Gegenüber legt auf, ohne dass hier jemand
// `Close` ruft. Was der Leser oder Schreiber dann bekommt, ist das Ende;
// **eine abgelaufene Frist ist keines**.
type zaehlstrom struct {
	net.Conn

	zahl   *atomic.Int64
	einmal sync.Once
}

func (z *zaehlstrom) Read(puffer []byte) (int, error) {
	n, fehler := z.Conn.Read(puffer)
	z.endetMit(fehler)

	return n, fehler
}

func (z *zaehlstrom) Write(puffer []byte) (int, error) {
	n, fehler := z.Conn.Write(puffer)
	z.endetMit(fehler)

	return n, fehler
}

func (z *zaehlstrom) Close() error {
	z.abzaehlen()

	return z.Conn.Close()
}

// endetMit zählt ab, wenn `fehler` das Ende des Stroms ist.
func (z *zaehlstrom) endetMit(fehler error) {
	var netzfehler net.Error
	if fehler == nil || (errors.As(fehler, &netzfehler) && netzfehler.Timeout()) {
		return
	}

	z.abzaehlen()
}

func (z *zaehlstrom) abzaehlen() { z.einmal.Do(func() { z.zahl.Add(-1) }) }

// anbinden zählt einen übernommenen Strom und gibt ihn in seiner Hülle.
func (h *Horcher) anbinden(roh net.Conn) net.Conn {
	h.angebunden.Add(1)

	return &zaehlstrom{Conn: roh, zahl: &h.angebunden}
}

// Nimm wartet auf die nächste Verbindung.
//
// **Und hier bricht der Kontext wirklich ab** — anders als bei
// [leitung.Horcher.Nimm], wo `Accept` keinen kennt. Der Unterschied steht
// hier, damit niemand ihn für einen Zufall hält: tailcat ruft uns, wir rufen
// nicht tailcat.
func (h *Horcher) Nimm(ctx context.Context) (*leitung.Leitung, error) {
	select {
	case roh := <-h.stroeme:
		strom := h.anbinden(roh)
		genommen, fehler := leitung.NimmStrom(ctx, strom, h.ich)
		if fehler != nil {
			// **Ein gescheiterter Handschlag bindet niemanden an.** Ein
			// Handschlag, der scheitert, schließt seinen Strom schon selbst
			// (`leitung/tls.go`); ein Zeugnis, das sich nicht ausstellen ließ,
			// nicht — und dann hielte der Strom die Zahl fest. Zweimal
			// schließen zählt einmal ab ([zaehlstrom]).
			_ = strom.Close() //nolint:errcheck // der Handschlag ist der Befund
		}

		return genommen, fehler

	case <-h.zu:
		return nil, domain.MitGrund(domain.ErrNotFound, context.Canceled,
			"der Horcher ist niedergelegt")

	case <-ctx.Done():
		return nil, domain.MitGrund(domain.ErrNotFound, ctx.Err(),
			"es kam keine Verbindung herein")
	}
}

// Schliesse legt den Horcher nieder.
//
// **Zweimal rufen ist erlaubt** — ein Lebenszyklus, der beim zweiten Mal
// klagt, zwingt jeden Aufrufer zu einem Merker.
func (h *Horcher) Schliesse() error {
	var fehler error

	h.einmal.Do(func() {
		close(h.zu)
		h.leeren()

		if schliessfehler := h.server.Close(); schliessfehler != nil {
			fehler = domain.MitGrund(domain.ErrValidation, schliessfehler,
				"den Fernhorcher niederlegen")
		}
	})

	return fehler
}

// leeren schließt, was noch in der Schlange wartet — **nach dem Niederlegen
// nimmt niemand mehr**, und ein Strom dort hielte die Zahl fest.
func (h *Horcher) leeren() {
	h.schlange.Lock()
	defer h.schlange.Unlock()

	h.geschlossen = true

	for {
		select {
		case strom := <-h.stroeme:
			_ = strom.Close() //nolint:errcheck // der Horcher ist zu
		default:
			return
		}
	}
}
