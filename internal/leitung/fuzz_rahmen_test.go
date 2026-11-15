package leitung_test

// **Was ein Fremder als Rahmen schickt** — QA-06
// (#1737),
// Fläche 31 aus `docs/FUZZZIELE.md`.
//
// # Warum
//
// **Der Längenpräfix ist das Erste, was ein Gegenüber bestimmt** — vor dem
// Handschlag, vor jeder Prüfung einer Unterschrift. Vier Bytes behaupten eine
// Länge, und wer ihr glaubt, teilt zu.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - **eine angenommene Nachricht, die nicht genau so lang ist, wie ihr
//     Rahmen sagt**, oder die andere Bytes trägt als die, die darauf folgten
//   - **eine behauptete Länge von null oder über der Grenze, die angenommen
//     wird** — dann hätte ein Fremder die Zuteilung bestimmt
//   - **mehr gelesen, als dastand**: die Nachrichten zusammen mit ihren Köpfen
//     sind höchstens so lang wie die Eingabe
//
//	go test ./internal/leitung -run Fuzz -fuzz FuzzRahmen -fuzztime 2m

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// bytestrom ist eine Verbindung, die genau diese Bytes hergibt und dann
// endet — ohne Netz und ohne zweite Goroutine.
type bytestrom struct{ *bytes.Reader }

func (bytestrom) Write(b []byte) (int, error)      { return len(b), nil }
func (bytestrom) Close() error                     { return nil }
func (bytestrom) LocalAddr() net.Addr              { return &net.TCPAddr{} }
func (bytestrom) RemoteAddr() net.Addr             { return &net.TCPAddr{} }
func (bytestrom) SetDeadline(time.Time) error      { return nil }
func (bytestrom) SetReadDeadline(time.Time) error  { return nil }
func (bytestrom) SetWriteDeadline(time.Time) error { return nil }

func rahmen(nachricht []byte) []byte {
	kopf := make([]byte, 4, 4+len(nachricht))
	binary.BigEndian.PutUint32(kopf, uint32(len(nachricht))) //nolint:gosec // nur kurze Saaten

	return append(kopf, nachricht...)
}

func FuzzRahmen(f *testing.F) {
	f.Add(rahmen([]byte("hallo")))
	f.Add(append(rahmen([]byte("eins")), rahmen([]byte("zwei"))...))
	f.Add([]byte{0, 0, 0, 0})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 'x'})
	f.Add([]byte{0, 0x10, 0, 1, 'x'})
	f.Add([]byte{0, 0, 0, 9, 'k', 'u', 'r', 'z'})
	f.Add([]byte{0, 0})

	f.Fuzz(func(t *testing.T, roh []byte) {
		ctx, abbrechen := context.WithTimeout(context.Background(), 5*time.Second)
		defer abbrechen()

		strom := leitung.Ueber(bytestrom{bytes.NewReader(roh)}, "fuzz")
		gelesen := 0

		for {
			nachricht, fehler := strom.Empfange(ctx)
			if fehler != nil {
				break
			}

			if len(roh) < gelesen+4 {
				t.Fatalf("eine Nachricht ohne vollständigen Kopf bei %d von %d", gelesen, len(roh))
			}

			behauptet := binary.BigEndian.Uint32(roh[gelesen : gelesen+4])
			if behauptet == 0 || behauptet > peer.Hoechstlaenge {
				t.Fatalf("eine behauptete Länge von %d angenommen", behauptet)
			}

			ende := gelesen + 4 + len(nachricht)
			if uint32(len(nachricht)) != behauptet || ende > len(roh) || //nolint:gosec // höchstens ein Mebibyte
				!bytes.Equal(nachricht, roh[gelesen+4:ende]) {
				t.Fatalf("der Rahmen sagt %d Byte, gelesen wurden %d andere", behauptet, len(nachricht))
			}

			gelesen = ende
		}
	})
}
