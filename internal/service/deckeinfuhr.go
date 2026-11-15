package service

// **Eine Deckliste wird ein Deck** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Warum es dafür einen eigenen Anwendungsfall gibt
//
// **Weil [Listeneinfuhr] die Sammlung schreibt.** Sie legt Exemplare an —
// Besitz, mit Zustand, Sprache und Ausführung. Eine Deckliste ist keine
// Besitzmehrung: wer sie einfügt, sagt „so sieht mein Deck aus", nicht „das
// habe ich gekauft". Das Issue sagt es ausdrücklich: „Bestehende
// Listeneinfuhr schreibt Sammlung und darf nicht direkt als Deckimport
// verwendet werden."
//
// **Geteilt wird trotzdem, was dasselbe ist**: der Zerteiler
// (`internal/sammelliste`) und die Namensauflösung ([Listeneinfuhr.Deuten]).
// Zwei Fassungen desselben Deutens liefen auseinander.
//
// # Warum sie legt und nicht ersetzt
//
// **Weil Wegnehmen niemand bestellt hat.** Bestätigt wird eine Liste, die
// hinzukommt; ein Deck leerzuräumen, weil jemand eine Liste einfügt, wäre
// eine Wirkung, über die niemand entschieden hat. Wer ersetzen will, nimmt
// vorher heraus — das geht über den vorhandenen Weg und ist sichtbar.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// Deckdeutung ist, was die Einfuhr vom Listenleser braucht.
//
// **Eine Schnittstelle und nicht [Listeneinfuhr] selbst** (§2): so hängt
// dieser Anwendungsfall nicht daran, dass die Deutung Scryfall fragt — und
// ein Lauf braucht kein Netz.
type Deckdeutung interface {
	Deuten(ctx context.Context, text string) (Listendeutung, error)
}

// Deckhalter ist, was sie vom Deck braucht — ein Ausschnitt aus
// [Deckspeicher].
type Deckhalter interface {
	Deck(ctx context.Context, kennung int64) (domain.Deck, error)
	PlatzSetzen(ctx context.Context, deck int64, platz domain.Deckplatz) (
		domain.Deckplatz, error)
}

// Deckeinfuhr liest eine Deckliste in ein Deck.
type Deckeinfuhr struct {
	deutung Deckdeutung
	decks   Deckhalter
}

// NeueDeckeinfuhr baut sie.
func NeueDeckeinfuhr(deutung Deckdeutung, decks Deckhalter) *Deckeinfuhr {
	return &Deckeinfuhr{deutung: deutung, decks: decks}
}

// Deckvorschauzeile ist eine Zeile, wie sie ins Deck käme.
type Deckvorschauzeile struct {
	Bereich domain.Deckbereich
	Menge   int

	// Name, Setcode und Nummer sind, was dastand — nicht, was gefunden wurde.
	//
	// **Damit jemand seine Zeile wiedererkennt** (§8): wer „Blitzschlag"
	// geschrieben hat, sucht „Blitzschlag" und nicht „Lightning Bolt".
	Name    string
	Setcode string
	Nummer  string

	// Printing ist der Druck — **leer, solange die Zeile nicht eindeutig
	// ist**.
	Printing string

	// Treffer sind die Drucke zur Wahl, Drucke ihre Gesamtzahl.
	Treffer []Suchtreffer
	Drucke  int

	// Befund sagt, warum diese Zeile nicht übernommen wird — leer, wenn sie
	// es wird.
	Befund string

	// Uebernehmbar heißt: diese Zeile hat genau einen Druck und einen
	// Bereich.
	Uebernehmbar bool
}

// Deckvorschau ist, was aus einer Liste würde — **ohne dass etwas geschieht**.
type Deckvorschau struct {
	DeckID int64

	// Stand ist der Abdruck des Decks, aus dem diese Vorschau entstand
	// ([domain.Deckstand]).
	//
	// **Er geht bei der Bestätigung zurück.** Wer dazwischen eine Karte
	// hineinlegt, bekommt einen Konflikt statt einer Deckliste über einer
	// anderen (§7, Regel 2).
	Stand string

	Zeilen  []Deckvorschauzeile
	Befunde []domain.Befund

	// Hauptdeck und Sideboard sind die Stückzahlen, die hinzukämen.
	Hauptdeck int
	Sideboard int
}

