package storage

// **Ob eine Zusage zum Vertrag passt — in der Transaktion des Eintragens** —
// #2492.
//
// Bis dahin kam hier eine Zahl an: wie viel der Handel von dieser Partei
// nimmt. Ob das gewählte Exemplar eine der Zeilen **erfüllt**, fragte
// niemand, und eine Zeile in anderem Zustand ging beim Verbuchen weg, während
// drüben die abgesprochene ankam. **Jetzt kommen die Zeilen**, und gerechnet
// wird je Variante ([domain.Zusagevertrag]).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vertragPruefenInTx sagt, ob das Exemplar eine Zeile des Vertrags erfüllt
// und ob von seiner Variante noch so viel frei ist.
func vertragPruefenInTx(
	ctx context.Context, wer bestandsfrager, profilID int64,
	exemplar domain.Exemplar, zusage domain.Reservierung,
	vertrag domain.Zusagevertrag,
) error {
	grenze := vertrag.GrenzeFuer(exemplar)
	if grenze == 0 {
		return domain.NewFieldError("exemplar", domain.SchluesselZusageUnpassend,
			"Dieses Exemplar erfüllt keine Zeile, die du gibst — Zustand, "+
				"Sprache oder Ausführung weichen ab. Wähle ein passendes oder ergänze die Sammlung.")
	}

	gebunden, fehler := gebundenInTx(ctx, wer, profilID, zusage.Vorgang, exemplar)
	if fehler != nil {
		return fehler
	}

	if gebunden+zusage.Menge > grenze {
		return domain.MitGrund(domain.ErrConflict, nil,
			"dieser Handel nimmt %d Stück dieser Art von dir, festgehalten "+
				"sind schon %d — mehr bindet er nicht", grenze, gebunden)
	}

	return nil
}

// gebundenInTx zählt, was dieser Vorgang von der Variante des Exemplars
// schon hält — über alle Besitzzeilen dieser Variante.
func gebundenInTx(
	ctx context.Context, wer bestandsfrager, profilID int64, vorgang string,
	exemplar domain.Exemplar,
) (int, error) {
	zeilen, fehler := wer.QueryContext(ctx,
		`SELECT r.menge, s.printing, coalesce(s.druck_id, 0), s.zustand,
		        s.sprache, s.finish
		   FROM reservierungen r
		   JOIN sammlung s ON s.id = r.exemplar_id AND s.profil_id = r.profil_id
		  WHERE r.profil_id = ? AND r.vorgang = ?`,
		profilID, vorgang)
	if fehler != nil {
		return 0, fmt.Errorf("die Zusagen des Vorgangs ließen sich nicht lesen: %w", fehler)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen ist gelesen

	gebunden := 0

	for zeilen.Next() {
		var (
			menge  int
			andere domain.Exemplar
		)

		if fehler := zeilen.Scan(&menge, &andere.Printing, &andere.DruckID,
			&andere.Zustand, &andere.Sprache, &andere.Finish); fehler != nil {
			return 0, fmt.Errorf("eine Zusage ließ sich nicht lesen: %w", fehler)
		}

		if andere.GleicheVariante(exemplar) {
			gebunden += menge
		}
	}

	if fehler := zeilen.Err(); fehler != nil {
		return 0, fmt.Errorf("die Zusagen des Vorgangs ließen sich nicht zu Ende lesen: %w", fehler)
	}

	return gebunden, nil
}
