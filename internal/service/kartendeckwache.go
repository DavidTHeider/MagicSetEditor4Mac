package service

// **Eine Karte, die in einem Deck liegt, verschwindet nicht unbemerkt** —
// #1369.
//
// # Warum es diese Wache gibt
//
// `deckplaetze.karte_id` trägt ein `ON DELETE CASCADE`. Wer eine Karte löscht,
// die in einem Deck liegt, verliert damit **den Platz** — ohne Meldung, ohne
// Rückfrage.
//
// **Die Kaskade selbst ist richtig.** Der erste Anlauf ließ sie weg, damit ein
// Löschen auffällt; `TestJedeAbhaengigeTabelleKehrtHinterSich` hat ihn
// abgewiesen, und er hatte recht: ohne sie bliebe eine verwaiste Zeile liegen
// — in der Datei und in jeder Sicherung. Der Schutz gehört eine Schicht höher,
// und hier ist er.
//
// # Was sie nicht ist
//
// **Kein Verbot.** Wer bestätigt, löscht — und weiß, was er verliert. Das ist
// dieselbe Haltung wie beim Löschen eines Sets, wo der Name bestätigt.

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Deckwache sagt, in welchen Decks eine Karte liegt.
//
// **Eine eigene Schnittstelle mit einem Verfahren** (§2): die Aufsicht liest
// Decks und schreibt keine.
type Deckwache interface {
	DecksMitKarte(ctx context.Context, karteID int64) ([]domain.Deckbezug, error)
}

// MitDeckwache hängt der Aufsicht die Decks an.
//
// **Wahlfrei wie [Kartenaufsicht.MitFeldweg]**: ein Aufrufer ohne Decks soll
// nicht an einer fehlenden Abhängigkeit scheitern. Was er verliert, ist die
// Rückfrage, nicht die Funktion.
func (a *Kartenaufsicht) MitDeckwache(wache Deckwache) *Kartenaufsicht {
	a.decks = wache

	return a
}

// deckhindernis prüft, ob Karten in Decks liegen, und meldet es.
//
// **Ein Fehler beim Nachsehen hält an.** „Ich konnte nicht nachsehen" ist
// nicht dasselbe wie „sie liegt in keinem Deck" — und der Unterschied ist
// genau der, den diese Wache halten soll.
func (a *Kartenaufsicht) deckhindernis(
	ctx context.Context, karteIDs []int64, trotzdem bool,
) error {
	if a.decks == nil || trotzdem {
		return nil
	}

	var betroffen []string

	for _, karteID := range karteIDs {
		bezuege, err := a.decks.DecksMitKarte(ctx, karteID)
		if err != nil {
			return fmt.Errorf("die Decks zur Karte %d lesen: %w", karteID, err)
		}

		for _, bezug := range bezuege {
			betroffen = append(betroffen, bezug.Name)
		}
	}

	if len(betroffen) == 0 {
		return nil
	}

	namen := strings.Join(betroffen, ", ")

	return domain.NewFieldError("karteID", domain.SchluesselKarteInDeck,
		"Diese Karte liegt in %d Deck(s) — %s. Das Löschen nimmt sie auch dort "+
			"heraus; bestätige es ausdrücklich.", len(betroffen), namen).
		MitWerten(map[string]string{
			"anzahl": fmt.Sprintf("%d", len(betroffen)),
			"decks":  namen,
		})
}
