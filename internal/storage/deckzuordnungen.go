package storage

// **Die Zuordnung im Bestand** — Werkstück U9 von Phase 5.1
// (#1420),
// Wanderung 0052. Was eine Zuordnung ist und was sie nicht ist, steht in
// `domain/deckzuordnung.go`; hier steht nur, wie sie liegt.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Deckzuordnungen ist die Ablage der Zuordnungen.
type Deckzuordnungen struct{ db *sql.DB }

// NeueDeckzuordnungen baut sie über einer Datenbank.
func NeueDeckzuordnungen(db *sql.DB) *Deckzuordnungen { return &Deckzuordnungen{db: db} }

// zuordnungsspalten sind die Spalten in der Reihenfolge, in der gelesen wird.
const zuordnungsspalten = `z.id, z.deck_id, z.platz_id, z.exemplar_id, z.menge`

// Setzen legt eine Zuordnung an oder ersetzt ihre Menge.
//
// **`ON CONFLICT … DO UPDATE` und `RETURNING`**: dieselbe Zeile ist dieselbe
// Kennung mit neuer Menge — und ein Weg weniger, den kein echter Treiber
// hergibt (§3).
func (z *Deckzuordnungen) Setzen(
	ctx context.Context, zuordnung domain.Deckzuordnung,
) (domain.Deckzuordnung, error) {
	if befunde := zuordnung.Validate(); len(befunde) > 0 {
		return domain.Deckzuordnung{}, befunde[0]
	}

	fehler := imGeschaeft(ctx, z.db, "die Zuordnung setzen", func(tx *sql.Tx) error {
		// **Platz, Deck und Exemplar aus dem gezeigten Profil, und der Platz
		// aus diesem Deck** (#1815). Der Bestand hält keines davon selbst
		// (0052); der Dienst fragt es vorher, die Ablage sagt es zu.
		if fehler := platzImDeck(ctx, tx, zuordnung.DeckID, zuordnung.PlatzID); fehler != nil {
			return fehler
		}

		if fehler := exemplarImProfil(ctx, tx, zuordnung.ExemplarID); fehler != nil {
			return fehler
		}

		return uebersetze(tx.QueryRowContext(ctx,
			`INSERT INTO deckzuordnungen (deck_id, platz_id, exemplar_id, menge)
			 VALUES (?, ?, ?, ?)
			 ON CONFLICT (platz_id, exemplar_id) DO UPDATE SET menge = excluded.menge
			 RETURNING id`,
			zuordnung.DeckID, zuordnung.PlatzID, zuordnung.ExemplarID, zuordnung.Menge).
			Scan(&zuordnung.ID), "die Zuordnung setzen")
	})
	if fehler != nil {
		return domain.Deckzuordnung{}, fehler
	}

	return zuordnung, nil
}

// Loesen nimmt eine Zuordnung aus ihrem Deck — und nur aus ihrem.
//
// **Über das Deck und die Kennung**, damit ein fremdes Deck keine Zuordnung
// eines anderen löst; `RETURNING` sagt, ob es sie gab.
func (z *Deckzuordnungen) Loesen(ctx context.Context, deckID, id int64) error {
	return imGeschaeft(ctx, z.db, "die Zuordnung lösen", func(tx *sql.Tx) error {
		// **Nur aus einem Deck des gezeigten Profils** (#1815).
		if fehler := deckImProfil(ctx, tx, deckID); fehler != nil {
			return fehler
		}

		var geloest int64

		err := tx.QueryRowContext(ctx,
			`DELETE FROM deckzuordnungen WHERE deck_id = ? AND id = ? RETURNING id`, deckID, id).
			Scan(&geloest)

		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: Zuordnung %d gibt es in Deck %d nicht", domain.ErrNotFound, id, deckID)
		}

		return uebersetze(err, "die Zuordnung lösen")
	})
}

// JeDeck gibt die Zuordnungen eines Decks in der Reihenfolge seiner Plätze.
func (z *Deckzuordnungen) JeDeck(ctx context.Context, deckID int64) ([]domain.Deckzuordnung, error) {
	if fehler := deckImProfil(ctx, z.db, deckID); fehler != nil {
		return nil, fehler
	}

	return z.lesen(ctx,
		`SELECT `+zuordnungsspalten+` FROM deckzuordnungen z
		  WHERE z.deck_id = ? ORDER BY z.platz_id, z.id`, deckID)
}

// JeExemplar gibt die Zuordnungen eines Exemplars — der Rückweg von der
// Sammlung zu den Decks, die es benutzen.
func (z *Deckzuordnungen) JeExemplar(ctx context.Context, exemplarID int64) ([]domain.Deckzuordnung, error) {
	if fehler := exemplarImProfil(ctx, z.db, exemplarID); fehler != nil {
		return nil, fehler
	}

	return z.lesen(ctx,
		`SELECT `+zuordnungsspalten+` FROM deckzuordnungen z
		  WHERE z.exemplar_id = ? ORDER BY z.deck_id, z.id`, exemplarID)
}

// ZugeordnetJeExemplar summiert die Zuordnungen je Exemplar über alle Decks
// des Profils.
//
// **Über das Deck auf das Profil geklemmt** (Vermerk 0066): die Zuordnung
// ist ein Abkömmling und trägt kein Profil; das Deck trägt es.
func (z *Deckzuordnungen) ZugeordnetJeExemplar(ctx context.Context) (map[int64]int, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := z.db.QueryContext(ctx,
		`SELECT z.exemplar_id, SUM(z.menge) FROM deckzuordnungen z
		   JOIN decks d ON d.id = z.deck_id
		  WHERE d.profil_id = ?
		  GROUP BY z.exemplar_id`, profilID)
	if err != nil {
		return nil, uebersetze(err, "die Zuordnungen summieren")
	}

	defer func() { _ = zeilen.Close() }()

	summen := map[int64]int{}

	for zeilen.Next() {
		var (
			exemplarID int64
			menge      int
		)

		if err := zeilen.Scan(&exemplarID, &menge); err != nil {
			return nil, uebersetze(err, "eine Zuordnungssumme lesen")
		}

		summen[exemplarID] = menge
	}

	return summen, uebersetze(zeilen.Err(), "die Zuordnungen summieren")
}

// lesen führt eine Abfrage aus und baut die Zuordnungen — nie `nil`.
func (z *Deckzuordnungen) lesen(ctx context.Context, abfrage string, werte ...any) (
	[]domain.Deckzuordnung, error,
) {
	zeilen, err := z.db.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "die Zuordnungen lesen")
	}

	defer func() { _ = zeilen.Close() }()

	zuordnungen := []domain.Deckzuordnung{}

	for zeilen.Next() {
		var zuordnung domain.Deckzuordnung

		if err := zeilen.Scan(&zuordnung.ID, &zuordnung.DeckID, &zuordnung.PlatzID,
			&zuordnung.ExemplarID, &zuordnung.Menge); err != nil {
			return nil, uebersetze(err, "eine Zuordnung lesen")
		}

		zuordnungen = append(zuordnungen, zuordnung)
	}

	return zuordnungen, uebersetze(zeilen.Err(), "die Zuordnungen lesen")
}
