package leitung

// **Ein Rahmen, der angekündigt ist, kommt binnen seiner Frist** — D05
// (#2236).
//
// Die langen Sitzungen empfangen mit einem Kontext ohne Frist
// (`internal/service/partienetz.go`): eine Sitzung darf ruhen. Ein Gegenüber,
// das einen Kopf über ein Mebibyte schickte und den Rumpf dann tröpfelte,
// hielt aber Leser und Puffer, bis die Sitzung endete. Die Überlänge war
// begrenzt, der langsame Sender nicht.

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitRahmenfrist setzt die Frist für die Dauer eines Tests.
func mitRahmenfrist(t *testing.T, frist time.Duration) {
	t.Helper()

	vorher := rahmenfrist
	rahmenfrist = frist

	t.Cleanup(func() { rahmenfrist = vorher })
}

// **Byte für Byte, und trotzdem ganz** — `io.ReadFull` setzt zusammen, was
// in Stücken kommt; dieser Lauf hält es fest.
func TestEinRahmenInEinzelnenBytesKommtGanzAn(t *testing.T) {
	mitRahmenfrist(t, 5*time.Second)

	hier, dort := net.Pipe()
	t.Cleanup(func() { _ = hier.Close(); _ = dort.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	inhalt := []byte("in Stücken")
	go func() {
		var kopf [Rahmenlaenge]byte
		binary.BigEndian.PutUint32(kopf[:], uint32(len(inhalt))) //nolint:gosec // eine kleine Probe

		for _, b := range append(kopf[:], inhalt...) {
			if _, fehler := dort.Write([]byte{b}); fehler != nil {
				return
			}

			time.Sleep(2 * time.Millisecond)
		}
	}()

	gelesen, fehler := Ueber(hier, "probe").Empfange(context.Background())
	if fehler != nil || !bytes.Equal(gelesen, inhalt) {
		t.Errorf("gelesen %q (%v), geschrieben %q", gelesen, fehler, inhalt)
	}
}

// **Ein Rumpf, der tröpfelt, endet mit der Rahmenfrist** — auch ohne Frist
// im Kontext.
func TestEinLangsamerSenderEndetMitDerRahmenfrist(t *testing.T) {
	mitRahmenfrist(t, 150*time.Millisecond)

	hier, dort := net.Pipe()
	t.Cleanup(func() { _ = hier.Close(); _ = dort.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	go func() {
		var kopf [Rahmenlaenge]byte
		binary.BigEndian.PutUint32(kopf[:], 64)

		if _, fehler := dort.Write(kopf[:]); fehler != nil {
			return
		}

		for range 64 {
			time.Sleep(50 * time.Millisecond)

			if _, fehler := dort.Write([]byte{'x'}); fehler != nil {
				return
			}
		}
	}()

	ergebnis := make(chan error, 1)
	beginn := time.Now()

	go func() {
		_, fehler := Ueber(hier, "probe").Empfange(context.Background())
		ergebnis <- fehler
	}()

	select {
	case fehler := <-ergebnis:
		if !errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("der langsame Sender endete mit %v, erwartet ein Konflikt", fehler)
		}

		if dauer := time.Since(beginn); dauer > time.Second {
			t.Errorf("er endete erst nach %v", dauer)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("der langsame Sender hält den Leser fest — keine Rahmenfrist")
	}
}

// abgebrochenOhneSignal gilt als abgebrochen, signalisiert es aber nie — so
// kommt der Kopf sicher an, und erst danach zählt der Abbruch.
type abgebrochenOhneSignal struct{ context.Context }

func (abgebrochenOhneSignal) Done() <-chan struct{} { return nil }
func (abgebrochenOhneSignal) Err() error            { return context.Canceled }

// **Ein Abbruch zwischen Kopf und Rumpf geht vor** — die Rahmenfrist
// überschreibt ihn nicht.
func TestEinAbbruchZwischenKopfUndRumpfGehtVor(t *testing.T) {
	mitRahmenfrist(t, time.Minute)

	hier, dort := net.Pipe()
	t.Cleanup(func() { _ = hier.Close(); _ = dort.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	go func() {
		var kopf [Rahmenlaenge]byte
		binary.BigEndian.PutUint32(kopf[:], 8)
		_, _ = dort.Write(kopf[:]) //nolint:errcheck // der Leser bricht ab
	}()

	beginn := time.Now()

	_, fehler := Ueber(hier, "probe").Empfange(abgebrochenOhneSignal{context.Background()})
	if !errors.Is(fehler, domain.ErrConflict) || time.Since(beginn) > time.Second {
		t.Errorf("nach %v: %v — erwartet ein sofortiger Abbruch", time.Since(beginn), fehler)
	}
}
