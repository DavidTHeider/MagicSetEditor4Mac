package leitung_test

// **Was nur ein Doppel hergibt** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Wozu ein eigener Strom
//
// **`net.Pipe` ist zu gutmütig.** Sein `Close` scheitert nie, und seine
// Fristen sind von außen nicht zu sehen. Beides gehört zu den Zusagen dieser
// Schicht: dass eine klagende Leitung klagt, und dass eine Frist **je
// Richtung** gesetzt und danach **wieder weggeräumt** wird.
//
// # Und warum die Fristen hier stehen und nicht bei den Uhren
//
// **Weil ein Lauf über echte Zeit die Maschine misst.** „Es hat gewartet" und
// „es war langsam" sehen gleich aus. Das Doppel schreibt auf, **was gesetzt
// wurde** — und das ist eine Aussage über den Code, nicht über den Rechner.

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

// probestrom schreibt auf, was mit ihm geschieht.
type probestrom struct {
	schloss sync.Mutex

	// haelt lässt ein Lesen warten, bis jemand es freigibt.
	//
	// **Damit ein Abbruch überhaupt etwas zu unterbrechen hat.** Ein Lesen,
	// das sofort zurückkommt, ist längst fertig, bevor jemand abbricht.
	haelt  chan struct{}
	liest  chan struct{}
	einmal sync.Once

	schliessfehler error
	geschlossen    bool

	// lesefristen und schreibfristen halten jede gesetzte Frist fest.
	lesefristen    []time.Time
	schreibfristen []time.Time
}

func (p *probestrom) Read([]byte) (int, error) {
	// **Ohne `haelt` ist der Strom sofort leer** — das brauchen die Läufe,
	// die nur die Fristen ansehen.
	if p.haelt == nil {
		return 0, errors.New("leer")
	}

	p.einmal.Do(func() { close(p.liest) })

	<-p.haelt

	return 0, errors.New("freigegeben")
}
func (p *probestrom) Write(b []byte) (int, error) { return len(b), nil }

func (p *probestrom) Close() error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.geschlossen = true

	return p.schliessfehler
}

func (p *probestrom) LocalAddr() net.Addr  { return nil }
func (p *probestrom) RemoteAddr() net.Addr { return nil }

func (p *probestrom) SetDeadline(time.Time) error {
	return errors.New("diese Leitung kennt keine gemeinsame Frist")
}

func (p *probestrom) SetReadDeadline(wann time.Time) error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.lesefristen = append(p.lesefristen, wann)

	return nil
}

func (p *probestrom) SetWriteDeadline(wann time.Time) error {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	p.schreibfristen = append(p.schreibfristen, wann)

	return nil
}

func (p *probestrom) gesetzt() (lesend, schreibend []time.Time) {
	p.schloss.Lock()
	defer p.schloss.Unlock()

	return append([]time.Time{}, p.lesefristen...),
		append([]time.Time{}, p.schreibfristen...)
}

// **Eine Leitung, die sich beim Niederlegen beschwert, wird gemeldet.**
func TestEineKlagendeLeitungWirdGemeldet(t *testing.T) {
	t.Parallel()

	strom := &probestrom{schliessfehler: errors.New("die Leitung klagt")}
	verbindung := leitung.Ueber(strom, "kanal-probe")

	if fehler := verbindung.Schliesse(); fehler == nil {
		t.Fatal("die klagende Leitung schwieg")
	}

	if !strom.geschlossen {
		t.Error("sie wurde gar nicht erst geschlossen")
	}

	// **Und ein zweites Schließen klagt nicht mehr** — das Ziel ist
	// erreicht, und der Strom wird kein zweites Mal angefasst.
	if fehler := verbindung.Schliesse(); fehler != nil {
		t.Fatalf("das zweite Schließen klagte: %v", fehler)
	}
}

