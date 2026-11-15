package app

// **Das Unified Log über einen laufenden `logger`, begrenzt** — PERF03
// (#2353),
// Befund A-20 des Audits 0.9.0.
//
// # Warum es das gibt
//
// **Bis PERF03 startete jede Zeile einen eigenen `logger`**, synchron im
// Schreibenden und mit einer Frist von fünf Sekunden — 2,7 ms je Zeile, und
// ein hängender `logger` hielt jeden auf, der protokollierte, also gerade den,
// bei dem es ohnehin klemmte.
//
// # Wie
//
// **Ein Kanal mit [systemkanalplaetze] Plätzen vor einem laufenden Prozess.**
// Wer schreibt, legt die Zeile ab und geht; ist der Kanal voll, wird sie
// verworfen, gezählt, und der Schreibende hört es (die Weiche nennt den
// Kanal dann nicht „an"). Ein Arbeiter gibt die Zeilen an die
// Standardeingabe des Prozesses — eine Zeile, ein Eintrag — und startet ihn
// neu, wenn er stirbt. Beim Schluss gehen die Zeilen noch hinaus, höchstens
// [systemkanalfrist] lang; danach wird der Prozess beendet.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// systemkanalplaetze ist, wie viele Zeilen warten dürfen.
const systemkanalplaetze = 256

// systemkanalfrist ist, wie lange der Schluss auf die letzten Zeilen wartet.
var systemkanalfrist = 5 * time.Second

// laufenderProzess ist ein gestarteter Schreiber: seine Eingabe, das Warten
// auf sein Ende und sein Abbruch.
type laufenderProzess struct {
	ein       io.WriteCloser
	warten    func() error
	abbrechen func() error
}

// systemkanal schreibt Zeilen über einen laufenden Prozess.
type systemkanal struct {
	starten func() (laufenderProzess, error)

	zeilen chan []byte
	fertig chan struct{}

	schloss     sync.Mutex
	zu          bool
	gestartet   bool
	abgebrochen bool
	verloren    int64
	grund       error
	lauf        *laufenderProzess
}

// neuerSystemkanal baut den Kanal über einem Starter.
func neuerSystemkanal(starten func() (laufenderProzess, error)) *systemkanal {
	return &systemkanal{
		starten: starten,
		zeilen:  make(chan []byte, systemkanalplaetze),
		fertig:  make(chan struct{}),
	}
}

// errSystemkanalZu ist der Fehler einer Zeile nach dem Schluss.
var errSystemkanalZu = errors.New("das Unified Log ist geschlossen")

// Write legt eine Zeile ab — **ohne zu warten**. Ein Fehler des Prozesses
// kommt mit der nächsten Zeile zurück: er geschah im Arbeiter.
func (k *systemkanal) Write(zeile []byte) (int, error) {
	text := bytes.TrimRight(zeile, "\n")
	if len(text) == 0 {
		return len(zeile), nil
	}

	k.schloss.Lock()
	defer k.schloss.Unlock()

	if k.zu {
		return 0, errSystemkanalZu
	}

	if !k.gestartet {
		k.gestartet = true

		go k.arbeiten()
	}

	select {
	case k.zeilen <- bytes.Clone(text):
	default:
		k.verloren++

		return 0, fmt.Errorf("das Unified Log nimmt gerade nichts an — %d Zeile(n) verworfen", k.verloren)
	}

	return len(zeile), k.grund
}

