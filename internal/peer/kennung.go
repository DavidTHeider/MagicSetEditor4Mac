package peer

// **Die Textform einer Peerkennung** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum eine eigene Datei
//
// **Weil sie eine andere Frage beantwortet als `peer.go`.** Dort steht, was
// eine Identität **kann**; hier, wie sie **heißt** — und das ist eine fremde
// Festlegung, kein eigener Entwurf. Die Fundstellen stehen im Paketkopf.
//
// # Der Bauplan in einer Zeile
//
//	b + base32( 0x01 | 0x72 | 0x00 | 36 | 0x08 0x01 0x12 0x20 | schluessel )
//	    └CIDv1┘ └key┘ └identity-multihash┘ └─Protobuf PublicKey─┘
//
// **Vier Hüllen um 32 Byte**, und jede hat ihren Grund: das Protobuf sagt,
// **welche** Art Schlüssel es ist; der Multihash sagt, dass **nicht gehasht**
// wurde; der CID sagt, dass es ein **libp2p-Schlüssel** ist; die Multibase
// sagt, in welcher Schrift der Rest steht.
//
// # Gelesen wird **eine** Textform, und das ist eine Entscheidung
//
// **libp2p kennt zwei**: diese hier — CIDv1 in Multibase `b` — und die
// ältere, rohe Base58btc-Form, die als `12D3Koo…` oder `Qm…` dasteht. Dieses
// Paket **schreibt** die erste und **liest** nur sie.
//
// **Der Grund ist kein technischer, sondern einer über die Gegenstellen.**
// Der Transport dieses Projekts ist seit
// [0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md) ausdrücklich
// **nicht** libp2p — es ist TCP mit TLS 1.3, und über fremde Netze tailcat.
// Wer `12D3Koo…` aus einem libp2p-Werkzeug einfügte, hätte ein Gegenüber, mit
// dem hier ohnehin kein Handschlag zustande kommt: ein Leser dafür wäre eine
// Fläche ohne Aufrufer.
//
// **Eine fremde Kennung kommt** über REST (`internal/rest/peerkontakt.go`),
// über MCP beim Sperren, aus der Nachbarschaft im Netz und als Urheber eines
// Umschlags — alle von **diesem Programm auf einem anderen Rechner**, und das
// schreibt die Form, die hier gelesen wird. Der tragende Grund ist trotzdem
// der Transport, nicht die Herkunft der Zeichenkette.
//
// **Was der Preis ist, steht im Fehlersatz** (§10: ein Nein braucht Deckung
// **und** einen Weg): wer die andere Form einfügt, bekommt gesagt, welche
// gemeint ist — nicht bloß, dass seine nicht passt. Ein Wächter hält den
// Satz, damit er nicht zu „ist keine Peerkennung" verkümmert.
//
// **Und wenn ein Gegenüber dazukommt, das sie spricht**, ist das eine neue
// Entscheidung und keine neue Zeile: dann kommt ein Base58-Leser dazu, und
// alles, was hereinkommt, wird **normalisiert** — sonst stünde derselbe Peer
// zweimal in den Kontakten und verglichen würde nichts.

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Bytes der Hülle, in der Reihenfolge, in der sie stehen.
const (
	// multibaseBase32 ist das Präfix `b` der Textform.
	multibaseBase32 = "b"

	// cidVersion1 ist die erste Zahl des CID.
	cidVersion1 = 0x01

	// multicodecLibp2pKey ist `libp2p-key`.
	multicodecLibp2pKey = 0x72

	// multihashIdentity heißt: hier wurde **nicht** gehasht.
	multihashIdentity = 0x00

	// protobufTypEd25519 ist `Type = 1` als Feld 1, Varint — `0x08 0x01`.
	protobufTypEd25519 = "\x08\x01"

	// protobufDaten32 ist `Data` als Feld 2, längenbehaftet, 32 Byte —
	// `0x12 0x20`.
	protobufDaten32 = "\x12\x20"

	// protobufLaenge ist, wie schwer der kodierte Schlüssel wiegt: vier Byte
	// Hülle und 32 Byte Schlüssel.
	//
	// **Eine Konstante und kein `len(...)`**: die Zahl steht fest, und eine
	// Umrechnung `int -> byte` an dieser Stelle liest gosec als möglichen
	// Überlauf (G115) — zu Recht, denn die Stelle allein sagt nicht, dass
	// die Länge nie über 255 geht. **Unter 42 muss sie ohnehin bleiben**,
	// sonst wäre es kein identity-Multihash mehr.
	protobufLaenge = 4 + ed25519.PublicKeySize
)

