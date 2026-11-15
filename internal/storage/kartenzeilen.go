// **Die Kartenliste, ohne alles mitzubringen** —
// #891,
// Werkstück 5 von Phase Y.
//
// # Was gemessen wurde
//
// `Black.mse-set` hat **3.788 Karten**. `Karten` gibt sie mit **jedem**
// Feldwert zurück: **4.761.182 Bytes JSON** je Aufruf, 35,4 ms allein auf der
// Go-Seite — und danach über die Wails-Naht und durch den JSON-Zerteiler des
// Fensters (gemessen am 2026-08-25, `BenchmarkKartenliste`).
//
// **Die Liste braucht davon einen Namen je Karte.** Der Rest ist für die
// **eine** Karte da, die gerade offen ist.
//
// # Warum die Zahl der Felder mitkommt
//
// Die Liste zeigt „ohne Felder" an einer Karte, die keine hat — das ist der
// Unterschied zwischen „heißt nicht" und „ist leer". Sie zu zählen kostet die
// Datenbank nichts; sie mitzuschicken wäre der ganze Aufwand gewesen.

package storage

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kartenzeilenLesen gibt je Karte eine Zeile.
//
// **Ein Unterausdruck für den Namen und einer für die Zahl.** Der Verbund
// darüber gäbe je Feldwert eine Zeile — genau das, was hier vermieden werden
// soll.
const kartenzeilenLesen = `
SELECT cards.id, cards.reihenfolge,
       coalesce((SELECT wert FROM card_values
                  WHERE card_values.card_id = cards.id
                    AND card_values.feld = 'name'), ''),
       (SELECT count(*) FROM card_values WHERE card_values.card_id = cards.id)
  FROM cards
 WHERE cards.set_id = ?
 ORDER BY cards.reihenfolge`

// Kartenzeilen gibt die Karten eines Sets als Liste — ohne ihre Feldwerte.
func (l *SetLeser) Kartenzeilen(ctx context.Context, setID int64) (
	[]domain.CardLine, error,
) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx, kartenzeilenLesen, setID)
	if err != nil {
		return nil, uebersetze(err, "die Kartenliste lesen")
	}

	defer func() { _ = zeilen.Close() }()

	liste := []domain.CardLine{}

	for zeilen.Next() {
		var zeile domain.CardLine

		if scanfehler := zeilen.Scan(
			&zeile.ID, &zeile.Number, &zeile.Name, &zeile.Fields,
		); scanfehler != nil {
			return nil, uebersetze(scanfehler, "eine Kartenzeile lesen")
		}

		liste = append(liste, zeile)
	}

	if abschluss := zeilen.Err(); abschluss != nil {
		return nil, uebersetze(abschluss, "die Kartenliste lesen")
	}

	return liste, nil
}
