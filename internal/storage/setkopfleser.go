// **Den Setkopf aus dem Bestand lesen** —
// #1325.
//
// Eigene Datei, weil `setleser.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort stehen die Sets und ihre Karten, hier die fünf
// Formatangaben eines einzelnen.

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const setkopfLesen = `
SELECT mse_fassung, spielfassung, stilfassung, versionskontrolle, apprentice_code
  FROM sets
 WHERE id = ? AND profil_id = ?`

// Setkopf gibt die fünf Formatangaben eines Sets.
//
// **Ein Set ohne Zeile ist ein Fehler und kein leerer Kopf.** Wer nach einem
// Set fragt, das es nicht gibt, hat sich vertan; ein leerer Kopf sähe aus wie
// „steht nicht in der Datei" und wäre eine Falschauskunft (§8).
func (l *SetLeser) Setkopf(ctx context.Context, setID int64) (domain.Setkopf, error) {
	var kopf domain.Setkopf

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.Setkopf{}, profilfehler
	}

	err := l.db.QueryRowContext(ctx, setkopfLesen, setID, profilID).Scan(
		&kopf.MSEFassung, &kopf.Spielfassung, &kopf.Stilfassung,
		&kopf.Versionskontrolle, &kopf.Apprenticecode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Setkopf{}, fmt.Errorf("%w: das Set %d gibt es nicht",
				domain.ErrNotFound, setID)
		}

		return domain.Setkopf{}, uebersetze(err, "den Setkopf lesen")
	}

	return kopf, nil
}
