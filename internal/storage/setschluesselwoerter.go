// Die **Schlüsselwörter eines Sets** im Bestand
// (#1062).
//
// # Warum eine eigene Tabelle
//
// Weil es mehrere sind und jedes fünf Felder hat. Als Setwert lägen sie flach
// nebeneinander (`keyword.0.match`), und jede Frage nach „dem dritten" wäre
// eine Zeichenkettenrechnung. Die Begründung steht ausführlich in der
// Wanderung `0016_setschluesselwoerter.sql` — seit Vermerk 0189 unter
// `git show v0.7.0:internal/storage/migrationen/…`.

package storage

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schluesselwoerter gibt die Schlüsselwörter eines Sets — in ihrer Reihenfolge.
//
// **Die des Spiels stehen hier nicht.** Sie kommen aus dem Spielpaket, gehören
// keinem Set und sind unveränderlich (`Keyword::fixed`, `src/data/keyword.hpp`,
// Zeile 104, Fassung f9d9356 vom 2021-07-06).
func (l *SetLeser) Schluesselwoerter(ctx context.Context, setID int64) (
	[]*domain.Keyword, error,
) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx,
		`SELECT id, reihenfolge, wort, muster, erinnerung, regeln, modus
		   FROM set_keywords WHERE set_id = ? ORDER BY reihenfolge`, setID)
	if err != nil {
		return nil, uebersetze(err, "Schlüsselwörter lesen")
	}

	defer func() { _ = zeilen.Close() }()

	woerter := []*domain.Keyword{}

	for zeilen.Next() {
		wort := &domain.Keyword{}
		if err := zeilen.Scan(&wort.ID, &wort.Reihenfolge, &wort.Word, &wort.Match,
			&wort.Reminder, &wort.Rules, &wort.Mode); err != nil {
			return nil, uebersetze(err, "ein Schlüsselwort lesen")
		}

		woerter = append(woerter, wort)
	}

	return woerter, uebersetze(zeilen.Err(), "Schlüsselwörter lesen")
}

// schluesselwoerterSchreiben legt die Schlüsselwörter eines Sets ab.
//
// **In der Reihenfolge, in der sie ankommen** — sie ist die der Datei, und sie
// entscheidet, welches zuerst auf einen Text passt.
func schluesselwoerterSchreiben(
	ctx context.Context, tx *sql.Tx, setID int64, woerter []*domain.Keyword,
) error {
	for stelle, wort := range woerter {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO set_keywords
			     (set_id, reihenfolge, wort, muster, erinnerung, regeln, modus)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			setID, stelle, wort.Word, wort.Match,
			wort.Reminder, wort.Rules, wort.Mode); err != nil {
			return uebersetze(err, "ein Schlüsselwort schreiben")
		}
	}

	return nil
}