// **Eine Frist gilt einer Richtung — und wird danach weggeräumt.**
//
// # Der Fund
//
// **`SetDeadline` setzt beide Richtungen.** Der erste Entwurf nahm es, und
// damit hätte ein Senden einem gleichzeitigen Empfangen die Frist gebrochen.
// Dieses Doppel weist `SetDeadline` deshalb **ab**: wer es doch nähme, fiele
// hier auf.
//
// **Und ohne das Wegräumen bliebe sie stehen.** Der nächste Ruf ohne Frist
// scheiterte dann an einer Uhr, die ihn nichts angeht — sichtbar nur unter
// Last, und das ist die Sorte Fehler, die man nicht nachstellt.
func TestEineFristGiltEinerRichtungUndWirdWeggeraeumt(t *testing.T) {
	t.Parallel()

	strom := &probestrom{}
	verbindung := leitung.Ueber(strom, "kanal-probe")

	wann := time.Now().Add(time.Hour)

	ctx, abbrechen := context.WithDeadline(t.Context(), wann)
	defer abbrechen()

	if fehler := verbindung.Sende(ctx, []byte("hallo")); fehler != nil {
		t.Fatalf("Sende: %v", fehler)
	}

	lesend, schreibend := strom.gesetzt()

	if len(lesend) != 0 {
		t.Errorf("das Senden hat eine Lesefrist gesetzt: %v", lesend)
	}

	// **Zweimal: die Frist und ihr Wegräumen.** Ein Nullzeitpunkt heißt
	// „keine Frist" — so schreibt `net.Conn` das.
	if len(schreibend) != 2 || !schreibend[0].Equal(wann) ||
		!schreibend[1].IsZero() {
		t.Fatalf("gesetzt wurde %v", schreibend)
	}
}

// **Ein Abbruch setzt die Frist in die Vergangenheit.**
//
// **Das ist der Weg, den `net.Conn` kennt**: es gibt keinen Abbruch, nur eine
// Frist, die schon um ist. Ohne ihn hinge ein Empfangen an einem Gegenüber,
// das nichts mehr schickt — und §16 verlangt für jede Anfrage beides.
//
// # Warum das Lesen dafür warten muss
//
// **Weil ein Abbruch nur unterbricht, was läuft.** Der erste Anlauf dieses
// Laufs brach **nach** dem Ruf ab — da ist der Wächter längst abgestellt, und
// nichts geschah. Er war rot, und er hatte recht.
func TestEinAbbruchSetztDieFristInDieVergangenheit(t *testing.T) {
	t.Parallel()

	strom := &probestrom{haelt: make(chan struct{}), liest: make(chan struct{})}
	verbindung := leitung.Ueber(strom, "kanal-probe")

	ctx, abbrechen := context.WithCancel(t.Context())

	fertig := make(chan struct{})

	go func() {
		defer close(fertig)

		if _, fehler := verbindung.Empfange(ctx); fehler == nil {
			t.Error("das Empfangen kam ohne Fehler zurück")
		}
	}()

	<-strom.liest

	fruehestens := time.Now()

	abbrechen()

	vergangen := false

	for range 200 {
		lesend, _ := strom.gesetzt()

		for _, wann := range lesend {
			if !wann.IsZero() && wann.Before(fruehestens) {
				vergangen = true
			}
		}

		if vergangen {
			break
		}

		time.Sleep(5 * time.Millisecond)
	}

	close(strom.haelt)
	<-fertig

	if !vergangen {
		t.Fatal("der Abbruch hat keine vergangene Frist gesetzt")
	}
}

// **Und ein Abbruch danach fasst den Strom nicht mehr an.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Sonst brächte ein alter Kontext einem neuen Ruf die Frist.** Der Wächter
// gehört dem Ruf, der ihn gesetzt hat; wer ihn stehen ließe, hätte einen
// Abbruch, der später jemand anderen trifft — sichtbar nur unter Last.
func TestEinAbbruchDanachFasstDenStromNichtMehrAn(t *testing.T) {
	t.Parallel()

	strom := &probestrom{}
	verbindung := leitung.Ueber(strom, "kanal-probe")

	ctx, abbrechen := context.WithCancel(t.Context())

	if _, fehler := verbindung.Empfange(ctx); fehler == nil {
		t.Fatal("das Empfangen kam ohne Fehler zurück")
	}

	vorher, _ := strom.gesetzt()

	abbrechen()

	// **Ein wenig Zeit für einen Wächter, den es nicht geben soll.**
	time.Sleep(50 * time.Millisecond)

	nachher, _ := strom.gesetzt()

	if len(nachher) != len(vorher) {
		t.Fatalf("nach dem Abbruch kam eine Frist dazu: %v", nachher)
	}
}
