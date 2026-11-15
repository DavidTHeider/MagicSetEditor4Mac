// **Was einem Deck fehlt, auf die Wunschliste** — DOM-03.AK-02
// (#1698).
//
// # Die Regel
//
// **Ein Druck, den es noch nicht als Wunsch gibt, kommt mit der fehlenden
// Menge darauf; ein Druck, der schon gewünscht ist, bleibt, wie er ist** —
// die Antwort nennt ihn samt beider Zahlen. Eine Menge, die jemand selbst
// gesetzt hat, still zu verschieben, wäre genau die unkontrollierte
// Mengenänderung, die AK-03 ausschließt; und ein zweiter Ruf legt nichts
// doppelt an.
//
// **Marke `new`** (AGENTS.md §15): MSE2 kennt weder Decks noch Wunschliste —
// gesucht am 2026-09-27 unter `/Users/david/mse2-quelle/src` (Fassung
// f9d9356 vom 2021-07-06) nach `deck` und `wishlist`: ein Treffer, und der
// ist ein Kommentar; `collection` meint dort Skriptsammlungen.
//
// # Warum der Abgleich die Wünsche zählt
//
// Er liest die Exemplare ohnehin und überspringt die Wünsche, weil sie kein
// Besitz sind (0057). Sie dort mitzuzählen kostet eine Zeile; ein zweites
// Lesen hier wären zwei Stände derselben Sammlung in einer Antwort.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sammelanlage legt mehrere Exemplare an — **alle oder keines**.
//
// **Die Schnittstelle steht hier und wird in `storage` erfüllt** (§2).
type Sammelanlage interface {
	AlleAnlegen(ctx context.Context, exemplare []*domain.Exemplar) ([]int64, error)
}

// Decknamen gibt ein Deck mit den Namen seiner Plätze — `Deckbau.Deck`.
//
// **Für die Namen und nur für sie**: ein neuer Wunsch ohne Namen stünde mit
// seiner Kennung in der Liste, und die sagt niemandem etwas.
type Decknamen interface {
	Deck(ctx context.Context, kennung int64) (domain.Deck, error)
}

// Vorhandenerwunsch ist ein Druck, der fehlt und schon gewünscht ist.
type Vorhandenerwunsch struct {
	Habenzeile
}

// Fehlkartenwuensche ist die Antwort: was neu darauf steht, was schon
// darauf stand und was keinen Druck hat.
type Fehlkartenwuensche struct {
	Neu       []domain.Exemplar
	Vorhanden []Vorhandenerwunsch

	// Unbekannt ist die Zahl der Karten, die eigene Entwürfe ohne Verweis
	// sind — ohne Druck gibt es keinen Wunsch.
	Unbekannt int
}

// Fehlkartenwunsch ist der Anwendungsfall „was fehlt, auf die Wunschliste".
type Fehlkartenwunsch struct {
	abgleich *Habenabgleich
	decks    Decknamen
	ablage   Sammelanlage
}

// NeuerFehlkartenwunsch baut ihn.
func NeuerFehlkartenwunsch(
	abgleich *Habenabgleich, decks Decknamen, ablage Sammelanlage,
) *Fehlkartenwunsch {
	return &Fehlkartenwunsch{abgleich: abgleich, decks: decks, ablage: ablage}
}

// Wuenschen setzt, was dem Deck fehlt und noch nicht gewünscht ist, auf die
// Wunschliste — **in einem Zug**.
func (f *Fehlkartenwunsch) Wuenschen(ctx context.Context, deckID int64) (
	Fehlkartenwuensche, error,
) {
	haben, err := f.abgleich.Fuer(ctx, deckID)
	if err != nil {
		return Fehlkartenwuensche{}, err
	}

	deck, err := f.decks.Deck(ctx, deckID)
	if err != nil {
		return Fehlkartenwuensche{}, err
	}

	namen := map[string]string{}

	for _, platz := range deck.Plaetze {
		namen[domain.Druckschluessel(platz.DruckID, platz.Printing)] = platz.Bezeichnung
	}

	ergebnis := Fehlkartenwuensche{
		Neu: []domain.Exemplar{}, Vorhanden: []Vorhandenerwunsch{}, Unbekannt: haben.Unbekannt,
	}
	neue := []*domain.Exemplar{}

	for _, zeile := range haben.Zeilen {
		// **Die Zeile der Unbekannten fehlt nicht** — sie hat keinen Druck,
		// und ihre Zahl steht in `Unbekannt`.
		if zeile.Fehlend == 0 {
			continue
		}

		if zeile.Gewuenscht > 0 {
			ergebnis.Vorhanden = append(ergebnis.Vorhanden, Vorhandenerwunsch{Habenzeile: zeile})

			continue
		}

		wunsch := &domain.Exemplar{
			Printing: zeile.Printing, DruckID: zeile.DruckID, Menge: zeile.Fehlend,
			Wunsch: true, Name: namen[domain.Druckschluessel(zeile.DruckID, zeile.Printing)],
		}
		wunsch.Normalisieren()

		neue = append(neue, wunsch)
	}

	if len(neue) == 0 {
		return ergebnis, nil
	}

	kennungen, err := f.ablage.AlleAnlegen(ctx, neue)
	if err != nil {
		return Fehlkartenwuensche{}, err
	}

	for stelle, wunsch := range neue {
		wunsch.ID, wunsch.Fassung = kennungen[stelle], 1
		ergebnis.Neu = append(ergebnis.Neu, *wunsch)
	}

	return ergebnis, nil
}
