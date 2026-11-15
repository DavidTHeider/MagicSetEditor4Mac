package domain

// **Bedarf ist nicht Zuordnung** — Werkstück U9 von Phase 5.1
// (#1420).
//
// # Zwei Fragen, zwei Antworten
//
// Der **Bedarf** eines Decks ist eine Rechnung: vier Blitze im Hauptdeck und
// einer im Sideboard sind fünf Stück Papier, und der Abgleich sagt, wie viele
// davon in der Sammlung liegen ([Habenabgleich]). Eine Deckliste **reserviert
// nichts** — zwei Decks mit denselben drei Blitzen sind beide „vollständig",
// und das ist richtig, solange keines gespielt wird.
//
// Die **Zuordnung** ist eine Entscheidung: diese drei Exemplare gehören in
// dieses Deck. Sie nimmt sie aus dem Freien, und das zweite Deck sieht sie
// nicht mehr. Was frei ist, ist Besitz minus Reservierungen (Handeln, T2)
// minus Zuordnungen — nie unter null.
//
// # Der Vertrag mit T2
//
// Reservierungen kommen aus dem Handeln (#1467), das in dieser Fassung noch
// nicht gebaut ist. Die Rechnung kennt sie schon: [Verfuegbarkeit.Reserviert]
// ist die Zahl, die eine Reservierung dort einträgt, und heute ist sie null.
// Das ist der Vertrag, den das Issue verlangt — als Zahl in der Rechnung,
// nicht als Annahme im Kopf.

import "strings"

// Verfuegbarkeit sagt, wie viel von einem Exemplar frei ist.
type Verfuegbarkeit struct {
	// Besitz ist die Menge der Zeile — ein Wunsch hat keinen.
	Besitz int

	// Reserviert ist, was ein Handelsvorgang festhält (T2).
	Reserviert int

	// Zugeordnet ist, was Decks ausdrücklich bei sich haben.
	Zugeordnet int
}

// Frei ist, was noch zu vergeben ist — nie negativ.
//
// **Nie negativ**, aus demselben Grund wie [Habenzeile.Fehlend]: „minus zwei
// frei" wäre eine Zahl, die jede Summe verfälschte.
func (v Verfuegbarkeit) Frei() int {
	frei := v.Besitz - v.Reserviert - v.Zugeordnet
	if frei < 0 {
		return 0
	}

	return frei
}

// Erlaubt sagt, ob eine Menge noch ins Freie passt.
func (v Verfuegbarkeit) Erlaubt(menge int) bool { return menge <= v.Frei() }

// Deckzuordnung ordnet einem Deckplatz Exemplare der Sammlung zu.
//
// **Je Platz und je Exemplar eine Zeile**: ein Platz kann aus zwei
// Sammlungszeilen bedient werden (zwei Near Mint, zwei Played), und dieselbe
// Zeile kann zwei Decks bedienen, solange die Summe in ihren Besitz passt.
type Deckzuordnung struct {
	// ID ist die Kennung im Bestand — 0 für eine, die es dort noch nicht gibt.
	ID int64

	DeckID     int64
	PlatzID    int64
	ExemplarID int64

	// Menge ist die Stückzahl — mindestens eine.
	Menge int
}

// Validate gibt einen Befund je Feld, das nicht stimmt.
func (z Deckzuordnung) Validate() []*FieldError {
	var befunde []*FieldError

	if z.DeckID == 0 {
		befunde = append(befunde, NewFieldError("deckID", SchluesselPflichtfeld,
			"Eine Zuordnung gehört zu einem Deck."))
	}

	if z.PlatzID == 0 {
		befunde = append(befunde, NewFieldError("platzID", SchluesselPflichtfeld,
			"Eine Zuordnung nennt einen Deckplatz."))
	}

	if z.ExemplarID == 0 {
		befunde = append(befunde, NewFieldError("exemplarID", SchluesselPflichtfeld,
			"Eine Zuordnung nennt ein Exemplar der Sammlung."))
	}

	if z.Menge < 1 {
		befunde = append(befunde, NewFieldError("menge", SchluesselValidierung,
			"Die Menge ist %d — mindestens eines muss es sein.", z.Menge))
	}

	return befunde
}

// Abweichend sagt, ob ein Exemplar einen **anderen** Druck trägt als der Platz.
//
// **Keine Namensgleichheit als Identität** (#1420): „Lightning Bolt" aus
// Alpha ist nicht der aus M11. Wer ihn trotzdem zuordnet, tut es ausdrücklich
// — und die Zuordnung sagt es.
func Abweichend(platz Deckplatz, exemplar Exemplar) bool {
	return strings.TrimSpace(platz.Druckschluessel()) != strings.TrimSpace(exemplar.Druckschluessel())
}

// Zuordenbar sagt, ob ein Exemplar Zuordnungen tragen kann.
//
// **Ein Wunsch ist kein Besitz** (0057): wer die Karte nicht hat, kann sie
// keinem Deck geben.
func (e Exemplar) Zuordenbar() bool { return e.Besitzt() }

// Zuordenbar sagt, ob ein Platz Zuordnungen tragen kann.
//
// **Eine eigene Karte ohne Verweis hat keinen Druck** — es gibt kein Stück
// Papier, das ihr entspräche, und jede Zuordnung wäre eine Abweichung von
// nichts.
func (p Deckplatz) Zuordenbar() bool { return p.Druckschluessel() != "" }

// Druckschluessel sagt, welchen Druck der Platz meint — für Gleichheit.
func (p Deckplatz) Druckschluessel() string { return Druckschluessel(p.DruckID, p.Printing) }
