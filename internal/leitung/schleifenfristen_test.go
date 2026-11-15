package leitung

// **Was ein Warten an der gepufferten Schleife beendet** — Werkstück G5 von
// Phase 5.5 (#1462).
//
// **Aus `schleifenpuffer_test.go` herausgelöst**, als sie an die Zeilengrenze
// aus §2 stieß. Der Schnitt folgt der Frage: dort steht, **was sie trägt** —
// Stücke, Überhang, Adressen —, hier, **was ein Warten beendet**: eine Frist,
// ein Niederlegen, ein voller Puffer.
//
// # Und warum das die schwierige Hälfte ist
//
// **Weil `internal/leitung` einen Abbruch über eine Frist in der Vergangenheit
// macht** — so kennt `net.Conn` das. Wer sie nur beim Eintritt läse, hinge
// weiter; wer sie nur im `select` läse, würfelte gegen einen freien Puffer.
// Beide Wege stehen hier, und beide werden betreten.

import (
	"errors"
	"net"
	"os"
	"testing"
	"time"
)

// **Eine verstrichene Frist lässt nichts warten.**
func TestEineVerstricheneFristSchlaegtSofort(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	t.Cleanup(func() {
		_ = hier.Close() //nolint:errcheck // nur ein Lauf
	})

	vergangen := time.Now().Add(-time.Second)

	if fehler := dort.SetReadDeadline(vergangen); fehler != nil {
		t.Fatalf("Lesefrist: %v", fehler)
	}

	if _, fehler := dort.Read(make([]byte, 4)); !errors.Is(
		fehler, os.ErrDeadlineExceeded) {
		t.Errorf("lesen kam mit %v", fehler)
	}

	// **Und beim Schreiben ebenso** — dafür muss der Puffer voll sein, sonst
	// geht es einfach durch.
	if fehler := hier.SetWriteDeadline(vergangen); fehler != nil {
		t.Fatalf("Schreibfrist: %v", fehler)
	}

	if _, fehler := hier.Write([]byte("hallo")); !errors.Is(
		fehler, os.ErrDeadlineExceeded) {
		t.Errorf("schreiben kam mit %v", fehler)
	}
}

// **Eine Frist, die erst später gesetzt wird, bricht ein laufendes Warten ab.**
//
// # Warum das der Kern dieser Datei ist
//
// **Weil `internal/leitung` einen Abbruch genau so macht**: es setzt eine
// Frist in die **Vergangenheit**, während jemand schon wartet. Eine Schleife,
// die ihre Frist nur beim Eintritt läse, hinge weiter — und ein abgebrochener
// Ruf käme nie zurück.
func TestEineSpaeterGesetzteFristBrichtDasWartenAb(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	t.Cleanup(func() {
		_ = hier.Close() //nolint:errcheck // nur ein Lauf
	})

	fertig := make(chan error, 1)

	ab := bisher(t, dort)

	go func() {
		_, fehler := dort.Read(make([]byte, 4))
		fertig <- fehler
	}()

	// **Warten, bis der Leser wirklich wartet** — sonst prüfte dieser Lauf
	// den Fall aus dem Lauf darüber (§3).
	bisErWartet(t, dort, ab)

	if fehler := dort.SetDeadline(time.Unix(1, 0)); fehler != nil {
		t.Fatalf("Frist: %v", fehler)
	}

	select {
	case fehler := <-fertig:
		if !errors.Is(fehler, os.ErrDeadlineExceeded) {
			t.Errorf("das Lesen kam mit %v", fehler)
		}

	case <-time.After(2 * time.Second):
		t.Error("das Lesen hing trotz Frist")
	}
}

