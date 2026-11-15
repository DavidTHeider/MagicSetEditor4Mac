package storage

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KarteAnlegenMitWerten legt eine Karte samt ihren gerechneten Werten an —
// **in einem Schritt im Verlauf** (#2516): wer sie zurücknimmt, nimmt die
// Werte mit, und ein zweites ⌘Z für sie gäbe es nicht. Ohne Werte ist es
// [Kartenschreiber.KarteAnlegen].
func (k *Kartenschreiber) KarteAnlegenMitWerten(
	ctx context.Context, setID int64, werte []domain.Value,
) (int64, error) {
	tx, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, uebersetze(err, "eine Karte im Set %d anlegen", setID)
	}

	defer func() { _ = tx.Rollback() }() //nolint:errcheck // nach Commit ein ErrTxDone

	if fehler := setImProfil(ctx, tx, setID); fehler != nil {
		return 0, fehler
	}

	felder := make([]domain.Feldstand, 0, len(werte))
	for i, wert := range werte {
		felder = append(felder, domain.Feldstand{Feld: wert.Field, Wert: wert.Text, Reihenfolge: int64(i)})
	}

	kennung, fehler := k.karteEinreihen(ctx, tx, setID, felder)
	if fehler != nil {
		return 0, fehler
	}

	if fehler := tx.Commit(); fehler != nil {
		return 0, uebersetze(fehler, "eine Karte im Set %d anlegen", setID)
	}

	return kennung, nil
}
