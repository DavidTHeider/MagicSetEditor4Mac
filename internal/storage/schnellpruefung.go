package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// schnellPruefen fragt SQLite, ob seine Datei heil ist — `PRAGMA
// quick_check`, F05 (#2344),
// Befund A-21.
//
// **Jede Seite und nicht nur die, die das Öffnen berührt**: eine Sicherung
// mit einer zerschriebenen Datenseite ging bis hierher durch das Öffnen und
// durch die Buchführung — gemessen in `sicherungsprobe_test.go`.
// `quick_check` und nicht `integrity_check`: es lässt den Abgleich der
// Indizes mit ihren Tabellen aus und ist dafür um ein Vielfaches schneller;
// eine zerschriebene Seite findet es trotzdem.
//
// **Ein Fehler der Prüfung ist kein Bestanden.** Wer nicht prüfen konnte,
// weiß nicht, dass die Datei heil ist.
func schnellPruefen(ctx context.Context, db *sql.DB) error {
	var ergebnis string

	if err := db.QueryRowContext(ctx, "PRAGMA quick_check(1)").Scan(&ergebnis); err != nil {
		return einordnen(fmt.Errorf("quick_check: %w", err))
	}

	if ergebnis != "ok" {
		return domain.MitGrund(domain.ErrBeschaedigt, fmt.Errorf("quick_check: %s", ergebnis),
			"die Datei ist beschädigt")
	}

	return nil
}

// geschriebeneSicherungPruefen öffnet eine eben geschriebene Sicherung nur
// lesend und prüft sie.
func geschriebeneSicherungPruefen(ctx context.Context, pfad string) error {
	db, err := oeffner(AlsURI(pfad)+"?mode=ro", "")
	if err != nil {
		return fmt.Errorf("die Sicherung zum Prüfen öffnen: %w", err)
	}

	return errors.Join(schnellPruefen(ctx, db), db.Close())
}