// **Eine gesetzte Frist in der Zukunft hält niemanden auf.**
func TestEineFristInDerZukunftHaeltNiemandenAuf(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	t.Cleanup(func() {
		_ = hier.Close() //nolint:errcheck // nur ein Lauf
	})

	if fehler := dort.SetDeadline(time.Now().Add(time.Minute)); fehler != nil {
		t.Fatalf("Frist: %v", fehler)
	}

	if _, fehler := hier.Write([]byte("hallo")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if _, fehler := dort.Read(make([]byte, 5)); fehler != nil {
		t.Errorf("lesen kam mit %v", fehler)
	}
}

// **Die Adressen sagen, was sie sind.**
//
// **`net.Conn` verlangt sie**, und sie behaupten nichts: hier ist kein
// Rechner, hier ist ein Prozess.
func TestDieSchleifeNenntIhreAdresse(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	t.Cleanup(func() {
		_ = hier.Close() //nolint:errcheck // nur ein Lauf
	})

	for _, adresse := range []net.Addr{
		hier.LocalAddr(), dort.RemoteAddr(),
	} {
		if adresse.Network() != "schleife" || adresse.String() == "" {
			t.Errorf("die Adresse ist %q/%q",
				adresse.Network(), adresse.String())
		}
	}
}

// **Ein voller Puffer lässt warten — und eine Frist beendet das Warten.**
//
// # Warum dieser Lauf dasteht
//
// **Weil der Puffer eine Grenze hat und sie eine Zusage ist.** Wer mehr
// schickt, als unterwegs sein darf, wartet — genau wie an einem echten
// Socket. Ohne diesen Lauf stünde die Zusage im Kopf der Datei und nirgends
// sonst (§10).
func TestEinVollerPufferLaesstWarten(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	t.Cleanup(func() {
		_ = dort.Close() //nolint:errcheck // nur ein Lauf
	})

	for range Schleifenpuffer {
		if _, fehler := hier.Write([]byte("x")); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	if fehler := hier.SetWriteDeadline(
		time.Now().Add(50 * time.Millisecond)); fehler != nil {
		t.Fatalf("Frist: %v", fehler)
	}

	if _, fehler := hier.Write([]byte("einer zu viel")); !errors.Is(
		fehler, os.ErrDeadlineExceeded) {
		t.Errorf("schreiben kam mit %v", fehler)
	}
}

// **Ein Warten endet an der Frist, die schon stand.**
//
// **Nicht dasselbe wie der Lauf darüber**: dort wird die Frist **während** des
// Wartens gesetzt, hier steht sie vorher und der Wecker schlägt mitten im
// Warten. Zwei Wege in denselben Ausgang, und beide gehören betreten (§10).
func TestEineLaufendeFristBeendetDasWarten(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	t.Cleanup(func() {
		_ = hier.Close() //nolint:errcheck // nur ein Lauf
	})

	if fehler := dort.SetReadDeadline(
		time.Now().Add(30 * time.Millisecond)); fehler != nil {
		t.Fatalf("Frist: %v", fehler)
	}

	if _, fehler := dort.Read(make([]byte, 4)); !errors.Is(
		fehler, os.ErrDeadlineExceeded) {
		t.Errorf("lesen kam mit %v", fehler)
	}

	// **Und beim Schreiben ebenso, wenn der Puffer voll ist.**
	for range Schleifenpuffer {
		if _, fehler := hier.Write([]byte("x")); fehler != nil {
			t.Fatalf("füllen: %v", fehler)
		}
	}

	if fehler := hier.SetWriteDeadline(
		time.Now().Add(30 * time.Millisecond)); fehler != nil {
		t.Fatalf("Frist: %v", fehler)
	}

	if _, fehler := hier.Write([]byte("einer zu viel")); !errors.Is(
		fehler, os.ErrDeadlineExceeded) {
		t.Errorf("schreiben kam mit %v", fehler)
	}
}

// **Eine Frist, die sich beim Schreiben ändert, lässt weiterwarten.**
//
// **Die Gegenprobe zum Abbruch**: eine **neue** Frist ist kein Ende — wer sie
// als eines läse, bräche ein Warten ab, das noch laufen darf.
func TestEineNeueSchreibfristBrichtNichtAb(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	t.Cleanup(func() {
		_ = hier.Close() //nolint:errcheck // nur ein Lauf
	})

	for range Schleifenpuffer {
		if _, fehler := hier.Write([]byte("x")); fehler != nil {
			t.Fatalf("füllen: %v", fehler)
		}
	}

	fertig := make(chan error, 1)
	ab := bisher(t, hier)

	go func() {
		_, fehler := hier.Write([]byte("einer zu viel"))
		fertig <- fehler
	}()

	bisErWartet(t, hier, ab)

	if fehler := hier.SetWriteDeadline(
		time.Now().Add(time.Minute)); fehler != nil {
		t.Fatalf("Frist: %v", fehler)
	}

	// **Und jetzt liest jemand** — damit geht der Wartende durch.
	if _, fehler := dort.Read(make([]byte, 1)); fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if fehler := warteAufEnde(t, fertig); fehler != nil {
		t.Errorf("schreiben kam mit %v", fehler)
	}
}

// starteLeser liest in einer eigenen Goroutine.
func starteLeser(woraus net.Conn) chan error {
	fertig := make(chan error, 1)

	go func() {
		_, fehler := woraus.Read(make([]byte, 4))
		fertig <- fehler
	}()

	return fertig
}

// warteAufEnde holt den Ausgang — oder bricht ab.
func warteAufEnde(t *testing.T, fertig chan error) error {
	t.Helper()

	select {
	case fehler := <-fertig:
		return fehler

	case <-time.After(2 * time.Second):
		t.Fatal("es hing")

		return nil
	}
}

// bisher nennt, wie oft an diesem Ende schon gewartet wurde — **vor** dem
// Ruf, auf dessen Warten ein Lauf gleich wartet. Das Füllen des Puffers geht
// auch durch das `select` und zählt mit.
func bisher(t *testing.T, ende net.Conn) int64 {
	t.Helper()

	return gepuffert(t, ende).gewartet.Load()
}

// bisErWartet wartet, bis an diesem Ende nach `ab` ein Lesen oder Schreiben
// im `select` steht — **auf die Wirkung und nicht auf eine Frist** (§3).
func bisErWartet(t *testing.T, ende net.Conn, ab int64) {
	t.Helper()

	gezaehlt := &gepuffert(t, ende).gewartet
	bis := time.Now().Add(2 * time.Second)

	for gezaehlt.Load() <= ab {
		if time.Now().After(bis) {
			t.Fatal("niemand ging ins Warten")
		}

		time.Sleep(time.Millisecond)
	}
}

// gepuffert gibt das Ende der Schleife hinter der Verbindung.
func gepuffert(t *testing.T, ende net.Conn) *schleifenende {
	t.Helper()

	schleife, ist := ende.(*schleifenende)
	if !ist {
		t.Fatalf("%T ist kein Ende der Schleife", ende)
	}

	return schleife
}
