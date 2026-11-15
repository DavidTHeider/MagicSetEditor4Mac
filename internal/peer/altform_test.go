package peer_test

// **Die andere Textform bekommt einen Satz, der weiterhilft** — Befund 2 des
// Zweitblicks über Phase 5.4, geprüft und **zur Hälfte** übernommen.
//
// # Was der Befund sagte und was davon gilt
//
// **Er sagte, der Leser verenge einen Interoperabilitätsstandard auf die
// selbst erzeugte Teilmenge** — und das stimmt: libp2p kennt zwei Textformen,
// dieses Paket liest eine.
//
// **Was daraus nicht folgt, ist ein zweiter Leser.** Der Transport ist seit
// [0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md) ausdrücklich
// nicht libp2p; wer `12D3Koo…` einfügte, hätte ein Gegenüber, mit dem hier
// ohnehin kein Handschlag zustande kommt — ein Leser dafür wäre eine Fläche
// ohne Aufrufer.
//
// **Und der Grund ist der Transport und nicht die Herkunft.** Der erste
// Anlauf schrieb, fremde Kennungen erreichten dieses Programm gar nicht; das
// stimmte nicht (REST, MCP, Nachbarschaft, Umschlagurheber), und der Nachlauf
// hat es angestrichen. Eine richtige Entscheidung mit einem falschen Grund
// ist keine (§10).
//
// **Was bleibt, ist der Satz.** „Fängt nicht mit b an" schickt jemanden auf
// die Suche nach einem Tippfehler, den es nicht gibt.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Die ältere Form wird benannt, nicht bloß abgewiesen.**
func TestDieAeltereTextformBekommtIhrenEigenenSatz(t *testing.T) {
	t.Parallel()

	faelle := []string{
		"12D3KooWA5C1Bq2mCkxqFSzcBQJFhqhLsZkuEEvHiUAoBLVMLhRi",
		"QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
	}

	for _, alt := range faelle {
		t.Run(alt[:6], func(t *testing.T) {
			t.Parallel()

			fehler := peer.Pruefe(peer.Kennung(alt), []byte("x"), []byte("y"))
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}

			// **Der Satz nennt die Form, die gemeint ist** — sonst wäre er
			// ein Nein ohne Weg (§10).
			for _, muss := range []string{"Base58", "\"b\"", "Peeridentität"} {
				if !strings.Contains(fehler.Error(), muss) {
					t.Errorf("der Satz nennt %s nicht: %v", muss, fehler)
				}
			}
		})
	}
}

// **Und alles andere bekommt weiterhin den kurzen Satz.**
//
// **Die Gegenprobe**: ohne sie stünde nur, dass ein Satz für einen Fall
// kommt — nicht, dass er für die anderen ausbleibt. Ein Hinweis auf Base58
// über einem Tippfehler wäre eine falsche Fährte.
func TestWasKeineAeltereFormIstBekommtDenKurzenSatz(t *testing.T) {
	t.Parallel()

	for _, krumm := range []string{"", "hallo", "12D3", "qm-klein", "Bxyz"} {
		t.Run(krumm, func(t *testing.T) {
			t.Parallel()

			fehler := peer.Pruefe(peer.Kennung(krumm), []byte("x"), []byte("y"))
			if fehler == nil {
				t.Fatalf("%q kam durch", krumm)
			}

			if strings.Contains(fehler.Error(), "Base58") {
				t.Errorf("%q bekam den Base58-Satz: %v", krumm, fehler)
			}
		})
	}
}

// **Was dieses Programm schreibt, liest es auch** — die Form, die gilt.
func TestDieEigeneFormWirdGelesen(t *testing.T) {
	t.Parallel()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if !strings.HasPrefix(string(ich.Kennung()), "b") {
		t.Fatalf("die eigene Kennung ist %q", ich.Kennung())
	}

	if fehler := peer.Pruefe(ich.Kennung(),
		[]byte("x"), ich.Signiere([]byte("x"))); fehler != nil {
		t.Errorf("die eigene Kennung wurde abgewiesen: %v", fehler)
	}
}
