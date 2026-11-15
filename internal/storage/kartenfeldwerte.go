// **Ein Feld über ein ganzes Set** —
// #891.
//
// Die Massenänderung zeigt neben jeder angehakten Karte den Wert, den das
// gewählte Feld **heute** hat; ohne ihn wäre eine Rücknahme unsichtbar — sie
// stellte etwas wieder her, was niemand gesehen hat.
//
// Bis zum 2026-08-25 kam dieser Wert mit der Kartenliste, also mit **jedem**
// Feld **jeder** Karte. Diese Abfrage gibt einen kurzen Wert je Karte, die ihn
// hat, in **einem** Aufruf.

package storage

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kartenfeldwerteLesen gibt je Karte **mit** diesem Feld eine Zeile.
//
// **Ein Verbund und kein Unterausdruck**, anders als in `kartenzeilenLesen`:
// dort ist ein Feld je Karte gesucht und der Verbund gäbe jedes; hier ist
// genau eines gemeint, und der Verbund ist die kürzere Antwort.
const kartenfeldwerteLesen = `
SELECT cards.id, card_values.wert
  FROM cards
  JOIN card_values ON card_values.card_id = cards.id
 WHERE cards.set_id = ? AND card_values.feld = ?
 ORDER BY cards.reihenfolge`

// Kartenfeldwerte gibt den Wert eines Feldes über alle Karten eines Sets.
func (l *SetLeser) Kartenfeldwerte(ctx context.Context, setID int64, feld string) (
	[]domain.CardField, error,
) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx, kartenfeldwerteLesen, setID, feld)
	if err != nil {
		return nil, uebersetze(err, "die Feldwerte lesen")
	}

	defer func() { _ = zeilen.Close() }()

	werte := []domain.CardField{}

	for zeilen.Next() {
		var wert domain.CardField

		if scanfehler := zeilen.Scan(&wert.ID, &wert.Text); scanfehler != nil {
			return nil, uebersetze(scanfehler, "einen Feldwert lesen")
		}

		werte = append(werte, wert)
	}

	if abschluss := zeilen.Err(); abschluss != nil {
		return nil, uebersetze(abschluss, "die Feldwerte lesen")
	}

	return werte, nil
}