// arbeiten gibt die Zeilen an den Prozess und startet ihn, wenn keiner läuft.
func (k *systemkanal) arbeiten() {
	defer close(k.fertig)

	for zeile := range k.zeilen {
		lauf, fehler := k.laufender()
		if fehler == nil {
			_, fehler = lauf.ein.Write(append(zeile, '\n'))
		}

		k.schloss.Lock()
		k.grund = nil

		// **Die Zeile ist verloren, und das wird gezählt** (Nachlauf über
		// Phase 3, Befund 6): `grund` steht nur bis zur nächsten geglückten
		// Zeile, die Zahl bis zum Schluss — und der nennt sie.
		if fehler != nil {
			k.verloren++
			k.grund = fmt.Errorf("das Unified Log: %w", fehler)
			k.beenden()
		}
		k.schloss.Unlock()
	}

	// **Auf sein Ende gewartet wird ohne das Schloss** (Zweitblick über Phase
	// 3, Befund 5): endet der Prozess nicht, braucht der Schluss nach seiner
	// Frist dasselbe Schloss, um ihn abzubrechen — und wartete sonst mit.
	// Der Lauf bleibt dabei eingetragen, damit [systemkanal.beenden] ihn
	// findet.
	k.schloss.Lock()
	lauf := k.lauf
	k.schloss.Unlock()

	if lauf != nil {
		_ = lauf.ein.Close() //nolint:errcheck // die Zeilen sind geschrieben
		_ = lauf.warten()    //nolint:errcheck // sein Ende sagt nichts mehr
	}

	k.schloss.Lock()
	k.lauf = nil
	k.schloss.Unlock()
}

// laufender gibt den laufenden Prozess — oder startet einen.
func (k *systemkanal) laufender() (*laufenderProzess, error) {
	k.schloss.Lock()
	defer k.schloss.Unlock()

	if k.lauf != nil {
		return k.lauf, nil
	}

	// **Nach dem Abbruch beim Schluss startet keiner mehr** — die übrigen
	// Zeilen gehen verloren, und das hat der Schluss gesagt.
	if k.abgebrochen {
		return nil, errSystemkanalZu
	}

	neu, fehler := k.starten()
	if fehler != nil {
		return nil, fehler
	}

	neu.warten = einmal(neu.warten)
	k.lauf = &neu

	return k.lauf, nil
}

// beenden bricht den laufenden Prozess ab — unter dem Schloss gerufen.
func (k *systemkanal) beenden() {
	if k.lauf == nil {
		return
	}

	_ = k.lauf.ein.Close()                                       //nolint:errcheck // er ist ohnehin fort
	_ = k.lauf.abbrechen()                                       //nolint:errcheck // dito
	go func(warten func() error) { _ = warten() }(k.lauf.warten) //nolint:errcheck // nur abgeholt

	k.lauf = nil
}

// Close lässt die abgelegten Zeilen noch hinaus und wartet höchstens
// [systemkanalfrist]; danach wird der Prozess beendet.
func (k *systemkanal) Close() error {
	k.schloss.Lock()

	if k.zu {
		k.schloss.Unlock()

		return nil
	}

	k.zu = true
	gestartet := k.gestartet
	close(k.zeilen)
	k.schloss.Unlock()

	if !gestartet {
		return nil
	}

	select {
	case <-k.fertig:
		return k.verlust()
	case <-time.After(systemkanalfrist):
		k.schloss.Lock()
		k.abgebrochen = true
		k.beenden()
		k.schloss.Unlock()

		return fmt.Errorf("das Unified Log nahm in %s nicht alles an", systemkanalfrist)
	}
}

// einmal macht aus dem Warten auf einen Prozess einen Ruf, den zwei Stellen
// tun dürfen — `exec.Cmd.Wait` selbst darf nur einmal laufen (os/exec), und
// der Schluss nach seiner Frist wartete neben dem Arbeiter ein zweites Mal
// (Nachlauf über Phase 3, Befund 5). Jeder Rufer hört dasselbe Ende.
func einmal(warten func() error) func() error {
	var (
		riegel sync.Once
		ende   error
	)

	return func() error {
		riegel.Do(func() { ende = warten() })

		return ende
	}
}

// verlust nennt, wie viele Zeilen das Unified Log nicht erreichten — `nil`,
// wenn alle ankamen.
func (k *systemkanal) verlust() error {
	k.schloss.Lock()
	defer k.schloss.Unlock()

	if k.verloren == 0 {
		return nil
	}

	return fmt.Errorf("%d Zeile(n) erreichten das Unified Log nicht", k.verloren)
}
