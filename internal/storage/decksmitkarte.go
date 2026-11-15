package storage

// **In welchen Decks liegt diese Karte?** —
// #1369.
//
// # Warum es diese Frage gibt
//
// `deckplaetze.karte_id` trägt ein `ON DELETE CASCADE`. Wer eine Karte löscht,
// die in einem Deck liegt, verliert damit den Platz — und ohne diese Abfrage
// **ohne es zu merken**. Die Kaskade selbst ist richtig: ohne sie bliebe eine
// verwaiste Zeile liegen, in der Datei und in jeder Sicherung.
//
// **Der Wächter sitzt deshalb darüber**, im Dienst, und diese Abfrage ist das,
// was er braucht.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// DecksMitKarte nennt die Decks, in denen eine eigene Karte liegt.
//
// **Jedes Deck einmal**, auch wenn die Karte in Hauptdeck **und** Sideboard
// steht: zweimal genannt sähe aus wie zwei Decks, und die Rückfrage zählte
// falsch.
//
// **Geordnet vom Bestand**, sonst nennt jede Fläche sie anders.
func (a *Deckablage) DecksMitKarte(ctx context.Context, karteID int64) (
	[]domain.Deckbezug, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := a.db.QueryContext(ctx,
		`SELECT DISTINCT d.id, d.name
		   FROM deckplaetze p
		   JOIN decks d ON d.id = p.deck_id AND d.profil_id = ?
		  WHERE p.karte_id = ?
		  ORDER BY d.name`, profilID, karteID)
	if err != nil {
		return nil, uebersetze(err, "die Decks zur Karte %d lesen", karteID)
	}

	defer zeilen.Close() //nolint:errcheck // der Fehler steht in zeilen.Err()

	bezuege := []domain.Deckbezug{}

	for zeilen.Next() {
		var bezug domain.Deckbezug

		if err := zeilen.Scan(&bezug.ID, &bezug.Name); err != nil {
			return nil, uebersetze(err, "ein Deck zur Karte lesen")
		}

		bezuege = append(bezuege, bezug)
	}

	return bezuege, uebersetze(zeilen.Err(), "die Decks zur Karte lesen")
}
