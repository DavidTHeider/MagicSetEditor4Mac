// **„Habe ich das?"** — Werkstück 6 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md).
//
// # Warum je Druck und nicht je Platz
//
// **Weil es dieselben Karten sind.** Vier im Hauptdeck und eine im Sideboard
// verlangen fünf Stück Papier; CR 100.4a rechnet die Höchstzahl ebenso über
// beide Bereiche. Ein Abgleich je Platz gäbe den Vorrat zweimal aus und
// meldete „alles da", wo eine fehlt.
//
// # Und dieselben drei Antworten wie bei der Formatprüfung
//
// Ein echter Druck wird gezählt. Eine eigene Karte **mit** Verweis zählt über
// ihren Druck. Eine eigene Karte **ohne** Verweis hat kein Exemplar — und das
// ist **kein Fehlbestand**, sondern eine Karte, über die nichts bekannt ist.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammlungsbestand ist, was der Abgleich von der Sammlung braucht.
//
// **Eine eigene Schnittstelle mit einem Verfahren** (§2): der Abgleich liest
// und schreibt nichts.
type Sammlungsbestand interface {
	Fuer(ctx context.Context, printing string) ([]*domain.Exemplar, error)

	// NachDruck gibt die Exemplare einer Druckzeile — für einen Druck ohne
	// Alias (A2, #1441): seit ihm kann ein Platz auf eine Zeile zeigen, die
	// Scryfall nicht kennt, und `Fuer("")` fände jede Zeile ohne Alias.
	NachDruck(ctx context.Context, druckID int64) ([]*domain.Exemplar, error)
}

// Habenzeile ist die Antwort für **einen Druck** im Deck.
type Habenzeile struct {
	// Printing ist der Druck — leer bei einer eigenen Karte ohne Verweis und
	// bei einem Druck ohne Scryfall-Alias.
	Printing string

	// DruckID ist die Zeile des Drucks in diesem Bestand (A2, #1441).
	DruckID int64

	// Vorlaeufig sagt, dass der Druck keinen Alias hat — die Zeile zählt
	// dann die Exemplare **ihrer Druckzeile**. Ein Zustand, kein Fehler.
	Vorlaeufig bool

	// Bezeichnung ist, wie die Sammlung ihn nennt — soweit sie ihn kennt.
	Bezeichnung string

	Gebraucht int
	Vorhanden int

	// ExemplarIDs sind die Zeilen der Sammlung, die hier gezählt wurden —
	// **ohne die Wünsche**. U9 (#1420) rechnet über sie, was von diesem Druck
	// noch frei ist; ohne sie müsste es die Sammlung ein zweites Mal lesen.
	ExemplarIDs []int64

	// Fehlend ist, was zu besorgen wäre — nie negativ.
	Fehlend int

	// Gewuenscht ist, wie viele davon schon auf der Wunschliste stehen —
	// **getrennt vom Besitz** (0057): sie fehlen weiter. DOM-03.AK-02
	// (#1698) fragt danach, bevor es einen Wunsch anlegt.
	Gewuenscht int

	// Grund steht, wenn nichts bekannt ist.
	Grund string
}

// Habenstand ist der ganze Abgleich eines Decks.
type Habenstand struct {
	Zeilen []Habenzeile

	// Fehlend ist die Zahl der **Karten**, die zu besorgen wären.
	Fehlend int

	// Unbekannt ist die Zahl der Karten, über die nichts bekannt ist.
	//
	// **Getrennt von [Habenstand.Fehlend]**, und das ist die Zusage: „vier
	// fehlen" und „über vier ist nichts bekannt" sind zwei Auskünfte.
	Unbekannt int
}

// Habenabgleich ist der Anwendungsfall „was von diesem Deck habe ich?".
type Habenabgleich struct {
	decks    Deckspeicher
	sammlung Sammlungsbestand
	verweise Deckherkunft
}

// NeuerHabenabgleich baut ihn.
func NeuerHabenabgleich(
	decks Deckspeicher, sammlung Sammlungsbestand, verweise Deckherkunft,
) *Habenabgleich {
	return &Habenabgleich{decks: decks, sammlung: sammlung, verweise: verweise}
}

// Fuer gleicht ein Deck gegen die Sammlung ab.
//
// **Ein Fehler hält an und wird nicht durchgewunken.** „Ich konnte nicht
// nachsehen" ist nicht „du hast nichts" — anders als bei der Formatprüfung,
// wo eine schweigende **fremde** Gegenstelle einen Befund gibt und die Arbeit
// weitergeht. Hier ist es der eigene Bestand: schweigt der, stimmt die ganze
// Antwort nicht.
func (h *Habenabgleich) Fuer(ctx context.Context, deckID int64) (Habenstand, error) {
	deck, err := h.decks.Deck(ctx, deckID)
	if err != nil {
		return Habenstand{}, err
	}

	bedarf, reihe, unbekannt, lesefehler := h.bedarf(ctx, deck)
	if lesefehler != nil {
		return Habenstand{}, lesefehler
	}

	stand := Habenstand{
		Zeilen:    make([]Habenzeile, 0, len(reihe)+1),
		Unbekannt: unbekannt,
	}

	for _, schluessel := range reihe {
		zeile, zaehlfehler := h.zeile(ctx, bedarf[schluessel])
		if zaehlfehler != nil {
			return Habenstand{}, zaehlfehler
		}

		stand.Fehlend += zeile.Fehlend
		stand.Zeilen = append(stand.Zeilen, zeile)
	}

	if unbekannt > 0 {
		stand.Zeilen = append(stand.Zeilen, Habenzeile{
			Gebraucht: unbekannt,
			Grund: "Diese Karten sind eigene Entwürfe ohne Verweis auf einen Druck; " +
				"die Sammlung führt sie nicht.",
		})
	}

	return stand, nil
}

