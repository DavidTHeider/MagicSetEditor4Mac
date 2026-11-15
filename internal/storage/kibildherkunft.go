package storage

// **Was über ein übernommenes Bild zusätzlich feststeht** — Werkstück A12 von
// Phase 5.3 (#1451).
//
// # Warum das Schreiben wieder in einer fremden Transaktion steht
//
// **Aus demselben Grund wie in A11**: die Angaben gehören zum Bild, und das
// Bild steht im Feld. Eine eigene Verbindung hätte eine eigene Transaktion —
// und dann gäbe es einen Stand, in dem das Bild da ist und die Auskunft
// darüber nicht.
//
// **Und sie stehen nach der Feldherkunft**, nicht davor: der Fremdschlüssel
// zeigt auf sie, und eine Angabe ohne ihre Herkunft ist eine Angabe über
// nichts.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bildangabenSchreiben legt eine Angabe an — und ersetzt die vorige.
const bildangabenSchreiben = `
	INSERT INTO ki_bildherkunft (profil_id, karte_id, feld, name, wert)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT (profil_id, karte_id, feld, name) DO UPDATE SET
		wert = excluded.wert`

// bildangabenRaeumen wirft weg, was von der vorigen Übernahme übrig ist.
//
// **Weil das Ersetzen einer Zeile die anderen stehen ließe.** Nennt der
// zweite Lauf keinen Seed, stünde der erste noch da — und niemand sähe, dass
// er zu einem Bild gehört, das es nicht mehr gibt.
const bildangabenRaeumen = `
	DELETE FROM ki_bildherkunft
	WHERE profil_id = ? AND karte_id = ? AND feld = ?`

// bildangabenLesen gibt die Angaben eines Feldes.
const bildangabenLesen = `
	SELECT name, wert FROM ki_bildherkunft
	WHERE profil_id = ? AND karte_id = ? AND feld = ?
	ORDER BY name`

// bildangabenSetzen schreibt die Angaben eines Feldes in eine **bestehende**
// Transaktion.
func bildangabenSetzen(
	ctx context.Context, tx *sql.Tx, profilID, karteID int64, feld string,
	angaben []domain.KIBildparameter,
) error {
	if _, fehler := tx.ExecContext(ctx, bildangabenRaeumen,
		profilID, karteID, feld); fehler != nil {
		return uebersetze(fehler, "die Bildangaben von %q räumen", feld)
	}

	for _, angabe := range angaben {
		if angabe.Name == "" {
			return domain.MitGrund(domain.ErrValidation, nil,
				"eine Bildangabe von %q hat keinen Namen", feld)
		}

		if _, fehler := tx.ExecContext(ctx, bildangabenSchreiben,
			profilID, karteID, feld, angabe.Name, angabe.Wert); fehler != nil {
			return uebersetze(fehler, "die Bildangabe %q von %q schreiben",
				angabe.Name, feld)
		}
	}

	return nil
}

// Bildangaben gibt die Angaben zu einem Bildfeld — sortiert nach Namen.
func (h *Feldherkunft) Bildangaben(ctx context.Context, karteID int64, feld string) (
	[]domain.KIBildparameter, error,
) {
	profilID, fehler := ProfilAus(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen, fehler := h.db.QueryContext(ctx, bildangabenLesen,
		profilID, karteID, feld)
	if fehler != nil {
		return nil, uebersetze(fehler, "die Bildangaben von %q lesen", feld)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird oben

	gelesen := make([]domain.KIBildparameter, 0)

	for zeilen.Next() {
		var angabe domain.KIBildparameter

		if fehler := zeilen.Scan(&angabe.Name, &angabe.Wert); fehler != nil {
			return nil, uebersetze(fehler, "eine Bildangabe von %q lesen", feld)
		}

		gelesen = append(gelesen, angabe)
	}

	// **Der Iterationsfehler steht im Rückgabewert und nicht in einem `if`**
	// — dieselbe Gestalt wie nebenan (§10).
	return gelesen, uebersetze(zeilen.Err(), "die Bildangaben von %q lesen", feld)
}
