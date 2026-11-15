package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// WerteSetzen schreibt **verschiedene** Werte auf verschiedene Karten — als
// **ein** Schritt im Verlauf.
//
// Werkstück 2 von Phase 3.7
// (#1168).
//
// # Der Unterschied zu [Kartenschreiber.FelderSetzen] daneben
//
// Jene schreibt **einen** Wert auf viele Karten — die Massenänderung. Eine
// Ersetzung schreibt auf jede Karte einen **anderen**: ihren eigenen Text mit
// der ersetzten Stelle darin. Ein Aufruf je Karte gäbe je Karte einen Schritt,
// und ⌘Z nähme dreihundertmal eine einzelne zurück.
//
// # Alles oder nichts
//
// **In einer Transaktion**, und die Prüfung, ob es alle Karten gibt, steht
// darin — dieselbe Zusage und derselbe Grund wie bei der Massenänderung: eine
// Ersetzung, die zur Hälfte durchgeht, hinterlässt einen Bestand, von dem
// niemand sagen kann, welche Hälfte, und die Rücknahme müsste raten.
func (k *Kartenschreiber) WerteSetzen(
	ctx context.Context, wechsel []domain.Feldwechsel,
) error {
	if len(wechsel) == 0 {
		return nil
	}

	for _, einer := range wechsel {
		// Dieselbe Regel wie auf jedem anderen Weg (#89) — und vor dem ersten
		// Schreiben, wie jede Vorprüfung dieser Transaktion.
		if befund := domain.PruefeFeldname(einer.Feld); befund != nil {
			return befund
		}
	}

	transaktion, err := k.db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "Ersetzung beginnen")
	}

	defer func() { _ = transaktion.Rollback() }()

	if err := fehlendeMelden(ctx, transaktion, kartenAus(wechsel)); err != nil {
		return err
	}

	// **Eine Gruppe für alle** — dieselbe wie bei der Massenänderung, und aus
	// demselben Grund (§7 und #52).
	gruppe := neueGruppe()

	for _, einer := range wechsel {
		if err := vorwertHalten(ctx, transaktion, einer); err != nil {
			return err
		}

		if err := k.massenzeile(
			ctx, transaktion, einer.KarteID, einer.Feld, einer.Wert, gruppe); err != nil {
			return err
		}
	}

	return uebersetze(transaktion.Commit(), "Ersetzung abschließen")
}

// kartenAus gibt die berührten Kartenkennungen — je einmal.
//
// **Je einmal**, weil eine Ersetzung mehrere Felder derselben Karte trifft und
// die Fehlermeldung sie sonst mehrfach nennte.
func kartenAus(wechsel []domain.Feldwechsel) []int64 {
	gesehen := make(map[int64]struct{}, len(wechsel))
	karten := make([]int64, 0, len(wechsel))

	for _, einer := range wechsel {
		if _, schon := gesehen[einer.KarteID]; schon {
			continue
		}

		gesehen[einer.KarteID] = struct{}{}

		karten = append(karten, einer.KarteID)
	}

	return karten
}

// vorwertHalten prüft **in der Transaktion, die schreibt**, dass im Feld noch
// der Text steht, aus dem die Ersetzung rechnete (#2534) — sonst schriebe sie
// über eine Änderung, die ein anderer dazwischen gemacht hat, und beide
// meldeten Erfolg. Ein fehlendes Feld zählt als leer.
func vorwertHalten(ctx context.Context, tx *sql.Tx, einer domain.Feldwechsel) error {
	var jetzt sql.NullString

	err := tx.QueryRowContext(ctx,
		`SELECT wert FROM card_values WHERE card_id = ? AND feld = ?`, einer.KarteID, einer.Feld).Scan(&jetzt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return &domain.BulkError{
			CardIDs: []int64{einer.KarteID},
			Err:     uebersetze(err, "den Text von %q lesen", einer.Feld),
		}
	}

	if jetzt.String != einer.Vorher {
		return &domain.BulkError{
			CardIDs: []int64{einer.KarteID},
			Err: fmt.Errorf("%w: in %q steht inzwischen ein anderer Text — die Ersetzung "+
				"gilt einem Stand, den es so nicht mehr gibt; noch einmal suchen", domain.ErrConflict, einer.Feld),
		}
	}

	return nil
}
