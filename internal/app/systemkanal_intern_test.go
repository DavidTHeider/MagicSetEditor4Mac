package app

// **Das Unified Log über einen laufenden Prozess, begrenzt** — PERF03
// (#2353),
// Befund A-20. Der Prozess ist hier eine Leitung, an deren anderem Ende der
// Lauf liest — oder nicht liest.

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// probeprozess ist ein nachgebauter `logger`: was ankommt, liest `aus`.
type probeprozess struct {
	aus         *bufio.Reader
	leser       *io.PipeReader
	abgebrochen atomic.Bool
}

// loggerprobe startet bei jedem Ruf einen neuen Probeprozess und reicht ihn
// über `gestartet` hinaus.
func loggerprobe(gestartet chan<- *probeprozess) func() (laufenderProzess, error) {
	return func() (laufenderProzess, error) {
		leser, schreiber := io.Pipe()
		probe := &probeprozess{aus: bufio.NewReader(leser), leser: leser}
		gestartet <- probe

		return laufenderProzess{
			ein:    schreiber,
			warten: func() error { return nil },
			abbrechen: func() error {
				probe.abgebrochen.Store(true)

				return leser.Close()
			},
		}, nil
	}
}

// naechsterProzess wartet auf den nächsten gestarteten Probeprozess — mit
// Frist, damit ein Kanal, der nie startet, rot wird statt zu hängen.
func naechsterProzess(t *testing.T, gestartet <-chan *probeprozess) *probeprozess {
	t.Helper()

	select {
	case probe := <-gestartet:
		return probe
	case <-time.After(2 * time.Second):
		t.Fatal("kein Prozess gestartet")

		return nil
	}
}

func TestDerSystemkanalGibtJedeZeileEinmalWeiter(t *testing.T) {
	t.Parallel()

	gestartet := make(chan *probeprozess, 4)
	kanal := neuerSystemkanal(loggerprobe(gestartet))

	for _, zeile := range []string{"erste\n", "\n", "zweite\n"} {
		if _, err := kanal.Write([]byte(zeile)); err != nil {
			t.Fatal(err)
		}
	}

	probe := naechsterProzess(t, gestartet)

	for _, erwartet := range []string{"erste\n", "zweite\n"} {
		if gelesen, err := probe.aus.ReadString('\n'); err != nil || gelesen != erwartet {
			t.Errorf("gelesen %q (%v), erwartet %q — eine leere Zeile geht nicht hinaus", gelesen, err, erwartet)
		}
	}

	ende := make(chan error, 1)

	go func() { ende <- kanal.Close() }()

	// **Der Schluss schließt die Eingabe**: danach liest das andere Ende das
	// Ende der Leitung.
	if _, err := probe.aus.ReadString('\n'); !errors.Is(err, io.EOF) {
		t.Errorf("nach dem Schluss: %v", err)
	}

	if err := <-ende; err != nil {
		t.Errorf("Schluss: %v", err)
	}

	if _, err := kanal.Write([]byte("danach\n")); !errors.Is(err, errSystemkanalZu) {
		t.Errorf("nach dem Schluss geschrieben: %v", err)
	}

	if err := kanal.Close(); err != nil {
		t.Errorf("ein zweiter Schluss: %v", err)
	}
}

// **Wer schreibt, wartet nicht** — auch wenn der Prozess nichts annimmt. Was
// nicht mehr hineinpasst, wird verworfen und gezählt; der Schluss wartet
// höchstens seine Frist und beendet den Prozess dann.
func TestEinHaengenderLoggerHaeltNiemandenAuf(t *testing.T) { //nolint:paralleltest // die Frist ist eine Paketvariable
	vorher := systemkanalfrist
	systemkanalfrist = 100 * time.Millisecond

	t.Cleanup(func() { systemkanalfrist = vorher })

	gestartet := make(chan *probeprozess, 4)
	kanal := neuerSystemkanal(loggerprobe(gestartet))

	anfang := time.Now()

	var verworfen error

	for range systemkanalplaetze + 8 {
		if _, err := kanal.Write([]byte("eine Zeile\n")); err != nil {
			verworfen = err
		}
	}

	if dauer := time.Since(anfang); dauer > time.Second {
		t.Errorf("das Schreiben wartete %s", dauer)
	}

	if verworfen == nil || !strings.Contains(verworfen.Error(), "verworfen") {
		t.Errorf("über dem Kanal: %v", verworfen)
	}

	probe := naechsterProzess(t, gestartet)

	anfang = time.Now()
	if err := kanal.Close(); err == nil {
		t.Error("der Schluss meldete nicht, dass Zeilen liegen blieben")
	}

	if dauer := time.Since(anfang); dauer > time.Second || !probe.abgebrochen.Load() {
		t.Errorf("Schluss nach %s, abgebrochen %v", dauer, probe.abgebrochen.Load())
	}
}

