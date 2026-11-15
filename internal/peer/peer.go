// Package peer hält die Identität dieses Programms gegenüber anderen.
//
// **Ein Blatt** (§2): es kennt `internal/domain` und sonst nichts aus diesem
// Projekt — kein `service`, kein `storage`, keinen Schlüsselbund und kein
// libp2p. Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum die Kennung die eines libp2p-Peers ist
//
// **Weil sie es war, als dieses Paket entstand**, und weil die Gestalt gut
// ist: sie trägt den Schlüssel, sie ist nachgelesen, und sie ist prüfbar.
// Seit [0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md) vom
// 2026-09-14 setzt kein libp2p-Adapter darauf auf, und damit ist die Gestalt
// eine **Wahl** statt einer Notwendigkeit. Geändert wird sie trotzdem nicht:
// jede bestätigte Kennung zeigte danach ins Leere.
//
// **Die Gestalt ist nachgelesen** (§16, Satz 1), am 2026-09-14 aus
// `libp2p/specs`, `peer-ids/peer-ids.md`:
//
//   - Eine Peerkennung ist der **Multihash des kodierten öffentlichen
//     Schlüssels**. Kodiert wird er als Protobuf `PublicKey{Type, Data}` mit
//     `KeyType.Ed25519 = 1`.
//   - **Unter 42 Byte wird nicht gehasht**, sondern der `identity`-Multihash
//     genommen — der Schlüssel steht dann unverändert in der Kennung. Ein
//     Ed25519-Schlüssel wiegt 32 Byte, mit Protobuf-Hülle 36.
//   - Die **Textform** ist ein CIDv1 mit dem Multicodec `libp2p-key` (0x72),
//     in `base32` ohne Auffüllung und mit dem Multibase-Präfix `b`.
//   - Ein **privater** Ed25519-Schlüssel steht als 64 Byte da: die privaten
//     gefolgt von den öffentlichen. Das ist genau [ed25519.PrivateKey].
//
// # Was daraus folgt, und was es kostet
//
// **Die Kennung enthält den ganzen öffentlichen Schlüssel.** Wer zwei
// Kennungen vergleicht, vergleicht zwei Schlüssel — ein Umweg über eine
// Prüfsumme ist nicht nötig und wäre schwächer. Der [Identitaet.Fingerabdruck]
// ist deshalb eine **Lesehilfe** und kein zweiter Nachweis.
//
// **Was hier nicht geprüft ist**: dass eine echte libp2p-Fassung dieselbe
// Kennung errechnet. Dieses Paket hat keine Abhängigkeit auf libp2p, und seit
// [0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md) bekommt es auch
// keine — der Abgleich hat damit **keinen geplanten Ort mehr**. Für den
// Betrieb braucht es ihn nicht: beide Seiten dieses Programms rechnen dieselbe
// nachgelesene Gestalt. Er stünde an, wenn libp2p je dazukäme.
package peer

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kennung ist die öffentliche Peerkennung in ihrer Textform.
//
// **Ein eigener Typ und kein `string`**: sie reist durch `service` und `app`,
// und eine Zeichenkette neben einer Profilkennung ist beim ersten vertauschten
// Argument nicht mehr zu unterscheiden.
type Kennung string

// Identitaet ist ein Schlüsselpaar samt seiner Kennung.
//
// **Der private Schlüssel steht hier und nirgends sonst.** Wer ihn ablegen
// will, nimmt [Identitaet.Geheimnis]; wo das Geheimnis liegt, entscheidet
// nicht dieses Paket (§2).
type Identitaet struct {
	kennung Kennung
	privat  ed25519.PrivateKey
}

// Erzeugen legt ein neues Schlüsselpaar an.
//
// **Der Zufall wird eingesetzt** (§16) — nicht, weil jemand einen anderen
// wollte, sondern damit ein Lauf denselben Schlüssel zweimal bekommen kann
// und die Kennung prüfbar ist. `nil` heißt: [crypto/rand].
func Erzeugen(zufall io.Reader) (Identitaet, error) {
	if zufall == nil {
		zufall = rand.Reader
	}

	oeffentlich, privat, fehler := ed25519.GenerateKey(zufall)
	if fehler != nil {
		return Identitaet{}, domain.MitGrund(domain.ErrValidation, fehler,
			"ein Schlüsselpaar erzeugen")
	}

	return Identitaet{kennung: kennungAus(oeffentlich), privat: privat}, nil
}

// Geheimnis gibt den privaten Schlüssel in seiner ablegbaren Form.
//
// **Hex und nicht roh**: der Schlüsselbund nimmt Text
// ([schluesselbund.Bund.Schreibe]), und 64 Byte werden dabei 128 Zeichen —
// weit unter seiner Grenze.
func (i Identitaet) Geheimnis() string {
	return hex.EncodeToString(i.privat)
}

// Kennung gibt die öffentliche Kennung.
func (i Identitaet) Kennung() Kennung { return i.kennung }

// Oeffentlich gibt den öffentlichen Schlüssel.
//
// **Geschnitten und nicht zugesichert.** `ed25519.PrivateKey.Public` gäbe
// dasselbe, aber als `crypto.PublicKey` — und die Zusicherung zurück wäre ein
// Zweig, den kein Lauf betreten kann. Die Form steht fest: ein privater
// Ed25519-Schlüssel sind die privaten Bytes gefolgt von den öffentlichen
// (`libp2p/specs`, `peer-ids/peer-ids.md`, und dieselbe Festlegung in
// [ed25519.PrivateKey]).
func (i Identitaet) Oeffentlich() ed25519.PublicKey {
	return ed25519.PublicKey(i.privat[ed25519.SeedSize:])
}