// schrift ist die Multibase `base32`: RFC 4648, klein, ohne Auffüllung.
//
// **`base32.StdEncoding` schreibt groß** — die Multibase `b` ist die kleine
// Schreibweise, und `B` wäre die große. Gesenkt wird deshalb hier und nicht
// beim Leser.
var schrift = base32.StdEncoding.WithPadding(base32.NoPadding)

// alsProtobuf gibt den öffentlichen Schlüssel in libp2ps Hülle.
func alsProtobuf(oeffentlich ed25519.PublicKey) []byte {
	hülle := make([]byte, 0, 4+ed25519.PublicKeySize)
	hülle = append(hülle, protobufTypEd25519...)
	hülle = append(hülle, protobufDaten32...)

	return append(hülle, oeffentlich...)
}

// kennungAus rechnet die Textform aus einem öffentlichen Schlüssel.
func kennungAus(oeffentlich ed25519.PublicKey) Kennung {
	geschluesselt := alsProtobuf(oeffentlich)

	roh := make([]byte, 0, 4+len(geschluesselt))
	roh = append(roh, cidVersion1, multicodecLibp2pKey,
		multihashIdentity, protobufLaenge)
	roh = append(roh, geschluesselt...)

	return Kennung(multibaseBase32 + strings.ToLower(schrift.EncodeToString(roh)))
}

// schluesselAus holt den öffentlichen Schlüssel aus einer Kennung zurück.
//
// **Jede Hülle wird geprüft und nicht übersprungen.** Eine Kennung, die
// aussieht wie eine und einen anderen Schlüsseltyp trägt, ist keine — und wer
// sie durchließe, verglich später zwei Dinge, die nicht dasselbe sind.
func schluesselAus(kennung Kennung) (ed25519.PublicKey, error) {
	text := string(kennung)

	if !strings.HasPrefix(text, multibaseBase32) {
		// **Der Satz nennt die gemeinte Form**, und bei der alten sagt er
		// es ausdrücklich: „fängt nicht mit b an" schickt jemanden auf die
		// Suche nach einem Tippfehler, den es nicht gibt. Der Grund für die
		// eine Form steht im Kopf dieser Datei.
		if altform(text) {
			return nil, fmt.Errorf(
				"%w: %q ist die ältere Base58-Form einer Peerkennung — "+
					"dieses Programm liest die Form, die mit %q anfängt, "+
					"und genau die steht unter „Meine Peeridentität“",
				domain.ErrValidation, kennung, multibaseBase32)
		}

		return nil, fmt.Errorf(
			"%w: %q ist keine Peerkennung — sie fängt nicht mit %q an",
			domain.ErrValidation, kennung, multibaseBase32)
	}

	roh, fehler := schrift.DecodeString(
		strings.ToUpper(text[len(multibaseBase32):]))
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Peerkennung lesen")
	}

	kopf := make([]byte, 0, 4+len(protobufTypEd25519)+len(protobufDaten32))
	kopf = append(kopf, cidVersion1, multicodecLibp2pKey,
		multihashIdentity, protobufLaenge)
	kopf = append(kopf, protobufTypEd25519...)
	kopf = append(kopf, protobufDaten32...)

	if len(roh) != len(kopf)+ed25519.PublicKeySize ||
		!bytes.Equal(roh[:len(kopf)], kopf) {
		return nil, fmt.Errorf(
			"%w: %q trägt keinen Ed25519-Schlüssel in der Gestalt, die "+
				"libp2p für eine Peerkennung festlegt",
			domain.ErrValidation, kennung)
	}

	return ed25519.PublicKey(roh[len(kopf):]), nil
}

// altform sagt, ob ein Text nach der älteren Base58btc-Form aussieht.
//
// **Es ist eine Erkennung und keine Prüfung.** Sie entscheidet, welcher Satz
// zurückkommt, und nicht, ob etwas gültig ist — deshalb reichen die zwei
// Vorsätze, die libp2p für einen Ed25519- und einen RSA-Peer nennt. Wer sie
// zu einer Prüfung ausbaute, hätte einen Leser gebaut, und der wäre die
// Entscheidung, die im Kopf dieser Datei mit Grund nicht getroffen ist.
func altform(text string) bool {
	return strings.HasPrefix(text, "12D3Koo") || strings.HasPrefix(text, "Qm")
}
