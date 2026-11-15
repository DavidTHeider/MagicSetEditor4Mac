package app

// **Zwei gegenläufige Schaltrufe enden in einem Zustand** — N2 des Nachlaufs
// über Phase 6.10 (#1671). Der Schalter der Anwendung und der des Dienstes
// wurden in zwei Schritten gesetzt; ein zweiter Ruf dazwischen ließ die
// Anwendung „an" sagen, während der Dienst jeden Start aufhielt.

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/applefm"
)

func TestGegenlaeufigeSchaltrufeEndenEinig(t *testing.T) {
	t.Parallel()

	gestartet := 0
	a := geraeteapp(t, applefm.EinrichtungBereit, `{}`, &gestartet)

	// **Der erste Ruf hält an, und er hält sicher an** — er ist der
	// Ausschaltruf, denn der Gegenruf beginnt erst nach `<-drin`. Hier stand
	// ein nicht blockierendes Senden: kam der erste Ruf an der Naht an, bevor
	// der Test bei `<-drin` stand, schaltete er durch, und der Test wartete
	// ewig — in der Pipeline vom 2026-10-05 dreißig Minuten; nachgestellt
	// mit 300 Wiederholungen unter `-race`, die in ihre Frist liefen.
	//
	// **Und jedes Warten hat eine örtliche Frist** — T07 (#2100). Ohne sie
	// wartete der Test auf eine Naht, die nie rief, bis zum Ende der ganzen
	// Teststufe; jetzt nennt er die Phase. `drin` hat einen Platz, damit der
	// Ruf nicht hängt, wenn der Test schon gescheitert ist, und `weiter` geht
	// beim Aufräumen auf, auch nach einem Fehlschlag.
	drin, weiter := make(chan struct{}, 1), make(chan struct{})

	var auf sync.Once

	oeffnen := func() { auf.Do(func() { close(weiter) }) }
	t.Cleanup(oeffnen)

	var angehalten atomic.Bool

	a.geraetnaht().zwischenSchalten = func() {
		if angehalten.CompareAndSwap(false, true) {
			drin <- struct{}{}
			<-weiter
		}
	}

	aus := make(chan KIGeraetSitzungAntwort, 1)
	go func() { aus <- a.KIGeraetemodellFuerSitzung(t.Context(), false) }()

	empfangeIn(t, drin, "der Ausschaltruf erreicht die Naht zwischen den zwei Schritten nicht")

	an := make(chan KIGeraetSitzungAntwort, 1)
	go func() { an <- a.KIGeraetemodellFuerSitzung(t.Context(), true) }()

	// **Der zweite Ruf darf vor dem ersten nicht fertig werden** — er wartet
	// auf ihn. Gewartet wird höchstens so lange, wie er ohne Ordnung
	// bräuchte; danach geht der erste weiter.
	var antwort KIGeraetSitzungAntwort

	fertig := false

	select {
	case antwort = <-an:
		fertig = true

		t.Error("der Gegenruf wurde fertig, während der erste noch schaltete")
	case <-time.After(200 * time.Millisecond):
	}

	oeffnen()
	empfangeIn(t, aus, "der Ausschaltruf kehrt nach dem Öffnen nicht zurück")

	if !fertig {
		antwort = empfangeIn(t, an, "der Gegenruf kehrt nach dem ersten nicht zurück")
	}

	if !antwort.An {
		t.Fatalf("%+v", antwort)
	}

	if a.geraetSitzungAus() || a.geraetedienst().Abgeschaltet() {
		t.Errorf("uneinig: Anwendung aus %v, Dienst abgeschaltet %v",
			a.geraetSitzungAus(), a.geraetedienst().Abgeschaltet())
	}
}

// empfangeIn wartet auf einen Wert, höchstens bis zur örtlichen Frist, und
// nennt sonst die Phase, in der er ausblieb.
func empfangeIn[T any](t *testing.T, kanal <-chan T, phase string) T {
	t.Helper()

	frist := oertlicheFrist(t)

	select {
	case wert := <-kanal:
		return wert
	case <-time.After(frist):
		t.Fatalf("%s — nach %v", phase, frist)
	}

	var leer T

	return leer
}

// oertlicheFrist ist zehn Sekunden, höchstens aber die Hälfte dessen, was
// dem Test bis zu seiner Frist bleibt — so meldet er sich, bevor `go test`
// ihn mit einem Stapelabzug beendet.
func oertlicheFrist(t *testing.T) time.Duration {
	t.Helper()

	frist := 10 * time.Second
	if ende, gesetzt := t.Deadline(); gesetzt {
		frist = min(frist, time.Until(ende)/2)
	}

	return frist
}