// bedarf zählt, wie viele Karten je Druck gebraucht werden.
//
// **In der Reihenfolge der Plätze und nicht der Map** — sonst käme dieselbe
// Eingabe zweimal anders heraus (#1289).
func (h *Habenabgleich) bedarf(ctx context.Context, deck domain.Deck) (
	bedarf map[string]bedarfsposten, reihe []string, unbekannt int, fehler error,
) {
	bedarf = map[string]bedarfsposten{}

	for _, platz := range deck.Plaetze {
		druck, lesefehler := h.druckZu(ctx, platz)
		if lesefehler != nil {
			return nil, nil, 0, lesefehler
		}

		// **Der Schlüssel ist die Druckzeile, wo es keinen Alias gibt**
		// (A2, #1441). Ohne beides ist es eine eigene Karte ohne Verweis.
		schluessel := domain.Druckschluessel(druck.DruckID, druck.Printing)
		if schluessel == "" {
			unbekannt += platz.Menge

			continue
		}

		posten, schon := bedarf[schluessel]
		if !schon {
			reihe = append(reihe, schluessel)
			posten = druck
		}

		posten.Menge += platz.Menge
		bedarf[schluessel] = posten
	}

	return bedarf, reihe, unbekannt, nil
}

// bedarfsposten ist ein Druck mit der Zahl, die ein Deck von ihm braucht.
type bedarfsposten struct {
	Printing string
	DruckID  int64
	Menge    int
}

// druckZu sagt, welcher Druck für einen Platz zu zählen ist.
//
// **Alias und Zeile zusammen**: ein Platz auf einen echten Druck trägt seit
// A2 beides, eine eigene Karte mit Verweis erbt beides vom Verweis.
func (h *Habenabgleich) druckZu(ctx context.Context, platz domain.Deckplatz) (
	bedarfsposten, error,
) {
	if platz.Printing != "" || platz.DruckID != 0 {
		return bedarfsposten{Printing: platz.Printing, DruckID: platz.DruckID}, nil
	}

	verweis, err := h.verweise.Fuer(ctx, platz.KarteID)
	if err != nil {
		return bedarfsposten{}, fmt.Errorf("den Verweis der Karte %d lesen: %w", platz.KarteID, err)
	}

	if verweis == nil {
		return bedarfsposten{}, nil
	}

	return bedarfsposten{Printing: verweis.Printing, DruckID: verweis.DruckID}, nil
}

// zeile zählt, was von einem Druck da ist.
//
// **Über den Alias, wo es einen gibt, sonst über die Druckzeile** (A2,
// #1441): ein vorläufiger Druck hat keinen Alias, und `Fuer("")` fände jede
// Zeile ohne Alias statt seiner.
func (h *Habenabgleich) zeile(ctx context.Context, posten bedarfsposten) (
	Habenzeile, error,
) {
	exemplare, err := h.exemplare(ctx, posten)
	if err != nil {
		return Habenzeile{}, err
	}

	zeile := Habenzeile{
		Printing: posten.Printing, DruckID: posten.DruckID,
		Vorlaeufig: posten.Printing == "", Gebraucht: posten.Menge,
		ExemplarIDs: []int64{},
	}
	gebraucht := posten.Menge

	for _, exemplar := range exemplare {
		// **Ein Wunsch ist kein Besitz** (0057). Wer eine Karte auf die
		// Wunschliste gesetzt hat, hat sie nicht — und ein Abgleich, der sie
		// mitzählte, sagte „vollständig" über ein Deck, das niemand spielen
		// kann.
		if !exemplar.Besitzt() {
			zeile.Gewuenscht += exemplar.Menge

			continue
		}

		zeile.Vorhanden += exemplar.Menge
		zeile.ExemplarIDs = append(zeile.ExemplarIDs, exemplar.ID)

		if zeile.Bezeichnung == "" {
			zeile.Bezeichnung = exemplar.Bezeichnung()
		}
	}

	// **Nie negativ.** Wer sechs hat und vier braucht, dem fehlt nichts —
	// „minus zwei" wäre eine Zahl, die jede Summe verfälschte.
	if gebraucht > zeile.Vorhanden {
		zeile.Fehlend = gebraucht - zeile.Vorhanden
	}

	return zeile, nil
}

// exemplare holt, was zu einem Posten zählt.
func (h *Habenabgleich) exemplare(ctx context.Context, posten bedarfsposten) (
	[]*domain.Exemplar, error,
) {
	if posten.Printing == "" {
		exemplare, err := h.sammlung.NachDruck(ctx, posten.DruckID)
		if err != nil {
			return nil, fmt.Errorf("die Sammlung zur Druckzeile %d lesen: %w", posten.DruckID, err)
		}

		return exemplare, nil
	}

	exemplare, err := h.sammlung.Fuer(ctx, posten.Printing)
	if err != nil {
		return nil, fmt.Errorf("die Sammlung zum Druck %s lesen: %w", posten.Printing, err)
	}

	return exemplare, nil
}
