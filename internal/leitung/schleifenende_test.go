package leitung

// **Ein Niederlegen beendet ein laufendes Warten** — und welches Ende es war,
// entscheidet, was der Wartende hört (#2504): drüben zu ist das Ende des
// Stroms, das eigene Ende zu heißt `net.ErrClosed`. Aus
// `schleifenfristen_test.go` herausgelöst, als es mit den zwei neuen Fällen
// über die Zeilengrenze ging (§2).

import (
	"errors"
	"io"
	"net"
	"testing"
)

// **Ein Niederlegen beendet ein laufendes Warten — lesend wie schreibend.**
func TestEinNiederlegenBeendetDasWarten(t *testing.T) {
	t.Parallel()

	t.Run("lesend", func(t *testing.T) {
		t.Parallel()

		hier, dort := gepuffertesPaar()

		ab := bisher(t, dort)
		fertig := starteLeser(dort)

		// **Warten, bis wirklich jemand wartet** — sonst führe dieser Lauf
		// durch [schleifenende.vorbei] und nicht durch das `select` (§3).
		bisErWartet(t, dort, ab)

		if fehler := hier.Close(); fehler != nil {
			t.Fatalf("schließen: %v", fehler)
		}

		// **Drüben zu heißt das Ende des Stroms**, wie bei einem Socket
		// (#2504) — und nicht „dieses Ende ist zu".
		if fehler := warteAufEnde(t, fertig); !errors.Is(fehler, io.EOF) {
			t.Errorf("lesen kam mit %v", fehler)
		}
	})

	t.Run("lesend, das eigene Ende", func(t *testing.T) {
		t.Parallel()

		_, dort := gepuffertesPaar()

		ab := bisher(t, dort)
		fertig := starteLeser(dort)

		bisErWartet(t, dort, ab)

		if fehler := dort.Close(); fehler != nil {
			t.Fatalf("schließen: %v", fehler)
		}

		if fehler := warteAufEnde(t, fertig); !errors.Is(
			fehler, net.ErrClosed) {
			t.Errorf("lesen kam mit %v", fehler)
		}
	})

	t.Run("schreibend, das eigene Ende", func(t *testing.T) {
		t.Parallel()

		hier, _ := gepuffertesPaar()

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

		if fehler := hier.Close(); fehler != nil {
			t.Fatalf("schließen: %v", fehler)
		}

		if fehler := warteAufEnde(t, fertig); !errors.Is(
			fehler, net.ErrClosed) {
			t.Errorf("schreiben kam mit %v", fehler)
		}
	})

	t.Run("schreibend", func(t *testing.T) {
		t.Parallel()

		hier, dort := gepuffertesPaar()

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

		if fehler := dort.Close(); fehler != nil {
			t.Fatalf("schließen: %v", fehler)
		}

		if fehler := warteAufEnde(t, fertig); !errors.Is(
			fehler, net.ErrClosed) {
			t.Errorf("schreiben kam mit %v", fehler)
		}
	})
}

// **Ein Stück, das während des Wartens kommt, weckt den Leser** — der Weg
// durch das `select` und nicht durch den Blick in den Puffer davor
// ([schleifenende.bereit]).
func TestEinStueckWecktDenWartendenLeser(t *testing.T) {
	t.Parallel()

	hier, dort := gepuffertesPaar()

	ab := bisher(t, dort)
	fertig := starteLeser(dort)

	bisErWartet(t, dort, ab)

	if _, fehler := hier.Write([]byte("da")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := warteAufEnde(t, fertig); fehler != nil {
		t.Errorf("lesen kam mit %v", fehler)
	}
}
