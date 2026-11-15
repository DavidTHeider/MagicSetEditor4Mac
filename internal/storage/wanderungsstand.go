package storage

// **Ob ein laufender Prozess den Bestand noch kennt** — M02
// (#2346),
// Befund A-11 des Audits 0.9.0.
//
// Beim Öffnen weist der Migrator einen Bestand ab, der eine unbekannte
// Wanderung führt. **Ein Prozess, der ihn schon offen hält, fragte nie
// wieder**: ein Dienst unter launchd lief nach einer Aktualisierung mit dem
// alten Code weiter, und die neue Fassung wanderte den Bestand unter ihm —
// gemessen in `TestEinDienstEndetVorEinemNeuerenBestand`: er legte danach
// noch ein Set an. Dieselbe Frage wie beim Öffnen, gestellt vor dem Schreiben.

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"slices"
)

// PruefeWanderungsstand gibt einen Fehler mit [domain.NeuereFassung], wenn
// der Bestand eine Wanderung führt, die diese Fassung nicht kennt.
func PruefeWanderungsstand(ctx context.Context, db *sql.DB) error {
	return pruefeWanderungsstandGegen(ctx, db, migrationen, grundlinie)
}

func pruefeWanderungsstandGegen(ctx context.Context, db *sql.DB, quelle, basis fs.FS) error {
	namen, err := migrationsnamen(quelle)
	if err != nil {
		return err
	}

	umfang, _, err := grundlinienumfang(basis)
	if err != nil {
		return err
	}

	verbucht, err := angewendeteMitSummen(ctx, db)
	if err != nil {
		return fmt.Errorf("die Buchführung lesen: %w", einordnen(err))
	}

	return unbekannteAbweisen(append(slices.Clone(umfang), namen...), verbucht)
}
