package peer_test

// **Die Peeridentität** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass die Kennung die eines libp2p-Peers ist**, und zwar Byte für Byte
// gegen die Festlegung aus `libp2p/specs`, `peer-ids/peer-ids.md` (nachgelesen
// am 2026-09-14). Ein Lauf, der nur den Rundgang prüfte — erzeugen, ablegen,
// zurücklesen —, wäre grün, gleich welche Gestalt sich das Paket ausdenkt.
//
// **Und dass eine beschädigte Ablage keine neue Identität ergibt.** Das ist
// die gefährliche Richtung: wer aus einem krummen Geheimnis still ein frisches
// Schlüsselpaar machte, tauschte den Peer aus, und jeder bestätigte
// Fingerabdruck zeigte danach ins Leere.

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base32"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// festerZufall gibt immer dieselben Bytes.
//
// **Damit die Kennung eine feste ist**: der Vergleich gegen die Festlegung
// braucht einen Schlüssel, den auch ein zweiter Lauf hat.
type festerZufall struct{ stand byte }

func (f *festerZufall) Read(ziel []byte) (int, error) {
	for i := range ziel {
		ziel[i] = f.stand
		f.stand++
	}

	return len(ziel), nil
}

// eineIdentitaet gibt eine mit festem Schlüssel.
func eineIdentitaet(t *testing.T) peer.Identitaet {
	t.Helper()

	ich, fehler := peer.Erzeugen(&festerZufall{})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return ich
}

// **Die Kennung trägt genau die Hüllen, die libp2p festlegt.**
//
// **Der Fund**: ein anderer Multicodec, eine andere Multibase oder ein
// gehashter statt eines identity-Multihash — jedes davon ergäbe eine Kennung,
// die aussieht wie eine und auf einen anderen Peer zeigt.
func TestDieKennungTraegtDieHuellenAusDerFestlegung(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)
	kennung := string(ich.Kennung())

	// **Multibase `b`**: base32, RFC 4648, klein, ohne Auffüllung.
	if !strings.HasPrefix(kennung, "b") {
		t.Fatalf("die Kennung fängt mit %q an", kennung[:1])
	}

	if kennung != strings.ToLower(kennung) {
		t.Errorf("die Kennung ist nicht durchgehend klein: %q", kennung)
	}

	roh, fehler := base32.StdEncoding.WithPadding(base32.NoPadding).
		DecodeString(strings.ToUpper(kennung[1:]))
	if fehler != nil {
		t.Fatalf("die Kennung ist kein base32: %v", fehler)
	}

	// **CIDv1, `libp2p-key`, identity-Multihash über 36 Byte, dann das
	// Protobuf `PublicKey{Type: Ed25519, Data: …}`.**
	erwartet := []byte{0x01, 0x72, 0x00, 0x24, 0x08, 0x01, 0x12, 0x20}

	if len(roh) != len(erwartet)+ed25519.PublicKeySize {
		t.Fatalf("die Kennung wiegt %d Byte, erwartet %d",
			len(roh), len(erwartet)+ed25519.PublicKeySize)
	}

	if !bytes.Equal(roh[:len(erwartet)], erwartet) {
		t.Errorf("die Hüllen sind % x, erwartet % x", roh[:len(erwartet)], erwartet)
	}

	// **Und dahinter steht der Schlüssel selbst** — nicht sein Hash. Das ist
	// die Aussage des identity-Multihash, und sie ist der Grund, warum ein
	// Vergleich zweier Kennungen ein Vergleich zweier Schlüssel ist.
	if !bytes.Equal(roh[len(erwartet):], ich.Oeffentlich()) {
		t.Error("hinter den Hüllen steht nicht der öffentliche Schlüssel")
	}
}

// **Dieselbe Ablage gibt dieselbe Identität.**
func TestDasGeheimnisStelltDieselbeIdentitaetHer(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	zurueck, fehler := peer.AusGeheimnis(ich.Geheimnis())
	if fehler != nil {
		t.Fatalf("AusGeheimnis: %v", fehler)
	}

	if zurueck.Kennung() != ich.Kennung() {
		t.Fatalf("zurück kam %s, abgelegt war %s",
			zurueck.Kennung(), ich.Kennung())
	}

	// **Und sie unterschreibt wie die erste** — eine Kennung allein sagte
	// nicht, dass auch der private Schlüssel derselbe ist.
	if fehler := peer.Pruefe(
		ich.Kennung(), []byte("halt"), zurueck.Signiere([]byte("halt")),
	); fehler != nil {
		t.Errorf("die zurückgelesene Identität unterschreibt anders: %v", fehler)
	}
}

// **Das Geheimnis ist die Form, die libp2p für Ed25519 bevorzugt.**
//
// 64 Byte: die privaten, dann die öffentlichen — und das ist genau
// `ed25519.PrivateKey`.
func TestDasGeheimnisWiegtVierundsechzigByte(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	if laenge := len(ich.Geheimnis()); laenge != 2*ed25519.PrivateKeySize {
		t.Fatalf("das Geheimnis hat %d Zeichen, erwartet %d",
			laenge, 2*ed25519.PrivateKeySize)
	}
}

