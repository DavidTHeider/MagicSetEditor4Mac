package service

// **Die Partie über eine offene Sitzung** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Was hier dazukommt
//
// `Partiesitzung` verpackt und schickt, `Partieempfang` urteilt über das, was
// ankommt — **niemand hat bisher gelauscht**. Diese Stelle ist die Schleife
// dazwischen: sie liest von der Leitung, hält jede Nachricht gegen den Stand
// und reicht weiter, was gilt.
//
// # Die vier Urteile und was sie hier bewirken
//
// Das Issue sagt es in einem Satz: „Doppelte/geänderte IDs, Lücken, falsche
// Hashes und alte Epochs **blockieren statt weiterzuspielen**." Daraus folgt
// je Urteil etwas anderes, und der Unterschied ist der ganze Punkt:
//
//   - **Neu** — weiterreichen.
//   - **Schon da** — **still fallen lassen**. Eine zweite Zustellung ist auf
//     einer Leitung, die wiederholt, der Normalfall; sie darf nur nicht
//     zweimal wirken („erzeugen keinen doppelten Draw").
//   - **Lücke** — **anhalten**. Weiterzuspielen hieße, einen Stand
//     fortzuschreiben, dem ein Schritt fehlt. Der Ruf endet, und der Aufrufer
//     nimmt über [domain.Wiederaufnahme] wieder auf.
//   - **Alte Epoche** — fallen lassen **und melden**. Sie ist kein
//     Normalfall: sie heißt, dass am anderen Ende jemand einen Stand hat, den
//     es hier nicht mehr gibt.
//
// # Was sie **nicht** hält
//
//   - **Nicht die Kette.** Sie kommt mit und wird fortgeschrieben; **geprüft**
//     wird sie dort, wo das Ereignis wieder entsteht — nur dort liegt der
//     öffentliche Teil vor, über den [domain.Ereignis.Abdruck] rechnet. Wer
//     hier prüfte, prüfte eine Zeichenkette gegen sich selbst.
//   - **Nicht, was eine Nachricht bedeutet.** Ob aus einem Ereignis ein
//     gezogenes Blatt wird, entscheidet die Partieleitung; ob ein Chat
//     angezeigt wird, das Fenster. Diese Stelle kennt vier Arten und keine
//     Fachregel.
//   - **Nicht den Verlust des Wirts.** Dass die Leitung abreißt, gibt sie als
//     Fehler zurück; **was das für die Partie heißt**, sagt der Gast
//     (`app.App.tischVerloren`) — und das ist eine fachliche Entscheidung
//     und keine des Transports.

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Partiezugang ist die offene Sitzung, über die eine Partie läuft.
//
// **Eine Schnittstelle und nicht `*Sitzungsgriff`** — so lässt sich die
// Schleife prüfen, ohne eine Leitung aufzubauen (§2). `*Sitzungsgriff`
// erfüllt sie.
type Partiezugang interface {
	Partienaht

	// Empfange liest die nächste Nachricht — blockierend bis zum Kontext.
	Empfange(ctx context.Context) ([]byte, error)

	// Empfangen sagt, ob eine eingetroffene Nachricht noch gilt.
	Empfangen(roh []byte) bool

	// Leere schickt hinaus, was wartet.
	//
	// # Warum die Schleife sie ruft und nicht der Bote
	//
	// **Weil `Sende` einreiht und nicht schickt.** Der Griff sammelt, damit
	// eine volle Warteschlange den Absender erreicht statt still zu
	// wachsen — geleert wird auf Ruf. Ohne diesen Ruf legt der Wirt seine
	// Ereignisse und das Tischbild in die Schlange, und der Gast wartet auf
	// etwas, das schon fertig ist.
	//
	// **Genau das ist beim ersten Anlauf passiert**, und gefunden hat es
	// nicht ein Test neben dem Code, sondern der Lauf über beide Enden:
	// der Zug kam beim Wirt an, und beim Gast kam nichts zurück.
	Leere(ctx context.Context) error
}

// Angenommene ist eine Sitzung, wie eine Aufnahme sie vorfindet.
//
// **Zugang plus Auskunft plus Ende.** Wer eine hereinkommende Sitzung
// bedient, muss lesen und schreiben können, wissen **wer** und **wozu** — und
// sie niederlegen, wenn es nichts zu tun gibt.
//
// **Eine Schnittstelle und nicht `*Sitzungsgriff`**: so lässt sich eine
// Aufnahme prüfen, ohne zwei Identitäten und eine Leitung aufzubauen (§2) —
// und die Schicht darüber bekommt nicht den ganzen Griff in die Hand, sondern
// das, was sie braucht.
type Angenommene interface {
	Partiezugang

	Stand() Sitzungsstand
	Schliesse() error
}

