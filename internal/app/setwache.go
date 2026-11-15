// **Gibt es dieses Set? — für die Flächen ohne Fenster** —
// #1336.
//
// # Warum es diese Naht gibt
//
// Die Regel steht in `internal/service` (`SetDaSein`), und dort gehört sie
// hin. **Die Kommandozeile darf `service` aber nicht kennen** — der
// Schichtenwächter führt für `cmd/mse` genau `app`, `domain` und `storage`.
//
// Ohne diese Naht müsste sie die Frage selbst beantworten, und dann stünde
// dieselbe Regel an zwei Stellen (§2).

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// SetDaSein weist eine Setkennung ab, die auf nichts zeigt.
//
// **`ErrNotFound`**: die Kennung ist wohlgeformt, sie zeigt nur auf nichts.
func SetDaSein(ctx context.Context, db *sql.DB, setID int64) error {
	return service.SetDaSein(ctx, storage.NeuerSetLeser(db), setID)
}

// symbolsetDa fragt, ob das Set dem gezeigten Profil gehört, bevor eine
// Zeichnung neben ihm gelesen oder geschrieben wird (#1815).
//
// **Ohne Bestandsleser gibt es kein fremdes Set** — dieselbe Regel wie bei
// [App.ueberNutzer]. Das Fenster setzt ihn beim Start; eine Anwendung, die
// keinen Bestand kennt, kennt auch kein Set, das einem anderen gehörte.
func (a *App) symbolsetDa(ctx context.Context, setID int64) error {
	if a.bestand == nil {
		return nil
	}

	return service.SetDaSein(ctx, a.bestand, setID)
}
