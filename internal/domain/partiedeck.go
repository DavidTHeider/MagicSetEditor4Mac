package domain

// **Das Deck, wie es in dieser Partie gilt** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum ein eigener Abzug und nicht der aus dem Turnier
//
// **Das Issue sagt es in einem Halbsatz**: „Turnier-Deckabzug allein genügt
// nicht." Ein [Abzugsplatz] hält, was jemand **eingereicht** hat — Karte,
// Druck, Menge, Bereich; das ist ein Dokument und reicht für die Regelprüfung
// am Tisch.
//
// **Eine Partie braucht mehr**: sie muss dieselbe Karte in einem Monat noch
// einmal zeichnen können. Dafür gehören die **Definitionen**, die **Seiten**,
// die **Bild-, Schrift- und Zeichenverweise** und die **auflösbaren
// Abhängigkeiten** dazu — sonst sieht ein Wiedereinstieg anders aus als die
// Partie, die er fortsetzt.
//
// # Und deshalb ändert ein Deck sich nicht mehr, wenn es einmal liegt
//
// **Wer sein Deck nach dem Anfang ändert oder löscht, ändert diese Partie
// nicht.** Das ist keine Höflichkeit gegenüber dem Gegenüber, sondern die
// Bedingung dafür, dass derselbe Startabzug mit derselben Befehlsfolge
// denselben Zustand ergibt — die erste Abnahme dieses Werkstücks.

import (
	"fmt"
	"strings"
)

// Partiedeck ist der eingefrorene Stand eines Decks für eine Partie.
type Partiedeck struct {
	// Herkunft ist die Kennung des Decks, aus dem es entstand — **eine
	// Spur und kein Verweis**: das Deck darf sich danach ändern.
	Herkunft string

	// Quelle ist der Bestand, auf den die Kartenkennungen dieses Abzugs
	// zeigen — eine Peerkennung, **leer heißt: der, in dem die Partie
	// liegt** (P9, #2377, Vermerk 0198). Ein Gastdeck oder eine Leihe nennt
	// Zeilen eines **anderen** Bestands; dort nachgeschlagen fände man eine
	// andere Karte.
	Quelle string

	// Abdruck ist der Hash über den Abzug.
	//
	// **Er sagt, dass zwei Seiten dasselbe meinen**, und er ist der einzige
	// Vergleich, den es dafür gibt: zwei Decks mit denselben Karten sind
	// derselbe Abzug, zwei mit einer anderen Bilddatei nicht.
	Abdruck string

	Karten []Abzugskarte

	// Beigaben sind die auflösbaren Abhängigkeiten — Bilder, Schriften,
	// Symbolschriften, Stilpakete.
	//
	// **Sie stehen als Verweise da und nicht als Bytes.** Ein Abzug, der
	// jedes Bild mitschleppte, wöge mehr als die Partie, und §11 erlaubt es
	// ohnehin nicht, fremde Vorlagen weiterzugeben.
	Beigaben []Beigabe
}

// Abzugskarte ist eine Karte, wie sie beim Anfang dastand.
type Abzugskarte struct {
	// Kennung ist innerhalb des Abzugs eindeutig.
	Kennung string

	Name  string
	Menge int

	// Druck ist die Druckidentität aus A2 — leer heißt: keine gewählt.
	Druck string

	// Seiten sind die Seiten dieser Karte; eine einseitige hat eine.
	Seiten []Abzugsseite

	// Bereich trennt Hauptdeck von Seitendeck — leer heißt Hauptdeck.
	Bereich Deckbereich
}

