package storage

// **Eine Massenänderung, die rechnet** — #2516 (Rest aus #2396, Vermerk 0203).
//
// [Kartenschreiber.FelderSetzen] schreibt **einen** Rohwert auf viele Karten;
// ein gerechneter Wert fällt aber je Karte anders aus, und was ihn liest,
// zieht nach. Hier steht deshalb je Karte ein eigener Schritt — und doch
// **ein** Schritt im Verlauf, alles oder nichts in einer Transaktion.
//
// **Zwei Bedingungen, beide in der Transaktion, die schreibt**: die Erwartung
// aus der Vorschau (U6, #1417, [abweichendeMelden]) und der Stand, aus dem
// je Karte gerechnet wurde ([standHalten], #2533). MSE2 hat keine
// Massenänderung; die Marke ist `new`.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// FelderGerechnetSetzen schreibt je Karte ihre gerechneten Werte und gibt die
// Gruppe des Schritts zurück — leer, wenn es keine Karte gibt. `erwartet`
// nennt die Altwerte des Felds `feld` aus der Vorschau; `nil` heißt: keine
// Erwartung, wie bei [Kartenschreiber.FelderSetzen].
func (k *Kartenschreiber) FelderGerechnetSetzen(
	ctx context.Context, schritte []domain.Kartenschritt, erwartet []domain.Feldvorwert, feld string,
) (string, error) {
	if len(schritte) == 0 {
		return "", nil
	}

	if befund := domain.PruefeFeldname(feld); befund != nil {
		return "", befund
	}

	karten := make([]int64, 0, len(schritte))
	for _, schritt := range schritte {
		karten = append(karten, schritt.KarteID)
	}

	// **Ein Schritt für alle** (#52), wie bei der Gruppe aus der KI.
	einer := k.inEinemSchritt()

	fehler := imGeschaeft(ctx, k.db, "Massenänderung", func(tx *sql.Tx) error {
		if fehlend := fehlendeMelden(ctx, tx, karten); fehlend != nil {
			return fehlend
		}

		if erwartet != nil {
			if abweichend := abweichendeMelden(ctx, tx, erwartet, feld); abweichend != nil {
				return abweichend
			}
		}

		for _, schritt := range schritte {
			if gescheitert := einer.gerechneteKarte(ctx, tx, schritt); gescheitert != nil {
				return &domain.BulkError{CardIDs: []int64{schritt.KarteID}, Err: gescheitert}
			}
		}

		return nil
	})
	if fehler != nil {
		return "", fehler
	}

	return einer.gruppe, nil
}

// gerechneteKarte schreibt die Werte einer Karte in die laufende Transaktion
// — **auf dem Stand, aus dem gerechnet wurde**.
func (k *Kartenschreiber) gerechneteKarte(
	ctx context.Context, tx *sql.Tx, schritt domain.Kartenschritt,
) error {
	if fehler := standHalten(ctx, tx, schritt.KarteID, schritt.Erwartet); fehler != nil {
		return fehler
	}

	for _, wert := range schritt.Werte {
		// **Nachgezogenes ohne Bedingung**, wie beim einzelnen Feld.
		lauf := ctx
		if wert.Nachgezogen {
			lauf = ohneBekannte(ctx)
		}

		if _, _, _, fehler := k.imLauf(lauf, tx, schritt.KarteID, wert.Field,
			func(tx *sql.Tx) error {
				_, fehler := tx.ExecContext(ctx, feldSetzen,
					schritt.KarteID, wert.Field, wert.Text, schritt.KarteID)

				return fehler
			}); fehler != nil {
			return fehler
		}
	}

	return nil
}