// **Eine beschädigte Ablage gibt keine neue Identität, sondern einen Befund.**
func TestEineBeschaedigteAblageGibtKeineNeueIdentitaet(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"kein Hex":             "keinhex",
		"zu kurz":              "abcd",
		"leer":                 "",
		"eine Hälfte zu wenig": strings.Repeat("ab", ed25519.PrivateKeySize/2),
	}

	for name, geheimnis := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := peer.AusGeheimnis(geheimnis); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("der Fehler ist %v", fehler)
			}
		})
	}
}

// **Eine fremde Unterschrift wird abgewiesen.**
func TestEineFremdeUnterschriftWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	andere, fehler := peer.Erzeugen(&festerZufall{stand: 99})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if ich.Kennung() == andere.Kennung() {
		t.Fatal("zwei Schlüssel, eine Kennung")
	}

	// **Die Unterschrift der anderen gegen meine Kennung.**
	if fehler := peer.Pruefe(
		ich.Kennung(), []byte("halt"), andere.Signiere([]byte("halt")),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("die fremde Unterschrift kam durch: %v", fehler)
	}

	// **Und eine geänderte Nachricht.**
	if fehler := peer.Pruefe(
		ich.Kennung(), []byte("halt!"), ich.Signiere([]byte("halt")),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("die geänderte Nachricht kam durch: %v", fehler)
	}
}

// **Eine Kennung, die keine ist, prüft nichts.**
func TestEineKennungDieKeineIstPrueftNichts(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)
	unterschrift := ich.Signiere([]byte("halt"))

	echt := string(ich.Kennung())

	faelle := map[string]peer.Kennung{
		"ohne Multibase":    peer.Kennung(echt[1:]),
		"falsche Multibase": peer.Kennung("z" + echt[1:]),
		"kein base32":       "b!!!!",
		"leer":              "",

		// **Eine Kennung mit einem anderen Multicodec** — dieselbe Länge,
		// dieselbe Schrift, ein anderer Sinn.
		"fremder Multicodec": fremderMulticodec(t, ich),
	}

	for name, kennung := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := peer.Pruefe(
				kennung, []byte("halt"), unterschrift,
			); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			if abdruck := peer.Fingerabdruck(kennung); abdruck != "" {
				t.Errorf("eine unlesbare Kennung hat den Abdruck %q", abdruck)
			}
		})
	}
}

// **Der Fingerabdruck ist stabil, lesbar und gehört zu seiner Kennung.**
func TestDerFingerabdruckIstLesbarUndGehoertZurKennung(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)
	abdruck := ich.Fingerabdruck()

	// **Acht Gruppen zu vier** — 128 Bit, und der Paketkopf sagt, was das
	// kostet.
	gruppen := strings.Split(abdruck, "-")
	if len(gruppen) != 8 {
		t.Fatalf("der Abdruck hat %d Gruppen: %q", len(gruppen), abdruck)
	}

	for _, gruppe := range gruppen {
		if len(gruppe) != 4 {
			t.Fatalf("eine Gruppe ist %q", gruppe)
		}
	}

	if abdruck != peer.Fingerabdruck(ich.Kennung()) {
		t.Error("der Abdruck der Identität und der ihrer Kennung sind zwei")
	}

	// **Zwei Schlüssel, zwei Abdrücke.**
	andere, fehler := peer.Erzeugen(&festerZufall{stand: 7})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	if andere.Fingerabdruck() == abdruck {
		t.Error("zwei Identitäten teilen einen Abdruck")
	}
}

// fremderMulticodec baut eine Kennung mit `dag-pb` statt `libp2p-key`.
func fremderMulticodec(t *testing.T, ich peer.Identitaet) peer.Kennung {
	t.Helper()

	roh := make([]byte, 0, 8+ed25519.PublicKeySize)
	roh = append(roh, 0x01, 0x70, 0x00, 0x24, 0x08, 0x01, 0x12, 0x20)
	roh = append(roh, ich.Oeffentlich()...)

	return peer.Kennung("b" + strings.ToLower(
		base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(roh)))
}

// **Und wenn der Zufall versiegt, entsteht keine Identität.**
//
// **Die gefährliche Richtung ist die andere**: ein Schlüsselpaar aus einem
// Zufall, der nichts mehr hergibt, wäre vorhersagbar — und eine Identität,
// die jemand erraten kann, ist keine.
func TestOhneZufallEntstehtKeineIdentitaet(t *testing.T) {
	t.Parallel()

	if _, fehler := peer.Erzeugen(versiegterZufall{}); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// versiegterZufall gibt nichts und sagt warum.
type versiegterZufall struct{}

func (versiegterZufall) Read([]byte) (int, error) {
	return 0, errors.New("der Zufall ist zu Ende")
}

// **Ohne eigenen Zufall nimmt sie den des Betriebssystems.**
//
// **Die Zeile ist die Vorgabe** und nicht die Ausnahme: im Betrieb ruft
// niemand mit einem eigenen Zufall. Ein Lauf, der sie ausließe, prüfte nur
// den Weg der Tests.
func TestOhneEigenenZufallEntstehtEineEchteIdentitaet(t *testing.T) {
	t.Parallel()

	erste, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	zweite, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	// **Zweimal derselbe Schlüssel wäre kein Zufall** — und genau das prüft
	// diese Zeile: dass die Vorgabe eine Quelle nimmt und nicht eine Null.
	if erste.Kennung() == zweite.Kennung() {
		t.Fatal("zwei Läufe gaben dieselbe Identität")
	}
}