// AusGeheimnis stellt eine Identität aus ihrem abgelegten Geheimnis her.
//
// **Sie wird nicht geraten.** Eine falsche Länge oder ein krummes Hex sind ein
// Befund und keine neue Identität — wer aus einem beschädigten Geheimnis still
// ein frisches Schlüsselpaar machte, hätte den Peer unbemerkt getauscht, und
// jeder bestätigte Fingerabdruck zeigte ins Leere.
func AusGeheimnis(geheimnis string) (Identitaet, error) {
	roh, fehler := hex.DecodeString(geheimnis)
	if fehler != nil {
		return Identitaet{}, domain.MitGrund(domain.ErrValidation, fehler,
			"das abgelegte Geheimnis lesen")
	}

	if len(roh) != ed25519.PrivateKeySize {
		return Identitaet{}, fmt.Errorf(
			"%w: das abgelegte Geheimnis wiegt %d Byte und nicht %d — es "+
				"gehört zu keinem Ed25519-Schlüssel",
			domain.ErrValidation, len(roh), ed25519.PrivateKeySize)
	}

	privat := ed25519.PrivateKey(roh)

	ich := Identitaet{privat: privat}
	ich.kennung = kennungAus(ich.Oeffentlich())

	return ich, nil
}

// Signierer gibt die Identität in der Form, die TLS erwartet.
//
// # Warum ein `crypto.Signer` und kein Schlüssel
//
// **Weil ein TLS-Zeugnis unterschrieben werden muss, nicht ausgelesen.**
// `x509.CreateCertificate` nimmt einen Signierer und ruft ihn; wer ihm das
// Geheimnis gäbe, hätte es einmal mehr in einer Veränderlichen liegen, ohne
// etwas zu gewinnen.
//
// **Sie ist da, weil dieser Stapel nicht verschlüsselt.** Der Grund und sein
// Preis stehen in
// [0113](../../docs/entscheidungen/0113-kein-libp2p-im-lan.md).
func (i Identitaet) Signierer() crypto.Signer { return i.privat }

// KennungFuer rechnet die Kennung zu einem öffentlichen Schlüssel.
//
// # Wofür die Gegenrichtung gebraucht wird
//
// **Für das Festnageln.** Aus einem TLS-Zeugnis kommt ein Schlüssel heraus,
// und die Frage lautet dann: gehört er zu der Kennung, die ich erwarte? Ohne
// diesen Weg bliebe nur, das Zeugnis einer Zertifizierungsstelle zu glauben —
// und eine solche gibt es hier nicht und soll es nicht geben.
//
// **Was kein Ed25519-Schlüssel ist, bekommt keine Kennung.** Sonst entstünde
// eine, die aussieht wie eine und zu niemandem gehört; der Vergleich dagegen
// schlüge fehl, ohne zu sagen, warum.
func KennungFuer(oeffentlich ed25519.PublicKey) (Kennung, error) {
	if len(oeffentlich) != ed25519.PublicKeySize {
		return "", fmt.Errorf(
			"%w: ein öffentlicher Ed25519-Schlüssel wiegt %d Byte und nicht %d",
			domain.ErrValidation, ed25519.PublicKeySize, len(oeffentlich))
	}

	return kennungAus(oeffentlich), nil
}

// Signiere unterschreibt, was hinausgeht.
func (i Identitaet) Signiere(daten []byte) []byte {
	return ed25519.Sign(i.privat, daten)
}

// Pruefe hält eine Unterschrift gegen die Kennung, die sie behauptet.
//
// **Die Kennung trägt den Schlüssel** — deshalb braucht es hier keinen
// zweiten Weg, ihn zu holen, und keine Liste bekannter Peers: was nicht zu
// seiner eigenen Kennung passt, ist keine gültige Unterschrift.
func Pruefe(kennung Kennung, daten, unterschrift []byte) error {
	oeffentlich, fehler := schluesselAus(kennung)
	if fehler != nil {
		return fehler
	}

	if !ed25519.Verify(oeffentlich, daten, unterschrift) {
		return fmt.Errorf(
			"%w: die Unterschrift gehört nicht zu %s", domain.ErrValidation,
			kennung)
	}

	return nil
}

// Fingerabdruck ist die Kennung in einer Form, die zwei Menschen vergleichen.
//
// # Warum es ihn gibt, obwohl die Kennung schon der Schlüssel ist
//
// **Weil niemand 52 Zeichen am Telefon vorliest.** Das Issue verlangt eine
// Bestätigung des Gegenübers, und ein Anzeigename ist keine; verglichen wird
// etwas, das aus dem Schlüssel entsteht.
//
// # Was er kostet
//
// **128 Bit von 256.** Wer einen zweiten Schlüssel mit demselben Abdruck
// finden will, sucht ein zweites Urbild über 128 Bit — nicht über 256. Das ist
// die Reichweite dieser Zahl (§10), und sie ist die Grenze der Bestätigung
// und nicht die der Verbindung: gebunden wird an die **ganze** Kennung.
func (i Identitaet) Fingerabdruck() string { return Fingerabdruck(i.kennung) }

// Fingerabdruck gibt den Abdruck einer fremden Kennung.
func Fingerabdruck(kennung Kennung) string {
	oeffentlich, fehler := schluesselAus(kennung)
	if fehler != nil {
		// **Eine unlesbare Kennung hat keinen Abdruck** — und einen zu
		// erfinden hieße, zwei kaputte gleich aussehen zu lassen.
		return ""
	}

	summe := sha256.Sum256(alsProtobuf(oeffentlich))

	gruppen := make([]string, 0, 8)
	for i := 0; i < 16; i += 2 {
		gruppen = append(gruppen, hex.EncodeToString(summe[i:i+2]))
	}

	return strings.Join(gruppen, "-")
}