// Deckzeile ist eine Zeile, für die entschieden ist.
type Deckzeile struct {
	Printing string
	Menge    int
	Bereich  domain.Deckbereich
}

// Deckeinfuhrbericht sagt, was gelegt wurde.
type Deckeinfuhrbericht struct {
	// Gelegt ist die Zahl der Plätze, Stueck die der Karten.
	Gelegt int
	Stueck int

	// Stand ist der Abdruck **nach** dem Legen — wer weitermachen will,
	// nimmt ihn.
	Stand string
}

// Vorschau liest die Liste und zeigt, was ins Deck käme — **ohne zu
// schreiben**.
func (d *Deckeinfuhr) Vorschau(
	ctx context.Context, deckID int64, text string,
) (Deckvorschau, error) {
	deck, fehler := d.decks.Deck(ctx, deckID)
	if fehler != nil {
		return Deckvorschau{}, fehler
	}

	gedeutet, fehler := d.deutung.Deuten(ctx, text)
	if fehler != nil {
		return Deckvorschau{}, fehler
	}

	vorschau := Deckvorschau{
		DeckID: deckID, Stand: domain.Deckstand(deck),
		Zeilen: []Deckvorschauzeile{}, Befunde: gedeutet.Befunde,
	}

	// **Gleiche Karte, gleicher Bereich: eine Zeile.** Eine Liste nennt
	// dieselbe Karte oft in zwei Zeilen; zwei Plätze über denselben Druck
	// wären zwei Zeilen über dasselbe — und der Bestand legte sie ohnehin
	// zusammen ([storage.Deckablage.PlatzSetzen]).
	stellen := map[string]int{}

	for _, gedeuteteZeile := range gedeutet.Zeilen {
		vorschauzeileEintragen(&vorschau, stellen, gedeuteteZeile)
	}

	return vorschau, nil
}

// vorschauzeileEintragen legt eine gedeutete Zeile in die Vorschau.
func vorschauzeileEintragen(
	vorschau *Deckvorschau, stellen map[string]int, gedeutet Zeilendeutung,
) {
	zeile := vorschauzeile(gedeutet)

	if !zeile.Uebernehmbar {
		vorschau.Zeilen = append(vorschau.Zeilen, zeile)

		return
	}

	schluessel := string(zeile.Bereich) + "|" + zeile.Printing

	if stelle, schon := stellen[schluessel]; schon {
		vorschau.Zeilen[stelle].Menge += zeile.Menge
	} else {
		stellen[schluessel] = len(vorschau.Zeilen)
		vorschau.Zeilen = append(vorschau.Zeilen, zeile)
	}

	if zeile.Bereich == domain.BereichSideboard {
		vorschau.Sideboard += zeile.Menge
	} else {
		vorschau.Hauptdeck += zeile.Menge
	}
}

// vorschauzeile macht aus einer Deutung eine Zeile der Vorschau.
func vorschauzeile(gedeutet Zeilendeutung) Deckvorschauzeile {
	zeile := Deckvorschauzeile{
		Menge: mengeAus(gedeutet.Zeile), Name: gedeutet.Zeile.Name,
		Setcode: gedeutet.Zeile.Setcode, Nummer: gedeutet.Zeile.Nummer,
		Treffer: gedeutet.Treffer, Drucke: gedeutet.Drucke,
		Befund: gedeutet.Befund,
	}

	bereich, imDeck := domain.DeckbereichAus(gedeutet.Zeile.Abschnitt)
	if !imDeck {
		zeile.Befund = fmt.Sprintf(
			"steht unter »%s« und gehört damit nicht ins Deck",
			gedeutet.Zeile.Abschnitt)

		return zeile
	}

	zeile.Bereich = bereich

	// **Kein geratener Druck** — das Issue sagt es so. Null Treffer heißt
	// „nicht gefunden", mehrere heißen „jemand muss wählen"; beides ist eine
	// Auskunft und keine Karte im Deck.
	if !gedeutet.Eindeutig() {
		if zeile.Befund == "" {
			zeile.Befund = befundOhneEindeutigkeit(gedeutet.Drucke)
		}

		return zeile
	}

	zeile.Printing = gedeutet.Treffer[0].Kennung
	zeile.Uebernehmbar = true

	return zeile
}

