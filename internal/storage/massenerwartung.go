package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// FelderSetzenMitErwartung schreibt ein Feld auf mehrere Karten — aber nur,
// wenn jede noch auf dem Wert steht, den die Vorschau gezeigt hat
// (Werkstück U6, #1417).
//
// **Vergleichen, dann schreiben, in einer Transaktion.** Zwischen Vorschau
// und Anwenden kann jemand anderes schreiben; wer den gezeigten Altwert
// dann überschriebe, überschriebe etwas, das niemand gesehen hat. Weicht
// eine Karte ab, wird **nichts** geschrieben, und der Fehler ist ein
// [domain.ErrConflict], der die abweichenden Karten nennt — alle, nicht die
// erste: wer zwölf Karten ändert, will nicht zwölfmal anlaufen.
//
// **Die Erwartung ist die Liste der Karten.** Eine Karte ohne Erwartung
// wird nicht geschrieben; die Vorschau lässt die ausgeschlossenen weg.
func (k *Kartenschreiber) FelderSetzenMitErwartung(
	ctx context.Context, erwartet []domain.Feldvorwert, feld, wert string,
) (string, error) {
	if len(erwartet) == 0 {
		return "", nil
	}

	if befund := domain.PruefeFeldname(feld); befund != nil {
		return "", befund
	}

	transaktion, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return "", uebersetze(err, "Massenänderung beginnen")
	}

	defer func() { _ = transaktion.Rollback() }()

	karten := make([]int64, 0, len(erwartet))
	for _, vorwert := range erwartet {
		karten = append(karten, vorwert.KarteID)
	}

	if err := fehlendeMelden(ctx, transaktion, karten); err != nil {
		return "", err
	}

	if err := abweichendeMelden(ctx, transaktion, erwartet, feld); err != nil {
		return "", err
	}

	gruppe := neueGruppe()

	for _, karte := range karten {
		if err := k.massenzeile(ctx, transaktion, karte, feld, wert, gruppe); err != nil {
			return "", err
		}
	}

	if err := transaktion.Commit(); err != nil {
		return "", uebersetze(err, "Massenänderung abschließen")
	}

	return gruppe, nil
}

// abweichendeMelden vergleicht jede Erwartung mit dem Stand und nennt alle
// Karten, die abweichen.
func abweichendeMelden(
	ctx context.Context, tx *sql.Tx, erwartet []domain.Feldvorwert, feld string,
) error {
	var abweichend []int64

	for _, vorwert := range erwartet {
		var stand sql.NullString

		err := tx.QueryRowContext(ctx,
			`SELECT wert FROM card_values WHERE card_id = ? AND feld = ?`,
			vorwert.KarteID, feld).Scan(&stand)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return &domain.BulkError{
				CardIDs: []int64{vorwert.KarteID},
				Err:     uebersetze(err, "den Stand von %q lesen", feld),
			}
		}

		if stand.Valid != vorwert.GabEs || (stand.Valid && stand.String != vorwert.Wert) {
			abweichend = append(abweichend, vorwert.KarteID)
		}
	}

	if len(abweichend) == 0 {
		return nil
	}

	return &domain.BulkError{
		CardIDs: abweichend,
		Err: fmt.Errorf("%w: das Feld %q steht nicht mehr auf dem Wert, den die Vorschau gezeigt hat",
			domain.ErrConflict, feld),
	}
}
