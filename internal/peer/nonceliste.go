package peer

// **Zufallszahlen und ihr Gedächtnis** — Werkstück N1 von Phase 5.4
// (#1454).

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Noncelaenge ist das Gewicht einer Zufallszahl in Byte.
//
// **32 Byte, und nicht weniger.** Sie soll nicht zu erraten und nicht zweimal
// zu treffen sein; bei 256 Bit ist beides kein Thema mehr, und der Preis sind
// 64 Zeichen in einer Nachricht.
const Noncelaenge = 32

// NoncelisteHoechstens ist, wie viele Zufallszahlen die Liste höchstens hält
// — bei 64 Zeichen je Zahl rund 15 MB.
const NoncelisteHoechstens = 100_000

// Nonceliste merkt sich, welche Zufallszahlen schon da waren.
//
// # Wogegen sie schützt — und wogegen der Kanal schützt
//
// **Sie fängt das Wiedereinspielen auf derselben Verbindung.** Wer eine
// mitgeschnittene, gültig unterschriebene Nachricht ein zweites Mal schickt,
// kommt an ihr nicht vorbei.
//
// **Über Verbindungen hinweg tut es die Kanalbindung.** Ein Angebot trägt den
// Kanal, auf dem es entstand; auf einer anderen Verbindung passt es nicht
// mehr. **Das ist eine Anforderung an den Transportadapter** (N1.5, N2): er
// muss eine Bindung liefern, die je Verbindung eine andere ist. Liefert er
// eine feste Zeichenkette, trägt allein diese Liste — und nur so lange, wie
// ihre Frist reicht.
//
// **Deshalb darf die Liste vergessen.** Ohne Frist wüchse sie unbegrenzt, und
// ein Peer, der viele Angebote schickt, füllte den Speicher eines Rechners,
// der ihm noch gar nicht traut.
type Nonceliste struct {
	schloss sync.Mutex
	gesehen map[string]time.Time
	uhr     func() time.Time
	frist   time.Duration

	// reihe hält die Zufallszahlen in der Reihenfolge, in der sie kamen —
	// so räumt die Frist von vorn, ohne die ganze Liste zu lesen.
	reihe []string

	// hoechstens ist die Obergrenze; was darüber käme, verdrängt die
	// älteste (Zweitblick über Phase 6.7, Z3).
	hoechstens int
}

// NeueNonceliste gibt eine mit eingesetzter Uhr.
//
// **Eingesetzt und nicht genommen** (§16): sonst bräuchte jeder Lauf über eine
// Frist echte Zeit, und ein Test, der wartet, ist ein Test, der flackert.
func NeueNonceliste(uhr func() time.Time, frist time.Duration) *Nonceliste {
	return &Nonceliste{
		gesehen:    map[string]time.Time{},
		uhr:        uhr,
		frist:      frist,
		hoechstens: NoncelisteHoechstens,
	}
}

// Merke trägt eine Zufallszahl ein und weist die zweite ab.
func (l *Nonceliste) Merke(nonce string) error {
	l.schloss.Lock()
	defer l.schloss.Unlock()

	jetzt := l.uhr()

	for len(l.reihe) > 0 && jetzt.Sub(l.gesehen[l.reihe[0]]) > l.frist {
		delete(l.gesehen, l.reihe[0])
		l.reihe = l.reihe[1:]
	}

	if _, schon := l.gesehen[nonce]; schon {
		return fmt.Errorf(
			"%w: diese Zufallszahl war schon einmal da",
			domain.ErrValidation)
	}

	// **Über der Grenze geht die älteste** — die Kanalbindung trägt den
	// Schutz vor einem weitergereichten Handschlag; diese Liste ist die
	// zweite Reihe, und eine, die ein Fremder beliebig wachsen ließe, wäre
	// ein Weg in den Speicher.
	//
	// **Und keine strikte Frist** (SEC-15, Vermerk 0176): eine volle Liste,
	// die jedes neue Angebot abwiese, sperrte jeden Kontakt aus, sobald ein
	// Fremder sie füllt. Die Zusage ist damit genau: keine Zahl zweimal in
	// einer Stunde, **solange in dieser Stunde weniger als
	// [NoncelisteHoechstens] neue kamen**.
	if len(l.reihe) >= l.hoechstens {
		delete(l.gesehen, l.reihe[0])
		l.reihe = l.reihe[1:]
	}

	l.gesehen[nonce] = jetzt
	l.reihe = append(l.reihe, nonce)

	return nil
}

// eineZufallszahl zieht eine neue.
//
// **`nil` heißt: die Quelle des Betriebssystems** — dieselbe Regel wie bei
// [Erzeugen], und aus demselben Grund: der natürliche Ruf eines Aufrufers, der
// nichts einsetzen will, ist `nil`. Bis N2 stürzte er hier ab, weil
// `io.ReadFull` mit einem leeren Leser nicht umgehen kann; aufgefallen ist es,
// als `service` den Handschlag zum ersten Mal wirklich führte
// (#1455).
func eineZufallszahl(zufall io.Reader) (string, error) {
	if zufall == nil {
		zufall = rand.Reader
	}

	roh := make([]byte, Noncelaenge)

	if _, fehler := io.ReadFull(zufall, roh); fehler != nil {
		return "", domain.MitGrund(domain.ErrConflict, fehler,
			"die Zufallsquelle gab nichts her")
	}

	return hex.EncodeToString(roh), nil
}

// pruefeZufallszahl hält eine fremde gegen ihre Form.
func pruefeZufallszahl(nonce string) error {
	roh, fehler := hex.DecodeString(nonce)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Zufallszahl ist kein Hex")
	}

	if len(roh) != Noncelaenge {
		return fmt.Errorf(
			"%w: eine Zufallszahl wiegt %d Byte, hier sind es %d",
			domain.ErrValidation, Noncelaenge, len(roh))
	}

	return nil
}
