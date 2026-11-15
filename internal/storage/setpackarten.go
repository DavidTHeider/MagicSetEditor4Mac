// Die **Packarten eines Sets** im Bestand
// (#1081).
//
// # Warum zwei Tabellen
//
// Weil eine Packart mehrere Posten hat und jeder drei Felder. Die Begründung
// steht ausführlich in der Wanderung `0017_setpackarten.sql` — seit Vermerk
// 0189 unter `git show v0.7.0:internal/storage/migrationen/…`.

package storage

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Setpackarten gibt die Packarten eines Sets — in ihrer Reihenfolge.
//
// **Die des Spiels stehen hier nicht.** Sie kommen aus dem Spielpaket, gehören
// keinem Set und sind unveränderlich; die Vorrangregel steht in
// `service.Packbereich`.
func (l *SetLeser) Setpackarten(ctx context.Context, setID int64) (
	[]*domain.PackArt, error,
) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	arten, err := l.packarten(ctx, setID)
	if err != nil {
		return nil, err
	}

	for _, art := range arten {
		posten, postenfehler := l.packposten(ctx, art.ID)
		if postenfehler != nil {
			return nil, postenfehler
		}

		art.Items = posten
	}

	return arten, nil
}

// packarten liest die Arten ohne ihre Posten.
func (l *SetLeser) packarten(ctx context.Context, setID int64) ([]*domain.PackArt, error) {
	zeilen, err := l.db.QueryContext(ctx,
		`SELECT id, reihenfolge, name, auswahl, filter, waehlbar
		   FROM set_pack_types WHERE set_id = ? ORDER BY reihenfolge`, setID)
	if err != nil {
		return nil, uebersetze(err, "Packarten lesen")
	}

	defer func() { _ = zeilen.Close() }()

	arten := []*domain.PackArt{}

	for zeilen.Next() {
		art := &domain.PackArt{}

		var waehlbar int
		if err := zeilen.Scan(&art.ID, &art.Reihenfolge, &art.Name,
			&art.Select, &art.Filter, &waehlbar); err != nil {
			return nil, uebersetze(err, "eine Packart lesen")
		}

		art.Selectable = waehlbar != 0
		art.Items = []domain.PackItem{}

		arten = append(arten, art)
	}

	return arten, uebersetze(zeilen.Err(), "Packarten lesen")
}

// packposten liest die Posten einer Art.
func (l *SetLeser) packposten(ctx context.Context, packID int64) ([]domain.PackItem, error) {
	zeilen, err := l.db.QueryContext(ctx,
		`SELECT name, anzahl, gewicht
		   FROM set_pack_items WHERE pack_id = ? ORDER BY reihenfolge`, packID)
	if err != nil {
		return nil, uebersetze(err, "Packposten lesen")
	}

	defer func() { _ = zeilen.Close() }()

	posten := []domain.PackItem{}

	for zeilen.Next() {
		var eines domain.PackItem
		if err := zeilen.Scan(&eines.Name, &eines.Amount, &eines.Weight); err != nil {
			return nil, uebersetze(err, "einen Packposten lesen")
		}

		posten = append(posten, eines)
	}

	return posten, uebersetze(zeilen.Err(), "Packposten lesen")
}

// setpackartenSchreiben legt die Packarten eines Sets ab.
//
// **In der Reihenfolge, in der sie ankommen** — sie ist die der Datei und
// damit der Ort, an den zurückgeschrieben wird.
func setpackartenSchreiben(
	ctx context.Context, tx *sql.Tx, setID int64, arten []*domain.PackArt,
) error {
	for stelle, art := range arten {
		waehlbar := 0
		if art.Selectable {
			waehlbar = 1
		}

		// **`RETURNING` und nicht `LastInsertId`** — derselbe Grund wie im
		// Kartenschreiber: der Rückgabewert von `LastInsertId` hat einen
		// Fehlerweg, den kein Test betreten kann, und ein Zweig, den niemand
		// betritt, ist keine Absicherung (§3). `RETURNING` gibt die Kennung
		// auf demselben Weg wie jeden anderen Lesefehler.
		var packID int64

		if err := tx.QueryRowContext(ctx,
			`INSERT INTO set_pack_types (set_id, reihenfolge, name, auswahl, filter, waehlbar)
			 VALUES (?, ?, ?, ?, ?, ?)
			 RETURNING id`,
			setID, stelle, art.Name, art.Select, art.Filter, waehlbar,
		).Scan(&packID); err != nil {
			return uebersetze(err, "eine Packart schreiben")
		}

		if postenfehler := packpostenSchreiben(ctx, tx, packID, art.Items); postenfehler != nil {
			return postenfehler
		}
	}

	return nil
}

// packpostenSchreiben legt die Posten einer Art ab.
func packpostenSchreiben(
	ctx context.Context, tx *sql.Tx, packID int64, posten []domain.PackItem,
) error {
	for stelle, eines := range posten {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO set_pack_items (pack_id, reihenfolge, name, anzahl, gewicht)
			 VALUES (?, ?, ?, ?, ?)`,
			packID, stelle, eines.Name, eines.Amount, eines.Weight); err != nil {
			return uebersetze(err, "einen Packposten schreiben")
		}
	}

	return nil
}
