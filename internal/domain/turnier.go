// **Turnier, Runde, Paarung, Ergebnis** — Werkstück 2 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Was hier **nicht** steht
//
// **Die Rangliste.** Sie ist abgeleitet und wird nirgends abgelegt (Vermerk
// 0062, Entscheidung 1): eine gespeicherte Tabelle und der Verlauf, aus dem
// sie folgt, laufen auseinander, sobald jemand ein Ergebnis berichtigt — und
// das ist am Turniertisch der Normalfall.

package domain

import (
	"fmt"
	"strings"
)

// Turnier ist ein Verlauf über die Zeit.
type Turnier struct {
	ID      int64
	Fassung int64

	Name  string
	Spiel string

	// Einzelspiel sagt, ob ein Match aus **einem** Spiel besteht.
	//
	// **Es entscheidet nur die Anzeige, nicht die Rechnung.** MTR §3.1 hält fest, dass
	// nicht jeder dieser Tiebreaker in Formaten mit einspieligen Matches
	// gilt. Spielquote und Gegnerspielquote sind dann Kopien der
	// Matchzahlen; gerechnet werden sie trotzdem, damit später keine Lücke in
	// den Daten steht (Vermerk 0062, Entscheidung 5).
	Einzelspiel bool

	// FormatID ist das Format, nach dem gespielt wird — 0 heißt: keines
	// gewählt. DOM-04.AK-01 (#1699): **eine Angabe und kein Zwang**, wie am
	// Deck. Dass es zum Spiel des Turniers gehört, hält der Dienst.
	FormatID int64

	// Abgeschlossen ist das Datum des Abschlusses — leer heißt: offen.
	// DOM-04.AK-04 (#1699), Vermerk 0139: **eine Handlung mit Datum** und
	// kein Zustandsfeld, das ein Absturz falsch hinterließe (0062/10).
	Abgeschlossen string

	Notiz string

	Teilnehmer []Teilnehmer
	Runden     []Runde
}

// Teilnehmer ist ein Name, kein Nutzerkonto.
//
// **Wer am Küchentisch ein Turnier führt, tippt Namen** (Vermerk 0062,
// Entscheidung 2). Nutzer und Profile gibt es erst in Phase 4.12, und ein
// Turnier, das darauf wartete, käme nie.
type Teilnehmer struct {
	ID        int64
	TurnierID int64

	Name string

	// DeckID ist die **Herkunft** der Deckliste, nicht ihre Quelle.
	//
	// Die Liste selbst ist ein Abzug (Vermerk 0062, Entscheidung 3): eine
	// Deckliste im Turnier ist ein **eingereichtes Dokument**, und sie
	// nachträglich zu ändern, ist genau das, was die MTR unter „Deck Problem"
	// führt. Der Verweis geht mit, damit man das Deck aufschlagen kann.
	DeckID int64

	// Deckabzug ist die Liste, wie sie beim Anmelden aussah.
	//
	// **Kein `[]Deckplatz`.** Ein Deckplatz verweist auf eine Karte, und ein
	// Verweis ändert sich mit ihr — dann wäre es kein Abzug. Was hier steht,
	// ist Text und überlebt jedes Löschen.
	Deckabzug []Abzugsplatz

	// Ausgestiegen heißt: wird nicht mehr gepaart, bleibt aber gezählt.
	//
	// **Das ist keine Freundlichkeit, sondern Appendix C**: seine Matchquote
	// geht weiter in die Gegnermatchquote **jedes** Gegners ein, den er
	// hatte. Wer ihn entfernte, veränderte die Rangliste aller anderen.
	Ausgestiegen bool

	// Stelle ist die Reihenfolge der Anmeldung — sie entscheidet bei
	// Gleichstand (Vermerk 0062, Entscheidung 8).
	Stelle int
}

// Abzugsplatz ist eine Zeile einer eingereichten Deckliste.
//
// **Zwei Felder, die einander nicht ausschließen**: ein echter Druck hat eine
// Kennung **und** einen Namen, und der Name ist es, den ein Mensch liest. Bei
// einer eigenen Karte gibt es nur den Namen.
type Abzugsplatz struct {
	ID int64

	Bezeichnung string
	Printing    string

	// DruckID ist die lokale Zeile des Drucks (A2, #1441, 0086).
	DruckID int64

	Menge   int
	Bereich Deckbereich
	Stelle  int
}

// Runde ist eine Menge Paarungen, die zusammen entstehen.
type Runde struct {
	ID        int64
	TurnierID int64

	// Nummer zählt ab 1.
	Nummer int

	Paarungen []Paarung
}

// Paarung ist ein Match zwischen zwei Teilnehmern — oder ein Bye.
type Paarung struct {
	ID      int64
	RundeID int64

	A int64

	// B ist **null bei einem Bye**.
	//
	// **Ein Bye ist ein Ergebnis, kein fehlender Gegner** (Vermerk 0062,
	// Entscheidung 8): so steht er im Verlauf, ist sichtbar und lässt sich
	// zurücknehmen.
	B int64

	// Ergebnis ist der Spielstand — **nil, solange nicht gespielt ist**.
	Ergebnis *Ergebnis

	Stelle int
}

// IstBye sagt, ob diese Paarung ein Bye ist.
func (p Paarung) IstBye() bool { return p.B == 0 }

// Ergebnis ist ein Spielstand, kein Sieger.
//
// **Weil Appendix C Spielpunkte braucht**, und die folgen aus dem Stand. Ein
// Modell, das nur den Sieger hielte, könnte Tiebreaker 3 und 4 nicht rechnen —
// und das merkte niemand, bis zwei Spieler gleichauf sind (Vermerk 0062,
// Entscheidung 7).
type Ergebnis struct {
	SpieleA       int
	SpieleB       int
	Unentschieden int
}

// Pruefe sagt, ob dieses Ergebnis eintragbar ist.
func (e Ergebnis) Pruefe(hoechstens int) error {
	switch {
	case e.SpieleA < 0 || e.SpieleB < 0 || e.Unentschieden < 0:
		return fmt.Errorf(
			"%w: ein Spielstand hat keine negativen Zahlen", ErrValidation)

	case e.SpieleA+e.SpieleB+e.Unentschieden == 0:
		return fmt.Errorf(
			"%w: ein Ergebnis ohne ein einziges Spiel ist kein Ergebnis — "+
				"wer nicht gespielt hat, hat noch keines", ErrValidation)

	case hoechstens > 0 && e.SpieleA+e.SpieleB+e.Unentschieden > hoechstens:
		return fmt.Errorf(
			"%w: %d Spiele, aber ein Match geht hier über höchstens %d",
			ErrValidation, e.SpieleA+e.SpieleB+e.Unentschieden, hoechstens)
	}

	return nil
}

// Pruefe sagt, ob dieses Turnier so angelegt werden darf.
func (t Turnier) Pruefe() error {
	if strings.TrimSpace(t.Name) == "" {
		return NewFieldError("name", SchluesselTurniernameLeer,
			"Ein Turnier braucht einen Namen.")
	}

	if strings.TrimSpace(t.Spiel) == "" {
		return NewFieldError("spiel", SchluesselTurnierspielLeer,
			"Ein Turnier gehört zu einem Spiel — sonst gibt es kein Format, "+
				"gegen das sich prüfen ließe.")
	}

	return nil
}
