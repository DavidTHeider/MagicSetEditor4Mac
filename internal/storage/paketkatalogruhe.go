// **Ob eine Transaktion etwas geschrieben hat** —
// #1325.
//
// Eigene Datei, weil `paketkatalog.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **was** in den Katalog kommt, hier, ob
// überhaupt etwas hineinkam.
//
// # Warum das zählt
//
// Jeder Commit hebt `PRAGMA data_version` für die übrigen Verbindungen. Das
// Fenster liest daraus, ob ein **anderer** Prozess geschrieben hat, und lädt
// dann alle Seiten neu. `App.Pakete` frischt bei jedem Aufruf auf — und die
// Startseite ruft es. Der Kreis war am 2026-09-03 im Fenster zu sehen: zehn
// Sekunden Leerlauf, zehnmal `Sets`, fünfmal `Pakete`, ein Flackern im
// Sekundentakt.

package storage

import (
	"context"
	"database/sql"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// aenderungenSeit sagt, ob diese Transaktion etwas geschrieben hat.
func aenderungenSeit(ctx context.Context, tx *sql.Tx, vorher int64) (bool, error) {
	nachher, err := aenderungen(ctx, tx)
	if err != nil {
		return false, err
	}

	return nachher != vorher, nil
}

// aenderungen liest den Schreibzähler der Verbindung.
func aenderungen(ctx context.Context, tx *sql.Tx) (int64, error) {
	var zahl int64

	if err := tx.QueryRowContext(ctx, `SELECT total_changes()`).Scan(&zahl); err != nil {
		return 0, uebersetze(err, "den Schreibzähler lesen")
	}

	return zahl, nil
}

// bedarfGleich sagt, ob die Abhängigkeiten schon so im Katalog stehen.
//
// **In der Reihenfolge und mit ihr.** `reihenfolge` ist Teil des Eintrags: zwei
// Listen mit denselben Paketen in anderer Ordnung sind nicht dieselbe Liste.
func bedarfGleich(ctx context.Context, tx *sql.Tx, kennung int64,
	braucht []domain.Bedarf,
) (bool, error) {
	zeilen, err := tx.QueryContext(ctx,
		`SELECT braucht, version FROM package_deps
		  WHERE package_id = ? ORDER BY reihenfolge`, kennung)
	if err != nil {
		return false, uebersetze(err, "die Abhängigkeiten lesen")
	}
	defer func() { _ = zeilen.Close() }() //nolint:errcheck // der Fehler steht unten

	stand := make([]domain.Bedarf, 0, len(braucht))

	for zeilen.Next() {
		var eintrag domain.Bedarf

		if err := zeilen.Scan(&eintrag.Paket, &eintrag.Version); err != nil {
			return false, uebersetze(err, "eine Abhängigkeit lesen")
		}

		stand = append(stand, eintrag)
	}

	if err := zeilen.Err(); err != nil {
		return false, uebersetze(err, "die Abhängigkeiten durchgehen")
	}

	return slices.Equal(stand, braucht), nil
}
