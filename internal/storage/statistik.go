package storage

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein leerer Wert ist keine Belegung.** Sonst sähe ein Set vollständig aus,
// dessen Felder alle leer sind — und die Statistik behauptete das Gegenteil
// dessen, wofür sie da ist.
const nichtLeer = `trim(card_values.wert) <> ''`

const kennzahlen = `
SELECT (SELECT count(*) FROM cards WHERE set_id = :set),
       (SELECT count(DISTINCT card_values.feld)
          FROM card_values
          JOIN cards ON cards.id = card_values.card_id
         WHERE cards.set_id = :set),
       (SELECT count(*) FROM cards
         WHERE set_id = :set
           AND id NOT IN (SELECT card_values.card_id
                            FROM card_values
                           WHERE ` + nichtLeer + `))`

const belegung = `
SELECT card_values.feld, count(*)
  FROM card_values
  JOIN cards ON cards.id = card_values.card_id
 WHERE cards.set_id = ? AND ` + nichtLeer + `
 GROUP BY card_values.feld
 ORDER BY count(*) DESC, card_values.feld`

// verteilung gibt **eine Zeile je Karte**, geordnet wie die Werte: der
// häufigste zuerst, dann nach Wert, dann nach Karte.
//
// **Nicht `GROUP BY` mit `group_concat`**: das gäbe die Karten als Text, den
// jemand wieder zerlegen müsste, und `group_concat` sagt über die Reihenfolge
// seiner Zeilen nichts zu. Die Fensterfunktion `count(*) OVER` zählt je Wert,
// ohne die Zeilen zu verlieren — die Karten kommen in der Ordnung der Abfrage.
const verteilung = `
SELECT card_values.wert,
       count(*) OVER (PARTITION BY card_values.wert) AS karten,
       card_values.card_id
  FROM card_values
  JOIN cards ON cards.id = card_values.card_id
 WHERE cards.set_id = ? AND card_values.feld = ? AND ` + nichtLeer + `
 ORDER BY karten DESC, card_values.wert, card_values.card_id`

// Kennzahlen gibt die Statistik eines Sets.
//
// **Aus Abfragen und nicht aus einer Schleife über die Karten.** Der
// Unterschied ist nicht Geschmack: eine Schleife lädt 3.000 Karten in den
// Arbeitsspeicher, um vier Zahlen zu bilden, und tut das bei jedem Blick.
func (l *SetLeser) Kennzahlen(ctx context.Context, setID int64) (domain.SetStats, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return domain.SetStats{}, fehler
	}

	var zahlen domain.SetStats

	if err := l.db.QueryRowContext(ctx, kennzahlen, sql.Named("set", setID)).
		Scan(&zahlen.Cards, &zahlen.Fields, &zahlen.Empty); err != nil {
		return domain.SetStats{}, uebersetze(err, "Kennzahlen lesen")
	}

	return zahlen, nil
}

// Belegung sagt je Feld, wie viele Karten es nichtleer tragen.
//
// **Die Auskunft, für die jemand herkommt:** ein Feld, das nur die Hälfte der
// Karten führt, ist entweder Absicht oder eine halbe Arbeit — und beides sieht
// man nur an dieser Zahl.
func (l *SetLeser) Belegung(ctx context.Context, setID int64) ([]domain.FieldCount, error) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx, belegung, setID)
	if err != nil {
		return nil, uebersetze(err, "die Belegung lesen")
	}

	defer func() { _ = zeilen.Close() }()

	felder := []domain.FieldCount{}

	for zeilen.Next() {
		var feld domain.FieldCount
		if err := zeilen.Scan(&feld.Field, &feld.Cards); err != nil {
			return nil, uebersetze(err, "eine Belegung lesen")
		}

		felder = append(felder, feld)
	}

	return felder, uebersetze(zeilen.Err(), "die Belegung lesen")
}

// Verteilung sagt, wie viele Karten welchen Wert eines Felds tragen — und
// welche.
//
// **Die Karten kommen aus derselben Abfrage** (UXR-F6, #2013): MSE2 behält zu
// jedem Element den `original_index` seiner Karte, und die Wahl einer Gruppe
// gibt diese Karten zurück (`GraphData::indices`, `src/gui/control/graph.cpp`
// Zeilen 249–258, Fassung f9d9356 vom 2021-07-06). Eine zweite Abfrage dafür
// sähe einen Bestand, der sich zwischen beiden geändert haben kann.
func (l *SetLeser) Verteilung(ctx context.Context, setID int64, feld string) (
	[]domain.ValueCount, error,
) {
	// **Nur aus dem gezeigten Profil** (#1815).
	if fehler := setImProfil(ctx, l.db, setID); fehler != nil {
		return nil, fehler
	}

	zeilen, err := l.db.QueryContext(ctx, verteilung, setID, feld)
	if err != nil {
		return nil, uebersetze(err, "die Verteilung lesen")
	}

	defer func() { _ = zeilen.Close() }()

	werte := []domain.ValueCount{}

	for zeilen.Next() {
		var (
			wert    string
			karten  int
			kennung int64
		)

		if err := zeilen.Scan(&wert, &karten, &kennung); err != nil {
			return nil, uebersetze(err, "eine Verteilung lesen")
		}

		// **Die Zeilen eines Werts stehen beieinander** — das ordnet die
		// Abfrage —, also gehört die Karte zum Wert davor oder beginnt einen.
		if letzter := len(werte) - 1; letzter >= 0 && werte[letzter].Value == wert {
			werte[letzter].CardIDs = append(werte[letzter].CardIDs, kennung)

			continue
		}

		werte = append(werte, domain.ValueCount{
			Value: wert, Cards: karten, CardIDs: []int64{kennung},
		})
	}

	return werte, uebersetze(zeilen.Err(), "die Verteilung lesen")
}
