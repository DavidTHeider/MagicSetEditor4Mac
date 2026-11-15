package service_test

// **Was die Tiefe einer Sitzung zusagt** — Befund 2 des Nachlaufs über Phase
// 6.2.
//
// Die Tiefe ist die Zahl der Nachrichten, die warten dürfen, und **die eine,
// die gerade hinausgeht, zählt mit**: sie ist angenommen, und scheitert sie,
// steht sie wieder an. Seit dem Zweitblick nahm `Leere` sie vor dem Schicken
// aus der Schlange — ein `Sende` währenddessen belegte ihren Platz, und nach
// einem Fehler standen mehr Nachrichten an, als die Tiefe erlaubt.
//
// **Und eine angenommene zuverlässige Nachricht fällt nie weg.** Bis hierher
// verdrängte eine verwerfbare bei voller Schlange die älteste — gleich
// welcher Art. „Zuverlässig heißt: der Absender erfährt es"; eine Nachricht,
// die angenommen und dann still verdrängt wird, hat ihr niemand gesagt.

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestWasGeradeHinausgehtZaehltZurTiefe(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	sperre := &sperrbar{}
	paar.hiesigerWaehler.umhuellen = func(leitung service.Verbindung) service.Verbindung {
		sperre.Verbindung = leitung

		return sperre
	}

	hier, _ := paar.eineSitzung(t, domain.ZweckPartie)

	if fehler := hier.Sende(t.Context(), []byte("A"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende A: %v", fehler)
	}

	// **A geht hinaus und hängt** — bis der Lauf es loslässt.
	haengt, loslassen := sperre.anhalten()
	geleert := make(chan error, 1)

	go func() { geleert <- hier.Leere(t.Context()) }()

	<-haengt

	// **Die Tiefe ist zwei**: A und B passen, C nicht.
	if fehler := hier.Sende(t.Context(), []byte("B"), service.Zuverlaessig); fehler != nil {
		t.Fatalf("Sende B: %v", fehler)
	}

	if fehler := hier.Sende(
		t.Context(), []byte("C"), service.Zuverlaessig,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("C kam hinein, obwohl A noch unterwegs ist: %v", fehler)
	}

	// **A scheitert.** Danach stehen A und B an, nicht mehr: A zählte die
	// ganze Zeit mit.
	loslassen()

	if fehler := <-geleert; fehler == nil {
		t.Fatal("A ging hinaus, obwohl die Leitung riss")
	}

	if wartend := hier.Stand().Wartend; wartend != 2 {
		t.Errorf("nach dem Scheitern warten %d — die Tiefe ist zwei", wartend)
	}
}

func TestEineVerwerfbareVerdraengtNieEineZuverlaessige(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	hier, _ := paar.eineSitzung(t, domain.ZweckPartie)

	for _, was := range []string{"eins", "zwei"} {
		if fehler := hier.Sende(t.Context(), []byte(was), service.Zuverlaessig); fehler != nil {
			t.Fatalf("Sende %q: %v", was, fehler)
		}
	}

	// **Die Schlange ist voll, und in ihr steht nichts Verwerfbares** — dann
	// ist die neue verwerfbare die, die weicht.
	if fehler := hier.Sende(t.Context(), []byte("praesenz"), service.Verwerfbar); fehler != nil {
		t.Fatalf("Sende verwerfbar: %v", fehler)
	}

	angekommen := zweiEmpfangen(t, paar)

	if fehler := hier.Leere(t.Context()); fehler != nil {
		t.Fatalf("Leere: %v", fehler)
	}

	if erste, zweite := <-angekommen, <-angekommen; erste != "eins" || zweite != "zwei" {
		t.Errorf("es kamen %q und %q — eine zuverlässige wurde verdrängt", erste, zweite)
	}

	if hier.Stand().Verworfen != 1 {
		t.Errorf("der Stand ist %+v", hier.Stand())
	}
}

func TestEineVerwerfbareVerdraengtDieAeltesteVerwerfbare(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	hier, _ := paar.eineSitzung(t, domain.ZweckPartie)

	for _, was := range []string{"eins", "p1", "p2"} {
		art := service.Verwerfbar
		if was == "eins" {
			art = service.Zuverlaessig
		}

		if fehler := hier.Sende(t.Context(), []byte(was), art); fehler != nil {
			t.Fatalf("Sende %q: %v", was, fehler)
		}
	}

	angekommen := zweiEmpfangen(t, paar)

	if fehler := hier.Leere(t.Context()); fehler != nil {
		t.Fatalf("Leere: %v", fehler)
	}

	// **„eins" bleibt, `p1` weicht `p2`** — der neuere Stand ersetzt den
	// älteren derselben Art.
	if erste, zweite := <-angekommen, <-angekommen; erste != "eins" || zweite != "p2" {
		t.Errorf("es kamen %q und %q", erste, zweite)
	}
}

// **Und die verwerfbare, die gerade hinausgeht, weicht nicht** — sonst
// nähme `Leere` nach dem Schicken eine andere aus der Schlange als die, die
// ankam.
func TestDieVerwerfbareUnterwegsWeichtNicht(t *testing.T) {
	t.Parallel()

	paar := zweiSeiten(t, true)
	sperre := &sperrbar{}
	paar.hiesigerWaehler.umhuellen = func(leitung service.Verbindung) service.Verbindung {
		sperre.Verbindung = leitung

		return sperre
	}

	hier, _ := paar.eineSitzung(t, domain.ZweckPartie)

	if fehler := hier.Sende(t.Context(), []byte("A"), service.Verwerfbar); fehler != nil {
		t.Fatalf("Sende A: %v", fehler)
	}

	haengt, loslassen := sperre.anhalten()
	geleert := make(chan error, 1)

	go func() { geleert <- hier.Leere(t.Context()) }()

	<-haengt

	// **B füllt die Schlange, C verdrängt B** — nicht A, die unterwegs ist.
	for _, was := range []string{"B", "C"} {
		if fehler := hier.Sende(t.Context(), []byte(was), service.Verwerfbar); fehler != nil {
			t.Fatalf("Sende %s: %v", was, fehler)
		}
	}

	loslassen()

	if fehler := <-geleert; fehler == nil {
		t.Fatal("A ging hinaus, obwohl die Leitung riss")
	}

	// **Jetzt geht es durch**: A steht wieder vorn, dahinter C.
	sperre.durchlassen()

	angekommen := zweiEmpfangen(t, paar)

	if fehler := hier.Leere(t.Context()); fehler != nil {
		t.Fatalf("Leere: %v", fehler)
	}

	if erste, zweite := <-angekommen, <-angekommen; erste != "A" || zweite != "C" {
		t.Errorf("es kamen %q und %q", erste, zweite)
	}
}

// zweiEmpfangen liest nebenher zwei Nachrichten am anderen Ende.
func zweiEmpfangen(t *testing.T, paar *paarung) <-chan string {
	t.Helper()

	angekommen := make(chan string, 2)

	go func() {
		for range 2 {
			gelesen, fehler := paar.dortigeLeitung.Empfange(t.Context())
			if fehler != nil {
				t.Errorf("Empfange: %v", fehler)

				return
			}

			angekommen <- string(gelesen)
		}
	}()

	return angekommen
}

// sperrbar hält ein Senden an, sobald man es verlangt, und lässt es dann
// scheitern.
//
// **Weil die Schleife der Tests gepuffert ist**: ein Senden dort kehrt
// sofort zurück, und „gerade unterwegs" gäbe es nicht. Bis zum Anhalten lässt
// sie alles durch — der Handschlag läuft über dieselbe Verbindung.
type sperrbar struct {
	service.Verbindung

	schloss sync.Mutex
	halt    chan struct{}
	haengt  chan struct{}
}

func (s *sperrbar) anhalten() (haengt <-chan struct{}, loslassen func()) {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	s.halt = make(chan struct{})
	s.haengt = make(chan struct{}, 1)
	halt := s.halt

	return s.haengt, func() { close(halt) }
}

func (s *sperrbar) durchlassen() {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	s.halt = nil
}

func (s *sperrbar) Sende(ctx context.Context, nachricht []byte) error {
	s.schloss.Lock()
	halt, haengt := s.halt, s.haengt
	s.schloss.Unlock()

	if halt == nil {
		return s.Verbindung.Sende(ctx, nachricht) //nolint:wrapcheck // durchgereicht
	}

	haengt <- struct{}{}
	<-halt

	return errors.New("die Leitung ist gerissen")
}