// **Und der Griff erfüllt beide.**
//
// **Eine Zusicherung beim Übersetzen und kein Lauf**: sie kostet nichts, sie
// kann nicht flackern, und sie bricht **dort**, wo jemand eine der drei
// Signaturen ändert — statt weiter unten in einem Aufrufer, der dann wie der
// Schuldige aussieht.
var _ Angenommene = (*Sitzungsgriff)(nil)

// Partienbote bekommt, was gilt.
//
// **Eine Methode und keine vier.** Die Aufteilung nach Art gehört dem
// Verbraucher: was mit einem Chat geschieht, weiß das Fenster, und was mit
// einem Ereignis geschieht, die Partieleitung. Vier Methoden zwängen jeden
// Verbraucher, drei leere zu schreiben — und die Art steht ohnehin in der
// Nachricht.
type Partienbote interface {
	Angekommen(ctx context.Context, nachricht domain.Partienachricht) error
}

// Partienetz lauscht auf einer Sitzung und reicht weiter, was gilt.
type Partienetz struct {
	zugang  Partiezugang
	empfang *Partieempfang
	bote    Partienbote

	melden func(error)
}

// NeuesPartienetz baut die Schleife.
func NeuesPartienetz(
	zugang Partiezugang, empfang *Partieempfang, bote Partienbote,
) *Partienetz {
	return &Partienetz{zugang: zugang, empfang: empfang, bote: bote}
}

// MitMeldung setzt, wohin Befunde gehen.
//
// **Befunde und keine Fehler** (§8): eine unlesbare Zustellung und eine alte
// Epoche halten die Sitzung nicht an, aber sie verschwinden auch nicht.
func (n *Partienetz) MitMeldung(melden func(error)) *Partienetz {
	n.melden = melden

	return n
}

// Bediene liest, bis der Kontext endet, die Leitung abreißt oder eine Lücke
// aufgeht.
//
// **Ein Ruf und keine Goroutine** — derselbe Grund wie bei
// [Sitzungsgriff.Leere]: wer hier eine startete, hätte einen Lauf, der auf
// echte Zeit wartet (§3). Wer sie nebenher haben will, startet sie selbst und
// weiß dann auch, wem sie gehört.
func (n *Partienetz) Bediene(ctx context.Context) error {
	for {
		if fehler := ctx.Err(); fehler != nil {
			return fehler
		}

		roh, fehler := n.zugang.Empfange(ctx)
		if fehler != nil {
			return fehler
		}

		// **Was zu einem Stand gehört, den es nicht mehr gibt, fällt still
		// weg** (§7, Regel 2). Ein Fehler zwänge jeden Aufrufer, ihn
		// wegzufangen — und ein Profilwechsel ist kein Fehlschlag.
		if !n.zugang.Empfangen(roh) {
			continue
		}

		ausgang, fehler := n.empfang.Nimm(roh)
		if fehler != nil {
			// **Eine unlesbare Zustellung hält die Sitzung nicht an.** Sie
			// ist ein Befund über **eine** Nachricht; der Strom daneben ist
			// unberührt, und wer dafür die Leitung niederlegte, ließe ein
			// fremdes Programm die Partie beenden (§8).
			n.melde(fehler)

			continue
		}

		if fehler := n.urteilen(ctx, ausgang); fehler != nil {
			return fehler
		}

		// **Was der Bote hinausgelegt hat, geht hinaus** — bevor die nächste
		// Nachricht gelesen wird. Der Grund steht bei [Partiezugang.Leere].
		if fehler := n.zugang.Leere(ctx); fehler != nil {
			return fehler
		}
	}
}

// urteilen tut, was das Urteil verlangt.
func (n *Partienetz) urteilen(
	ctx context.Context, ausgang Empfangsausgang,
) error {
	switch ausgang.Urteil {
	case domain.FolgeNeu:
		return n.bote.Angekommen(ctx, ausgang.Nachricht)

	case domain.FolgeSchonDa:
		// **Still, und das ist die Zusage.** Wiederholung ist der Normalfall.
		return nil

	case domain.FolgeLuecke:
		return ausgang.Urteil.Fehler(
			ausgang.Nachricht.Epoche, ausgang.Nachricht.Folge)

	case domain.FolgeAlteEpoche:
		n.melde(ausgang.Urteil.Fehler(
			ausgang.Nachricht.Epoche, ausgang.Nachricht.Folge))

		return nil
	}

	// **Ein Urteil, das es nicht gibt, hält an.** Wer es fallen ließe, hätte
	// bei der nächsten neuen Art eine Nachricht, die still verschwindet.
	return fmt.Errorf("%w: unbekanntes Urteil %q",
		domain.ErrValidation, ausgang.Urteil)
}

// melde gibt einen Befund weiter, wenn jemand zuhört.
func (n *Partienetz) melde(fehler error) {
	if n.melden != nil && !errors.Is(fehler, context.Canceled) {
		n.melden(fehler)
	}
}
