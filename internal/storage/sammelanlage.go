package storage

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// AlleAnlegen legt mehrere Exemplare an — **alle oder keines**, DOM-03.AK-02
// (#1698).
//
// **Ein Anwendungsfall, ein Zug** (#1820): die Fehlkarten eines Decks kommen
// zusammen auf die Wunschliste. Stößt eine an eine Zeile, die es schon gibt,
// geht die ganze Liste zurück — eine halbe sähe aus wie eine vollständige,
// und der nächste Versuch fände die erste Hälfte schon da.
//
// **Welche Zeilen es werden, entscheidet der Dienst** (§2); hier wird
// geschrieben, was entschieden ist.
func (s *Sammlungsspeicher) AlleAnlegen(ctx context.Context, exemplare []*domain.Exemplar) (
	[]int64, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	kennungen := make([]int64, 0, len(exemplare))

	fehler := imGeschaeft(ctx, s.db, "Exemplare anlegen", func(zug *sql.Tx) error {
		for _, exemplar := range exemplare {
			id, anlagefehler := exemplarEinfuegen(ctx, zug, exemplar, profilID)
			if anlagefehler != nil {
				return anlagefehler
			}

			kennungen = append(kennungen, id)
		}

		return nil
	})
	if fehler != nil {
		return nil, fehler
	}

	return kennungen, nil
}