// befundOhneEindeutigkeit sagt, warum eine Zeile offen bleibt.
func befundOhneEindeutigkeit(drucke int) string {
	if drucke == 0 {
		return "kein Druck gefunden"
	}

	return fmt.Sprintf("%d Drucke passen — welcher ist gemeint?", drucke)
}

// mengeAus gibt die Menge einer Zeile — **mindestens eine**.
//
// **Weil eine Zeile ohne Zahl eine Karte meint.** Der Zerteiler lässt die
// Menge weg, wo keine dasteht; eine Null im Deck wäre ein Platz, den
// [domain.Deckplatz.Validate] abwiese.
func mengeAus(zeile sammelliste.Zeile) int {
	if zeile.Menge < 1 {
		return 1
	}

	return zeile.Menge
}

// Uebernehmen legt die bestätigten Zeilen ins Deck.
//
// **Der Stand ist die Bedingung** (§7, Regel 2): er stammt aus der Vorschau,
// und hat sich das Deck seitdem geändert, wird nichts gelegt. Das ist zugleich
// die Zusage gegen den zweiten Lauf — wer dieselbe Vorschau zweimal
// bestätigt, legt nicht zweimal.
func (d *Deckeinfuhr) Uebernehmen(
	ctx context.Context, deckID int64, stand string, zeilen []Deckzeile,
) (Deckeinfuhrbericht, error) {
	if len(zeilen) == 0 {
		return Deckeinfuhrbericht{}, fmt.Errorf(
			"%w: es ist keine Zeile bestätigt", domain.ErrValidation)
	}

	deck, fehler := d.decks.Deck(ctx, deckID)
	if fehler != nil {
		return Deckeinfuhrbericht{}, fehler
	}

	if jetzt := domain.Deckstand(deck); jetzt != stand {
		return Deckeinfuhrbericht{}, fmt.Errorf(
			"%w: das Deck hat sich seit der Vorschau geändert",
			domain.ErrConflict)
	}

	// **Erst prüfen, dann schreiben.** Eine Zeile, die der Platz abwiese,
	// ließe sonst die Hälfte der Liste im Deck und die andere draußen (§8).
	plaetze := make([]domain.Deckplatz, 0, len(zeilen))

	for _, zeile := range zeilen {
		platz := domain.Deckplatz{
			Printing: zeile.Printing, Menge: zeile.Menge,
			Bereich: zeile.Bereich,
		}

		if befunde := platz.Validate(); len(befunde) > 0 {
			return Deckeinfuhrbericht{}, befunde[0]
		}

		plaetze = append(plaetze, platz)
	}

	return d.legen(ctx, deckID, plaetze)
}

// legen schreibt die geprüften Plätze und zählt mit.
func (d *Deckeinfuhr) legen(
	ctx context.Context, deckID int64, plaetze []domain.Deckplatz,
) (Deckeinfuhrbericht, error) {
	bericht := Deckeinfuhrbericht{}

	for _, platz := range plaetze {
		if _, fehler := d.decks.PlatzSetzen(ctx, deckID, platz); fehler != nil {
			return Deckeinfuhrbericht{}, fehler
		}

		bericht.Gelegt++
		bericht.Stueck += platz.Menge
	}

	// **Der neue Stand geht mit** — wer eine zweite Liste einfügt, braucht
	// ihn, und ein zweiter Ruf dafür wäre einer zu viel.
	deck, fehler := d.decks.Deck(ctx, deckID)
	if fehler != nil {
		return Deckeinfuhrbericht{}, fehler
	}

	bericht.Stand = domain.Deckstand(deck)

	return bericht, nil
}