// **Ein Prozess, der stirbt, wird neu gestartet** — und sein Fehler kommt mit
// der nächsten Zeile beim Schreibenden an, damit die Weiche den Kanal nicht
// „an" nennt.
func TestEinGestorbenerLoggerWirdNeuGestartet(t *testing.T) {
	t.Parallel()

	gestartet := make(chan *probeprozess, 4)
	kanal := neuerSystemkanal(loggerprobe(gestartet))

	t.Cleanup(func() { _ = kanal.Close() }) //nolint:errcheck // Aufräumen

	if _, err := kanal.Write([]byte("vor dem Sterben\n")); err != nil {
		t.Fatal(err)
	}

	erster := naechsterProzess(t, gestartet)
	if _, err := erster.aus.ReadString('\n'); err != nil {
		t.Fatal(err)
	}

	// **Er stirbt**: das andere Ende der Leitung geht zu.
	if err := erster.leser.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := kanal.Write([]byte("ins Leere\n")); err != nil {
		t.Fatal(err)
	}

	// **Gewartet wird auf die Wirkung** (§3): bis eine Zeile den Fehler trägt.
	var gemeldet error

	for range 200 {
		if _, gemeldet = kanal.Write([]byte("nachgefragt\n")); gemeldet != nil {
			break
		}

		time.Sleep(5 * time.Millisecond)
	}

	if gemeldet == nil || !strings.Contains(gemeldet.Error(), "Unified Log") {
		t.Errorf("der Fehler kam nicht an: %v", gemeldet)
	}

	select {
	case zweiter := <-gestartet:
		if gelesen, err := zweiter.aus.ReadString('\n'); err != nil || gelesen == "" {
			t.Errorf("der neue Prozess: %q, %v", gelesen, err)
		}
	case <-time.After(2 * time.Second):
		t.Error("kein neuer Prozess")
	}
}

// **Ein Starter, der scheitert, meldet sich** — und der nächste Versuch
// startet neu.
func TestEinScheiternderStartMeldetSich(t *testing.T) {
	t.Parallel()

	kanal := neuerSystemkanal(func() (laufenderProzess, error) { return laufenderProzess{}, errors.New("kein logger") })

	t.Cleanup(func() { _ = kanal.Close() }) //nolint:errcheck // Aufräumen

	var gemeldet error

	for range 200 {
		if _, gemeldet = kanal.Write([]byte("eine Zeile\n")); gemeldet != nil {
			break
		}

		time.Sleep(5 * time.Millisecond)
	}

	if gemeldet == nil || !strings.Contains(gemeldet.Error(), "kein logger") {
		t.Errorf("der fehlende logger blieb stumm: %v", gemeldet)
	}
}

// **Ein Logger, der nach dem Schluss seiner Eingabe nicht endet, hält den
// Schluss nicht auf** — Befund 5 des Zweitblicks über Phase 3 von 0.9.0: der
// Arbeiter wartete unter dem Schloss auf sein Ende, und der Abbruch nach der
// Frist wartete auf dasselbe Schloss.
func TestEinLoggerDerNichtEndetHaeltDenSchlussNichtAuf(t *testing.T) { //nolint:paralleltest // die Frist ist eine Paketvariable
	vorher := systemkanalfrist
	systemkanalfrist = 100 * time.Millisecond

	t.Cleanup(func() { systemkanalfrist = vorher })

	angekommen := make(chan struct{}, 1)
	ende := make(chan struct{})

	var abgebrochen atomic.Bool

	kanal := neuerSystemkanal(func() (laufenderProzess, error) {
		return laufenderProzess{
			ein:    nimmtAlles{angekommen: angekommen},
			warten: func() error { <-ende; return nil },
			abbrechen: func() error {
				if abgebrochen.CompareAndSwap(false, true) {
					close(ende)
				}

				return nil
			},
		}, nil
	})

	if _, err := kanal.Write([]byte("eine Zeile\n")); err != nil {
		t.Fatal(err)
	}

	<-angekommen

	geschlossen := make(chan error, 1)

	go func() { geschlossen <- kanal.Close() }()

	select {
	case fehler := <-geschlossen:
		// **Und er sagt, dass nicht alles ankam** — der Logger endete nicht
		// in seiner Frist (`systemkanalfrist`).
		if fehler == nil {
			t.Error("der Schluss verschwieg, dass das Unified Log nicht alles annahm")
		}
	case <-time.After(2 * time.Second):
		close(ende)
		t.Fatal("der Schluss hing am Ende des Loggers")
	}

	if !abgebrochen.Load() {
		t.Error("der Logger wurde nicht beendet")
	}
}

// nimmtAlles nimmt jede Zeile und sagt es einmal.
type nimmtAlles struct{ angekommen chan struct{} }

func (n nimmtAlles) Write(zeile []byte) (int, error) {
	select {
	case n.angekommen <- struct{}{}:
	default:
	}

	return len(zeile), nil
}

func (nimmtAlles) Close() error { return nil }
