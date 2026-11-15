package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kopieren legt eine Kopie eines Decks an — **Zeile und Plätze in einem
// Zug**, DOM-03.AK-02
// (#1698).
//
// # Was mitkommt und was nicht
//
// **Mit**: Spiel, Notiz, Format und jeder Platz an seiner Stelle — wer ein
// Deck kopiert, will dasselbe Deck unter einem zweiten Namen.
//
// **Nicht**: die Herkunft, denn die Kopie ist hier entstanden; Zuordnungen
// und Leihen, denn sie hängen an Stücken Papier, und eines liegt nicht in
// zwei Decks zugleich; Marken, die jemand dem Original gegeben hat.
//
// # Warum in SQL und nicht Platz für Platz
//
// **Die Plätze sind schon geprüft** — als sie ins Original kamen, samt Profil
// und Spiel. Ein `INSERT … SELECT` behält ihre Stelle, statt sie einzeln neu
// zu vergeben, und `RETURNING` sagt im selben Satz, ob es das Original im
// gezeigten Profil gab: ohne Zeile keine Kopie, und kein zweiter Ruf dafür.
func (a *Deckablage) Kopieren(ctx context.Context, quelle int64, name string) (
	domain.Deck, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Deck{}, profilfehler
	}

	var kopie domain.Deck

	fehler := imGeschaeft(ctx, a.db, "das Deck kopieren", func(zug *sql.Tx) error {
		var format sql.NullInt64

		err := zug.QueryRowContext(ctx,
			`INSERT INTO decks (name, spiel, notiz, format_id, profil_id)
			 SELECT ?, spiel, notiz, format_id, profil_id
			   FROM decks WHERE id = ? AND profil_id = ?
			 RETURNING id, fassung, name, spiel, notiz, format_id`,
			name, quelle, profilID).Scan(&kopie.ID, &kopie.Fassung, &kopie.Name,
			&kopie.Spiel, &kopie.Notiz, &format)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: ein Deck mit dieser Kennung gibt es nicht", domain.ErrNotFound)
		}

		if err != nil {
			return uebersetze(err, "das Deck kopieren")
		}

		kopie.FormatID = format.Int64

		if _, err := zug.ExecContext(ctx,
			`INSERT INTO deckplaetze (deck_id, karte_id, printing, druck_id, menge, bereich, stelle)
			 SELECT ?, karte_id, printing, druck_id, menge, bereich, stelle
			   FROM deckplaetze WHERE deck_id = ? ORDER BY stelle, id`,
			kopie.ID, quelle); err != nil {
			return uebersetze(err, "die Plätze des Decks kopieren")
		}

		plaetze, lesefehler := plaetzeIn(ctx, zug, kopie.ID)
		kopie.Plaetze = plaetze

		return lesefehler
	})
	if fehler != nil {
		return domain.Deck{}, fehler
	}

	kopie.Zaehlen()

	return kopie, nil
}
