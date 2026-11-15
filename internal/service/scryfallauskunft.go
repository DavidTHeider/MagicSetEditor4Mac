package service

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Fremdauskuenfte ist, was die Auskunft von einem fremden Dienst braucht.
//
// **Getrennt von [Fremdkarten] und [Fremdbilder]**, obwohl derselbe Klient
// alle drei erfüllt: die Schnittstelle sagt, was **gebraucht** wird (§2), und
// wer nur Regelhinweise zeigt, soll keine Schreibfähigkeit binden.
type Fremdauskuenfte interface {
	KarteNachName(ctx context.Context, name string) (
		karte *scryfall.Karte, alter time.Duration, befunde []domain.Befund, fehler error)

	Rulings(ctx context.Context, kartenID string) (
		hinweise []scryfall.Regelhinweis, alter time.Duration,
		befunde []domain.Befund, fehler error)

	Setangaben(ctx context.Context, code string) (
		satz *scryfall.Setangabe, alter time.Duration,
		befunde []domain.Befund, fehler error)

	Symbologie(ctx context.Context) (
		symbole []scryfall.Symbol, alter time.Duration,
		befunde []domain.Befund, fehler error)
}

// Auskunft holt, was Scryfall über Karten, Sätze und Symbole weiß.
//
// # Warum das ein eigener Anwendungsfall ist und nicht Teil des Abgleichs
//
// **Weil er nichts schreibt.** Der Abgleich zeigt einen Vorschlag und
// übernimmt ihn; die Auskunft ist reines Lesen. Sie in denselben Typ zu legen
// hieße, dass jeder Aufrufer einer Regelauskunft eine Schreibfähigkeit
// mitbrächte — und §2 sagt, wonach eine Schnittstelle geschnitten wird.
type Auskunft struct{ fremd Fremdauskuenfte }

// NeueAuskunft baut sie.
func NeueAuskunft(fremd Fremdauskuenfte) *Auskunft { return &Auskunft{fremd: fremd} }

// Regelauskunft sind die Hinweise zu einer Karte samt ihrer Herkunft.
type Regelauskunft struct {
	// Kartenname ist der Name, unter dem Scryfall die Karte führt — **nicht**
	// der eingegebene. Wer sich vertippt hat, sieht es hier.
	Kartenname string

	Hinweise []scryfall.Regelhinweis

	// Adresse ist die Kartenseite bei Scryfall.
	Adresse string

	Befunde []domain.Befund
	Alter   time.Duration
}

// Regelhinweise holt die Hinweise zu einer Karte über ihren Namen.
//
// # Warum über den Namen und nicht über die Kennung
//
// Weil niemand die Kennung von Scryfall kennt. Sie steht in keiner
// MSE-Vorlage und in keinem Set; wer sie verlangte, machte die Auskunft
// unbenutzbar. Der Name kostet einen zusätzlichen Abruf — der aber aus dem
// Zwischenspeicher kommt, sobald jemand vorher abgeglichen hat.
func (a *Auskunft) Regelhinweise(ctx context.Context, name string) (*Regelauskunft, error) {
	karte, kartenalter, kartenbefunde, fehler := a.fremd.KarteNachName(ctx, name)
	if fehler != nil {
		return nil, fehler
	}

	hinweise, alter, befunde, hinweisfehler := a.fremd.Rulings(ctx, karte.Kennung)
	if hinweisfehler != nil {
		return nil, hinweisfehler
	}

	// **Das ältere der beiden Alter gilt.** Eine Auskunft ist so frisch wie
	// ihr ältester Bestandteil; das jüngere zu nennen behauptete mehr, als
	// gemessen ist (§16, Regel 5).
	if kartenalter > alter {
		alter = kartenalter
	}

	return &Regelauskunft{
		Kartenname: karte.Name, Hinweise: hinweise, Adresse: karte.Adresse,
		Befunde: append(kartenbefunde, befunde...), Alter: alter,
	}, nil
}

// Satzauskunft ist, was Scryfall über einen Satz sagt.
type Satzauskunft struct {
	Satz    *scryfall.Setangabe
	Befunde []domain.Befund
	Alter   time.Duration
}

// Satz holt die Angaben zu einem Satz über sein Kürzel.
func (a *Auskunft) Satz(ctx context.Context, code string) (*Satzauskunft, error) {
	if code == "" {
		return nil, fmt.Errorf("%w: es wurde kein Satzkürzel genannt", domain.ErrValidation)
	}

	satz, alter, befunde, fehler := a.fremd.Setangaben(ctx, code)
	if fehler != nil {
		return nil, fehler
	}

	return &Satzauskunft{Satz: satz, Befunde: befunde, Alter: alter}, nil
}

// Symbolauskunft ist die Liste aller Kartensymbole.
type Symbolauskunft struct {
	Symbole []scryfall.Symbol
	Befunde []domain.Befund
	Alter   time.Duration
}

// Symbole holt die Symbologie.
func (a *Auskunft) Symbole(ctx context.Context) (*Symbolauskunft, error) {
	symbole, alter, befunde, fehler := a.fremd.Symbologie(ctx)
	if fehler != nil {
		return nil, fehler
	}

	return &Symbolauskunft{Symbole: symbole, Befunde: befunde, Alter: alter}, nil
}

// Manakostenpruefung sagt, welche Symbole eines Kostens unbekannt sind.
//
// # Wofür das gut ist
//
// **Ein getippter Manakosten ist die häufigste Fehlerquelle einer eigenen
// Karte.** `{2}{R}` ist richtig, `{2}{r}` nicht, und MSE zeichnet für das
// zweite ein leeres Kästchen — ohne zu sagen, warum. Die Symbologie weiß,
// welche Schreibweisen es gibt.
//
// **Sie prüft gegen `appears_in_mana_costs`**, nicht gegen „ist ein Symbol":
// `{20}` ist ein Manasymbol und kommt trotzdem in keinem Manakosten vor. Wer
// das übergeht, lässt einen Kosten durch, den es nicht gibt.
func (a *Auskunft) Manakostenpruefung(ctx context.Context, kosten string) ([]string, error) {
	auskunft, fehler := a.Symbole(ctx)
	if fehler != nil {
		return nil, fehler
	}

	erlaubt := map[string]bool{}
	for _, eines := range auskunft.Symbole {
		if eines.InManakosten {
			erlaubt[eines.Zeichen] = true
		}
	}

	var unbekannt []string

	for _, zeichen := range domain.Manasymbole(kosten) {
		if !erlaubt[zeichen] {
			unbekannt = append(unbekannt, zeichen)
		}
	}

	return unbekannt, nil
}
