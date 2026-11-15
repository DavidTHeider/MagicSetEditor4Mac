package storage

import (
	"context"
)

// Kartenstile gibt die eigenen Stile der Karten eines Sets — jeden Wert von
// `stylesheet:` an einer seiner Karten, einmal (#2423).
//
// **Für den Abdruck einer Problemprüfung** (PERF02, #2352): er faltet die
// Pakete jedes Stils, aus dem gezeichnet wird, und ein Stil an der Karte geht
// vor dem des Sets (`service.StilFuerKarte`, `src/data/set.cpp`, Zeilen
// 90–93, Fassung f9d9356 vom 2021-07-06).
//
// **Über die Karten des Sets und den Schlüssel der Werte** — `cards_nach_set`,
// dann `(card_id, feld)`; kein Gang über alle Werte.
func (l *SetLeser) Kartenstile(ctx context.Context, setID int64) ([]string, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx,
		`SELECT DISTINCT v.wert
		   FROM cards c
		   JOIN card_values v ON v.card_id = c.id AND v.feld = 'stylesheet'
		  WHERE c.set_id = ?`, setID)
	if err != nil {
		return nil, uebersetze(err, "die Kartenstile lesen")
	}

	defer func() { _ = zeilen.Close() }()

	var stile []string

	for zeilen.Next() {
		var stil string
		if err := zeilen.Scan(&stil); err != nil {
			return nil, uebersetze(err, "einen Kartenstil lesen")
		}

		stile = append(stile, stil)
	}

	return stile, uebersetze(zeilen.Err(), "die Kartenstile lesen")
}
