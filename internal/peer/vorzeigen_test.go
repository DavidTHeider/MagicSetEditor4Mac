package peer_test

// **Wie sich eine Identität einem Transport vorzeigt** — Werkstück N2 von
// Phase 5.4
// (#1455).
//
// # Wofür die zwei Wege gebraucht werden
//
// **Für TLS, und der Grund steht in
// [0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md).** Dieser
// Stapel unterschreibt und verschlüsselt nicht; libp2p hätte Noise oder TLS
// mitgebracht, und ohne libp2p gingen Kartenbestände im Klartext hinaus.
//
// **Ein TLS-Zeugnis braucht zweierlei**: etwas, das unterschreibt, ohne den
// Schlüssel herzugeben — und einen Weg von einem öffentlichen Schlüssel zu der
// Kennung, gegen die festgenagelt wird. Beides gehört hierher, denn hier steht
// schon die Rechnung von der einen Form in die andere.

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// **Die Kennung zu einem öffentlichen Schlüssel ist die eigene.**
//
// **Die Gegenrichtung ist der Punkt.** Aus einem TLS-Zeugnis kommt ein
// Schlüssel heraus, und die Frage lautet dann: gehört er zu dem, den ich
// erwarte? Ohne diesen Weg müsste die Antwort geraten werden.
func TestDieKennungZuEinemSchluesselIstDieEigene(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	gerechnet, fehler := peer.KennungFuer(ich.Oeffentlich())
	if fehler != nil {
		t.Fatalf("KennungFuer: %v", fehler)
	}

	if gerechnet != ich.Kennung() {
		t.Fatalf("gerechnet ist %q, eigen ist %q", gerechnet, ich.Kennung())
	}
}

// **Was kein Ed25519-Schlüssel ist, bekommt keine Kennung.**
//
// **Sonst entstünde eine, die aussieht wie eine.** Ein Zeugnis mit einem
// RSA-Schlüssel gäbe sonst eine Kennung, die zu niemandem gehört — und der
// Vergleich dagegen schlüge fehl, ohne zu sagen, warum.
func TestWasKeinEd25519SchluesselIstBekommtKeineKennung(t *testing.T) {
	t.Parallel()

	for name, schluessel := range map[string]ed25519.PublicKey{
		"gar keiner":      nil,
		"zu kurz":         make([]byte, ed25519.PublicKeySize-1),
		"zu lang":         make([]byte, ed25519.PublicKeySize+1),
		"ganz leer":       {},
		"RSA-verdaechtig": make([]byte, 256),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := peer.KennungFuer(schluessel); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Der Signierer unterschreibt, ohne den Schlüssel herzugeben.**
//
// **`crypto.Signer` ist die Form, die TLS erwartet** — und sie reicht einen
// öffentlichen Schlüssel samt einer Unterschriftsfunktion, nicht das
// Geheimnis.
func TestDerSigniererUnterschreibtFuerDieEigeneKennung(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	signierer := ich.Signierer()

	oeffentlich, taugt := signierer.Public().(ed25519.PublicKey)
	if !taugt {
		t.Fatalf("der Signierer zeigt %T", signierer.Public())
	}

	if !oeffentlich.Equal(ich.Oeffentlich()) {
		t.Fatal("der Signierer gehört zu einem anderen Schlüssel")
	}

	// **Und seine Unterschrift gilt unter der eigenen Kennung** — gemessen
	// über denselben Weg, den die Hülle geht.
	//
	// **`crypto.Hash(0)` und nicht `nil`**: Ed25519 unterschreibt die
	// Nachricht selbst und keinen Streuwert, und [ed25519.PrivateKey.Sign]
	// liest dafür `opts.HashFunc()`. Der erste Anlauf gab `nil` und stürzte
	// mit einem Zeiger ins Leere — die Form gehört hierher, weil sie an der
	// Signatur nicht abzulesen ist.
	unterschrift, fehler := signierer.Sign(
		rand.Reader, []byte("hallo"), crypto.Hash(0))
	if fehler != nil {
		t.Fatalf("Sign: %v", fehler)
	}

	if fehler := peer.Pruefe(
		ich.Kennung(), []byte("hallo"), unterschrift); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}
}