// Abzugsseite ist eine Seite samt dem, was sie zum Zeichnen braucht.
type Abzugsseite struct {
	Name string

	// Stil ist das Stilpaket, mit dem sie gezeichnet wurde.
	Stil string

	// Felder sind die Feldwerte, wie sie beim Anfang standen.
	Felder map[string]string

	// KarteID ist die **auflösbare** Abhängigkeit — die Zeile im Bestand,
	// aus der diese Seite stammt (G2,
	// #1459).
	//
	// # Warum eine Kennung, wo daneben die Felder stehen
	//
	// **Weil ein Bild mehr braucht als Feldwerte.** Zum Zeichnen gehört der
	// Stil mit seinen Skripten, seinen Rahmen und seinen Schriften; die
	// Felder sagen, **was** darauf steht, nicht **wie**. Wer nur sie hätte,
	// baute eine Karte nach und keine Karte.
	//
	// **Null heißt: es gibt keine eigene Karte.** Ein Platz kann statt einer
	// eigenen Karte einen **echten Druck** nennen; dann steht der Name in
	// [Abzugskarte.Name], und ein Bild dieses Programms gibt es nicht.
	// Genau eines von beiden — dieselbe Regel wie an [Deckplatz].
	KarteID int64
}

// Beigabe ist eine auflösbare Abhängigkeit des Abzugs.
type Beigabe struct {
	// Art ist `bild`, `schrift`, `symbolschrift` oder `stil`.
	Art string

	// Kennung ist, womit sie sich auflösen lässt.
	Kennung string

	// Abdruck ist der Hash ihres Inhalts — **damit auffällt, wenn sie sich
	// ändert**, statt dass dieselbe Kennung etwas anderes zeigt.
	Abdruck string
}

// Abzuggrenze ist die größte Zahl von Karten eines Abzugs — und damit auch
// von Plätzen (Zweitblick über Phase 7, Befund 2).
//
// **Weil je Exemplar ein Objekt entsteht.** Ein Abzug kommt seit P9b auch
// von einem Gast; die Grenze der Leitung begrenzt seine Nachricht, nicht die
// Zahl der Objekte, die eine Menge daraus macht. Tausend trägt jedes Format
// samt Würfel und bleibt weit unter dem, was ein Tisch zeichnet.
const Abzuggrenze = 1000

// Pruefe sagt, ob dieser Abzug eine Partie tragen kann.
func (d Partiedeck) Pruefe() error {
	if strings.TrimSpace(d.Abdruck) == "" {
		return fmt.Errorf(
			"%w: ein Deckabzug ohne Abdruck lässt sich mit keinem zweiten "+
				"vergleichen", ErrValidation)
	}

	if len(d.Karten) == 0 {
		return fmt.Errorf(
			"%w: ein Deckabzug ohne Karten ist kein Deck", ErrValidation)
	}

	if len(d.Karten) > Abzuggrenze {
		return fmt.Errorf("%w: ein Deckabzug hat höchstens %d Plätze",
			ErrValidation, Abzuggrenze)
	}

	gesehen := make(map[string]bool, len(d.Karten))
	zusammen := 0

	for _, karte := range d.Karten {
		if strings.TrimSpace(karte.Kennung) == "" {
			return fmt.Errorf(
				"%w: eine Abzugskarte ohne Kennung lässt sich nicht "+
					"ansprechen", ErrValidation)
		}

		if gesehen[karte.Kennung] {
			return fmt.Errorf(
				"%w: die Abzugskarte %q kommt zweimal vor", ErrValidation,
				karte.Kennung)
		}

		gesehen[karte.Kennung] = true

		if karte.Menge < 1 {
			return fmt.Errorf(
				"%w: die Abzugskarte %q liegt %dmal im Deck", ErrValidation,
				karte.Kennung, karte.Menge)
		}

		// **Gegen den Rest geprüft und nicht gegen die Summe** — eine
		// Summe zweier großer Mengen liefe über.
		if karte.Menge > Abzuggrenze-zusammen {
			return fmt.Errorf("%w: ein Deckabzug führt höchstens %d Karten",
				ErrValidation, Abzuggrenze)
		}

		zusammen += karte.Menge
	}

	return nil
}

// Zaehle gibt, wie viele Karten der Abzug führt.
//
// **Die Menge und nicht die Zeilen** — vier Wälder sind vier Karten.
func (d Partiedeck) Zaehle() int {
	var zusammen int

	for _, karte := range d.Karten {
		zusammen += karte.Menge
	}

	return zusammen
}
